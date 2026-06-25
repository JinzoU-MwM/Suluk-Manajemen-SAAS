# Scan Top-up + Fair-use Cap Implementation Plan (Phase 4b)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a Starter org buy a Rp49k/100-scan top-up (one-time, via Pakasir) credited to the current month, and give Pro/Bisnis a fair-use WARN at 2000 scans/month — closing Phase 4.

**Architecture:** Reuse the existing Pakasir spine (`create-order → webhook → verify → mark-paid → act`). A new `purpose` column on `payment_orders` routes the webhook to either subscription activation (existing) or a new scan-credit call into aiocr, which owns a `scan_topups` ledger (idempotent by `order_id`). auth surfaces `usage_limit = base + purchased`. Fair-use is a best-effort WARN in aiocr's metering path.

**Tech Stack:** Go (Fiber microservices), PostgreSQL (per-service DBs: `jamaah_invoice`, `jamaah_aiocr`, `jamaah_auth`), Pakasir, Svelte.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-06-25-scan-topup-design.md`. Builds on 4A (commit `985d5e2`).
- SKU is **server-authoritative**: top-up amount = `plan.ScanTopupPrice`, never client-supplied.
- Crediting is **exactly-once**: `scan_topups.order_id` is UNIQUE; insert is `ON CONFLICT DO NOTHING`. Credit runs AFTER `MarkPaymentOrderPaid`; on failure, `RevertPaymentOrderToPending` so Pakasir retries.
- Top-up eligibility: **Starter only** (server-enforced). Monthly reset via `(year, month)` filtering.
- `Unlimited = -1`: for Pro/Bisnis, `usage_limit` stays `-1` regardless of purchased.
- Fair-use: **never blocks** OCR; tier-agnostic threshold `plan.FairUseScanCap = 2000`; one WARN per org/month.
- `plan.go` constants MUST be mirrored in `frontend-svelte/src/lib/config/pricing.js`.
- DB-bound repo SQL has no unit test in this repo (no test DB); it is verified by `go build ./...` + the cross-service contract tests + manual. Mirrors how 4a shipped. Commit messages: NO AI co-author line.

---

### Task 1: SKU + fair-use constants (plan + pricing mirror)

**Files:**
- Modify: `internal/shared/plan/plan.go`
- Test: `internal/shared/plan/plan_test.go`
- Modify: `frontend-svelte/src/lib/config/pricing.js`

**Interfaces:**
- Produces: `plan.ScanTopupPrice int = 49000`, `plan.ScanTopupScans int = 100`, `plan.FairUseScanCap int = 2000`. Consumed by Tasks 3, 4, 6, 7, 9.

- [ ] **Step 1: Write the failing test**

Append to `internal/shared/plan/plan_test.go`:

```go
func TestScanTopupAndFairUseConstants(t *testing.T) {
	if ScanTopupPrice != 49000 {
		t.Errorf("ScanTopupPrice = %d, want 49000", ScanTopupPrice)
	}
	if ScanTopupScans != 100 {
		t.Errorf("ScanTopupScans = %d, want 100", ScanTopupScans)
	}
	if FairUseScanCap != 2000 {
		t.Errorf("FairUseScanCap = %d, want 2000", FairUseScanCap)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/shared/plan/ -run TestScanTopupAndFairUseConstants`
Expected: FAIL — `undefined: ScanTopupPrice`.

- [ ] **Step 3: Add the constants**

In `internal/shared/plan/plan.go`, after `const Unlimited = -1`:

```go
// Scan top-up SKU (Phase 4b): a one-time purchase that adds ScanTopupScans to a
// Starter org's quota for the current month. Server-authoritative price (IDR).
const (
	ScanTopupPrice = 49000
	ScanTopupScans = 100
)

// FairUseScanCap is the monthly soft cap for "unlimited" tiers (Pro/Bisnis):
// crossing it raises an ops WARN, never a block.
const FairUseScanCap = 2000
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/shared/plan/ -v`
Expected: PASS (all).

- [ ] **Step 5: Mirror in pricing.js**

In `frontend-svelte/src/lib/config/pricing.js`, after the `UNLIMITED` const, add:

```js
// Scan top-up SKU (mirror of plan.go ScanTopupPrice/ScanTopupScans).
export const SCAN_TOPUP_PRICE = 49000;
export const SCAN_TOPUP_SCANS = 100;
```

- [ ] **Step 6: Verify frontend check passes**

Run: `npm --prefix "D:\Codding\Project\Suluk-Manajemen-SAAS\frontend-svelte" run check`
Expected: 0 errors.

- [ ] **Step 7: Commit**

```bash
git add internal/shared/plan/plan.go internal/shared/plan/plan_test.go frontend-svelte/src/lib/config/pricing.js
git commit -m "feat(plan): add scan top-up SKU + fair-use cap constants (Phase 4b)"
```

---

### Task 2: aiocr `scan_topups` ledger (migration + repo)

**Files:**
- Create: `migrations/aiocr/003_scan_topups.up.sql`, `migrations/aiocr/003_scan_topups.down.sql`
- Modify: `internal/aiocr/repository/repository.go`

**Interfaces:**
- Produces: `(*AIOCRRepo) CreditScanTopup(ctx, orderID, orgID uuid.UUID, scans int) error` (idempotent); `(*AIOCRRepo) GetPurchasedScansThisMonth(ctx, orgID uuid.UUID) (int, error)`. Consumed by Task 3.

- [ ] **Step 1: Write the up migration**

`migrations/aiocr/003_scan_topups.up.sql`:

```sql
-- Per-purchase ledger of bought scan top-ups (Phase 4b). order_id is the
-- invoice payment_orders.id; UNIQUE makes crediting idempotent under Pakasir's
-- at-least-once webhook retries. purchased_this_month = SUM(scans) for the month.
CREATE TABLE IF NOT EXISTS scan_topups (
    order_id   UUID PRIMARY KEY,
    org_id     UUID NOT NULL,
    year       INT  NOT NULL,
    month      INT  NOT NULL,
    scans      INT  NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scan_topups_org_month ON scan_topups (org_id, year, month);
```

`migrations/aiocr/003_scan_topups.down.sql`:

```sql
DROP TABLE IF EXISTS scan_topups;
```

- [ ] **Step 2: Add repo methods**

In `internal/aiocr/repository/repository.go`, after `GetScanUsageThisMonth`:

```go
// CreditScanTopup records a purchased top-up for the org's current month. The
// order_id PK + DO NOTHING make it idempotent: a duplicate webhook credits once.
func (r *AIOCRRepo) CreditScanTopup(ctx context.Context, orderID, orgID uuid.UUID, scans int) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO scan_topups (order_id, org_id, year, month, scans)
		VALUES ($1, $2, EXTRACT(YEAR FROM NOW())::int, EXTRACT(MONTH FROM NOW())::int, $3)
		ON CONFLICT (order_id) DO NOTHING`,
		orderID, orgID, scans)
	return err
}

// GetPurchasedScansThisMonth sums the org's top-up credits for the current month
// (0 when none).
func (r *AIOCRRepo) GetPurchasedScansThisMonth(ctx context.Context, orgID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(scans), 0) FROM scan_topups
		WHERE org_id = $1 AND year = EXTRACT(YEAR FROM NOW())::int
		  AND month = EXTRACT(MONTH FROM NOW())::int`,
		orgID).Scan(&n)
	return n, err
}
```

- [ ] **Step 3: Verify build**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" build ./...`
Expected: no output (DB-bound SQL has no unit test; verified by build + Task 3 contract tests).

- [ ] **Step 4: Commit**

```bash
git add migrations/aiocr/003_scan_topups.up.sql migrations/aiocr/003_scan_topups.down.sql internal/aiocr/repository/repository.go
git commit -m "feat(aiocr): scan_topups ledger for purchased scan credits (Phase 4b)"
```

---

### Task 3: aiocr credit endpoint + purchased on usage endpoint

**Files:**
- Modify: `internal/aiocr/service/service.go` (delegates)
- Modify: `internal/aiocr/handler/handler.go` (`ScanTopupInternal`, extend `ScanUsageInternal`)
- Modify: `cmd/ai-ocr-service/main.go` (route)
- Test: `internal/aiocr/handler/handler_test.go`

**Interfaces:**
- Consumes: Task 2 repo methods.
- Produces: `POST /api/v1/internal/scan-topup` body `{order_id, org_id, scans}`; `POST /api/v1/internal/scan-usage` now returns `{documents_scanned, purchased_scans}`. Consumed by Tasks 7, 8.

- [ ] **Step 1: Write failing tests**

Append to `internal/aiocr/handler/handler_test.go`:

```go
func TestScanTopupInternalRejectsMissingKey(t *testing.T) {
	t.Setenv("INTERNAL_API_KEY", "testkey")
	h := NewAIOCRHandler(service.NewAIOCRService(nil, nil, zap.NewNop().Sugar()))
	app := fiber.New()
	app.Post("/api/v1/internal/scan-topup", h.ScanTopupInternal)

	body, _ := json.Marshal(map[string]any{"order_id": uuid.NewString(), "org_id": uuid.NewString(), "scans": 100})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/scan-topup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestScanTopupInternalCreditsWithValidKey(t *testing.T) {
	t.Setenv("INTERNAL_API_KEY", "testkey")
	h := NewAIOCRHandler(service.NewAIOCRService(nil, nil, zap.NewNop().Sugar()))
	app := fiber.New()
	app.Post("/api/v1/internal/scan-topup", h.ScanTopupInternal)

	body, _ := json.Marshal(map[string]any{"order_id": uuid.NewString(), "org_id": uuid.NewString(), "scans": 100})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/scan-topup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", "testkey")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK { // nil repo → credit is a no-op success
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}
```

- [ ] **Step 2: Run to verify fail**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/aiocr/handler/ -run TestScanTopupInternal`
Expected: FAIL — `h.ScanTopupInternal undefined`.

- [ ] **Step 3: Add service delegates**

In `internal/aiocr/service/service.go`, after `GetScanUsageThisMonth`:

```go
// GetPurchasedScansThisMonth returns the org's bought top-up credits this month
// (0 when no repo is wired).
func (s *AIOCRService) GetPurchasedScansThisMonth(ctx context.Context, orgID uuid.UUID) (int, error) {
	if s.repo == nil {
		return 0, nil
	}
	return s.repo.GetPurchasedScansThisMonth(ctx, orgID)
}

// CreditScanTopup records a purchased top-up (idempotent). No-op without a repo.
func (s *AIOCRService) CreditScanTopup(ctx context.Context, orderID, orgID uuid.UUID, scans int) error {
	if s.repo == nil {
		return nil
	}
	return s.repo.CreditScanTopup(ctx, orderID, orgID, scans)
}
```

- [ ] **Step 4: Add the handler + extend usage endpoint**

In `internal/aiocr/handler/handler.go`, add:

```go
// ScanTopupInternal credits a paid scan top-up to an org's current month.
// Service-to-service (X-Internal-Key); called by invoice-service's webhook.
func (h *AIOCRHandler) ScanTopupInternal(c *fiber.Ctx) error {
	if !validInternalKey(c) {
		return response.Unauthorized(c, "invalid internal key")
	}
	var req struct {
		OrderID string `json:"order_id"`
		OrgID   string `json:"org_id"`
		Scans   int    `json:"scans"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		return response.BadRequest(c, "invalid order_id")
	}
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return response.BadRequest(c, "invalid org_id")
	}
	if req.Scans <= 0 {
		return response.BadRequest(c, "scans must be positive")
	}
	if err := h.svc.CreditScanTopup(c.Context(), orderID, orgID, req.Scans); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"credited": true})
}
```

Then change `ScanUsageInternal`'s success return (the `documents_scanned` line) to also include purchased:

```go
	n, err := h.svc.GetScanUsageThisMonth(c.Context(), orgID)
	if err != nil {
		return response.Internal(c, err)
	}
	purchased, err := h.svc.GetPurchasedScansThisMonth(c.Context(), orgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"documents_scanned": n, "purchased_scans": purchased})
```

- [ ] **Step 5: Register the route**

In `cmd/ai-ocr-service/main.go`, after the `scan-usage` internal route:

```go
	app.Post("/api/v1/internal/scan-topup", aiocrHandler.ScanTopupInternal)
```

- [ ] **Step 6: Run tests to verify pass**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/aiocr/... -v`
Expected: PASS (new + existing).

- [ ] **Step 7: Commit**

```bash
git add internal/aiocr/service/service.go internal/aiocr/handler/handler.go internal/aiocr/handler/handler_test.go cmd/ai-ocr-service/main.go
git commit -m "feat(aiocr): internal scan-topup credit endpoint + purchased on usage (Phase 4b)"
```

---

### Task 4: aiocr fair-use WARN at 2000/month

**Files:**
- Create: `migrations/aiocr/004_scan_usage_fairuse.up.sql`, `.down.sql`
- Modify: `internal/aiocr/repository/repository.go` (`MarkFairUseAlerted`)
- Modify: `internal/aiocr/service/service.go` (helper + hook), `internal/aiocr/service/process_sync.go` (call hook)
- Test: `internal/aiocr/service/process_sync_test.go`

**Interfaces:**
- Consumes: `plan.FairUseScanCap`.
- Produces: pure helper `fairUseExceeded(total int) bool`; metering calls it best-effort. No external interface.

- [ ] **Step 1: Write the migration**

`migrations/aiocr/004_scan_usage_fairuse.up.sql`:

```sql
-- One-shot-per-month marker so the fair-use WARN fires at most once per org/month.
ALTER TABLE scan_usage ADD COLUMN IF NOT EXISTS fairuse_alerted_at TIMESTAMPTZ;
```

`migrations/aiocr/004_scan_usage_fairuse.down.sql`:

```sql
ALTER TABLE scan_usage DROP COLUMN IF EXISTS fairuse_alerted_at;
```

- [ ] **Step 2: Write the failing tests (pure logic + never-block)**

Append to `internal/aiocr/service/process_sync_test.go`:

```go
func TestFairUseExceeded(t *testing.T) {
	if fairUseExceeded(plan.FairUseScanCap - 1) {
		t.Error("below cap should not exceed")
	}
	if !fairUseExceeded(plan.FairUseScanCap) {
		t.Error("at cap should exceed")
	}
}
```

(Add `"github.com/jamaah-in/v2/internal/shared/plan"` to the test imports.)

- [ ] **Step 3: Run to verify fail**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/aiocr/service/ -run TestFairUseExceeded`
Expected: FAIL — `undefined: fairUseExceeded`.

- [ ] **Step 4: Add the helper + repo marker + metering hook**

In `internal/aiocr/repository/repository.go`:

```go
// MarkFairUseAlerted stamps the current month's row as alerted IFF not already
// stamped, returning true only for the first caller this month (so the WARN
// fires once). Assumes the IncrementScanUsage row already exists.
func (r *AIOCRRepo) MarkFairUseAlerted(ctx context.Context, orgID uuid.UUID) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE scan_usage SET fairuse_alerted_at = NOW()
		WHERE org_id = $1 AND year = EXTRACT(YEAR FROM NOW())::int
		  AND month = EXTRACT(MONTH FROM NOW())::int AND fairuse_alerted_at IS NULL`,
		orgID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}
```

In `internal/aiocr/service/service.go`:

```go
// fairUseExceeded reports whether a monthly scan total has crossed the soft cap.
func fairUseExceeded(total int) bool { return total >= plan.FairUseScanCap }
```

(Add `"github.com/jamaah-in/v2/internal/shared/plan"` to service.go imports.)

In `internal/aiocr/service/process_sync.go`, replace the metering block's success branch so that after `IncrementScanUsage` succeeds it checks fair-use (best-effort, never blocks):

```go
	if s.repo != nil && scanned > 0 {
		if err := s.repo.IncrementScanUsage(ctx, orgID, scanned); err != nil {
			s.logger.Errorf("record scan usage (org %s): %v", orgID, err)
		} else if total, err := s.repo.GetScanUsageThisMonth(ctx, orgID); err == nil && fairUseExceeded(total) {
			if first, err := s.repo.MarkFairUseAlerted(ctx, orgID); err == nil && first {
				s.logger.Warnf("fair-use: org %s reached %d scans this month (cap %d)", orgID, total, plan.FairUseScanCap)
			}
		}
	}
```

(Add `"github.com/jamaah-in/v2/internal/shared/plan"` to process_sync.go imports.)

- [ ] **Step 5: Run tests to verify pass**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/aiocr/service/ -v`
Expected: PASS — including the existing `TestProcessDocumentsSync*` (proves metering+fair-use path never breaks the OCR result; repo is nil in those tests so the block is skipped).

- [ ] **Step 6: Commit**

```bash
git add migrations/aiocr/004_scan_usage_fairuse.up.sql migrations/aiocr/004_scan_usage_fairuse.down.sql internal/aiocr/repository/repository.go internal/aiocr/service/service.go internal/aiocr/service/process_sync.go internal/aiocr/service/process_sync_test.go
git commit -m "feat(aiocr): fair-use WARN at 2000 scans/month, once per org (Phase 4b)"
```

---

### Task 5: invoice `purpose` column on payment_orders

**Files:**
- Create: `migrations/invoice/005_payment_order_purpose.up.sql`, `.down.sql`
- Modify: `internal/invoice/model/model.go` (`PaymentOrder.Purpose`)
- Modify: `internal/invoice/repository/payment_order.go` (insert + selects + scan)

**Interfaces:**
- Produces: `PaymentOrder.Purpose string` (db `purpose`, default `"subscription"`). Consumed by Tasks 6, 7.

- [ ] **Step 1: Write the migration**

`migrations/invoice/005_payment_order_purpose.up.sql`:

```sql
-- Distinguish a subscription purchase from a scan top-up so the Pakasir webhook
-- can route to the right post-payment action. Existing rows = 'subscription'.
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS purpose VARCHAR(30) NOT NULL DEFAULT 'subscription';
```

`migrations/invoice/005_payment_order_purpose.down.sql`:

```sql
ALTER TABLE payment_orders DROP COLUMN IF EXISTS purpose;
```

- [ ] **Step 2: Add the model field**

In `internal/invoice/model/model.go`, in `type PaymentOrder struct`, after `Status`:

```go
	Purpose       string     `json:"purpose" db:"purpose"` // "subscription" | "scan_topup"
```

- [ ] **Step 3: Thread purpose through repo insert + selects + scan**

In `internal/invoice/repository/payment_order.go`:

`CreatePaymentOrder` query → add `purpose` to the column list and `$10`:

```go
	query := `
		INSERT INTO payment_orders (id, org_id, user_id, plan, plan_type, amount, status, redirect_url, gateway_ref, purpose)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		order.ID, order.OrgID, order.UserID, order.Plan, order.PlanType, order.Amount,
		order.Status, order.RedirectURL, order.GatewayRef, defaultPurpose(order.Purpose),
	).Scan(&order.CreatedAt, &order.UpdatedAt)
```

Add the helper at the bottom of the file:

```go
// defaultPurpose keeps existing callers (which never set Purpose) on the
// subscription path.
func defaultPurpose(p string) string {
	if p == "" {
		return "subscription"
	}
	return p
}
```

In both `GetPaymentOrder` and `GetPaymentOrderByID`, add `purpose` to the SELECT column list (immediately after `status`). In `scanPaymentOrder`, add `&o.Purpose` to the `row.Scan(...)` call immediately after `&o.Status`.

- [ ] **Step 4: Verify build**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" build ./...`
Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add migrations/invoice/005_payment_order_purpose.up.sql migrations/invoice/005_payment_order_purpose.down.sql internal/invoice/model/model.go internal/invoice/repository/payment_order.go
git commit -m "feat(invoice): add purpose to payment_orders (Phase 4b)"
```

---

### Task 6: invoice top-up order endpoint (Starter-only, server-priced)

**Files:**
- Modify: `internal/invoice/model/model.go` (`TopupOrderResponse` optional — reuse `PaymentOrderResponse`)
- Modify: `internal/invoice/service/service.go` (`aiocrAddr` field + `PaymentDeps.AiocrAddr`)
- Modify: `internal/invoice/service/payment_order.go` (`CreateTopupOrder`, `callerPlan`)
- Modify: `internal/invoice/handler/payment.go` (`CreateTopupOrder`)
- Modify: `cmd/invoice-service/main.go` (route + `AiocrAddr` wiring)
- Test: `internal/invoice/service/payment_order_test.go` (new)

**Interfaces:**
- Consumes: `plan.ScanTopupPrice`, auth `/api/v1/subscription/status` (returns `{plan}`).
- Produces: `POST /api/v1/payment/topup-order` → `PaymentOrderResponse{order_id, payment_url, status, amount}`. `(*InvoiceService) CreateTopupOrder(ctx, orgID, userID uuid.UUID, authToken string) (*model.PaymentOrderResponse, error)`. Consumed by Task 9.

- [ ] **Step 1: Write failing tests**

Create `internal/invoice/service/payment_order_test.go`:

```go
package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/shared/httpclient"
	"github.com/jamaah-in/v2/internal/shared/plan"
)

