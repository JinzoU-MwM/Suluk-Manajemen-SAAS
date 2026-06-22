# Switchable AI Provider: Gemini ↔ OpenCode Zen — Design

- **Date:** 2026-06-22
- **Status:** Approved — proceeding to implementation plan
- **Area:** `internal/aiocr` (OCR/vision), `internal/shared/ai` (text), `internal/shared/config`, `cmd/ai-ocr-service`, `cmd/accounting-service`, `deployments/Dockerfile`

## 1. Context & problem

Suluk uses **Google Gemini** for two AI features:

1. **OCR scanner** (`internal/aiocr`) — `GeminiClient.AnalyzeDocument(ctx, imageData, mimeType) (*GeminiResult, error)` extracts structured fields (`doc_type`, `extracted_data`, `confidence`) from KTP / KK / Paspor / Visa images via `generativelanguage.googleapis.com/v1beta/models/{model}:generateContent` (auth `x-goog-api-key`, model default `gemini-2.5-flash`, inline base64 image, `responseMimeType: application/json`, `thinkingBudget: 0`). Constructed in `cmd/ai-ocr-service/main.go:63`; called in `internal/aiocr/service/process_sync.go:59`. The handler (`internal/aiocr/handler/handler.go:66-87`) accepts `image/jpeg|png|gif|webp` **and `application/pdf`**.
2. **Accounting insights** (`internal/shared/ai`) — `Client.GenerateText(ctx, prompt) (string, error)` for the accounting copilot. Constructed in `cmd/accounting-service/main.go:54` via `service.NewService(...).WithAI(ai.New(cfg.Gemini.APIKey))`.

The user wants to switch the AI provider to **OpenCode Zen** — an OpenAI-compatible chat-completions API at `https://opencode.ai/zen/v1/chat/completions` (Bearer auth, default model `gpt-5-nano`) already used for OCR in a sibling project (`opencode_ocr.py`). OpenCode Zen handles **both** vision (via `image_url` data-URI) and text.

## 2. Goals

- Migrate **both** AI features (OCR + insights) to OpenCode Zen.
- **Switchable** at runtime via `AI_PROVIDER` env (`opencode` default | `gemini`): both implementations live behind a small per-feature interface, selected by a factory. Flipping back to Gemini is an env change + restart — **no code change, no redeploy**.
- Preserve external behaviour: same extracted-field schema, same graceful-degradation (missing key → AI unavailable, non-fatal), same caching (the cache lives in the service layer, above the provider).
- Add **PDF support on OpenCode** by rasterizing the first page to an image (GPT vision cannot take a PDF).
- Everything env-tunable; reversible.

## 3. Non-goals

- Removing the Gemini code (kept as the fallback provider — that is the whole point of switchable).
- Changing prompts, the extracted-field schema, the cache layer, or any handler/route.
- Per-feature model selection (one shared `OPENCODE_MODEL` for now; can split into `OPENCODE_OCR_MODEL` / `OPENCODE_TEXT_MODEL` later if needed — YAGNI).
- Multi-page PDF OCR (only page 1 is rasterized, matching how single-page ID documents are handled today).

## 4. Config (`internal/shared/config/config.go`)

Add an `AIConfig` and keep `GeminiConfig`:

```go
type AIConfig struct {
    Provider        string // "opencode" (default) | "gemini"
    OpenCodeAPIKey  string
    OpenCodeModel   string // default "gpt-5-nano"
    OpenCodeBaseURL string // default "https://opencode.ai/zen/v1"
}
```

Loaded from env: `AI_PROVIDER` (default `opencode`), `OPENCODE_API_KEY`, `OPENCODE_MODEL` (default `gpt-5-nano`), `OPENCODE_BASE_URL` (default `https://opencode.ai/zen/v1`). `GEMINI_API_KEY` / `GEMINI_MODEL` stay as-is. `Provider` is normalized (lowercase/trim); an unknown value falls back to `opencode` with a warning.

## 5. OCR / vision (`internal/aiocr/service/`)

Extract a provider-agnostic interface (new `analyzer.go`):

```go
type DocumentAnalyzer interface {
    AnalyzeDocument(ctx context.Context, imageData []byte, mimeType string) (*OCRResult, error)
    Available() bool
}
```

- Rename `GeminiResult` → `OCRResult` (it is the OCR output, not Gemini-specific); `ExtractedFields`, `systemPrompts`, `cleanJSONString`, `detectMimeType` stay and are shared by both impls.
- `gemini.go`: existing `GeminiClient` keeps `AnalyzeDocument` (returns `*OCRResult`) and gains `Available()` (`c != nil && c.apiKey != ""`).
- `opencode.go` (new): `OpenCodeAnalyzer` implements the interface. It POSTs OpenAI-style to `{baseURL}/chat/completions`:
  ```json
  {"model": "<OPENCODE_MODEL>", "temperature": 0.1, "max_tokens": 4096,
   "messages": [{"role":"user","content":[
      {"type":"text","text": systemPrompts["auto"]},
      {"type":"image_url","image_url":{"url":"data:<mime>;base64,<b64>"}}]}]}
  ```
  Auth `Authorization: Bearer <OPENCODE_API_KEY>`. Parse `choices[0].message.content` → `cleanJSONString` → `json.Unmarshal` into `OCRResult`. **Retry** on HTTP 429 / 5xx / timeout / network error (exponential backoff, max 3 attempts), mirroring `opencode_ocr.py`.
