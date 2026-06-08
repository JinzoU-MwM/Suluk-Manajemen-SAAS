package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/jamaah-in/v2/internal/aiocr/model"
	"github.com/jamaah-in/v2/internal/aiocr/repository"
)

type Worker struct {
	repo    *repository.AIOCRRepo
	gemini  *GeminiClient
	logger  *zap.SugaredLogger
	httpCli *http.Client
}

func NewWorker(repo *repository.AIOCRRepo, gemini *GeminiClient, logger *zap.SugaredLogger) *Worker {
	return &Worker{
		repo:   repo,
		gemini: gemini,
		logger: logger,
		httpCli: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				IdleConnTimeout:     60 * time.Second,
				DisableCompression:  false,
			},
		},
	}
}

func (w *Worker) Start(ctx context.Context, interval time.Duration) {
	w.logger.Infof("starting AI scanner worker (interval: %s)", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	w.processBatch(ctx)

	for {
		select {
		case <-ticker.C:
			w.processBatch(ctx)
		case <-ctx.Done():
			w.logger.Info("AI scanner worker stopped")
			return
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	if w.gemini == nil {
		return
	}

	results, err := w.repo.GetPendingScanResults(ctx, 10)
	if err != nil {
		w.logger.Errorf("fetch pending results: %v", err)
		return
	}

	if len(results) == 0 {
		return
	}

	// jobKey → {orgID, processedCount} — fixed: use per-result orgID, not results[0]
	type jobKey struct {
		jobID string
		orgID string
	}
	jobCounts := map[jobKey]int{}

	for i := range results {
		sr := &results[i]

		w.logger.Infow("processing scan result",
			"id", sr.ID,
			"file", sr.FileName,
			"hash", sr.FileHash,
		)

		if err := w.repo.UpdateScanResultStatus(ctx, sr.ID, sr.OrgID, string(model.ResultStatusProcessing), ""); err != nil {
			w.logger.Errorf("mark result processing: %v", err)
			continue
		}

		jk := jobKey{jobID: sr.ScanJobID.String(), orgID: sr.OrgID.String()}

		cached, err := w.repo.CheckCache(ctx, sr.FileHash, sr.ModelUsed, sr.PromptVersion)
		if err == nil && cached != nil {
			sr.ExtractedData = cached.ResultJSON
			hit := true
			sr.CacheHit = &hit
			sr.Status = string(model.ResultStatusCompleted)

			docTypeStr := detectDocTypeFromCached(cached.ResultJSON)
			sr.DocType = &docTypeStr

			normalized := normalizeToSiskopatuh(sr.ExtractedData, docTypeStr)
			sr.NormalizedData = normalized

			if err := w.repo.UpdateScanResultCompleted(ctx, sr.ID, sr.OrgID, &docTypeStr,
				sr.ExtractedData, sr.NormalizedData, nil, &hit, 0); err != nil {
				w.logger.Errorf("update cached result: %v", err)
				continue
			}

			w.repo.IncrementCacheHit(ctx, sr.FileHash)

			jobCounts[jk]++
			w.logger.Infow("cache hit", "id", sr.ID, "file", sr.FileName)
			continue
		}

		imageData, mimeType, fetchErr := w.fetchImage(ctx, sr.FileURL, sr.FileName)
		if fetchErr != nil {
			errMsg := fetchErr.Error()
			w.logger.Errorf("fetch image %s: %v", sr.FileName, fetchErr)
			w.repo.UpdateScanResultFailed(ctx, sr.ID, sr.OrgID, errMsg)
			jobCounts[jk]++ // still count: the file was attempted
			continue
		}

		startTime := time.Now()
		result, geminiErr := w.gemini.AnalyzeDocument(imageData, mimeType)
		elapsed := int(time.Since(startTime).Milliseconds())

		if geminiErr != nil {
			errMsg := geminiErr.Error()
			w.logger.Errorf("gemini analysis failed for %s: %v", sr.FileName, geminiErr)
			w.repo.UpdateScanResultFailed(ctx, sr.ID, sr.OrgID, errMsg)
			jobCounts[jk]++
			continue
		}

		validationErrors := validateExtractedData(result.ExtractedData, result.DocType)

		docType := result.DocType
		sr.DocType = &docType
		sr.ExtractedData = result.ExtractedData
		sr.ValidationErrors = validationErrors
		// Use model name from the client, not a hardcoded string
		sr.ModelUsed = w.gemini.model
		sr.ProcessingTimeMs = &elapsed

		normalized := normalizeToSiskopatuh(result.ExtractedData, result.DocType)
		sr.NormalizedData = normalized

		if err := w.repo.UpdateScanResultCompleted(ctx, sr.ID, sr.OrgID, &docType,
			result.ExtractedData, normalized, validationErrors, nil, elapsed); err != nil {
			w.logger.Errorf("update completed result: %v", err)
		}

		cacheErr := w.repo.StoreInCache(ctx, sr.FileHash, sr.ModelUsed, sr.PromptVersion,
			"document_ocr", result.ExtractedData, &docType, 72*time.Hour)
		if cacheErr != nil {
			w.logger.Warnf("store cache: %v", cacheErr)
		}

		jobCounts[jk]++
		w.logger.Infow("scan completed",
			"id", sr.ID,
			"file", sr.FileName,
			"doc_type", docType,
			"elapsed_ms", elapsed,
		)
	}

	// Update job statuses using the correct orgID for each job
	for jk, processed := range jobCounts {
		jobID, err := model.ParseUUID(jk.jobID)
		if err != nil {
			continue
		}
		orgID, err := model.ParseUUID(jk.orgID)
		if err != nil {
			continue
		}
		if err := w.repo.UpdateScanJobStatus(ctx, jobID, orgID, string(model.ScanStatusCompleted), processed); err != nil {
			w.logger.Warnf("update job status: %v", err)
		}
	}
}

func (w *Worker) fetchImage(ctx context.Context, fileURL, fileName string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}

	resp, err := w.httpCli.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("fetch image returned status %d", resp.StatusCode)
	}

	const maxImageSize = 20 * 1024 * 1024 // 20 MB
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageSize+1))
	if err != nil {
		return nil, "", fmt.Errorf("read image data: %w", err)
	}
	if len(data) > maxImageSize {
		return nil, "", fmt.Errorf("image too large (>20MB): %s", fileName)
	}

	mimeType := detectMimeType(fileName)
	if resp.Header.Get("Content-Type") != "" {
		mimeType = resp.Header.Get("Content-Type")
	}

	return data, mimeType, nil
}