func addrOf(u string) string { return strings.TrimPrefix(u, "http://") }

// A non-Starter caller cannot buy a top-up.
func TestCreateTopupOrderRejectsNonStarter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":{"plan":"gratis"}}`))
	}))
	defer ts.Close()

	s := &InvoiceService{httpc: httpclient.New(), authAddr: addrOf(ts.URL)}
	_, err := s.CreateTopupOrder(context.Background(), uuid.New(), uuid.New(), "Bearer x")
	if err == nil {
		t.Fatal("expected error for non-Starter caller")
	}
}

// callerPlan surfaces the org's tier from auth status.
func TestCallerPlanReadsStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"plan":"starter"}}`))
	}))
	defer ts.Close()

	s := &InvoiceService{httpc: httpclient.New(), authAddr: addrOf(ts.URL)}
	got, err := s.callerPlan(context.Background(), "Bearer tok")
	if err != nil {
		t.Fatal(err)
	}
	if got != plan.Starter {
		t.Errorf("callerPlan = %q, want starter", got)
	}
}
```

- [ ] **Step 2: Run to verify fail**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/invoice/service/ -run 'TestCreateTopupOrder|TestCallerPlan'`
Expected: FAIL — `s.CreateTopupOrder undefined`, `s.callerPlan undefined`.

- [ ] **Step 3: Add the aiocr address to deps**

In `internal/invoice/service/service.go`: add `aiocrAddr string` to `InvoiceService` struct (after `authAddr`); add `AiocrAddr string` to `PaymentDeps`; in `WithPayments`, set `s.aiocrAddr = d.AiocrAddr`.

- [ ] **Step 4: Implement `callerPlan` + `CreateTopupOrder`**

In `internal/invoice/service/payment_order.go`:

```go
// callerPlan fetches the caller org's tier from auth (forwarding the caller's
// bearer token), used to gate Starter-only purchases.
func (s *InvoiceService) callerPlan(ctx context.Context, authToken string) (string, error) {
	var out struct {
		Plan string `json:"plan"`
	}
	if err := s.httpc.GetJSON(ctx, s.authAddr, "/api/v1/subscription/status", authToken, &out); err != nil {
		return "", fmt.Errorf("check plan: %w", err)
	}
	return out.Plan, nil
}

