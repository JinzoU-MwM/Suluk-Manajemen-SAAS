package service

import (
	"context"
	"errors"
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
func (s *AIOCRService) ProcessDocumentsSync(ctx context.Context, files []SyncFile, cacheMode string) (*ProcessDocumentsResult, error) {
	if s.analyzer == nil {
		return nil, ErrOCRUnavailable
	}

	res := &ProcessDocumentsResult{
		Data:               []any{},
		ValidationWarnings: []any{},
		FileResults:        []SyncFileResult{},
	}

	for _, f := range files {
		mimeType := f.ContentType
		if mimeType == "" {
			mimeType = detectMimeType(f.FileName)
		}

		result, err := s.analyzer.AnalyzeDocument(ctx, f.Data, mimeType)
		if err != nil {
			s.logger.Errorf("ocr analyze %s: %v", f.FileName, err)
			res.FileResults = append(res.FileResults, SyncFileResult{
				Filename: f.FileName,
				Status:   "failed",
				Error:    "gagal mengekstrak dokumen (coba foto yang lebih jelas/terang)",
			})
			continue
		}

		res.Data = append(res.Data, normalizeToSiskopatuh(result.ExtractedData, result.DocType))

		for _, ve := range validateExtractedData(result.ExtractedData, result.DocType) {
			res.ValidationWarnings = append(res.ValidationWarnings, map[string]any{
				"filename": f.FileName,
				"field":    ve.Field,
				"message":  ve.Message,
				"value":    ve.Value,
			})
		}

		res.FileResults = append(res.FileResults, SyncFileResult{
			Filename: f.FileName,
			Status:   "completed",
			DocType:  result.DocType,
		})
	}

	return res, nil
}
