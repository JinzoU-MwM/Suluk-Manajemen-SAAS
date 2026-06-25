# Renewal Reminders (H-7/H-3/H-1) Implementation Plan (Phase 5)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** An hourly auth-service scheduler that emails the org owner (+ in-app nudge) at 7/3/1 days before a paid subscription expires, once per threshold per cycle, idempotently.

**Architecture:** A ticker scheduler in auth-service (mirrors `StartCleanupScheduler`) queries paid subs expiring within 7 days; a pure `dueReminder` function picks the most-urgent unsent threshold; the owner is emailed (gated on `notify_expiry`) and an in-app notification raised; sends are recorded in a `subscription_reminders` table keyed by `(org_id, expires_at, threshold)` so it's idempotent and survives renewals.

**Tech Stack:** Go (Fiber), PostgreSQL (`jamaah_auth`), existing `internal/shared/email` client, in-app `CreateNotification`.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-06-25-renewal-reminders-design.md`. Closes the pricing restructure (Phases 1–5).
- Paid only: `status='active' AND plan <> 'gratis'`; skip `cancel_at_period_end=TRUE`; only `expires_at > NOW()`.
- Email + in-app, both gated on the owner's `notify_expiry` (default TRUE). Recipient = org owner (`team_members.role='owner'`).
- **Mark a threshold sent only after the email send succeeds** (or when suppressed by `notify_expiry=false`). Email failure → not marked → retried next tick. In-app is secondary (best-effort, only on email success).
- Idempotent + monotonic: each threshold emails at most once per `expires_at`; after downtime, one catch-up email (most-urgent), mark all applicable.
- Best-effort per sub: one org's failure never aborts the run. Commit messages: NO AI co-author line.

---

### Task 1: `dueReminder` pure logic

**Files:**
- Create: `internal/auth/service/reminders.go`
- Test: `internal/auth/service/reminders_test.go`

**Interfaces:**
- Produces: `dueReminder(daysLeft float64, sent []int) (emailThreshold int, markThresholds []int, ok bool)`. Consumed by Task 4.

- [ ] **Step 1: Write the failing test**

Create `internal/auth/service/reminders_test.go`:

```go
package service

import (
	"reflect"
	"testing"
)