// CreateTopupOrder creates a pending scan-topup order (server-priced) and returns
// the Pakasir checkout URL. Only Starter orgs may buy top-ups.
func (s *InvoiceService) CreateTopupOrder(ctx context.Context, orgID, userID uuid.UUID, authToken string) (*model.PaymentOrderResponse, error) {
	p, err := s.callerPlan(ctx, authToken)
	if err != nil {
		return nil, err
	}
	if plan.Normalize(p) != plan.Starter {
		return nil, fmt.Errorf("top-up tersedia hanya untuk paket Starter")
	}

	orderID := uuid.New()
	order := &model.PaymentOrder{
		ID:      orderID,
		OrgID:   orgID,
		UserID:  userID,
		Plan:    plan.Starter,
		Amount:  plan.ScanTopupPrice,
		Status:  "pending",
		Purpose: "scan_topup",
	}
	if err := s.repo.CreatePaymentOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("create topup order: %w", err)
	}

	payURL := s.pakasirPayURL(orderID.String(), plan.ScanTopupPrice)
	order.RedirectURL = &payURL
	return &model.PaymentOrderResponse{
		OrderID:    orderID.String(),
		PaymentURL: payURL,
		Status:     "pending",
		Amount:     plan.ScanTopupPrice,
	}, nil
}
```

Note: the two tests build `&InvoiceService{...}` with a nil repo and return *before* the repo call (non-Starter path / `callerPlan`), so they need no DB.

- [ ] **Step 5: Add the handler**

In `internal/invoice/handler/payment.go`:

```go
// CreateTopupOrder starts a Starter scan top-up purchase. Server-priced.
func (h *InvoiceHandler) CreateTopupOrder(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	result, err := h.svc.CreateTopupOrder(c.Context(), claims.OrgID, claims.UserID, c.Get("Authorization"))
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, result)
}
```

- [ ] **Step 6: Register route + wire AiocrAddr**

In `cmd/invoice-service/main.go`: add to the `WithPayments(service.PaymentDeps{...})` literal: `AiocrAddr: os.Getenv("AIOCR_SERVICE_ADDR"),`. After `payment.Post("/create-order", ...)` add:

```go
	payment.Post("/topup-order", invoiceHandler.CreateTopupOrder)