func validateExtractedData(data any, docType string) []ValidationError {
	var errors []ValidationError

	extracted, ok := data.(ExtractedFields)
	if !ok {
		return nil
	}

	switch docType {
	case "ktp":
		if extracted.NIK == "" {
			errors = append(errors, ValidationError{Field: "nik", Message: "NIK tidak terdeteksi"})
		} else if len(extracted.NIK) != 16 {
			errors = append(errors, ValidationError{Field: "nik", Message: "NIK harus 16 digit", Value: extracted.NIK})
		}
		if extracted.Nama == "" {
			errors = append(errors, ValidationError{Field: "nama", Message: "Nama tidak terdeteksi"})
		}
	case "paspor":
		if extracted.NoPaspor == "" {
			errors = append(errors, ValidationError{Field: "no_paspor", Message: "Nomor paspor tidak terdeteksi"})
		}
		if extracted.NamaPaspor == "" && extracted.Nama == "" {
			errors = append(errors, ValidationError{Field: "nama", Message: "Nama tidak terdeteksi"})
		}
	case "visa":
		if extracted.NoVisa == "" {
			errors = append(errors, ValidationError{Field: "no_visa", Message: "Nomor visa tidak terdeteksi"})
		}
	}

	return errors
}

func normalizeToSiskopatuh(data any, docType string) any {
	extracted, ok := data.(ExtractedFields)
	if !ok {
		return data
	}

	normalized := map[string]any{}

	if extracted.Nama != "" {
		normalized["nama"] = extracted.Nama
	}
	if extracted.NIK != "" {
		normalized["no_identitas"] = extracted.NIK
		normalized["jenis_identitas"] = "NIK"
	}
	if extracted.NoPaspor != "" {
		normalized["no_paspor"] = extracted.NoPaspor
	}
	if extracted.NamaPaspor != "" {
		normalized["nama_paspor"] = extracted.NamaPaspor
	} else if extracted.Nama != "" {
		normalized["nama_paspor"] = extracted.Nama
	}
	if extracted.TempatLahir != "" {
		normalized["tempat_lahir"] = extracted.TempatLahir
	}
	if extracted.TanggalLahir != "" {
		normalized["tanggal_lahir"] = extracted.TanggalLahir
	}
	if extracted.JenisKelamin != "" {
		normalized["gender"] = extracted.JenisKelamin
	}
	if extracted.Alamat != "" {
		normalized["alamat"] = extracted.Alamat
	}
	if extracted.Provinsi != "" {
		normalized["provinsi"] = extracted.Provinsi
	}
	if extracted.Kabupaten != "" {
		normalized["kabupaten"] = extracted.Kabupaten
	}
	if extracted.Kecamatan != "" {
		normalized["kecamatan"] = extracted.Kecamatan
	}
	if extracted.Kelurahan != "" {
		normalized["kelurahan"] = extracted.Kelurahan
	}
	if extracted.NoTelepon != "" {
		normalized["no_telepon"] = extracted.NoTelepon
	}
	if extracted.NoHP != "" {
		normalized["no_hp"] = extracted.NoHP
	}
	if extracted.Kewarganegaraan != "" {
		normalized["kewarganegaraan"] = extracted.Kewarganegaraan
	}
	if extracted.GolonganDarah != "" {
		normalized["golongan_darah"] = extracted.GolonganDarah
	}
	if extracted.Pendidikan != "" {
		normalized["pendidikan"] = extracted.Pendidikan
	}
	if extracted.Pekerjaan != "" {
		normalized["pekerjaan"] = extracted.Pekerjaan
	}
	if extracted.StatusPerkawinan != "" {
		normalized["status_pernikahan"] = extracted.StatusPerkawinan
	}
	if extracted.Agama != "" {
		normalized["agama"] = extracted.Agama
	}

	if extracted.TanggalPaspor != "" {
		normalized["tanggal_paspor"] = extracted.TanggalPaspor
	}
	if extracted.TanggalExpired != "" {
		normalized["tanggal_expired_paspor"] = extracted.TanggalExpired
	}
	if extracted.KotaPaspor != "" {
		normalized["kota_paspor"] = extracted.KotaPaspor
	}

	if extracted.ProviderVisa != "" {
		normalized["provider_visa"] = extracted.ProviderVisa
	}
	if extracted.NoVisa != "" {
		normalized["no_visa"] = extracted.NoVisa
	}
	if extracted.TanggalVisa != "" {
		normalized["tanggal_visa"] = extracted.TanggalVisa
	}
	if extracted.TanggalVisaAkhir != "" {
		normalized["tanggal_visa_akhir"] = extracted.TanggalVisaAkhir
	}

	normalized["source_doc_type"] = docType
	normalized["siskopatuh_version"] = "2.0"

	return normalized
}

func detectDocTypeFromCached(data any) string {
	if data == nil {
		return "unknown"
	}

	switch v := data.(type) {
	case map[string]any:
		if dt, ok := v["doc_type"].(string); ok && dt != "" {
			return dt
		}
	case ExtractedFields:
		return inferDocType(v)
	}

	b, _ := json.Marshal(data)
	var m map[string]any
	if json.Unmarshal(b, &m) == nil {
		if dt, ok := m["doc_type"].(string); ok && dt != "" {
			return dt
		}
		if _, ok := m["nik"].(string); ok {
			return "ktp"
		}
		if _, ok := m["no_paspor"].(string); ok {
			return "paspor"
		}
		if _, ok := m["no_visa"].(string); ok {
			return "visa"
		}
	}
	return "unknown"
}

func inferDocType(e ExtractedFields) string {
	if e.NIK != "" {
		return "ktp"
	}
	if e.NoPaspor != "" {
		return "paspor"
	}
	if e.NoVisa != "" {
		return "visa"
	}
	return "unknown"
}
