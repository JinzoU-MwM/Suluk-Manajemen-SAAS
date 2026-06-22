# Switchable AI Provider (Gemini ↔ OpenCode Zen) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Suluk's AI provider switchable via `AI_PROVIDER` env (default `opencode`), migrating both the OCR scanner and the accounting-insights copilot from Gemini to OpenCode Zen, with Gemini kept as an instant env-flip fallback.

**Architecture:** Each AI feature gets a small provider-agnostic interface (`ai.Generator` for text, `service.DocumentAnalyzer` for OCR) with a Gemini impl (existing) and a new OpenCode Zen impl, chosen by a factory that reads `config.AIConfig`. OpenCode Zen is an OpenAI-compatible chat-completions API (Bearer auth, `choices[0].message.content`) that also does vision; PDFs are rasterized to PNG via `pdftoppm` before the vision call.

**Tech Stack:** Go 1.24, Fiber, `CGO_ENABLED=0` static builds on `alpine:3.19`. Tests use the standard `testing` package + `net/http/httptest` (no live API calls). Module path: `github.com/jamaah-in/v2`.

## Global Constraints

- **Tests are hermetic:** standard `testing` + `net/http/httptest`. No live OpenCode/Gemini calls, no live `pdftoppm` in unit tests (the PDF rasterizer is a stubbable package var).
- **Defaults (verbatim):** `AI_PROVIDER` → `opencode`; `OPENCODE_MODEL` → `gpt-5-nano`; `OPENCODE_BASE_URL` → `https://opencode.ai/zen/v1`; `OPENCODE_API_KEY` → `""`. `GEMINI_API_KEY` / `GEMINI_MODEL` unchanged.
- **OpenCode Zen wire format:** `POST {baseURL}/chat/completions`, header `Authorization: Bearer <key>`, body is OpenAI chat shape, reply read from `choices[0].message.content`. **Retry** on HTTP 429 / ≥500 / timeout / network error, max 3 attempts, capped exponential backoff.
- **Typed-nil rule:** factories return **untyped `nil`** when the selected provider's key is empty — check the key BEFORE constructing a client, so a nil `*GeminiClient`/`*Client` is never wrapped in a non-nil interface. Text consumer (`insights.go`) additionally nil-guards `s.ai` before calling methods.
- **Each task leaves `go build ./...` green.** (That is why the text interface + factory + its accounting wiring are one atomic task — renaming the package's `New` would break the accounting build mid-flight otherwise.)
- **Preserve behaviour:** OCR field schema (`OCRResult`/`ExtractedFields`), prompts (`systemPrompts["auto"]`, `insightPrompt`), the accounting narrative cache, and graceful degradation (missing key → unavailable, non-fatal) are unchanged.
- **Secrets:** only via env (`OPENCODE_API_KEY`); never write a key into any tracked file.
- Commit after every task. **No "Co-Authored-By" or AI co-author trailers.**

---

### Task 1: AIConfig in shared/config

**Files:**
- Modify: `internal/shared/config/config.go` (add `AI` to the `Config` struct ~line 11-23; define `AIConfig`; add the load block in `Load()` after the `Gemini:` block ~line 141-143)
- Test: `internal/shared/config/config_test.go` (create)

**Interfaces:**
- Produces: `config.AIConfig{ Provider, OpenCodeAPIKey, OpenCodeModel, OpenCodeBaseURL string }`; `config.Config.AI AIConfig`. Loaded from `AI_PROVIDER` (default `opencode`), `OPENCODE_API_KEY`, `OPENCODE_MODEL` (default `gpt-5-nano`), `OPENCODE_BASE_URL` (default `https://opencode.ai/zen/v1`). Consumed by Tasks 2, 4, 5.

- [ ] **Step 1: Write the failing test** — `internal/shared/config/config_test.go`:
```go
package config

import "testing"

func TestLoadAIDefaults(t *testing.T) {
	t.Setenv("AI_PROVIDER", "")
	t.Setenv("OPENCODE_API_KEY", "")
	t.Setenv("OPENCODE_MODEL", "")
	t.Setenv("OPENCODE_BASE_URL", "")
	c := Load()
	if c.AI.Provider != "opencode" {
		t.Errorf("Provider = %q, want opencode", c.AI.Provider)
	}
	if c.AI.OpenCodeModel != "gpt-5-nano" {
		t.Errorf("OpenCodeModel = %q, want gpt-5-nano", c.AI.OpenCodeModel)
	}
	if c.AI.OpenCodeBaseURL != "https://opencode.ai/zen/v1" {
		t.Errorf("OpenCodeBaseURL = %q", c.AI.OpenCodeBaseURL)
	}
	if c.AI.OpenCodeAPIKey != "" {
		t.Errorf("OpenCodeAPIKey = %q, want empty", c.AI.OpenCodeAPIKey)
	}
}

func TestLoadAIOverride(t *testing.T) {
	t.Setenv("AI_PROVIDER", "Gemini") // mixed case -> normalized
	t.Setenv("OPENCODE_API_KEY", "sk-test")
	t.Setenv("OPENCODE_MODEL", "gpt-5")
	c := Load()
	if c.AI.Provider != "gemini" {
		t.Errorf("Provider = %q, want gemini (normalized)", c.AI.Provider)
	}
	if c.AI.OpenCodeAPIKey != "sk-test" || c.AI.OpenCodeModel != "gpt-5" {
		t.Errorf("override not applied: %+v", c.AI)
	}
}
```

- [ ] **Step 2: Run, verify it fails**

Run: `cd "D:/Codding/Project/Suluk-Manajemen-SAAS" && go test ./internal/shared/config/`
Expected: FAIL — `c.AI` undefined (compile error).

- [ ] **Step 3: Implement** — in `config.go`:

Add to the `Config` struct (after `Gemini   GeminiConfig`):
```go
	AI       AIConfig
```
Define the struct (next to `GeminiConfig`):
```go
// AIConfig selects and configures the AI provider for OCR + accounting insights.
// Provider is "opencode" (default) or "gemini"; Gemini stays available for an
// instant env-flip fallback.
type AIConfig struct {
	Provider        string
	OpenCodeAPIKey  string
	OpenCodeModel   string
	OpenCodeBaseURL string
}
```
Add to `Load()` (after the `Gemini: GeminiConfig{...},` block):
```go
		AI: AIConfig{
			Provider:        normalizeProvider(envOr("AI_PROVIDER", "opencode")),
			OpenCodeAPIKey:  envOr("OPENCODE_API_KEY", ""),
			OpenCodeModel:   envOr("OPENCODE_MODEL", "gpt-5-nano"),
			OpenCodeBaseURL: strings.TrimRight(envOr("OPENCODE_BASE_URL", "https://opencode.ai/zen/v1"), "/"),
		},
```
Add the helper (below `Load()`):
```go
// normalizeProvider lowercases/trims the provider and falls back to "opencode"
// for any unrecognized value.
func normalizeProvider(p string) string {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case "gemini":
		return "gemini"
	default:
		return "opencode"
	}
}
```

- [ ] **Step 4: Run, verify pass**

Run: `cd "D:/Codding/Project/Suluk-Manajemen-SAAS" && go test ./internal/shared/config/`
Expected: PASS.

- [ ] **Step 5: Commit**
```bash
git add internal/shared/config/config.go internal/shared/config/config_test.go
git commit -m "feat(config): AIConfig with AI_PROVIDER + OpenCode settings"
```

---

### Task 2: Text provider — Generator interface + OpenCode impl + factory, wired into accounting

This whole task is one commit so the build never breaks: renaming the package's `New` and changing `WithAI`'s parameter type must land together with the accounting call site.

**Files:**
- Create: `internal/shared/ai/generator.go` (interface + factory)
- Create: `internal/shared/ai/opencode.go` (OpenCode text impl + retry)
- Modify: `internal/shared/ai/gemini.go` (rename exported `New` → unexported `newGemini`; nothing else)
- Modify: `internal/accounting/service/service.go` (`ai *ai.Client` → `ai ai.Generator` ~line 22; `WithAI(c *ai.Client)` → `WithAI(c ai.Generator)` ~line 33)
- Modify: `internal/accounting/service/insights.go` (nil-guard `s.ai` at lines 62 and 106)
- Modify: `cmd/accounting-service/main.go` (`ai.New(cfg.Gemini.APIKey)` → `ai.New(cfg)` ~line 54)
- Test: `internal/shared/ai/opencode_test.go` (create)

**Interfaces:**
- Consumes: `config.Config` (Task 1).
- Produces: `ai.Generator` interface `{ GenerateText(ctx, prompt) (string, error); Available() bool }`; `ai.New(cfg *config.Config) Generator` (returns nil when the selected provider's key is empty). Existing `*ai.Client` (Gemini) satisfies `Generator`.

- [ ] **Step 1: Write failing tests** — `internal/shared/ai/opencode_test.go`:
```go
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
```

- [ ] **Step 2: Run, verify fail**

Run: `cd "D:/Codding/Project/Suluk-Manajemen-SAAS" && go test ./internal/shared/ai/`
Expected: FAIL — `newOpenCode` / `New(cfg)` undefined.

- [ ] **Step 3: Implement the provider**

In `gemini.go`, rename the constructor so the package's public `New` can be the factory. Change `func New(apiKey string) *Client {` → `func newGemini(apiKey string) *Client {` (leave the body, `Available()`, `GenerateText` unchanged).

Create `internal/shared/ai/generator.go`:
```go
package ai

import (
	"context"

	"github.com/jamaah-in/v2/internal/shared/config"
)

// Generator is the provider-agnostic text-generation interface used by the
// accounting copilot. Both the Gemini *Client and the OpenCode generator satisfy it.
type Generator interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
	Available() bool
}

// New returns the configured text Generator, or nil when the selected provider's
// API key is empty (callers treat nil as "AI unavailable", a graceful non-fatal
// state). The key is checked BEFORE constructing, so a nil concrete pointer is
// never wrapped into a non-nil interface.
func New(cfg *config.Config) Generator {
	switch cfg.AI.Provider {
	case "opencode":
		if cfg.AI.OpenCodeAPIKey == "" {
			return nil
		}
		return newOpenCode(cfg.AI.OpenCodeAPIKey, cfg.AI.OpenCodeModel, cfg.AI.OpenCodeBaseURL)
	default: // "gemini"
		if cfg.Gemini.APIKey == "" {
			return nil
		}
		return newGemini(cfg.Gemini.APIKey)
	}
}
```

Create `internal/shared/ai/opencode.go`:
```go
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type openCodeGenerator struct {
	apiKey  string
	model   string
	baseURL string
	httpc   *http.Client
}

func newOpenCode(apiKey, model, baseURL string) *openCodeGenerator {
	if model == "" {
		model = "gpt-5-nano"
	}
	if baseURL == "" {
		baseURL = "https://opencode.ai/zen/v1"
	}
	return &openCodeGenerator{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		httpc:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (g *openCodeGenerator) Available() bool { return g != nil && g.apiKey != "" }

type ocReq struct {
	Model       string      `json:"model"`
	Messages    []ocMessage `json:"messages"`
	Temperature float64     `json:"temperature"`
	MaxTokens   int         `json:"max_tokens"`
}

type ocMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string (text) or []any (vision parts)
}

type ocResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (g *openCodeGenerator) GenerateText(ctx context.Context, prompt string) (string, error) {
	if !g.Available() {
		return "", fmt.Errorf("opencode generator not configured")
	}
	body, _ := json.Marshal(ocReq{
		Model:       g.model,
		Messages:    []ocMessage{{Role: "user", Content: prompt}},
		Temperature: 0.3,
		MaxTokens:   1024,
	})
	raw, err := ocPost(ctx, g.httpc, g.baseURL+"/chat/completions", g.apiKey, body)
	if err != nil {
		return "", err
	}
	var r ocResp
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", fmt.Errorf("parse opencode response: %w", err)
	}
	if r.Error != nil {
		return "", fmt.Errorf("opencode api error: %s", r.Error.Message)
	}
	if len(r.Choices) == 0 {
		return "", fmt.Errorf("opencode returned no choices")
	}
	return r.Choices[0].Message.Content, nil
}

// ocPost POSTs JSON with Bearer auth and retries on 429 / >=500 / network error
// (max 3 attempts, capped exponential backoff). Returns the final response body.
func ocPost(ctx context.Context, httpc *http.Client, url, apiKey string, body []byte) ([]byte, error) {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		resp, err := httpc.Do(req)
		if err != nil {
			lastErr = err
			ocBackoff(ctx, attempt)
			continue
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("opencode api status %d", resp.StatusCode)
			ocBackoff(ctx, attempt)
			continue
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("opencode api status %d: %s", resp.StatusCode, string(raw))
		}
		return raw, nil
	}
	return nil, fmt.Errorf("opencode api failed after retries: %w", lastErr)
}

func ocBackoff(ctx context.Context, attempt int) {
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
```

- [ ] **Step 4: Wire into accounting** (same commit — keeps the build green)

`internal/accounting/service/service.go`:
- Field: `ai     *ai.Client` → `ai     ai.Generator`
- Method:
```go
// WithAI attaches an optional text Generator for copilot narratives. A nil
// generator (no key for the selected provider) is fine — insights still return
// the rule-based findings.
func (s *Service) WithAI(c ai.Generator) *Service {
	s.ai = c
	return s
}
```

`internal/accounting/service/insights.go`:
- Line 62: `AIAvailable: s.ai.Available(),` → `AIAvailable: s.ai != nil && s.ai.Available(),`
- Line 106: `if s.ai.Available() {` → `if s.ai != nil && s.ai.Available() {`
(The `s.ai.GenerateText(...)` at line 113 now runs only when `s.ai != nil`, so it is safe.)

`cmd/accounting-service/main.go` line 54: `.WithAI(ai.New(cfg.Gemini.APIKey))` → `.WithAI(ai.New(cfg))`

- [ ] **Step 5: Run, verify pass + build**

Run: `cd "D:/Codding/Project/Suluk-Manajemen-SAAS" && go test ./internal/shared/ai/ && go build ./... && go test ./internal/accounting/...`
Expected: ai tests PASS; build clean; accounting tests PASS (the nil guard keeps `GenerateInsights` panic-free when no provider is configured).

- [ ] **Step 6: Commit**
```bash
git add internal/shared/ai/generator.go internal/shared/ai/opencode.go internal/shared/ai/gemini.go internal/shared/ai/opencode_test.go internal/accounting/service/service.go internal/accounting/service/insights.go cmd/accounting-service/main.go
git commit -m "feat(ai): switchable text Generator (Gemini/OpenCode) wired into accounting"
```

---

### Task 3: OCR DocumentAnalyzer interface + rename GeminiResult → OCRResult

Leaves the build green without touching `main.go`: `*GeminiClient` satisfies the new interface, so the existing `NewAIOCRService(repo, geminiClient, logger)` call still compiles.

**Files:**
- Create: `internal/aiocr/service/analyzer.go` (interface + compile-time assertion)
- Modify: `internal/aiocr/service/gemini.go` (rename `GeminiResult` → `OCRResult` at lines 274, 173, 266; add `Available()` to `*GeminiClient`)
- Modify: `internal/aiocr/service/service.go` (field `gemini *GeminiClient` → `analyzer DocumentAnalyzer` line 18; `NewAIOCRService` param line 22-27; `Available()` body line 32)
- Modify: `internal/aiocr/service/process_sync.go` (`s.gemini` → `s.analyzer` at lines 43, 59)
- Test: `internal/aiocr/service/analyzer_test.go` (create)

> The complete set of references to update is exactly: `service.go:18,25,32`, `process_sync.go:43,59`, `gemini.go:173,266,274` (verified by grep — there is no async worker path).

**Interfaces:**
- Produces: `service.DocumentAnalyzer` interface `{ AnalyzeDocument(ctx, imageData []byte, mimeType string) (*OCRResult, error); Available() bool }`; `OCRResult` (was `GeminiResult`). `*GeminiClient` satisfies it. Consumed by Tasks 4, 5.

- [ ] **Step 1: Write failing test** — `internal/aiocr/service/analyzer_test.go`:
```go
package service

import (
	"context"
	"testing"
)

// compile-time: the Gemini client satisfies the provider-agnostic interface.
var _ DocumentAnalyzer = (*GeminiClient)(nil)

func TestNilGeminiAnalyzeReturnsError(t *testing.T) {
	var c *GeminiClient
	if _, err := c.AnalyzeDocument(context.Background(), []byte("x"), "image/png"); err == nil {
		t.Error("nil client should return an error, not panic")
	}
	if c.Available() {
		t.Error("nil client should not be Available")
	}
}
```

- [ ] **Step 2: Run, verify fail**

Run: `cd "D:/Codding/Project/Suluk-Manajemen-SAAS" && go test ./internal/aiocr/service/`
Expected: FAIL — `DocumentAnalyzer` undefined; `(*GeminiClient).Available` undefined.

- [ ] **Step 3: Implement**

Create `internal/aiocr/service/analyzer.go`:
```go
package service

import "context"

// DocumentAnalyzer is the provider-agnostic OCR interface. Both the Gemini and
// OpenCode clients satisfy it; the factory NewAnalyzer (Task 4) picks one by config.
type DocumentAnalyzer interface {
	AnalyzeDocument(ctx context.Context, imageData []byte, mimeType string) (*OCRResult, error)
	Available() bool
}
```

In `gemini.go`:
- Rename `type GeminiResult struct {` → `type OCRResult struct {` (line 274).
- In `AnalyzeDocument`, change the return type `(*GeminiResult, error)` → `(*OCRResult, error)` (line 173) and the local `var result GeminiResult` → `var result OCRResult` (line 266).
- Add `Available()` (next to `NewGeminiClient`):
```go
// Available reports whether a usable Gemini client is configured (nil-safe).
func (c *GeminiClient) Available() bool { return c != nil && c.apiKey != "" }
```

In `service.go`:
- Field: `gemini *GeminiClient` → `analyzer DocumentAnalyzer`
- Constructor:
```go
func NewAIOCRService(repo *repository.AIOCRRepo, analyzer DocumentAnalyzer, logger *zap.SugaredLogger) *AIOCRService {
	return &AIOCRService{
		repo:     repo,
		analyzer: analyzer,
		logger:   logger,
	}
}
```
- `Available()`:
```go
func (s *AIOCRService) Available() bool {
	return s.analyzer != nil
}
```

In `process_sync.go`:
- Line 43: `if s.gemini == nil {` → `if s.analyzer == nil {`
- Line 59: `result, err := s.gemini.AnalyzeDocument(ctx, f.Data, mimeType)` → `result, err := s.analyzer.AnalyzeDocument(ctx, f.Data, mimeType)`

- [ ] **Step 4: Run, verify pass + build**

Run: `cd "D:/Codding/Project/Suluk-Manajemen-SAAS" && go test ./internal/aiocr/service/ && go build ./...`
Expected: PASS + build clean (`main.go` still passes a `*GeminiClient`, which implements `DocumentAnalyzer`).

- [ ] **Step 5: Commit**
```bash
git add internal/aiocr/service/analyzer.go internal/aiocr/service/gemini.go internal/aiocr/service/service.go internal/aiocr/service/process_sync.go internal/aiocr/service/analyzer_test.go
git commit -m "refactor(aiocr): DocumentAnalyzer interface + rename GeminiResult->OCRResult"
```

---

### Task 4: OpenCode OCR analyzer + PDF rasterization + factory

**Files:**
- Create: `internal/aiocr/service/opencode.go` (OpenCode vision impl + `rasterizePDF` var + `NewAnalyzer` factory)
- Test: `internal/aiocr/service/opencode_test.go` (create)

**Interfaces:**
- Consumes: `DocumentAnalyzer`, `OCRResult`, `systemPrompts`, `cleanJSONString`, `NewGeminiClient` (Task 3); `config.Config` (Task 1).
- Produces: `OpenCodeAnalyzer` (implements `DocumentAnalyzer`); `NewAnalyzer(cfg *config.Config) DocumentAnalyzer` (nil when selected provider's key is empty). Consumed by Task 5.

- [ ] **Step 1: Write failing tests** — `internal/aiocr/service/opencode_test.go`:
```go
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
```

- [ ] **Step 2: Run, verify fail**

Run: `cd "D:/Codding/Project/Suluk-Manajemen-SAAS" && go test ./internal/aiocr/service/ -run OpenCode`
Expected: FAIL — `NewOpenCodeAnalyzer` / `NewAnalyzer` / `rasterizePDF` undefined.

- [ ] **Step 3: Implement** — `internal/aiocr/service/opencode.go`:
```go
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
		model = "gpt-5-nano"
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
```

- [ ] **Step 4: Run, verify pass**

Run: `cd "D:/Codding/Project/Suluk-Manajemen-SAAS" && go test ./internal/aiocr/service/`
Expected: PASS.

- [ ] **Step 5: Commit**
```bash
git add internal/aiocr/service/opencode.go internal/aiocr/service/opencode_test.go
git commit -m "feat(aiocr): OpenCode vision analyzer + pdftoppm rasterization + factory"
```

---

### Task 5: Wire OCR provider into ai-ocr-service

**Files:**
- Modify: `cmd/ai-ocr-service/main.go` (lines 63-64: `service.NewGeminiClient(cfg.Gemini.APIKey)` → `service.NewAnalyzer(cfg)`)

**Interfaces:**
- Consumes: `service.NewAnalyzer(cfg)` (Task 4).

- [ ] **Step 1: Wire the factory** — `cmd/ai-ocr-service/main.go`, change lines 63-64 to:
```go
	analyzer := service.NewAnalyzer(cfg)
	aiocrService := service.NewAIOCRService(aiocrRepo, analyzer, logger)
```

- [ ] **Step 2: Verify build + tests**

Run: `cd "D:/Codding/Project/Suluk-Manajemen-SAAS" && go build ./... && go test ./internal/aiocr/...`
Expected: build succeeds; tests PASS.

- [ ] **Step 3: Commit**
```bash
git add cmd/ai-ocr-service/main.go
git commit -m "feat(aiocr): wire switchable analyzer into ai-ocr-service"
```

---

### Task 6: Deployment — poppler-utils + AI env vars

**Files:**
- Modify: `deployments/Dockerfile` (re-declare `ARG SERVICE` in the runtime stage; conditional `apk add poppler-utils` for `ai-ocr-service`)
- Modify: `deployments/docker-compose.yml` (add the 4 AI env vars to the `ai-ocr-service` env block after line 316 and the `accounting-service` env block after line 514)

**Interfaces:**
- Consumes: the `AI_PROVIDER` / `OPENCODE_*` env contract (Tasks 1, 4).

- [ ] **Step 1: Dockerfile** — in `deployments/Dockerfile`, change the runtime stage (currently lines 12-14) to:
```dockerfile
FROM alpine:3.19

ARG SERVICE
RUN apk --no-cache add ca-certificates tzdata && \
    if [ "$SERVICE" = "ai-ocr-service" ]; then apk --no-cache add poppler-utils; fi
```
(`pdftoppm` ships in `poppler-utils`; only the OCR image gets it, keeping `CGO_ENABLED=0` and other images lean.)

- [ ] **Step 2: docker-compose** — add to the `ai-ocr-service` `environment:` block (right after `GEMINI_API_KEY: ${GEMINI_API_KEY:-}`, line 316) AND to the `accounting-service` `environment:` block (right after `GEMINI_API_KEY: ${GEMINI_API_KEY:-} ...`, line 514) the same four lines:
```yaml
      AI_PROVIDER: ${AI_PROVIDER:-opencode}
      OPENCODE_API_KEY: ${OPENCODE_API_KEY:-}
      OPENCODE_MODEL: ${OPENCODE_MODEL:-gpt-5-nano}
      OPENCODE_BASE_URL: ${OPENCODE_BASE_URL:-https://opencode.ai/zen/v1}
```

- [ ] **Step 3: Verify**

Run: `cd "D:/Codding/Project/Suluk-Manajemen-SAAS" && go build ./... && docker compose -f deployments/docker-compose.yml config >/dev/null && echo COMPOSE_OK`
Expected: build clean; `COMPOSE_OK` printed (compose parses; env substitution valid). The image build itself (with `poppler-utils`) is verified at deploy time.

- [ ] **Step 4: Commit**
```bash
git add deployments/Dockerfile deployments/docker-compose.yml
git commit -m "build(aiocr): poppler-utils for OCR image + OpenCode env wiring"
```

---

## Final verification (after all tasks)
- [ ] `cd "D:/Codding/Project/Suluk-Manajemen-SAAS" && go build ./... && go test ./...` — all green.
- [ ] `go vet ./internal/aiocr/... ./internal/shared/ai/... ./internal/accounting/...` — clean.
- [ ] Manual (post-deploy, jni-server): set `OPENCODE_API_KEY` + `AI_PROVIDER=opencode` (keep `GEMINI_API_KEY`) in `deployments/.env`; rebuild + recreate `ai-ocr-service` and `accounting-service`; scan a real KTP/passport (incl. one PDF) and trigger an accounting-insights call; eyeball extraction quality vs Gemini. To roll back: `AI_PROVIDER=gemini` + restart (no code change).

## Deferred (future)
- Per-feature models (`OPENCODE_OCR_MODEL` / `OPENCODE_TEXT_MODEL`) if insights want a stronger model than OCR.
- Multi-page PDF OCR (currently page 1 only).