- **PDF rasterization:** when `mimeType == "application/pdf"`, `OpenCodeAnalyzer` first calls `rasterizePDF(data) ([]byte, error)` — shells out to `pdftoppm -png -r 200 -f 1 -l 1` (writes the PDF to an `os.CreateTemp` file, reads back the page-1 PNG, cleans up), then sends the PNG with mime `image/png`. On rasterization failure it returns a clear Indonesian error. The Gemini path is unchanged (sends the PDF natively).
- Factory `NewAnalyzer(cfg config.Config) DocumentAnalyzer`: returns `OpenCodeAnalyzer` when `cfg.AI.Provider == "opencode"` (and key present), else `GeminiClient`. A nil/empty-key result is the existing graceful "unavailable" state.

Wiring: `cmd/ai-ocr-service/main.go:63` `service.NewGeminiClient(cfg.Gemini.APIKey)` → `service.NewAnalyzer(cfg)`. The service field `s.gemini *GeminiClient` → `s.analyzer DocumentAnalyzer` (param type of `NewAIOCRService` and the call site `process_sync.go:59` `s.gemini.AnalyzeDocument` → `s.analyzer.AnalyzeDocument`).

## 6. Text / insights (`internal/shared/ai/`)

Add a provider-agnostic interface (new `generator.go`):

```go
type Generator interface {
    GenerateText(ctx context.Context, prompt string) (string, error)
    Available() bool
}
```

- `gemini.go`: existing `Client` already satisfies it (`GenerateText`, `Available`).
- `opencode.go` (new): `openCodeGenerator` POSTs `{baseURL}/chat/completions` with `messages: [{role:"user", content: prompt}]`, temperature 0.3, parse `choices[0].message.content`. Same Bearer auth + retry as the OCR client.
- Factory: change `New(apiKey string) *Client` → `New(cfg config.Config) Generator` returning the right impl by `cfg.AI.Provider`. (Keep an unexported `newGemini(apiKey)` for the gemini branch so the existing logic is intact.)

Wiring: `cmd/accounting-service/main.go:54` `ai.New(cfg.Gemini.APIKey)` → `ai.New(cfg)`. The accounting service's `WithAI(...)` param type changes `*ai.Client` → `ai.Generator` (it only calls `GenerateText` / `Available`).

## 7. Deployment (`deployments/Dockerfile`)

The shared multi-service Dockerfile builds `CGO_ENABLED=0` on `golang:1.24-alpine` → `alpine:3.19`, parameterized by the `SERVICE` build arg. Add `poppler-utils` (provides `pdftoppm`) to the runtime **only for the OCR service**, keeping `CGO_ENABLED=0` and other images lean:

```dockerfile
ARG SERVICE
RUN if [ "$SERVICE" = "ai-ocr-service" ]; then apk add --no-cache poppler-utils; fi
```

Env additions (`.env` / compose): `AI_PROVIDER=opencode`, `OPENCODE_API_KEY=...`, `OPENCODE_MODEL=gpt-5-nano`. Keep `GEMINI_API_KEY` set so a flip back to `AI_PROVIDER=gemini` works instantly.

## 8. Testing

Pure, hermetic `httptest` unit tests (Go's `net/http/httptest`), no live API calls:

- **OpenCode OCR client:** server asserts `Authorization: Bearer`, model, and the `messages[].content` shape (text + `image_url` data-URI); returns a canned `choices[0].message.content` JSON → assert it parses into `OCRResult`. A 429-then-200 case asserts the retry. A `application/pdf` input is covered by injecting a fake rasterizer (the `rasterizePDF` call is a package var / small interface so the test can stub it without invoking `pdftoppm`).
- **OpenCode text client:** asserts request shape + `choices[0].message.content` extraction + retry.
- **Factory:** `NewAnalyzer` / `ai.New` return the OpenCode impl when `AI_PROVIDER=opencode`, the Gemini impl when `=gemini`, and an unavailable impl when the selected provider's key is empty.
- Existing Gemini tests (if any) stay green; the rename `GeminiResult → OCRResult` is updated across references.
- Gates: `go build ./...`, `go test ./...`.

## 9. Edge cases

- Missing `OPENCODE_API_KEY` while `AI_PROVIDER=opencode` → factory returns an unavailable client; OCR/insights degrade gracefully (same as Gemini today).
- `pdftoppm` missing at runtime (image not rebuilt) → rasterize returns a clear error; the scan job fails with a readable message rather than crashing.
- Multi-page PDF → only page 1 is rasterized (documented limitation).
- OpenCode returns non-JSON / fenced JSON → `cleanJSONString` strips fences before unmarshal (reused from the Gemini path).
- Unknown `AI_PROVIDER` value → normalized to `opencode` with a startup warning.

## 10. Rollout

- Land behind `AI_PROVIDER` (default `opencode`). Set `OPENCODE_API_KEY` + keep `GEMINI_API_KEY` in the server `.env`.
- Rebuild + redeploy `ai-ocr-service` (needs the `poppler-utils` image layer) and `accounting-service`.
- Verify a real KTP/passport scan + an accounting-insights call on OpenCode; compare extraction quality against Gemini by eye.
- If OpenCode underperforms on Indonesian ID extraction, flip `AI_PROVIDER=gemini` + restart — no code change.
