# Phase 4A (cont.) — Surface Scan Usage on Subscription Status

> Sub-phase of **Phase 4** in `2026-06-24-pricing-tier-restructure.md`. Phase 4a (committed `8790413`) added the *meter* (write-path). This increment lights up the *read* side.

**Goal:** Make `GET /api/v1/subscription/status` return `usage_count` (org's scans this calendar month) and `usage_limit` (tier `MaxScansPerMonth`), so the **already-built** quota UI (`ScannerPage`, `SubscriptionBanner`, `ProfilePage`, super-admin `UserManagement`) renders live numbers instead of dark bars.

**Out of scope (Phase 4b):** Pakasir top-up SKU (Rp49k/100), the "kuota habis → beli tambahan" purchase prompt, and the Pro/Bisnis fair-use alert. This increment is read-only surfacing — no billing.

## Architecture decision

The counter lives in `jamaah_aiocr` (separate DB); `/subscription/status` is served by auth (`jamaah_auth`). **Chosen: auth enriches the status response** — it owns the limit (from `plan.go`) and fetches the count from a new AI-OCR internal endpoint. Mirrors the existing `limits.go` cross-service read: short per-org cache + fail-open, so an AI-OCR hiccup never breaks status. Zero frontend changes (the contract already expects usage on the subscription object). Alternative considered (AI-OCR exposes + frontend merges) rejected to keep the client contract single-call and reuse the blessed fail-open pattern.

```
Frontend (unchanged)
  └─ getSubscriptionStatus() ─▶ auth /subscription/status
                                   ├─ usage_limit = plan.Get(plan).MaxScansPerMonth   (local)
                                   └─ usage_count ─▶ POST aiocr /api/v1/internal/scan-usage  (X-Internal-Key)
                                                       └─ GetScanUsageThisMonth(org)   ← breadcrumb from 4a
```

## Global constraints

- `Unlimited = -1` for Pro/Bisnis/Enterprise `usage_limit`; the UI gates the quota bar behind `!pro`, so the sentinel never renders a broken bar.
- Cross-service call is best-effort: missing addr/key or any error → `usage_count = 0` (fail-open), cached 45s per org (matches `limits.go`).
- Internal endpoint guarded by `X-Internal-Key` (constant-time), **not** behind `AuthMiddleware` — same pattern as auth's `/api/v1/internal/*` and invoice's `/internal/settle`.
- Commit messages: NO AI co-author line.

## Tasks (TDD)

### Task A — AI-OCR: internal scan-usage endpoint
- `service.GetScanUsageThisMonth(ctx, orgID) (int, error)` — thin delegate; `repo == nil → 0, nil` (keeps the handler unit-testable without a DB).
- `handler.ScanUsageInternal` — `validInternalKey` guard → parse `{org_id}` → return `{documents_scanned}`.
- Route: `app.Post("/api/v1/internal/scan-usage", h.ScanUsageInternal)` outside `authMW`.
- Tests: 401 missing/bad key; 400 bad body; 200 happy path (nil repo → `0`).

### Task B — auth: `usage_limit` from plan
- Add `UsageCount int` / `UsageLimit int` to `SubscriptionStatusResponse`.
- `statusResponse` sets `UsageLimit = t.MaxScansPerMonth`.
- Test: `statusResponse(Starter)=100`, `(Pro)=-1`, `(Gratis)=5`.

### Task C — auth: fetch `usage_count` from AI-OCR
- `AuthService` gains `aiocrAddr`, `internalKey`, `scanUsageCache sync.Map` + `WithScanUsageSource(addr, key)` option.
- `scanUsageThisMonth(ctx, orgID) int` — `httpc.PostJSON` w/ `X-Internal-Key`, fail-open → 0, 45s cache.
- Test (httptest): returns count on 200; `0` on 5xx/network (fail-open); 2nd call from cache (1 server hit).

### Task D — wire + config
- `GetSubscriptionStatus`: `resp.UsageCount = s.scanUsageThisMonth(ctx, orgID)`.
- `cmd/auth-service/main.go`: `.WithScanUsageSource(os.Getenv("AIOCR_SERVICE_ADDR"), cfg.Internal.APIKey)`.
- Add `AIOCR_SERVICE_ADDR` to auth's env in `.env.example` + `deployments/docker-compose.yml` if missing.

### Task E — verify + commit
- `go build ./...`; `go test ./internal/auth/... ./internal/aiocr/... ./internal/shared/plan/...` green.
- Confirm field names match the frontend (`usage_count` / `usage_limit`).
- Commit: `feat(aiocr): surface monthly scan usage on subscription status (Phase 4a)`.

## Self-review
- **Contract:** frontend already reads `usage_count`/`usage_limit` off the subscription object → no FE change. ✓
- **Fail-open:** AI-OCR down → status still returns (count 0). ✓
- **Reuse:** uses 4a's `GetScanUsageThisMonth`; mirrors `limits.go` cache/fail-open + auth's `validInternalKey`. ✓