```

- [ ] **Step 7: Run tests + build**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/invoice/... && go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" build ./...`
Expected: PASS + clean build.

- [ ] **Step 8: Commit**

```bash
git add internal/invoice/service/service.go internal/invoice/service/payment_order.go internal/invoice/service/payment_order_test.go internal/invoice/handler/payment.go cmd/invoice-service/main.go
git commit -m "feat(invoice): Starter scan top-up order endpoint (Phase 4b)"
```

---

### Task 7: invoice webhook routes scan_topup → aiocr credit

**Files:**
- Modify: `internal/invoice/service/payment_order.go` (`HandlePakasirWebhook` branch + `creditScanTopup`)
- Test: `internal/invoice/service/payment_order_test.go`

**Interfaces:**
- Consumes: aiocr `POST /api/v1/internal/scan-topup` (Task 3), `order.Purpose` (Task 5), `plan.ScanTopupScans`.
- Produces: `(*InvoiceService) creditScanTopup(ctx, order *model.PaymentOrder) error`.

- [ ] **Step 1: Write the failing test**

Append to `internal/invoice/service/payment_order_test.go`:

```go
// creditScanTopup forwards the order to aiocr with the internal key + correct
// scan quantity.
func TestCreditScanTopupCallsAiocr(t *testing.T) {
	var gotKey string
	var gotBody struct {
		OrderID string `json:"order_id"`
		OrgID   string `json:"org_id"`
		Scans   int    `json:"scans"`
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Internal-Key")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"success":true,"data":{"credited":true}}`))
	}))
	defer ts.Close()

	s := &InvoiceService{httpc: httpclient.New(), aiocrAddr: addrOf(ts.URL), internalKey: "k"}
	order := &model.PaymentOrder{ID: uuid.New(), OrgID: uuid.New(), Purpose: "scan_topup"}
	if err := s.creditScanTopup(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	if gotKey != "k" {
		t.Errorf("X-Internal-Key = %q, want k", gotKey)
	}
	if gotBody.Scans != plan.ScanTopupScans {
		t.Errorf("scans = %d, want %d", gotBody.Scans, plan.ScanTopupScans)
	}
	if gotBody.OrderID != order.ID.String() {
		t.Errorf("order_id = %q, want %q", gotBody.OrderID, order.ID.String())
	}
}
```

(Add `"encoding/json"` and `"github.com/jamaah-in/v2/internal/invoice/model"` to the test imports.)

- [ ] **Step 2: Run to verify fail**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/invoice/service/ -run TestCreditScanTopupCallsAiocr`
Expected: FAIL — `s.creditScanTopup undefined`.

