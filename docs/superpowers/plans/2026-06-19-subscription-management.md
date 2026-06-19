# Subscription Management on Profile — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let users manage their subscription from the profile — upgrade to a higher tier via the existing Pakasir flow, and cancel ("don't renew") a paid plan so it keeps working until `expires_at` then drops to Gratis, with a resume/undo.

**Architecture:** Add one boolean column (`subscriptions.cancel_at_period_end`) and two authenticated endpoints (`/subscription/cancel`, `/subscription/resume`) on auth-service. Cancel/resume are thin service methods over pure validation helpers (mirroring SP1's `ApplyProfileUpdate` pattern). The frontend profile gains state-driven subscription controls; the existing `UpgradeModal` is filtered to higher tiers. No payment/webhook machinery changes.

**Tech Stack:** Go (Fiber, pgx v5), Postgres (`jamaah_auth`), golang-migrate, SvelteKit 5 (runes), Tailwind.

## Global Constraints

- Migration is **additive and backward-compatible**: `cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE`.
- **Cancel** is valid only for a **paid, active** sub (`status == "active"`) with `ExpiresAt != nil && ExpiresAt.After(now)`. **Trials are NOT cancelable** (they lapse on their own). On success the sub keeps its tier/limits until expiry — *do not* change the tier/limit/expiry logic in `GetSubscriptionStatus`.
- **Resume** is valid only when the sub exists, is unexpired, and `cancel_at_period_end` is currently `true`.
- **`ActivatePlan` must reset `cancel_at_period_end = false`** on renewal/activation (re-committing clears a pending cancel).
- **Cancel/resume endpoints return the fresh `SubscriptionStatusResponse`**; the frontend overwrites the `sub:status` cache with it via `cacheSet` (the cache helpers expose no per-key delete — same pattern as SP1's `updateProfile` → `cacheSet('auth:me', …)`).
- Error sentinels live in the **model** package (like SP1's `model.ErrNameRequired`); handlers map them via `errors.Is` to `response.BadRequest` (400), everything else `response.Internal` (500).
- New routes sit under the existing `/api/v1/subscription` group (`AuthMiddleware` + `RequireStaff`) — same guard as `/upgrade`.
- Upgrades reuse the existing Pakasir purchase. The `UpgradeModal` only offers tiers with **rank > current rank**; **Enterprise** (non-purchasable) is a "Hubungi sales" WhatsApp link.
- **Out of scope:** proration, refunds, credits, paid→lower-paid switching, recurring billing, any invoice-service/Pakasir change.
- UI copy in **Bahasa Indonesia**.

---

### Task 1: Migration `016_subscription_cancel`

**Files:**
- Create: `migrations/auth/016_subscription_cancel.up.sql`
- Create: `migrations/auth/016_subscription_cancel.down.sql`

**Interfaces:**
- Produces: a `subscriptions.cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE` column consumed by Tasks 2–4.

- [ ] **Step 1: Write the up migration**

`migrations/auth/016_subscription_cancel.up.sql`:
```sql
-- Cancel-at-period-end flag: a paid sub flagged true keeps its tier until
-- expires_at, then the existing auto-expire drops the org to Gratis.
ALTER TABLE subscriptions
  ADD COLUMN IF NOT EXISTS cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE;
```

- [ ] **Step 2: Write the down migration**

`migrations/auth/016_subscription_cancel.down.sql`:
```sql
ALTER TABLE subscriptions DROP COLUMN IF EXISTS cancel_at_period_end;
```

- [ ] **Step 3: Sanity-check the SQL** (idempotent `IF NOT EXISTS`, NOT NULL + default so existing rows backfill to `false`). No DB run here — applied at deploy.

- [ ] **Step 4: Commit**
```bash
git add migrations/auth/016_subscription_cancel.up.sql migrations/auth/016_subscription_cancel.down.sql
git commit -m "feat(auth): migration for subscription cancel_at_period_end"
```

---

### Task 2: Model fields + repository read/write

**Files:**
- Modify: `internal/auth/model/model.go` (`Subscription`, `SubscriptionStatusResponse`)
- Modify: `internal/auth/repository/subscription.go` (`GetSubscription`, `CreateSubscription`, `UpdateSubscription`)

**Interfaces:**
- Consumes: the `cancel_at_period_end` column (Task 1).
- Produces: `model.Subscription.CancelAtPeriodEnd bool`, `model.SubscriptionStatusResponse.CancelAtPeriodEnd bool`, and repo methods that round-trip the column. Consumed by Tasks 3–5.

- [ ] **Step 1: Add the model fields**

In `internal/auth/model/model.go`, add to the `Subscription` struct (after `TrialUsed`, before `CreatedAt`):
```go
	CancelAtPeriodEnd bool `json:"cancel_at_period_end" db:"cancel_at_period_end"`
```
And add to `SubscriptionStatusResponse` (after `MaxUsers`):
```go
	CancelAtPeriodEnd bool `json:"cancel_at_period_end"`
```

- [ ] **Step 2: Read the column in `GetSubscription`**

In `internal/auth/repository/subscription.go`, update the SELECT + Scan. New query (add `COALESCE(cancel_at_period_end, FALSE)` last):
```go
	query := `SELECT id, org_id, plan, status, starts_at, expires_at, trial_used, created_at, updated_at,
		COALESCE(cancel_at_period_end, FALSE)
		FROM subscriptions WHERE org_id = $1`
	var sub model.Subscription
	err := r.pool.QueryRow(ctx, query, orgID).Scan(
		&sub.ID, &sub.OrgID, &sub.Plan, &sub.Status,
		&sub.StartsAt, &sub.ExpiresAt, &sub.TrialUsed,
		&sub.CreatedAt, &sub.UpdatedAt,
		&sub.CancelAtPeriodEnd,
	)
```

- [ ] **Step 3: Write the column in `CreateSubscription`**

Update INSERT to include the new column as `$8`:
```go
	query := `INSERT INTO subscriptions (id, org_id, plan, status, starts_at, expires_at, trial_used, cancel_at_period_end)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		sub.ID, sub.OrgID, sub.Plan, sub.Status,
		sub.StartsAt, sub.ExpiresAt, sub.TrialUsed, sub.CancelAtPeriodEnd,
	).Scan(&sub.CreatedAt, &sub.UpdatedAt)
```

- [ ] **Step 4: Write the column in `UpdateSubscription`**

Update UPDATE to set the new column as `$6`:
```go
	query := `UPDATE subscriptions SET plan = $2, status = $3, expires_at = $4, trial_used = $5, cancel_at_period_end = $6, updated_at = NOW()
		WHERE org_id = $1`
	_, err := r.pool.Exec(ctx, query, sub.OrgID, sub.Plan, sub.Status, sub.ExpiresAt, sub.TrialUsed, sub.CancelAtPeriodEnd)
	return err
```

- [ ] **Step 5: Build**

Run: `go build ./cmd/...`
Expected: clean (a Scan/column mismatch here fails the build).

- [ ] **Step 6: Commit**
```bash
git add internal/auth/model/model.go internal/auth/repository/subscription.go
git commit -m "feat(auth): persist subscription cancel_at_period_end (model + repo)"
```

---

### Task 3: Pure cancel/resume validation helpers (TDD)

**Files:**
- Create: `internal/auth/model/subscription_cancel.go`
- Test: `internal/auth/model/subscription_cancel_test.go`

**Interfaces:**
- Consumes: `model.Subscription` (Task 2).
- Produces: `model.ErrNothingToCancel`, `model.ErrNothingToResume`, `model.CanCancelSubscription(sub *Subscription, now time.Time) error`, `model.CanResumeSubscription(sub *Subscription, now time.Time) error`. Consumed by Tasks 4 (service) and 5 (handler maps the sentinels).

- [ ] **Step 1: Write the failing tests**

`internal/auth/model/subscription_cancel_test.go`:
```go
package model

import (
	"testing"
	"time"
)

func TestCanCancelSubscription(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	cases := []struct {
		name string
		sub  *Subscription
		ok   bool
	}{
		{"nil sub", nil, false},
		{"active future", &Subscription{Status: "active", ExpiresAt: &future}, true},
		{"active no expiry", &Subscription{Status: "active", ExpiresAt: nil}, false},
		{"active expired", &Subscription{Status: "active", ExpiresAt: &past}, false},
		{"trial not cancelable", &Subscription{Status: "trial", ExpiresAt: &future}, false},
		{"expired status", &Subscription{Status: "expired", ExpiresAt: &future}, false},
	}
	for _, c := range cases {
		err := CanCancelSubscription(c.sub, now)
		if c.ok && err != nil {
			t.Fatalf("%s: want ok, got %v", c.name, err)
		}
		if !c.ok && err == nil {
			t.Fatalf("%s: want error, got nil", c.name)
		}
	}
}

func TestCanResumeSubscription(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	cases := []struct {
		name string
		sub  *Subscription
		ok   bool
	}{
		{"nil sub", nil, false},
		{"flagged future", &Subscription{ExpiresAt: &future, CancelAtPeriodEnd: true}, true},
		{"not flagged", &Subscription{ExpiresAt: &future, CancelAtPeriodEnd: false}, false},
		{"flagged but expired", &Subscription{ExpiresAt: &past, CancelAtPeriodEnd: true}, false},
		{"flagged no expiry", &Subscription{ExpiresAt: nil, CancelAtPeriodEnd: true}, false},
	}
	for _, c := range cases {
		err := CanResumeSubscription(c.sub, now)
		if c.ok && err != nil {
			t.Fatalf("%s: want ok, got %v", c.name, err)
		}
		if !c.ok && err == nil {
			t.Fatalf("%s: want error, got nil", c.name)
		}
	}
}
```

- [ ] **Step 2: Run the tests, verify they fail**

Run: `go test ./internal/auth/model/ -run 'TestCanCancelSubscription|TestCanResumeSubscription' -v`
Expected: FAIL — `undefined: CanCancelSubscription` / `CanResumeSubscription`.

- [ ] **Step 3: Write the implementation**

`internal/auth/model/subscription_cancel.go`:
```go
package model

import (
	"errors"
	"time"
)

// ErrNothingToCancel is returned when there is no paid, active, unexpired
// subscription to flag cancel-at-period-end. Handlers map it to 400.
var ErrNothingToCancel = errors.New("no active subscription to cancel")

// ErrNothingToResume is returned when there is no pending cancel to undo.
var ErrNothingToResume = errors.New("no cancellable subscription to resume")

// CanCancelSubscription reports whether the org's subscription may be set to
// cancel-at-period-end: it must be a paid, active sub that has not expired.
// Trials lapse on their own and are intentionally not cancelable.
func CanCancelSubscription(sub *Subscription, now time.Time) error {
	if sub == nil || sub.Status != "active" || sub.ExpiresAt == nil || !sub.ExpiresAt.After(now) {
		return ErrNothingToCancel
	}
	return nil
}

// CanResumeSubscription reports whether a pending cancel can be undone: the sub
// must exist, still be unexpired, and currently be flagged cancel-at-period-end.
func CanResumeSubscription(sub *Subscription, now time.Time) error {
	if sub == nil || sub.ExpiresAt == nil || !sub.ExpiresAt.After(now) || !sub.CancelAtPeriodEnd {
		return ErrNothingToResume
	}
	return nil
}
```

- [ ] **Step 4: Run the tests, verify they pass**

Run: `go test ./internal/auth/model/ -run 'TestCanCancelSubscription|TestCanResumeSubscription' -v`
Expected: PASS (all cases).

- [ ] **Step 5: Commit**
```bash
git add internal/auth/model/subscription_cancel.go internal/auth/model/subscription_cancel_test.go
git commit -m "feat(auth): pure cancel/resume validation helpers (TDD)"
```

---

### Task 4: Service methods + status flag + ActivatePlan reset

**Files:**
- Modify: `internal/auth/service/subscription.go` (`GetSubscriptionStatus`, `ActivatePlan`; add `CancelSubscription`, `ResumeSubscription`)

**Interfaces:**
- Consumes: `model.CanCancelSubscription`/`CanResumeSubscription` (Task 3), repo methods (Task 2).
- Produces: `AuthService.CancelSubscription(ctx, orgID uuid.UUID) (*model.SubscriptionStatusResponse, error)` and `ResumeSubscription(...)` with the same signature. Consumed by Task 5.

- [ ] **Step 1: Surface the flag in `GetSubscriptionStatus`**

In `internal/auth/service/subscription.go`, change the final active-path return (currently `return statusResponse(sub.Plan, sub.Status, sub.ExpiresAt), nil`) to:
```go
	resp := statusResponse(sub.Plan, sub.Status, sub.ExpiresAt)
	resp.CancelAtPeriodEnd = sub.CancelAtPeriodEnd
	return resp, nil
```
Leave the `sub == nil` (Gratis) and `expired/cancelled` returns unchanged — `CancelAtPeriodEnd` stays `false` there.

- [ ] **Step 2: Reset the flag on activation in `ActivatePlan`**

In the update branch of `ActivatePlan` (after `sub.ExpiresAt = &expiresAt`), add:
```go
	sub.CancelAtPeriodEnd = false // re-committing/renewing clears a pending cancel
```
(The create branch builds a fresh `model.Subscription` whose `CancelAtPeriodEnd` is already the `false` zero value — no change needed there.)

- [ ] **Step 3: Add `CancelSubscription` and `ResumeSubscription`**

Append to `internal/auth/service/subscription.go`:
```go
// CancelSubscription flags a paid, active subscription to not renew. The org
// keeps its tier and limits until expires_at, after which the existing
// auto-expire drops it to Gratis. Returns the refreshed status.
func (s *AuthService) CancelSubscription(ctx context.Context, orgID uuid.UUID) (*model.SubscriptionStatusResponse, error) {
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if err := model.CanCancelSubscription(sub, time.Now()); err != nil {
		return nil, err
	}
	sub.CancelAtPeriodEnd = true
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return nil, err
	}
	return s.GetSubscriptionStatus(ctx, orgID)
}

// ResumeSubscription undoes a pending cancel-at-period-end (before expiry).
func (s *AuthService) ResumeSubscription(ctx context.Context, orgID uuid.UUID) (*model.SubscriptionStatusResponse, error) {
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if err := model.CanResumeSubscription(sub, time.Now()); err != nil {
		return nil, err
	}
	sub.CancelAtPeriodEnd = false
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return nil, err
	}
	return s.GetSubscriptionStatus(ctx, orgID)
}
```

- [ ] **Step 4: Build + vet**

Run: `go build ./cmd/... && go vet ./internal/auth/...`
Expected: clean.

- [ ] **Step 5: Run the model tests (still green)**

Run: `go test ./internal/auth/...`
Expected: PASS (no regressions; the model helpers from Task 3 still pass).

- [ ] **Step 6: Commit**
```bash
git add internal/auth/service/subscription.go
git commit -m "feat(auth): cancel/resume subscription service + status flag + activate reset"
```

---

### Task 5: Handlers + routes (+ gateway proxy check)

**Files:**
- Modify: `internal/auth/handler/subscription.go` (add `CancelSubscription`, `ResumeSubscription` handlers)
- Modify: `cmd/auth-service/main.go` (register the two routes)
- Verify: `cmd/api-gateway/main.go` (subscription proxying)

**Interfaces:**
- Consumes: `AuthService.CancelSubscription`/`ResumeSubscription` (Task 4), `model.ErrNothingToCancel`/`ErrNothingToResume` (Task 3).
- Produces: `POST /api/v1/subscription/cancel` and `POST /api/v1/subscription/resume`.

- [ ] **Step 1: Add the handlers**

In `internal/auth/handler/subscription.go`, ensure `"errors"` and the `model` package are imported (the file already imports `model`; add `"errors"` if missing), then add:
```go
func (h *AuthHandler) CancelSubscription(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	status, err := h.svc.CancelSubscription(c.Context(), claims.OrgID)
	if err != nil {
		if errors.Is(err, model.ErrNothingToCancel) {
			return response.BadRequest(c, err.Error())
		}
		return response.Internal(c, err)
	}
	return response.OK(c, status)
}

