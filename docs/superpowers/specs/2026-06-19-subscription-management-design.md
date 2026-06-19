# Subscription Management on Profile — Design (Sub-project 2 of the Profile overhaul)

- **Date:** 2026-06-19
- **Status:** Approved — proceeding to implementation plan
- **Area:** `auth-service` (Go) + `frontend-svelte` (`ProfilePage`, `UpgradeModal`)
- **Part of:** the 5-part profile effort. This is **Sub-project 2**. Done: SP1 (profile core, live). Later: SP3 avatar photo upload, SP4 active-session management, SP5 real 2FA.

## 1. Context & problem

The original request: *"user can change upgrade or downgrade their subscription on the user pages too."*

Current state (verified):
- **Upgrades already work end-to-end.** A generalized `UpgradeModal` → `ApiService.createPaymentOrder(plan, period)` → **Pakasir** checkout → verified webhook (`/payment/webhook` in invoice-service) → `ActivatePlanInternal` (auth-service, internal-key guarded) → `ActivatePlan`. It already supports all purchasable tiers (Starter, Pro, Bisnis) monthly/annual.
- **But the profile entry point is hardcoded** to "Upgrade ke Pro" and the modal defaults to Pro — the profile never lets a user pick Starter or Bisnis directly.
- **No downgrade exists.** There is no payment-free path to move down, and the backend never checks direction.

**Billing model (decisive):** subscriptions are **prepaid and non-recurring** — each payment buys a fixed period (`expires_at = now + 1 month / 1 year`); there is no stored card or auto-renew. At expiry, `GetSubscriptionStatus` auto-expires the sub and reports **Gratis** limits. So an org already falls back to free at expiry unless it re-pays.

## 2. Decisions (from brainstorming)

| Topic | Decision |
|-------|----------|
| What "downgrade" means | **Cancel at period end** — a "don't renew" choice. The org keeps the tier it paid for until `expires_at`, then drops to Gratis (the existing auto-expire). No mid-period feature loss, no refund. |
| Paid → lower **paid** tier (e.g. Pro→Starter) | **Not offered** (the awkward "pay for less, forfeit Pro" case). |
| Upgrades | Reuse the existing Pakasir purchase, **generalized** so the profile offers any tier *higher* than the current rank. |
| Enterprise | **"Hubungi sales"** (WhatsApp) link — not purchasable (price 0). |
| Resume / undo | **Included** — a cancelled (not-yet-expired) sub can be resumed (clears the cancel flag). |
| Trial | **No cancel button** — trials lapse on their own after 14 days; users can upgrade to commit. |
| Proration / refunds / credits | **Out of scope.** |

## 3. Goals

- Turn the profile plan card into a small **subscription manager** that shows the current plan, status, and expiry, and offers the right actions per state.
- Let users **upgrade** to any higher purchasable tier from the profile (reusing the existing payment flow).
- Let users **cancel ("don't renew")** a paid subscription — it keeps working until `expires_at`, then drops to Gratis — and **resume** before expiry.

## 4. Non-goals

- Proration, refunds, credits, or any mid-period money movement.
- Switching to a lower **paid** tier.
- Auto-recurring billing / stored payment methods.
- Any change to the Pakasir order/webhook machinery in invoice-service.
- SP3 avatar upload, SP4 sessions, SP5 2FA.

## 5. Backend changes (`auth-service`)

### 5.1 Migration `migrations/auth/016_subscription_cancel.{up,down}.sql`
Add one column to `subscriptions`:
```sql
ALTER TABLE subscriptions
  ADD COLUMN IF NOT EXISTS cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE;
```
Down: `ALTER TABLE subscriptions DROP COLUMN IF EXISTS cancel_at_period_end;`

### 5.2 Model (`internal/auth/model/model.go`)
- `Subscription`: add `CancelAtPeriodEnd bool` (`db:"cancel_at_period_end"`).
- `SubscriptionStatusResponse`: add `CancelAtPeriodEnd bool json:"cancel_at_period_end"`.

### 5.3 Repository (`internal/auth/repository/repository.go`)
- `GetSubscription` SELECT/Scan: include `cancel_at_period_end` (use `COALESCE(cancel_at_period_end, FALSE)` for safety).
- `CreateSubscription` / `UpdateSubscription`: write `cancel_at_period_end`.

### 5.4 Service (`internal/auth/service/subscription.go`)
- **`CancelSubscription(ctx, orgID) (*model.SubscriptionStatusResponse, error)`**: load the sub; valid to cancel only when `sub != nil`, `status == "active"` (a **paid** sub — trials lapse on their own and are not cancelable), and `ExpiresAt != nil && ExpiresAt.After(now)`. Set `cancel_at_period_end = true`, persist, then return the fresh `GetSubscriptionStatus`. Otherwise return a sentinel `ErrNothingToCancel` (→ 400). Idempotent: cancelling an already-cancelled sub succeeds (stays true).
- **`ResumeSubscription(ctx, orgID) (*model.SubscriptionStatusResponse, error)`**: load the sub; valid only when `sub != nil`, not expired, and `cancel_at_period_end` is currently true. Set `cancel_at_period_end = false`, persist, return the fresh `GetSubscriptionStatus`. If nothing to resume → `ErrNothingToResume` (→ 400).
- **`GetSubscriptionStatus`**: surface `CancelAtPeriodEnd` on the response. **No change to the tier/expiry logic** — an active sub still reports its paid tier and limits until `expires_at`; the existing auto-expire still drops it to Gratis after. (The cancel flag is informational; it does not change limits before expiry.)
- **`ActivatePlan`**: on any activation/renewal, set `cancel_at_period_end = false` (re-committing clears a pending cancel). Applies to both the create and update branches.

