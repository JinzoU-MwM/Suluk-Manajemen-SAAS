# Scan Top-up + Fair-use Cap — Design (Phase 4b)

**Tanggal:** 2026-06-25
**Status:** Approved (pending spec review)
**Sub-phase of:** Phase 4 in `2026-06-24-pricing-tier-restructure.md` (spec §3.3). Builds on 4a (metering) and 4A (surface usage, commit `985d5e2`). **This closes Phase 4.**

---

## 1. Goal & decisions

Make the soft scan-cap real: when a **Starter** org exhausts its 100/month quota, it isn't blocked — it can buy a **Rp 49.000 / 100-scan** top-up (one-time, via Pakasir), credited to the current month and reset monthly. Pro/Bisnis stay "unlimited" but get a **fair-use** safety net.

Decisions locked in brainstorming:

| Decision | Choice |
|---|---|
| Scope | Top-up **and** Pro/Bisnis fair-use cap (closes Phase 4) |
| Top-up eligibility | **Starter only.** Gratis over-quota → upgrade prompt (not top-up); Pro/Bisnis unlimited → never prompted |
| Where credits live | **aiocr** owns a `scan_topups` ledger (idempotent by `order_id`) — Approach A |
| Fair-use alert | **WARN log** for ops (notify infra is org-scoped only; alerting the customer on an "unlimited" tier is bad UX) |
| Fair-use scope | **Tier-agnostic** threshold (any org > cap/month) — avoids aiocr needing the plan; in practice only Pro/Bisnis reach it |
| Scans are never hard-blocked | OCR always processes; the cap is a soft prompt / a log — consistent with the existing fail-open enforcement philosophy |

**Out of scope:** auto-renew/dunning for top-ups (one-time only), proactive top-up before quota is hit (allowed by the endpoint, but the UI only surfaces it on exhaustion), per-tier custom fair-use thresholds.

## 2. Global constraints

- **Server-authoritative price.** The top-up amount is set from `plan.ScanTopupPrice`, never from the client (mirrors how plan purchases use `plan.PriceFor`).
- **Idempotency.** Pakasir webhooks are at-least-once; crediting must be exactly-once. The ledger's `UNIQUE(order_id)` + `INSERT … ON CONFLICT DO NOTHING` guarantees it.
- **Fail-safe money.** Mirror `activateSubscription`: credit happens *after* the atomic `MarkPaymentOrderPaid`; if crediting fails, **revert the order to pending** so Pakasir retries (never paid-but-not-credited).
- **Monthly reset.** Ledger rows carry `(year, month)` stamped at insert; `purchased_this_month = SUM(scans)` for the current year/month, so credits auto-expire at month end (same model as `scan_usage`).
- `ScanTopupPrice` / `ScanTopupScans` in `plan.go` MUST stay mirrored in `pricing.js`.
- Commit messages: NO AI co-author line.

## 3. Architecture

Reuses the existing Pakasir spine (`create-order → checkout → PakasirWebhook → Verify → MarkPaid → activate`). The only new branch is on a new **order purpose**.

```
Starter exhausts quota (used ≥ limit)
  → ScannerPage: "Kuota scan bulan ini habis. Beli 100 scan (Rp49.000)?"
  → POST /api/v1/payment/topup-order            invoice-svc, price = plan.ScanTopupPrice (server)
       ├─ validate caller is Starter (auth /subscription/status) → else 400
       └─ PaymentOrder{purpose:"scan_topup", amount:49000} → Pakasir checkout URL (existing)
  → user pays → POST /api/v1/payment/webhook → HandlePakasirWebhook (existing)
       Verify (authoritative) → MarkPaymentOrderPaid (atomic claim, idempotent)
       └─ branch on order.purpose:
            "subscription" → activateSubscription                  (existing, unchanged)
            "scan_topup"   → creditScanTopup → POST aiocr /api/v1/internal/scan-topup
                               {order_id, org_id, scans:100}  (X-Internal-Key)
                             on error → RevertPaymentOrderToPending → Pakasir retries
  → next GET /subscription/status:
       usage_count = used (from scan_usage)
       usage_limit = base_limit + purchased_this_month   (finite tiers only; Unlimited stays -1)
```

### 3.1 Components & interfaces

**invoice-service**
- `payment_orders` migration: add `purpose TEXT NOT NULL DEFAULT 'subscription'`. Existing rows/flows unaffected.
- `model.PaymentOrder.Purpose string`; `CreateTopupOrderRequest` (no plan/period).
- Handler `CreateTopupOrder` → `POST /api/v1/payment/topup-order` (authMW + RequireStaff). Validates Starter via auth status (forwards caller token); creates `PaymentOrder{Purpose:"scan_topup", Amount:plan.ScanTopupPrice}`; returns the Pakasir checkout URL via the existing `pakasirPayURL`.
- `HandlePakasirWebhook`: after `MarkPaymentOrderPaid`, `switch order.Purpose` → `activateSubscription` | `creditScanTopup`. `creditScanTopup` POSTs to aiocr internal endpoint with the shared internal key; revert-on-failure as above.