func (h *AuthHandler) ResumeSubscription(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	status, err := h.svc.ResumeSubscription(c.Context(), claims.OrgID)
	if err != nil {
		if errors.Is(err, model.ErrNothingToResume) {
			return response.BadRequest(c, err.Error())
		}
		return response.Internal(c, err)
	}
	return response.OK(c, status)
}
```

- [ ] **Step 2: Register the routes**

In `cmd/auth-service/main.go`, in the `subscription` group (after the `/pricing` line, ~line 163), add:
```go
	subscription.Post("/cancel", authHandler.CancelSubscription)
	subscription.Post("/resume", authHandler.ResumeSubscription)
```

- [ ] **Step 3: Verify the api-gateway proxies the new routes**

Run: `grep -rn "subscription" cmd/api-gateway/`
- If the gateway proxies the `/api/v1/subscription` prefix (a group/wildcard, like the working `/status`, `/upgrade`, `/pricing`), no change is needed — the new sub-paths are covered.
- If it registers each subscription route explicitly, add `/subscription/cancel` and `/subscription/resume` proxy entries mirroring `/subscription/upgrade`.
Record which case applies in the task report.

- [ ] **Step 4: Build + vet**

Run: `go build ./cmd/... && go vet ./...`
Expected: clean.

- [ ] **Step 5: Commit**
```bash
git add internal/auth/handler/subscription.go cmd/auth-service/main.go cmd/api-gateway/
git commit -m "feat(auth): cancel/resume subscription routes + handlers"
```

---

### Task 6: Frontend — API methods + ProfilePage subscription management

**Files:**
- Modify: `frontend-svelte/src/lib/services/apiDomains/authSubscriptionApi.js` (add `cancelSubscription`, `resumeSubscription`)
- Modify: `frontend-svelte/src/lib/pages/ProfilePage.svelte` (subscription block + handlers)

**Interfaces:**
- Consumes: `POST /subscription/cancel`, `POST /subscription/resume` (Task 5); `subscription.cancel_at_period_end` from status.
- Produces: `ApiService.cancelSubscription()`, `ApiService.resumeSubscription()`.

- [ ] **Step 1: Add the API methods**

In `frontend-svelte/src/lib/services/apiDomains/authSubscriptionApi.js`, inside the object returned by `createAuthSubscriptionApi({ cacheGet, cacheSet })`, after `getSubscriptionStatus`, add:
```js
        async cancelSubscription() {
            const response = await apiFetch(`${API_URL}/subscription/cancel`, {
                method: 'POST',
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            const data = unwrapData(await response.json());
            cacheSet('sub:status', data, 20000);
            return data;
        },

        async resumeSubscription() {
            const response = await apiFetch(`${API_URL}/subscription/resume`, {
                method: 'POST',
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            const data = unwrapData(await response.json());
            cacheSet('sub:status', data, 20000);
            return data;
        },
```

- [ ] **Step 2: Confirm `ApiService` exposes the new methods**

Run: `grep -n "getSubscriptionStatus\|createAuthSubscriptionApi" frontend-svelte/src/lib/services/api.js`
Confirm the auth-subscription domain is spread into `ApiService` (the new methods ride along automatically, same as `getSubscriptionStatus`). No code change expected; note the finding.

- [ ] **Step 3: Add ProfilePage state + handlers**

In `frontend-svelte/src/lib/pages/ProfilePage.svelte` `<script>`, add near the other `$state` declarations:
```js
    let savingSub = $state(false);
    let showCancelConfirm = $state(false);
    let cancelAtPeriodEnd = $derived(!!subscription?.cancel_at_period_end);
```
And add the handlers (near `saveProfile`):
```js
    async function confirmCancelSubscription() {
        savingSub = true;
        try {
            subscription = await ApiService.cancelSubscription();
            showCancelConfirm = false;
        } catch (e) {
            alert(e.message || "Gagal membatalkan langganan.");
        } finally {
            savingSub = false;
        }
    }

    async function resumeSubscription() {
        savingSub = true;
        try {
            subscription = await ApiService.resumeSubscription();
        } catch (e) {
            alert(e.message || "Gagal melanjutkan langganan.");
        } finally {
            savingSub = false;
        }
    }
```

- [ ] **Step 4: Render the manage controls in the plan block**

In the `.plan-block` (replace the expiry conditional at lines ~500–504 and add the paid-tier controls). Replace:
```svelte
                            {#if subscription?.expires_at && subscription?.status !== "expired"}
                                <div class="plan-expiry">Berlaku hingga {formatDate(subscription.expires_at)}</div>
                            {:else if subscription?.status === "expired"}
                                <div class="plan-expiry plan-expiry-danger">Langganan kedaluwarsa</div>
                            {/if}
```
with:
```svelte
                            {#if subscription?.status === "expired"}
                                <div class="plan-expiry plan-expiry-danger">Langganan kedaluwarsa</div>
                            {:else if cancelAtPeriodEnd && subscription?.expires_at}
                                <div class="plan-expiry plan-expiry-danger">
                                    Berakhir {formatDate(subscription.expires_at)} · tidak diperpanjang
                                </div>
                            {:else if subscription?.expires_at}
                                <div class="plan-expiry">Berlaku hingga {formatDate(subscription.expires_at)}</div>
                            {/if}

                            {#if pro && subscription?.status !== "expired"}
                                {#if cancelAtPeriodEnd}
                                    <Button
                                        variant="soft"
                                        size="sm"
                                        full
                                        disabled={savingSub}
                                        onclick={resumeSubscription}
                                    >
                                        {savingSub ? "Memproses…" : "Lanjutkan langganan"}
                                    </Button>
                                {:else if showCancelConfirm}
                                    <div class="plan-confirm">
                                        <div class="plan-confirm-text">
                                            Anda tetap bisa memakai {planName} sampai
                                            {formatDate(subscription?.expires_at)}, lalu paket turun ke Gratis.
                                        </div>
                                        <div class="plan-confirm-actions">
                                            <Button variant="ghost" size="sm" onclick={() => (showCancelConfirm = false)}>
                                                Batal
                                            </Button>
                                            <Button variant="danger" size="sm" disabled={savingSub} onclick={confirmCancelSubscription}>
                                                {savingSub ? "Memproses…" : "Ya, batalkan"}
                                            </Button>
                                        </div>
                                    </div>
                                {:else}
                                    <button type="button" class="plan-cancel-link" onclick={() => (showCancelConfirm = true)}>
                                        Batalkan perpanjangan
                                    </button>
                                {/if}
                            {/if}
```
> Note: if the `Button` component has no `danger`/`ghost` variant, use the nearest existing variants (check `Button.svelte`'s `variant` prop) and record the substitution. Keep the upgrade buttons block (`{#if !pro} … Upgrade … {/if}`) as-is for now (Task 7 relabels/scopes the modal).

- [ ] **Step 5: Relabel the upgrade button generically**

Change the `!pro` upgrade button label from `Upgrade ke Pro` to `Upgrade paket` (the modal now lets the user pick the tier). Leave the `onclick={onUpgradeRequest}` wiring and the trial button untouched.

- [ ] **Step 6: Add minimal styles**

In the ProfilePage `<style>` (near `.plan-expiry`), add:
```css
    .plan-cancel-link {
        background: none;
        border: none;
        padding: 0;
        font-size: 0.78rem;
        color: #94a3b8;
        text-decoration: underline;
        cursor: pointer;
        align-self: flex-start;
    }
    .plan-confirm {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        padding: 0.625rem;
        border-radius: 0.625rem;
        background: rgba(244, 63, 94, 0.06);
    }
    .plan-confirm-text {
        font-size: 0.78rem;
        color: #475569;
        line-height: 1.4;
    }
    .plan-confirm-actions {
        display: flex;
        gap: 0.5rem;
        justify-content: flex-end;
    }
```

- [ ] **Step 7: Type-check + build**

Run (in `frontend-svelte/`): `npm run check && npm run build`
Expected: 0 errors; build succeeds.

- [ ] **Step 8: Commit**
```bash
git add frontend-svelte/src/lib/services/apiDomains/authSubscriptionApi.js frontend-svelte/src/lib/pages/ProfilePage.svelte
git commit -m "feat(profile): subscription cancel/resume controls + generic upgrade label"
```

---

### Task 7: Frontend — UpgradeModal higher-tier filter + Enterprise contact-sales

**Files:**
- Modify: `frontend-svelte/src/lib/components/UpgradeModal.svelte`

**Interfaces:**
- Consumes: `ApiService.getSubscriptionStatus()`, `rankOf` from `pricing.js`.
- Produces: a modal that only offers tiers above the current rank, with an Enterprise contact-sales fallback.

- [ ] **Step 1: Import `rankOf` and add current-rank state**

In `UpgradeModal.svelte` `<script>`, extend the pricing import and add state. Change:
```js
    import { PLANS, planMeta, formatIDR } from "../config/pricing.js";
```
to:
```js
    import { PLANS, planMeta, formatIDR, rankOf } from "../config/pricing.js";
```
Add after `purchasableTiers`:
```js
    // Current plan rank: only tiers strictly above it are offered (no paid→lower-paid).
    let currentRank = $state(0);
    $effect(() => {
        if (show) {
            ApiService.getSubscriptionStatus()
                .then((s) => { currentRank = rankOf(s?.plan); })
                .catch(() => { currentRank = 0; });
        }
    });
    let offeredTiers = $derived(purchasableTiers.filter((t) => rankOf(t.key) > currentRank));
```
Define the sales WhatsApp link (reuse the landing page's constant if exported; otherwise a local const):
```js
    const SALES_WA = "https://wa.me/6281234567890"; // TODO: replace with the real sales number / shared constant if one exists
```
> Before hardcoding, run `grep -rn "SALES_WA\|wa.me" frontend-svelte/src/lib` and reuse the existing sales constant if present.

- [ ] **Step 2: Keep `selectedTier` valid against the offered set**

Add an effect that snaps `selectedTier` to the first offered tier whenever the offered set changes and the current selection isn't in it:
```js
    $effect(() => {
        if (offeredTiers.length && !offeredTiers.some((t) => t.key === selectedTier)) {
            selectedTier = offeredTiers[0].key;
        }
    });
```

- [ ] **Step 3: Render from `offeredTiers`, with a top-tier fallback**

In the tier-selector block, change `{#each purchasableTiers as t}` to `{#each offeredTiers as t}`. Immediately above the tier selector, add the empty-state fallback:
```svelte
                        {#if offeredTiers.length === 0}
                            <div style="text-align:center; padding: 12px 0;">
                                <p style="font-size:14px; color:#64748b; margin-bottom:12px;">
                                    Anda sudah di paket tertinggi yang tersedia. Untuk Enterprise, hubungi tim sales kami.
                                </p>
                                <a href={SALES_WA} target="_blank" rel="noopener" class="wa-confirm-btn" style="background:#25d366;">
                                    Hubungi Sales (Enterprise)
                                </a>
                            </div>
                        {:else}
```
Close the `{:else}` with a matching `{/if}` after the pay button block (so the selector, price, features, and pay button only render when there *is* an offered tier). Verify brace/`{/if}` balance carefully.

- [ ] **Step 4: Type-check + build**

Run (in `frontend-svelte/`): `npm run check && npm run build`
Expected: 0 errors; build succeeds.

- [ ] **Step 5: Commit**
```bash
git add frontend-svelte/src/lib/components/UpgradeModal.svelte
git commit -m "feat(profile): UpgradeModal offers only higher tiers + Enterprise contact-sales"
```

---

## Final verification (after all tasks)

- [ ] `go build ./cmd/... && go vet ./... && go test ./...` — all green; `gofmt -l internal/auth/` empty.
- [ ] In `frontend-svelte/`: `npm run check` (0 errors), `npm test`, `npm run build`.
- [ ] Manual (post-deploy, super-admin or a test org): on a paid sub → "Batalkan perpanjangan" → confirm → shows "Berakhir {date} · tidak diperpanjang", tier/limits unchanged; "Lanjutkan langganan" → back to "Berlaku hingga"; upgrade modal from profile shows only higher tiers; Bisnis user sees the Enterprise contact-sales fallback; re-upgrade clears the cancel flag.