### 5.5 Handlers + routes
- `internal/auth/handler/subscription.go`: add `CancelSubscription` and `ResumeSubscription` handlers (read `claims.OrgID`; on success return the fresh `SubscriptionStatusResponse` via `response.OK`; map `ErrNothingToCancel`/`ErrNothingToResume` via `errors.Is` → `response.BadRequest`, else `response.Internal`).
- `cmd/auth-service/main.go`: register `POST /api/v1/subscription/cancel` and `POST /api/v1/subscription/resume` under the **same guard as the existing `/upgrade` route** (`AuthMiddleware` + `RequireStaff`) for consistency.
- **api-gateway:** verify the gateway proxies the whole `/api/v1/subscription/*` group to auth-service (prefix proxy). If routes are registered explicitly per-path, add the two new ones. (Confirm during the plan's first task.)

## 6. Frontend changes (`frontend-svelte`)

### 6.1 `ProfilePage.svelte` — subscription section
Render by state (using `subscription.plan`, `status`, `expires_at`, `cancel_at_period_end`, and `trialStatus`):

| State | Display | Actions |
|-------|---------|---------|
| Gratis / no paid sub | "Paket Gratis" + limits | **Upgrade** (opens modal) |
| Paid, will renew | "Pro · aktif hingga {date}" | **Ubah paket** (upgrade to higher) · **Batalkan perpanjangan** |
| Paid, `cancel_at_period_end` | "Pro · berakhir {date} · tidak diperpanjang" | **Lanjutkan langganan** (resume) · **Ubah paket** |
| Trial | "Uji coba Pro · berakhir {date}" | **Upgrade** |
| Expired | "Kedaluwarsa" (from SP1) | **Perpanjang / Upgrade** |

- **Upgrade entry:** replace the hardcoded "Upgrade ke Pro" with a plan-aware button that opens the existing `UpgradeModal`, passing the current plan so the modal **only offers tiers with rank > current** (Enterprise shown as a "Hubungi sales" WhatsApp link, reusing the landing page's `SALES_WA`).
- **Cancel:** a confirmation dialog before calling cancel — copy: *"Anda tetap bisa memakai {plan} sampai {date}, lalu paket turun ke Gratis. Lanjutkan?"* On confirm → `cancelSubscription()`, refresh status.
- **Resume:** one click → `resumeSubscription()`, refresh status (no confirmation needed).

### 6.2 API client (`apiDomains/authSubscriptionApi.js`)
- Add `cancelSubscription()` → `POST /subscription/cancel` and `resumeSubscription()` → `POST /subscription/resume`, alongside the existing `getSubscriptionStatus` (which owns the `sub:status` cache).
- Both endpoints return the fresh `SubscriptionStatusResponse`; on success **overwrite the cache** with `cacheSet('sub:status', data, 20000)` (the cache helpers expose no per-key delete — overwriting with the returned status is the same pattern as SP1's `updateProfile` → `cacheSet('auth:me', …)` fix, and avoids a global `cacheClear`).

### 6.3 `UpgradeModal.svelte`
- Accept the caller's current plan and **filter the tier selector to higher-rank purchasable tiers** (so the profile's "upgrade" can't present a lower/equal tier). Default selection = the lowest offered higher tier. No change to the payment/polling logic.

## 7. Edge cases

- Cancel then upgrade/renew → `ActivatePlan` clears the flag (re-committed).
- Cancel rejected (400) when already Gratis or expired (nothing to cancel).
- Resume only valid before expiry; after expiry the org is Gratis and must re-purchase.
- Over-limit on downgrade is **not a concern**: limits only drop at expiry, where existing quota enforcement already applies.
- Idempotent cancel/resume (repeat calls converge).
- Concurrent upgrade-in-another-tab: status is re-fetched after actions; the webhook remains the source of truth for paid activation.

## 8. Testing

- **Go (service):** `CancelSubscription` sets the flag only for an active/trial unexpired sub and rejects otherwise; `ResumeSubscription` clears it; `GetSubscriptionStatus` reports the flag and keeps paid limits until expiry; `ActivatePlan` clears the flag on create and update. Repo round-trip of the new column.
- **Handlers:** cancel/resume map sentinels to 400, success to 200.
- **Frontend:** `npm run check` (0 errors), `npm test`, `npm run build`. Logic-level checks where the vitest env allows (state→actions mapping, higher-tier filter).
- **Manual:** upgrade from profile (pick a tier) → pay (Pakasir sandbox) → active; cancel → "won't renew" + keeps tier until expiry; resume → back to will-renew; cancel→upgrade clears the flag.
- Gates: `go build ./cmd/...`, `go vet ./...`, `go test ./...`, `npm run check`, `npm test`, `npm run build`.

## 9. Rollout

- Migration `016` on `jamaah_auth`; rebuild + restart **auth-service** + **frontend** (surgical, as for SP1).
- Additive and backward-compatible (one defaulted boolean column; existing upgrade flow unchanged).
- No invoice-service or Pakasir changes.
