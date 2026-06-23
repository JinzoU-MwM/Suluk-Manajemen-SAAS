package service

import "context"

// DocumentAnalyzer is the provider-agnostic OCR interface. Both the Gemini and
// OpenCode clients satisfy it; the factory NewAnalyzer (Task 4) picks one by config.
type DocumentAnalyzer interface {
	AnalyzeDocument(ctx context.Context, imageData []byte, mimeType string) (*OCRResult, error)
	Available() bool
}
