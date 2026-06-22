package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jamaah-in/v2/internal/shared/config"
)

func TestOpenCodeAnalyzeImage(t *testing.T) {
	var gotURL, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Messages []struct {
				Content []struct {
					Type     string `json:"type"`
					ImageURL struct {
						URL string `json:"url"`
					} `json:"image_url"`
				} `json:"content"`
			} `json:"messages"`
		}
		_ = json.Unmarshal(body, &req)
		if len(req.Messages) > 0 {
			for _, p := range req.Messages[0].Content {
				if p.Type == "image_url" {
					gotURL = p.ImageURL.URL
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"choices":[{"message":{"content":"{\"doc_type\":\"ktp\",\"extracted_data\":{\"nama\":\"BUDI\"},\"confidence\":0.9}"}}]}`)
	}))
	defer srv.Close()

	a := NewOpenCodeAnalyzer("sk-test", "gpt-5-nano", srv.URL)
	res, err := a.AnalyzeDocument(context.Background(), []byte("\x89PNGfake"), "image/png")
	if err != nil {
		t.Fatalf("AnalyzeDocument: %v", err)
	}
	if res.DocType != "ktp" || res.ExtractedData.Nama != "BUDI" {
		t.Errorf("res = %+v", res)
	}
	if gotAuth != "Bearer sk-test" {
		t.Errorf("auth = %q", gotAuth)
	}
	if !strings.HasPrefix(gotURL, "data:image/png;base64,") {
		t.Errorf("image url = %q", gotURL)
	}
}

func TestOpenCodeAnalyzePDFRasterizes(t *testing.T) {
	orig := rasterizePDF
	defer func() { rasterizePDF = orig }()
	var rasterized bool
	rasterizePDF = func(ctx context.Context, data []byte) ([]byte, error) {
		rasterized = true
		return []byte("\x89PNGfake"), nil
	}

	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Messages []struct {
				Content []struct {
					Type     string `json:"type"`
					ImageURL struct {
						URL string `json:"url"`
					} `json:"image_url"`
				} `json:"content"`
			} `json:"messages"`
		}
		_ = json.Unmarshal(body, &req)
		for _, p := range req.Messages[0].Content {
			if p.Type == "image_url" {
				gotURL = p.ImageURL.URL
			}
		}
		io.WriteString(w, `{"choices":[{"message":{"content":"{\"doc_type\":\"paspor\"}"}}]}`)
	}))
	defer srv.Close()

	a := NewOpenCodeAnalyzer("sk-test", "gpt-5-nano", srv.URL)
	res, err := a.AnalyzeDocument(context.Background(), []byte("%PDF-1.4 fake"), "application/pdf")
	if err != nil {
		t.Fatalf("AnalyzeDocument(pdf): %v", err)
	}
	if !rasterized {
		t.Error("expected rasterizePDF to be called for application/pdf")
	}
	if !strings.HasPrefix(gotURL, "data:image/png;base64,") {
		t.Errorf("pdf should be sent as image/png, url = %q", gotURL)
	}
	if res.DocType != "paspor" {
		t.Errorf("res = %+v", res)
	}
}

func TestNewAnalyzerSelectsProvider(t *testing.T) {
	cfg := &config.Config{}
	cfg.AI = config.AIConfig{Provider: "opencode", OpenCodeAPIKey: "sk", OpenCodeModel: "gpt-5-nano", OpenCodeBaseURL: "https://x"}
	if a := NewAnalyzer(cfg); a == nil || !a.Available() {
		t.Error("opencode+key should be available")
	}
	cfg.AI = config.AIConfig{Provider: "opencode"} // no key
	if a := NewAnalyzer(cfg); a != nil {
		t.Errorf("opencode without key should be nil, got %T", a)
	}
	cfg.AI = config.AIConfig{Provider: "gemini"}
	cfg.Gemini = config.GeminiConfig{APIKey: "g"}
	if a := NewAnalyzer(cfg); a == nil || !a.Available() {
		t.Error("gemini+key should be available")
	}
}