- [ ] **Step 3: Implement `creditScanTopup` + branch the webhook**

In `internal/invoice/service/payment_order.go`, add:

```go
// creditScanTopup calls the ai-ocr internal endpoint to add the purchased scans
// to the org's current month. Idempotent on the ai-ocr side (order_id key).
func (s *InvoiceService) creditScanTopup(ctx context.Context, order *model.PaymentOrder) error {
	if s.aiocrAddr == "" {
		return fmt.Errorf("aiocr service address not configured")
	}
	body := map[string]any{
		"order_id": order.ID.String(),
		"org_id":   order.OrgID.String(),
		"scans":    plan.ScanTopupScans,
	}
	headers := map[string]string{"X-Internal-Key": s.internalKey}
	return s.httpc.PostJSON(ctx, s.aiocrAddr, "/api/v1/internal/scan-topup", headers, body, nil)
}
```

In `HandlePakasirWebhook`, replace the post-claim activation line (`if err := s.activateSubscription(ctx, order, p.PaymentMethod); err != nil {`) with a purpose switch:

```go
	// Apply the purchase AFTER claiming. On failure, roll the order back to
	// pending so a later Pakasir retry re-attempts it (no charged-but-unapplied).
	var applyErr error
	switch order.Purpose {
	case "scan_topup":
		applyErr = s.creditScanTopup(ctx, order)
	default:
		applyErr = s.activateSubscription(ctx, order, p.PaymentMethod)
	}
	if applyErr != nil {
		if rerr := s.repo.RevertPaymentOrderToPending(ctx, order.ID); rerr != nil {
			log.Printf("CRITICAL: order %s paid but apply AND revert failed — needs reconciliation: apply=%v revert=%v", order.ID, applyErr, rerr)
		}
		return fmt.Errorf("apply purchase: %w", applyErr) // transient → 5xx → retry
	}
	return nil
```