func TestDueReminder(t *testing.T) {
	cases := []struct {
		name      string
		daysLeft  float64
		sent      []int
		wantEmail int
		wantMark  []int
		wantOk    bool
	}{
		{"H-7 first", 7, nil, 7, []int{7}, true},
		{"H-3 after 7 sent", 3, []int{7}, 3, []int{3, 7}, true},
		{"H-1 after 7,3 sent", 1, []int{7, 3}, 1, []int{1, 3, 7}, true},
		{"between 7 and 3, already sent 7", 5, []int{7}, 0, nil, false},
		{"downtime catch-up at 2 days, none sent", 2, nil, 3, []int{3, 7}, true},
		{"all sent", 0.5, []int{1, 3, 7}, 0, nil, false},
		{"far out, nothing applicable", 10, nil, 0, nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			email, mark, ok := dueReminder(c.daysLeft, c.sent)
			if email != c.wantEmail || ok != c.wantOk || !reflect.DeepEqual(mark, c.wantMark) {
				t.Errorf("dueReminder(%v,%v) = (%d,%v,%v), want (%d,%v,%v)",
					c.daysLeft, c.sent, email, mark, ok, c.wantEmail, c.wantMark, c.wantOk)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/auth/service/ -run TestDueReminder`
Expected: FAIL — `undefined: dueReminder`.

- [ ] **Step 3: Write minimal implementation**

Create `internal/auth/service/reminders.go`:

```go
package service

// reminderThresholds are the days-before-expiry marks, ascending (most-urgent first).
var reminderThresholds = []int{1, 3, 7}

// dueReminder decides which renewal reminder is due for a sub daysLeft from
// expiry, given thresholds already sent this cycle. It returns the single
// most-urgent unsent threshold to EMAIL, the full set of currently-applicable
// thresholds to MARK (so a larger window entered late never fires out of order),
// and ok=false when nothing new is due.
func dueReminder(daysLeft float64, sent []int) (emailThreshold int, markThresholds []int, ok bool) {
	sentSet := make(map[int]bool, len(sent))
	for _, t := range sent {
		sentSet[t] = true
	}
	for _, t := range reminderThresholds { // ascending
		if daysLeft <= float64(t) {
			markThresholds = append(markThresholds, t)
			if emailThreshold == 0 && !sentSet[t] {
				emailThreshold = t // smallest unsent applicable
			}
		}
	}
	if emailThreshold == 0 {
		return 0, nil, false
	}
	return emailThreshold, markThresholds, true
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/auth/service/ -run TestDueReminder -v`
Expected: PASS (all subtests).

- [ ] **Step 5: Commit**

```bash
git add internal/auth/service/reminders.go internal/auth/service/reminders_test.go
git commit -m "feat(auth): dueReminder threshold logic for renewal reminders (Phase 5)"
```

---

### Task 2: `renderRenewalEmail` pure copy

**Files:**
- Modify: `internal/auth/service/reminders.go`
- Test: `internal/auth/service/reminders_test.go`

**Interfaces:**
- Produces: `renderRenewalEmail(planKey string, daysLeft int, expiresAt time.Time, renewURL string) (subject, html string)`. Consumed by Task 4.

- [ ] **Step 1: Write the failing test**

Append to `internal/auth/service/reminders_test.go` (and add `"strings"` + `"time"` to its imports):

```go
func TestRenderRenewalEmail(t *testing.T) {
	exp := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	subject, html := renderRenewalEmail("pro", 3, exp, "https://app.suluk.site/")

	if !strings.Contains(subject, "Pro") || !strings.Contains(subject, "3 hari") {
		t.Errorf("subject missing plan/days: %q", subject)
	}
	for _, want := range []string{"Pro", "3 hari", "01-07-2026", "https://app.suluk.site/", "Perpanjang"} {
		if !strings.Contains(html, want) {
			t.Errorf("html missing %q:\n%s", want, html)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/auth/service/ -run TestRenderRenewalEmail`
Expected: FAIL — `undefined: renderRenewalEmail`.

- [ ] **Step 3: Write minimal implementation**

Add to `internal/auth/service/reminders.go` (add imports `"fmt"`, `"time"`, and `"github.com/jamaah-in/v2/internal/shared/plan"`):

```go
// renderRenewalEmail builds the Indonesian renewal-reminder subject + HTML body.
func renderRenewalEmail(planKey string, daysLeft int, expiresAt time.Time, renewURL string) (subject, html string) {
	name := plan.Get(planKey).Name
	subject = fmt.Sprintf("Paket %s Anda berakhir dalam %d hari", name, daysLeft)
	html = fmt.Sprintf(`<p>Halo,</p>
<p>Paket <strong>%s</strong> Anda akan berakhir dalam <strong>%d hari</strong> (pada %s).</p>
<p>Perpanjang sekarang biar fitur gak keputus dan akun gak turun ke paket Gratis.</p>
<p><a href="%s">Perpanjang paket</a></p>`,
		name, daysLeft, expiresAt.Format("02-01-2006"), renewURL)
	return subject, html
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/auth/service/ -run TestRenderRenewalEmail -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/auth/service/reminders.go internal/auth/service/reminders_test.go
git commit -m "feat(auth): renewal reminder email template (Phase 5)"
```

---

### Task 3: tracking table + repo queries

**Files:**
- Create: `migrations/auth/017_subscription_reminders.up.sql`, `.down.sql`
- Modify: `internal/auth/model/model.go` (add `ExpiringSub`)
- Modify: `internal/auth/repository/repository.go`

**Interfaces:**
- Produces: `model.ExpiringSub{OrgID uuid.UUID; Plan string; ExpiresAt time.Time}`; `(*AuthRepo) ListExpiringSubscriptions(ctx) ([]model.ExpiringSub, error)`; `GetOrgOwner(ctx, orgID) (email string, notifyExpiry bool, err error)`; `SentReminderThresholds(ctx, orgID, expiresAt) ([]int, error)`; `MarkReminderSent(ctx, orgID, expiresAt, threshold int) error`. Consumed by Task 4.

- [ ] **Step 1: Write the migration**

`migrations/auth/017_subscription_reminders.up.sql`:

```sql
-- Idempotency ledger for renewal reminders (Phase 5). One row per
-- (org, expiry cycle, threshold); a renewal = new expires_at = fresh rows.
CREATE TABLE IF NOT EXISTS subscription_reminders (
    org_id     UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    threshold  INT NOT NULL,          -- 7 | 3 | 1 days before expiry
    sent_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (org_id, expires_at, threshold)
);
```

`migrations/auth/017_subscription_reminders.down.sql`:

```sql
DROP TABLE IF EXISTS subscription_reminders;
```

- [ ] **Step 2: Add the `ExpiringSub` model**

In `internal/auth/model/model.go`, after the `Subscription` struct, add:

```go
// ExpiringSub is the slice of a subscription the renewal-reminder job needs.
type ExpiringSub struct {
	OrgID     uuid.UUID
	Plan      string
	ExpiresAt time.Time
}
```

- [ ] **Step 3: Add the repo methods**

Append to `internal/auth/repository/repository.go` (uses existing imports `context`, `time`, `uuid`, `model`):

```go
// ListExpiringSubscriptions returns active PAID subs expiring within 7 days
// (the largest reminder threshold) that have not opted out of renewal.
func (r *AuthRepo) ListExpiringSubscriptions(ctx context.Context) ([]model.ExpiringSub, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT org_id, plan, expires_at FROM subscriptions
		WHERE status = 'active' AND plan <> 'gratis' AND cancel_at_period_end = FALSE
		  AND expires_at IS NOT NULL AND expires_at > NOW() AND expires_at <= NOW() + INTERVAL '7 days'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ExpiringSub
	for rows.Next() {
		var s model.ExpiringSub
		if err := rows.Scan(&s.OrgID, &s.Plan, &s.ExpiresAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// GetOrgOwner returns the org owner's email + expiry-notification preference.
func (r *AuthRepo) GetOrgOwner(ctx context.Context, orgID uuid.UUID) (email string, notifyExpiry bool, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT u.email, COALESCE(u.notify_expiry, TRUE)
		FROM users u JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.org_id = $1 AND tm.role = 'owner' AND tm.status = 'active'
		ORDER BY tm.joined_at LIMIT 1`,
		orgID).Scan(&email, &notifyExpiry)
	return email, notifyExpiry, err
}

// SentReminderThresholds lists the thresholds already recorded for this expiry cycle.
func (r *AuthRepo) SentReminderThresholds(ctx context.Context, orgID uuid.UUID, expiresAt time.Time) ([]int, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT threshold FROM subscription_reminders WHERE org_id = $1 AND expires_at = $2`,
		orgID, expiresAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int
	for rows.Next() {
		var t int
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// MarkReminderSent records a threshold as sent (idempotent).
func (r *AuthRepo) MarkReminderSent(ctx context.Context, orgID uuid.UUID, expiresAt time.Time, threshold int) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO subscription_reminders (org_id, expires_at, threshold)
		VALUES ($1, $2, $3) ON CONFLICT (org_id, expires_at, threshold) DO NOTHING`,
		orgID, expiresAt, threshold)
	return err
}
```

- [ ] **Step 4: Verify build**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" build ./...`
Expected: no output (DB-bound SQL verified by build + Task 4 integration; no test DB in this repo).

- [ ] **Step 5: Commit**

```bash
git add migrations/auth/017_subscription_reminders.up.sql migrations/auth/017_subscription_reminders.down.sql internal/auth/model/model.go internal/auth/repository/repository.go
git commit -m "feat(auth): subscription_reminders table + expiry/owner queries (Phase 5)"
```

---

### Task 4: scheduler orchestration + wiring

**Files:**
- Modify: `internal/auth/service/reminders.go` (`runRenewalReminders`, `StartRenewalReminderScheduler`)
- Modify: `cmd/auth-service/main.go` (start the scheduler)

**Interfaces:**
- Consumes: Task 1 `dueReminder`, Task 2 `renderRenewalEmail`, Task 3 repo methods, existing `s.email.Send`/`s.email.Enabled`, `s.CreateNotification`, `plan.Get`.
- Produces: `(*AuthService) StartRenewalReminderScheduler(ctx, publicURL string)`.

- [ ] **Step 1: Add orchestration + scheduler**

Add to `internal/auth/service/reminders.go` (add imports `"context"`, `"log"`, and `"github.com/jamaah-in/v2/internal/auth/model"`):

```go
// runRenewalReminders sends due H-7/3/1 reminders for paid subs nearing expiry.
// Best-effort per sub; a threshold is marked only after a successful email (or
// when the owner opted out via notify_expiry), so transient failures retry.
func (s *AuthService) runRenewalReminders(ctx context.Context, publicURL string) {
	subs, err := s.repo.ListExpiringSubscriptions(ctx)
	if err != nil {
		log.Printf("renewal reminders: list expiring subs: %v", err)
		return
	}
	now := time.Now()
	for _, sub := range subs {
		daysLeft := sub.ExpiresAt.Sub(now).Hours() / 24
		sent, err := s.repo.SentReminderThresholds(ctx, sub.OrgID, sub.ExpiresAt)
		if err != nil {
			log.Printf("renewal reminders: sent thresholds (org %s): %v", sub.OrgID, err)
			continue
		}
		emailT, markTs, ok := dueReminder(daysLeft, sent)
		if !ok {
			continue
		}
		email, notifyExpiry, err := s.repo.GetOrgOwner(ctx, sub.OrgID)
		if err != nil {
			log.Printf("renewal reminders: owner (org %s): %v", sub.OrgID, err)
			continue
		}

		marked := false
		if !notifyExpiry {
			marked = true // opted out: mark so we don't re-evaluate hourly
		} else if email != "" && s.email != nil && s.email.Enabled() {
			subject, html := renderRenewalEmail(sub.Plan, emailT, sub.ExpiresAt, publicURL)
			if err := s.email.Send(ctx, email, subject, html); err != nil {
				log.Printf("renewal reminders: email (org %s): %v", sub.OrgID, err)
			} else {
				_ = s.CreateNotification(ctx, &model.Notification{
					OrgID:    sub.OrgID,
					Severity: "warning",
					Title:    "Paket akan berakhir",
					Message:  fmt.Sprintf("Paket %s Anda berakhir dalam %d hari. Perpanjang biar fitur gak keputus.", plan.Get(sub.Plan).Name, emailT),
				})
				marked = true
			}
		}
		// else: email unavailable/no address — leave unmarked, retry next tick.

		if marked {
			for _, t := range markTs {
				if err := s.repo.MarkReminderSent(ctx, sub.OrgID, sub.ExpiresAt, t); err != nil {
					log.Printf("renewal reminders: mark sent (org %s, t=%d): %v", sub.OrgID, t, err)
				}
			}
		}
	}
}

// StartRenewalReminderScheduler runs runRenewalReminders hourly (and once at
// startup), mirroring StartCleanupScheduler. publicURL is the renew-link base.
func (s *AuthService) StartRenewalReminderScheduler(ctx context.Context, publicURL string) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.runRenewalReminders(ctx, publicURL)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	s.runRenewalReminders(ctx, publicURL)
}
```

- [ ] **Step 2: Wire it in main**

In `cmd/auth-service/main.go`, after `authService` is constructed (the `service.NewAuthService(...).WithEmail(...).WithScanUsageSource(...)` chain ending around line 80) and before/near `authHandler := handler.NewAuthHandler(authService)`, add:

```go
	authService.StartRenewalReminderScheduler(ctx, cfg.App.PublicURL)
```

- [ ] **Step 3: Build + full test**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" build ./... && go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./...`
Expected: clean build; all packages `ok` (including `TestDueReminder`, `TestRenderRenewalEmail`).

- [ ] **Step 4: Commit**

```bash
git add internal/auth/service/reminders.go cmd/auth-service/main.go
git commit -m "feat(auth): hourly renewal-reminder scheduler H-7/3/1 (Phase 5)"
```

---

## Self-Review

- **Spec coverage:** scheduler hourly in auth (Task 4) ✓; query paid+active, skip gratis/cancel/past-expiry (Task 3) ✓; `dueReminder` most-urgent-unsent + mark-all-applicable, downtime catch-up (Task 1) ✓; email + in-app gated on notify_expiry, mark-after-email-success (Task 4) ✓; owner recipient (Task 3 GetOrgOwner) ✓; idempotent table keyed by (org,expires_at,threshold) (Task 3) ✓; renew copy + link (Task 2) ✓. All §3–§7 mapped.
- **Placeholder scan:** every step shows real SQL/Go/tests; migration is `017` (next after `016_subscription_cancel`). ✓
- **Type consistency:** `dueReminder(float64,[]int) (int,[]int,bool)`, `renderRenewalEmail(string,int,time.Time,string) (string,string)`, `ListExpiringSubscriptions(ctx) ([]model.ExpiringSub,error)`, `GetOrgOwner(ctx,uuid) (string,bool,error)`, `MarkReminderSent(ctx,uuid,time.Time,int)` — consistent across producers (Tasks 1–3) and consumer (Task 4). ✓
- **DB-test honesty:** repo SQL (Task 3) + orchestration (Task 4) build-verified + integration, matching repo convention; the risk-bearing pure logic (Tasks 1–2) is fully TDD'd. ✓
