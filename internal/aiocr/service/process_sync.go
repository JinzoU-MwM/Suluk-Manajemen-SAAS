package service

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

// ErrOCRUnavailable is returned when no AI provider is configured, so the
// handler can surface a clean 503 instead of a generic 500.
var ErrOCRUnavailable = errors.New("OCR tidak tersedia: AI provider belum dikonfigurasi (atur OPENCODE_API_KEY atau GEMINI_API_KEY)")

// SyncFile is one uploaded document to OCR synchronously (bytes already read).
type SyncFile struct {
	FileName    string
	ContentType string
	Data        []byte
}

// SyncFileResult reports the per-file outcome to the scanner UI.
type SyncFileResult struct {
	Filename string `json:"filename"`
	Status   string `json:"status"` // "completed" | "failed"
	DocType  string `json:"doc_type,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ProcessDocumentsResult is the synchronous payload the scanner frontend
// expects: extracted/normalized records plus per-file status and warnings.
type ProcessDocumentsResult struct {
	Data               []any            `json:"data"`
	ValidationWarnings []any            `json:"validation_warnings"`
	FileResults        []SyncFileResult `json:"file_results"`
}

// ProcessDocumentsSync OCRs each uploaded file inline (OCR → validate →
// normalize) and returns jamaah-shaped records immediately. This is the
// synchronous counterpart to the async CreateScanJob/worker path; the scanner
// frontend was built against this synchronous contract.
//
// cacheMode is accepted for forward-compat with the UI's cache selector but is
// not yet wired to the OCR cache table (every file is processed fresh).
func (s *AIOCRService) ProcessDocumentsSync(ctx context.Context, orgID uuid.UUID, files []SyncFile, cacheMode string) (*ProcessDocumentsResult, error) {
	if s.analyzer == nil {
		return nil, ErrOCRUnavailable
	}

	res := &ProcessDocumentsResult{
		Data:               []any{},
		ValidationWarnings: []any{},
		FileResults:        []SyncFileResult{},
	}

	// Classify each file: a PDF whose text reads like an insurance policy goes to
	// the policy lane (extract manifest, fill insurance columns later by passport
	// join); everything else is an identity document. pdftotext runs at most once
	// per PDF here and the text is reused for extraction.
	type policyDoc struct {
		file SyncFile
		text string
	}
	var identity []SyncFile
	var policies []policyDoc
	for _, f := range files {
		mime := f.ContentType
		if mime == "" {
			mime = detectMimeType(f.FileName)
		}
		if mime == "application/pdf" {
			if text, err := extractPDFText(ctx, f.Data); err == nil && looksLikePolicy(text) {
				policies = append(policies, policyDoc{file: f, text: text})
				continue
			}
		}
		identity = append(identity, f)
	}

	// --- identity lane ---
	// Process files CONCURRENTLY (bounded) so a batch of N photos finishes in
	// ~max(per-file time) instead of the sum. A sequential batch easily exceeds
	// the proxy / Fiber WriteTimeout (~60s) and the browser gets a 502 while the
	// server is still grinding. Each goroutine writes only its own index, so the
	// response keeps the upload order with no data race.
	type fileOut struct {
		data       any
		hasData    bool
		warnings   []map[string]any
		fileResult SyncFileResult
	}
	outs := make([]fileOut, len(identity))

	const maxConcurrent = 5
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	for i := range identity {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			f := identity[i]
			mimeType := f.ContentType
			if mimeType == "" {
				mimeType = detectMimeType(f.FileName)
			}
			result, err := s.analyzer.AnalyzeDocument(ctx, f.Data, mimeType)
			if err != nil {
				s.logger.Errorf("ocr analyze %s: %v", f.FileName, err)
				outs[i].fileResult = SyncFileResult{
					Filename: f.FileName,
					Status:   "failed",
					Error:    "gagal mengekstrak dokumen (coba foto yang lebih jelas/terang)",
				}
				return
			}
			outs[i].data = normalizeToSiskopatuh(result.ExtractedData, result.DocType)
			outs[i].hasData = true
			for _, ve := range validateExtractedData(result.ExtractedData, result.DocType) {
				outs[i].warnings = append(outs[i].warnings, map[string]any{
					"filename": f.FileName,
					"field":    ve.Field,
					"message":  ve.Message,
					"value":    ve.Value,
				})
			}
			outs[i].fileResult = SyncFileResult{
				Filename: f.FileName,
				Status:   "completed",
				DocType:  result.DocType,
			}
		}(i)
	}
	wg.Wait()

	// Assemble in upload order.
	for i := range outs {
		if outs[i].hasData {
			res.Data = append(res.Data, outs[i].data)
		}
		for _, w := range outs[i].warnings {
			res.ValidationWarnings = append(res.ValidationWarnings, w)
		}
		res.FileResults = append(res.FileResults, outs[i].fileResult)
	}

	// Merge documents of the same jamaah (a passport and its visa share the
	// passport number) so one person yields one row, not one row per file.
	res.Data = mergeIdentityRows(res.Data)

	// --- policy lane: extract manifests, then enrich identity rows by passport ---
	entriesByPaspor := map[string]PolicyEntry{}
	var ordered []PolicyEntry // manifest order, deduped by passport key
	var docAsuransi, docTglInput string
	for _, pd := range policies {
		if s.policy == nil {
			res.FileResults = append(res.FileResults, SyncFileResult{
				Filename: pd.file.FileName, Status: "failed",
				Error: "ekstraksi polis tidak tersedia (AI provider belum dikonfigurasi)",
			})
			continue
		}
		m, err := s.policy.ExtractManifest(ctx, pd.file.Data, pd.text)
		if err != nil || m == nil {
			s.logger.Errorf("policy extract %s: %v", pd.file.FileName, err)
			res.FileResults = append(res.FileResults, SyncFileResult{
				Filename: pd.file.FileName, Status: "failed",
				Error: "gagal membaca data polis (pastikan PDF asli, bukan hasil scan foto)",
			})
			continue
		}
		if m.Asuransi != "" {
			docAsuransi = m.Asuransi
		}
		if m.TanggalInput != "" {
			docTglInput = m.TanggalInput
		}
		for _, e := range m.Entries {
			k := normPaspor(e.NoIdentitas)
			if _, seen := entriesByPaspor[k]; !seen {
				ordered = append(ordered, e)
			}
			entriesByPaspor[k] = e
		}
		res.FileResults = append(res.FileResults, SyncFileResult{
			Filename: pd.file.FileName, Status: "completed", DocType: "polis",
		})
	}
	if len(entriesByPaspor) > 0 {
		// Fill insurance columns on scanned rows that match a manifest entry…
		enrichRowsWithPolicy(res.Data, entriesByPaspor, docAsuransi, docTglInput)
		// …then seed a row for every jamaah listed in the policy who has not been
		// scanned, so uploading the policy alone still produces the full list.
		existing := existingPasporKeys(res.Data)
		for _, e := range ordered {
			if existing[normPaspor(e.NoIdentitas)] {
				continue
			}
			res.Data = append(res.Data, policyEntryToRow(e, docAsuransi, docTglInput))
		}
	}

	// Meter successful scans (best-effort) toward the org's monthly quota. Counts
	// every completed document (identity + policy); a failure here must not affect
	// the OCR response returned to the user.
	scanned := 0
	for _, fr := range res.FileResults {
		if fr.Status == "completed" {
			scanned++
		}
	}
	if s.repo != nil && scanned > 0 {
		if err := s.repo.IncrementScanUsage(ctx, orgID, scanned); err != nil {
			s.logger.Errorf("record scan usage (org %s): %v", orgID, err)
		}
	}

	return res, nil
}