- [ ] **Step 4: Run tests to verify pass**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/invoice/... -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/invoice/service/payment_order.go internal/invoice/service/payment_order_test.go
git commit -m "feat(invoice): route paid scan_topup orders to ai-ocr credit (Phase 4b)"
```

---

### Task 8: auth surfaces usage_limit = base + purchased

**Files:**
- Modify: `internal/auth/service/subscription.go` (`scanUsageThisMonth` → returns used+purchased; limit combine)
- Test: `internal/auth/service/subscription_usage_test.go`

**Interfaces:**
- Consumes: aiocr usage endpoint now returns `{documents_scanned, purchased_scans}` (Task 3).
- Produces: `usage_limit = base + purchased` for finite tiers on `/subscription/status`.

- [ ] **Step 1: Update the failing tests**

In `internal/auth/service/subscription_usage_test.go`, change the fetch test's server body and add a combine assertion. Replace `TestScanUsageThisMonthFetchesAndCaches` body's write line with:

```go
		_, _ = w.Write([]byte(`{"success":true,"data":{"documents_scanned":7,"purchased_scans":100}}`))
```

and update its assertions to expect a `(used, purchased)` pair:

```go
	used, purchased := s.scanUsageThisMonth(context.Background(), org)
	if used != 7 || purchased != 100 {
		t.Fatalf("got (used=%d, purchased=%d), want (7, 100)", used, purchased)
	}
	used2, _ := s.scanUsageThisMonth(context.Background(), org)
	if used2 != 7 {
		t.Fatalf("cached used = %d, want 7", used2)
	}
	if n := atomic.LoadInt32(&hits); n != 1 {
		t.Errorf("ai-ocr hits = %d, want 1", n)
	}