**aiocr-service**
- Migration `003_scan_topups`: `scan_topups(order_id UUID PRIMARY KEY, org_id UUID, year INT, month INT, scans INT, created_at TIMESTAMPTZ)`, index on `(org_id, year, month)`.
- Repo: `CreditScanTopup(ctx, orderID, orgID, scans)` → idempotent insert; `GetPurchasedScansThisMonth(ctx, orgID)` → `SUM` for current month.
- Handler `ScanTopupInternal` → `POST /api/v1/internal/scan-topup` (X-Internal-Key), parses `{order_id, org_id, scans}`, calls `CreditScanTopup`.
- **Extend the 4A endpoint** `POST /api/v1/internal/scan-usage` to return `{documents_scanned, purchased_scans}`.

**auth-service** (extends 4A)
- `scanUsageThisMonth` → returns `(used, purchased)` from the extended endpoint; cache entry holds both.
- `statusResponse` / `GetSubscriptionStatus`: `UsageCount = used`; `UsageLimit = base == Unlimited ? Unlimited : base + purchased`.

**plan (shared)**
- `ScanTopupPrice = 49000`, `ScanTopupScans = 100`, `FairUseScanCap = 2000`.

**frontend**
- `pricing.js`: mirror `scanTopupPrice` / `scanTopupScans`.
- ScannerPage over-quota CTA: **Starter** → call `topup-order` and redirect to Pakasir (reuse the existing plan-purchase redirect helper); **Gratis** → existing `UpgradeModal`. Quota bar already renders from 4A; no new display logic.

### 3.2 Fair-use cap (Pro/Bisnis)

- `scan_usage` migration: add `fairuse_alerted_at TIMESTAMPTZ NULL` (per org/month row).
- In `ProcessDocumentsSync`, after `IncrementScanUsage`: if the org's new monthly total ≥ `FairUseScanCap` **and** `fairuse_alerted_at IS NULL` for this month, emit a `logger.Warnf` (org + count) and stamp `fairuse_alerted_at = NOW()` (fires once per org/month). Best-effort — a failure here never affects the OCR response. **Never blocks.**
- Tier-agnostic by design: aiocr doesn't fetch the plan. Finite tiers can't reach 2000 without paying for it (a signal worth logging anyway).

## 4. Reuse vs new

**Reused as-is:** Pakasir checkout URL builder, `VerifyTransaction`, `MarkPaymentOrderPaid` (atomic/idempotent), `RevertPaymentOrderToPending`, the X-Internal-Key pattern, the 4A usage-fetch + 45s cache, the ScannerPage quota bar, the plan-purchase → Pakasir redirect on the frontend.

**New/changed:** `purpose` column + branch; `topup-order` endpoint + Starter validation; `scan_topups` ledger + credit/read repo + internal endpoint; extended scan-usage endpoint; `usage_limit = base + purchased`; SKU + fair-use constants; `fairuse_alerted_at` + the WARN check; frontend top-up CTA + price mirror.

## 5. Edge cases

- **Double webhook delivery:** `MarkPaymentOrderPaid` claims once; even if the credit POST is delivered twice, `ON CONFLICT(order_id) DO NOTHING` makes it +100 exactly once.
- **Credit POST fails after MarkPaid:** order reverted to pending → next Pakasir retry re-runs verify→credit. The conflict guard prevents a double credit if the first credit actually applied but its response was lost.
- **Non-Starter calls `topup-order`:** rejected 400 (server-side plan check), so Gratis can't limp on top-ups and Pro/Bisnis can't buy redundant credits.
- **Downgrade after purchase:** credits already bought this month still count (effective limit = new base + purchased) — the customer paid, so honor it.
- **Month rollover:** unused purchased credits do not carry over (ledger filtered by current year/month) — matches "reset tiap awal bulan."

## 6. Testing (TDD)

- **aiocr:** idempotent credit (two calls, one order_id → SUM = scans once); `GetPurchasedScansThisMonth` filters by month; internal endpoints reject bad key (401) / bad body (400); scan-usage endpoint returns both fields; fair-use warns once per org/month and **does not block** the scan result.
- **invoice:** `topup-order` is server-priced at 49k and rejects non-Starter; webhook routes `scan_topup` to credit (not activate) and `subscription` to activate; credit failure reverts the order.
- **auth:** `usage_limit = base + purchased` for finite tiers; stays `Unlimited` for Pro/Bisnis; `usage_count` unchanged.
- **plan:** SKU + fair-use constants present.

## 7. Definition of Done

- Starter over-quota can buy 100 scans for Rp49k via Pakasir; quota increases for the month and resets next month; purchase is exactly-once under retries.
- `/subscription/status` reflects `usage_limit = base + purchased`; ScannerPage shows the updated remaining and the top-up CTA (Gratis sees upgrade instead).
- Pro/Bisnis crossing 2000 scans/month produce one WARN log per org/month; scans are never blocked.
- `plan.go` ↔ `pricing.js` in sync; all new behavior covered by tests; `go build ./...` + `go test ./...` green.
