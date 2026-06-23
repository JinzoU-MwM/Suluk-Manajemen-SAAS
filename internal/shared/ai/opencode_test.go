package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jamaah-in/v2/internal/shared/config"
)

func TestOpenCodeGenerateText(t *testing.T) {
	var gotAuth, gotModel, gotPrompt string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.Unmarshal(body, &req)
		gotModel = req.Model
		if len(req.Messages) > 0 {
			gotPrompt = req.Messages[0].Content
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"choices":[{"message":{"content":"Halo dunia"}}]}`)
	}))
	defer srv.Close()

	g := newOpenCode("sk-test", "gpt-5-nano", srv.URL)
	out, err := g.GenerateText(context.Background(), "ringkas ini")
	if err != nil {
		t.Fatalf("GenerateText: %v", err)
	}
	if out != "Halo dunia" {
		t.Errorf("out = %q", out)
	}
	if gotAuth != "Bearer sk-test" {
		t.Errorf("auth = %q", gotAuth)
	}
	if gotModel != "gpt-5-nano" {
		t.Errorf("model = %q", gotModel)
	}
	if gotPrompt != "ringkas ini" {
		t.Errorf("prompt = %q", gotPrompt)
	}
}

func TestOpenCodeRetriesOn429(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		io.WriteString(w, `{"choices":[{"message":{"content":"ok"}}]}`)
	}))
	defer srv.Close()
	g := newOpenCode("sk-test", "gpt-5-nano", srv.URL)
	out, err := g.GenerateText(context.Background(), "x")
	if err != nil || out != "ok" {
		t.Fatalf("out=%q err=%v", out, err)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2 (1 retry)", calls)
	}
}

func TestNewFactorySelectsProvider(t *testing.T) {
	cfg := &config.Config{}
	// opencode with key -> available
	cfg.AI = config.AIConfig{Provider: "opencode", OpenCodeAPIKey: "sk", OpenCodeModel: "gpt-5-nano", OpenCodeBaseURL: "https://x"}
	if g := New(cfg); g == nil || !g.Available() {
		t.Errorf("opencode+key should be available")
	}
	// opencode without key -> nil
	cfg.AI = config.AIConfig{Provider: "opencode"}
	if g := New(cfg); g != nil {
		t.Errorf("opencode without key should be nil, got %T", g)
	}
	// gemini with key -> available
	cfg.AI = config.AIConfig{Provider: "gemini"}
	cfg.Gemini = config.GeminiConfig{APIKey: "g-key"}
	if g := New(cfg); g == nil || !g.Available() {
		t.Errorf("gemini+key should be available")
	}
	// gemini without key -> nil
	cfg.Gemini = config.GeminiConfig{APIKey: ""}
	if g := New(cfg); g != nil {
		t.Errorf("gemini without key should be nil")
	}
}
