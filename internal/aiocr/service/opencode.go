package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/jamaah-in/v2/internal/shared/config"
)

// NewAnalyzer returns the configured OCR DocumentAnalyzer, or nil when the
// selected provider's key is empty (callers treat nil as "OCR unavailable").
// The key is checked before constructing, so a nil concrete pointer is never
// wrapped into a non-nil interface.
func NewAnalyzer(cfg *config.Config) DocumentAnalyzer {
	switch cfg.AI.Provider {
	case "opencode":
		if cfg.AI.OpenCodeAPIKey == "" {
			return nil
		}
		return NewOpenCodeAnalyzer(cfg.AI.OpenCodeAPIKey, cfg.AI.OpenCodeModel, cfg.AI.OpenCodeBaseURL)
	default: // "gemini"
		if cfg.Gemini.APIKey == "" {
			return nil
		}
		return NewGeminiClient(cfg.Gemini.APIKey)
	}
}

type OpenCodeAnalyzer struct {
	apiKey  string
	model   string
	baseURL string
	httpc   *http.Client
}

func NewOpenCodeAnalyzer(apiKey, model, baseURL string) *OpenCodeAnalyzer {
	if model == "" {
		model = "claude-haiku-4-5"
	}
	if baseURL == "" {
		baseURL = "https://opencode.ai/zen/v1"
	}
	return &OpenCodeAnalyzer{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		httpc:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (a *OpenCodeAnalyzer) Available() bool { return a != nil && a.apiKey != "" }

// rasterizePDF is a package var so tests can stub it without invoking pdftoppm.
var rasterizePDF = rasterizePDFImpl

// rasterizePDFImpl renders page 1 of a PDF to PNG via poppler's `pdftoppm`
// (-singlefile => the output is exactly <prefix>.png, no page-number suffix).
func rasterizePDFImpl(ctx context.Context, data []byte) ([]byte, error) {
	in, err := os.CreateTemp("", "ocr-*.pdf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(in.Name())
	if _, err := in.Write(data); err != nil {
		in.Close()
		return nil, err
	}
	in.Close()

	prefix := in.Name() + "-out"
	out := prefix + ".png"
	defer os.Remove(out)
	cmd := exec.CommandContext(ctx, "pdftoppm", "-png", "-r", "200", "-f", "1", "-l", "1", "-singlefile", in.Name(), prefix)
	if combined, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm gagal: %v: %s", err, string(combined))
	}
	return os.ReadFile(out)
}

func (a *OpenCodeAnalyzer) AnalyzeDocument(ctx context.Context, imageData []byte, mimeType string) (*OCRResult, error) {
	if !a.Available() {
		return nil, fmt.Errorf("opencode analyzer not configured (OPENCODE_API_KEY missing)")
	}
	if mimeType == "application/pdf" {
		png, err := rasterizePDF(ctx, imageData)
		if err != nil {
			return nil, fmt.Errorf("gagal konversi PDF ke gambar: %w", err)
		}
		imageData, mimeType = png, "image/png"
	}

	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imageData))
	reqBody := map[string]any{
		"model":       a.model,
		"temperature": 0.1,
		"max_tokens":  4096,
		"messages": []map[string]any{{
			"role": "user",
			"content": []map[string]any{
				{"type": "text", "text": systemPrompts["auto"]},
				{"type": "image_url", "image_url": map[string]any{"url": dataURI}},
			},
		}},
	}
	body, _ := json.Marshal(reqBody)

	raw, err := a.post(ctx, body)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse opencode response: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("opencode api error: %s", resp.Error.Message)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("opencode returned no choices")
	}

	text := cleanJSONString(resp.Choices[0].Message.Content)
	var result OCRResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse extracted data: %w", err)
	}
	return &result, nil
}

// post POSTs with Bearer auth, retrying on 429 / >=500 / network error (max 3).
func (a *OpenCodeAnalyzer) post(ctx context.Context, body []byte) ([]byte, error) {
	url := a.baseURL + "/chat/completions"
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+a.apiKey)
		resp, err := a.httpc.Do(req)
		if err != nil {
			lastErr = err
			a.backoff(ctx, attempt)
			continue
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("opencode api status %d", resp.StatusCode)
			a.backoff(ctx, attempt)
			continue
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("opencode api status %d: %s", resp.StatusCode, string(raw))
		}
		return raw, nil
	}
	return nil, fmt.Errorf("opencode api failed after retries: %w", lastErr)
}

func (a *OpenCodeAnalyzer) backoff(ctx context.Context, attempt int) {
	d := time.Duration(1<<uint(attempt)) * time.Second
	if d > 10*time.Second {
		d = 10 * time.Second
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