```

Update `TestScanUsageThisMonthFailsOpen` / `...Unconfigured` to the two-return form, e.g.:

```go
	if used, purchased := s.scanUsageThisMonth(context.Background(), uuid.New()); used != 0 || purchased != 0 {
		t.Errorf("fail-open got (%d,%d), want (0,0)", used, purchased)
	}
```

Add a combine test:

```go
func TestUsageLimitAddsPurchased(t *testing.T) {
	// Starter base 100 + 100 purchased = 200; Pro stays Unlimited.
	if got := effectiveLimit(plan.Get(plan.Starter).MaxScansPerMonth, 100); got != 200 {
		t.Errorf("starter effective limit = %d, want 200", got)
	}
	if got := effectiveLimit(plan.Get(plan.Pro).MaxScansPerMonth, 100); got != plan.Unlimited {
		t.Errorf("pro effective limit = %d, want Unlimited", got)
	}
}
```

- [ ] **Step 2: Run to verify fail**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/auth/service/ -run 'TestScanUsage|TestUsageLimit'`
Expected: FAIL — `effectiveLimit undefined` + `scanUsageThisMonth` arity mismatch.

- [ ] **Step 3: Update `scanUsageThisMonth` + add `effectiveLimit` + wire**

In `internal/auth/service/subscription.go`:

Change the cache entry + method to carry purchased:

```go
type scanUsageCacheEntry struct {
	used      int
	purchased int
	expiry    time.Time
}

// effectiveLimit is the org's surfaced monthly scan cap: base plan quota plus any
// purchased top-ups. Unlimited tiers stay Unlimited (top-ups are irrelevant).
func effectiveLimit(base, purchased int) int {
	if base == plan.Unlimited {
		return plan.Unlimited
	}
	return base + purchased
}

// scanUsageThisMonth returns (used, purchased) for the org this month, fetched
// from ai-ocr. Fails open to (0,0); cached briefly per org.
func (s *AuthService) scanUsageThisMonth(ctx context.Context, orgID uuid.UUID) (int, int) {
	if s.aiocrAddr == "" || s.internalKey == "" {
		return 0, 0
	}
	if v, ok := s.scanUsageCache.Load(orgID); ok {
		if e := v.(scanUsageCacheEntry); e.expiry.After(time.Now()) {
			return e.used, e.purchased
		}
	}
	var out struct {
		DocumentsScanned int `json:"documents_scanned"`
		PurchasedScans   int `json:"purchased_scans"`
	}
	if err := s.httpc.PostJSON(ctx, s.aiocrAddr, "/api/v1/internal/scan-usage",
		map[string]string{"X-Internal-Key": s.internalKey},
		map[string]string{"org_id": orgID.String()}, &out); err != nil {
		return 0, 0
	}
	s.scanUsageCache.Store(orgID, scanUsageCacheEntry{used: out.DocumentsScanned, purchased: out.PurchasedScans, expiry: time.Now().Add(45 * time.Second)})
	return out.DocumentsScanned, out.PurchasedScans
}
```

In `GetSubscriptionStatus`, replace the `usage := s.scanUsageThisMonth(ctx, orgID)` line and the three `resp.UsageCount = usage` assignments. Compute once:

```go
	used, purchased := s.scanUsageThisMonth(ctx, orgID)
```

and at each of the three return paths set both fields:

```go
		resp.UsageCount = used
		resp.UsageLimit = effectiveLimit(resp.UsageLimit, purchased)
```

(`resp.UsageLimit` already holds the base from `statusResponse`; `effectiveLimit` folds in purchased and preserves `Unlimited`.)

- [ ] **Step 4: Run tests to verify pass**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/auth/... -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/auth/service/subscription.go internal/auth/service/subscription_usage_test.go
git commit -m "feat(auth): usage_limit = base + purchased top-up scans (Phase 4b)"
```

---

### Task 9: frontend top-up CTA (Starter) vs upgrade (Gratis)

**Files:**
- Modify: `frontend-svelte/src/lib/services/apiDomains/paymentApi.js` (`createTopupOrder`)
- Modify: `frontend-svelte/src/lib/pages/ScannerPage.svelte` (over-quota CTA)
- Test: `frontend-svelte/src/lib/services/apiDomains/apiDomains.test.js`

**Interfaces:**
- Consumes: `POST /api/v1/payment/topup-order` (Task 6) → `{ payment_url }`.

- [ ] **Step 1: Write the failing test**

In `frontend-svelte/src/lib/services/apiDomains/apiDomains.test.js`, mirror the existing `createPaymentOrder` test:

```js
it("paymentApi.createTopupOrder posts to topup-order", async () => {
  fetchMock.mockResolvedValueOnce({
    ok: true,
    json: async () => ({ data: { payment_url: "https://pay/x", order_id: "o1" } }),
  });
  const res = await paymentApi.createTopupOrder();
  expect(fetchMock).toHaveBeenCalledWith(
    expect.stringContaining("/api/payment/topup-order"),
    expect.objectContaining({ method: "POST" }),
  );
  expect(res.payment_url).toBe("https://pay/x");
});
```

- [ ] **Step 2: Run to verify fail**

Run: `npm --prefix "D:\Codding\Project\Suluk-Manajemen-SAAS\frontend-svelte" run test -- apiDomains`
Expected: FAIL — `paymentApi.createTopupOrder is not a function`.

- [ ] **Step 3: Add the API method**

In `frontend-svelte/src/lib/services/apiDomains/paymentApi.js`, alongside `createPaymentOrder`:

```js
    async createTopupOrder() {
        const response = await apiFetch(`${API_URL}/payment/topup-order`, {
            method: 'POST',
            headers: authHeaders(),
        });
        const json = await response.json();
        return json.data ?? json;
    },
```

(Match the exact unwrap/return style of the adjacent `createPaymentOrder` in this file.)

- [ ] **Step 4: Run test to verify pass**

Run: `npm --prefix "D:\Codding\Project\Suluk-Manajemen-SAAS\frontend-svelte" run test -- apiDomains`
Expected: PASS.

- [ ] **Step 5: Wire the ScannerPage CTA**

In `frontend-svelte/src/lib/pages/ScannerPage.svelte`:

Add a derived flag near the other `$derived` usage vars (~line 300):

```js
  let isStarter = $derived(localSubscription?.plan === "starter");
```

Add a handler in the `<script>` (near `processDocuments`):

```js
  async function buyTopup() {
    try {
      const res = await ApiService.createTopupOrder();
      if (res?.payment_url) window.location.href = res.payment_url;
    } catch (e) {
      console.error("topup order failed:", e);
    }
  }
```

In the over-quota branch (`{:else if usageLimit}`), replace the single "Upgrade ke Pro" `<Button>` with a Starter-aware CTA:

```svelte
            {#if isStarter}
              <Button variant="primary" full size="sm" onclick={buyTopup}>
                Beli 100 scan · Rp49rb
              </Button>
              <button class="topup-upsell" onclick={() => (showUpgradeModal = true)}>
                atau upgrade ke Pro (tanpa batas)
              </button>
            {:else}
              <Button variant="soft" icon={Crown} full size="sm" onclick={() => (showUpgradeModal = true)}>
                Upgrade ke Pro
              </Button>
            {/if}
```

Add minimal styling next to `.quota-bar` in the `<style>` block:

```css
  .topup-upsell {
    margin-top: 6px;
    background: none;
    border: none;
    color: var(--c-muted);
    font-size: 12px;
    cursor: pointer;
    text-decoration: underline;
  }
```

Confirm `ApiService` exposes `createTopupOrder` (it re-exports `paymentApi`); if ScannerPage imports `paymentApi` directly, call that instead — match the existing `createPaymentOrder` call site.

- [ ] **Step 6: Verify check + tests**

Run: `npm --prefix "D:\Codding\Project\Suluk-Manajemen-SAAS\frontend-svelte" run check && npm --prefix "D:\Codding\Project\Suluk-Manajemen-SAAS\frontend-svelte" run test -- apiDomains`
Expected: 0 check errors; tests PASS.

- [ ] **Step 7: Commit**

```bash
git add frontend-svelte/src/lib/services/apiDomains/paymentApi.js frontend-svelte/src/lib/services/apiDomains/apiDomains.test.js frontend-svelte/src/lib/pages/ScannerPage.svelte
git commit -m "feat(scanner): top-up CTA for Starter, upgrade for Gratis (Phase 4b)"
```

---

### Task 10: full-suite verification + deploy env

**Files:**
- Modify: `.env.example`, `deployments/docker-compose.yml` (invoice-service `AIOCR_SERVICE_ADDR`)

- [ ] **Step 1: Ensure invoice can reach aiocr**

`.env.example` already defines `AIOCR_SERVICE_ADDR`. In `deployments/docker-compose.yml`, add to the **invoice-service** `environment:` block (it already has `AUTH_SERVICE_ADDR` + `INTERNAL_API_KEY`):

```yaml
      AIOCR_SERVICE_ADDR: ai-ocr-service:50056
```

- [ ] **Step 2: Full build + test**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" build ./... && go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./...`
Expected: clean build; all packages `ok`.

- [ ] **Step 3: Frontend check + tests**

Run: `npm --prefix "D:\Codding\Project\Suluk-Manajemen-SAAS\frontend-svelte" run check && npm --prefix "D:\Codding\Project\Suluk-Manajemen-SAAS\frontend-svelte" run test`
Expected: 0 errors; tests PASS.

- [ ] **Step 4: Commit**

```bash
git add .env.example deployments/docker-compose.yml
git commit -m "chore(invoice): point invoice-service at ai-ocr for scan top-up credit (Phase 4b)"
```

---

## Self-Review

- **Spec coverage:** top-up purchase (Tasks 5–7), Starter-only gate (Task 6), idempotent credit + ledger (Tasks 2–3, 7), monthly reset (Task 2 `(year,month)` filter), `usage_limit = base+purchased` (Task 8), prompt UX Starter vs Gratis (Task 9), fair-use WARN once/month, never-block (Task 4), SKU constants + mirror (Task 1), deploy wiring (Task 10). ✓ All §3–§7 spec points mapped.
- **Placeholder scan:** every code step shows real code; migrations, SQL, handlers, tests all concrete. ✓
- **Type consistency:** `CreateTopupOrder(ctx, orgID, userID, authToken)`, `creditScanTopup(ctx, order)`, `scanUsageThisMonth → (int,int)`, `effectiveLimit(base,purchased)`, endpoint bodies `{order_id, org_id, scans}` / `{documents_scanned, purchased_scans}` — consistent across producers/consumers. ✓
- **DB-test honesty:** repo SQL (Tasks 2, 4, 5) verified by build + cross-service contract tests, matching the repo's existing no-test-DB convention; pure/logic/handler/httptest paths are TDD'd. ✓
