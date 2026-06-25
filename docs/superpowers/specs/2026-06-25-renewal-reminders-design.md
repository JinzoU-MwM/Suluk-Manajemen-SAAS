# Renewal Reminders (H-7/H-3/H-1) — Design (Phase 5)

**Tanggal:** 2026-06-25
**Status:** Approved (pending spec review)
**Sub-phase of:** Phase 5 in `2026-06-24-pricing-tier-restructure.md` (spec §3.4.3). **This closes the pricing restructure.**

---

## 1. Goal & decisions

Before a paid subscription lapses to Gratis, email the org owner a renewal reminder at **H-7 / H-3 / H-1** ("perpanjang biar gak turun ke Gratis"), plus an in-app nudge. Prepaid + manual renewal only; deeper retention (auto-renew/dunning) is a separate workstream.

Decisions locked in brainstorming:

| Decision | Choice |
|---|---|
| Scope | **Paid subs only** (`status='active'`, `plan <> 'gratis'`). Trials keep their in-app countdown banner; not emailed here. |
| Channel | **Email + in-app** (existing `CreateNotification`), both gated on the user's `notify_expiry` preference (default TRUE). |
| Tracking | **`subscription_reminders` table** keyed by `(org_id, expires_at, threshold)`, idempotent insert. A renewal = new `expires_at` = fresh rows, no reset logic. |
| Scheduler | **Hourly ticker in auth-service**, mirroring the existing `StartCleanupScheduler`. |
| Cancelled subs | **Skip** `cancel_at_period_end = TRUE` — they opted out of renewal. |
| Recipient | **Org owner** (`team_members.role='owner'` → user email/name/notify_expiry). |
| Quiet hours | **None** (MVP) — a reminder may send at any hour. |

**Out of scope:** auto-renew, dunning, saved cards, WhatsApp channel, trial-expiry emails, quiet-hours/timezone windowing, per-org reminder cadence customization.

## 2. Global constraints

- Reminders are **best-effort**: one org's email/notify failure never blocks the rest of the run.
- **Mark sent only after the email send succeeds** — a transient failure retries on the next hourly tick, still inside the window. (In-app is secondary; its failure is logged, not retried.)
- **Idempotent + monotonic**: each threshold emails at most once per `expires_at` cycle; after scheduler downtime a single catch-up email goes out (the most-urgent crossed threshold), never a burst or an out-of-order H-7-after-H-3.
- Only subs with `expires_at > NOW()` are considered (never remind after expiry); Gratis (no expiry) is excluded.
- Auth-service owns this (subscriptions, email client, notifications, users all live in `jamaah_auth`). Commit messages: NO AI co-author line.

## 3. Architecture & data flow

```
auth-service startup → StartRenewalReminderScheduler(ctx)   // hourly time.Ticker (like StartCleanupScheduler)
each tick → runRenewalReminders(ctx):

  subs = ListExpiringSubscriptions(within=7d):
     SELECT org_id, plan, expires_at FROM subscriptions
      WHERE status='active' AND plan <> 'gratis' AND cancel_at_period_end = FALSE
        AND expires_at > NOW() AND expires_at <= NOW() + INTERVAL '7 days'

  for each sub:
     daysLeft   = expires_at - now
     applicable = [T in {7,3,1} where daysLeft <= T]                 // windows currently inside
     sent       = SentReminderThresholds(org, expires_at)
     unsent     = applicable \ sent
     if unsent is empty: continue
     owner = GetOrgOwner(org)                                        // email, name, notify_expiry
     markAll = false
     if not owner.notify_expiry:
        markAll = true                                               // suppressed: mark so we don't re-check hourly
     else:
        T* = min(unsent)                                             // most-urgent threshold
        (subject, html) = renderRenewalEmail(plan, T*, expires_at, renewURL)
        if email.Send(owner.email, subject, html) succeeds:
           CreateNotification(org, "warning", title, msg)           // best-effort, secondary
           markAll = true
        else:
           log error                                                // do NOT mark → retry next tick
     if markAll:
        for T in applicable: MarkReminderSent(org, expires_at, T)   // mark all applicable windows
```

**Note on marking when suppressed:** if `notify_expiry` is false we still `MarkReminderSent` for all applicable thresholds so the opted-out org isn't re-evaluated every hour. If `notify_expiry` is true but the email errors, we mark nothing (retry next tick).

## 4. Components & interfaces

**Migration** `auth/0NN_subscription_reminders.up.sql`:
```sql
CREATE TABLE IF NOT EXISTS subscription_reminders (
    org_id     UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    threshold  INT NOT NULL,          -- 7 | 3 | 1 (days before expiry)
    sent_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (org_id, expires_at, threshold)
);
```

**Repo (`internal/auth/repository`):**
- `ListExpiringSubscriptions(ctx, within time.Duration) ([]ExpiringSub, error)` — `ExpiringSub{OrgID uuid.UUID; Plan string; ExpiresAt time.Time}`.
- `GetOrgOwner(ctx, orgID) (email, name string, notifyExpiry bool, err error)` — `JOIN users ON team_members` where `role='owner'` (active), `LIMIT 1`.
- `SentReminderThresholds(ctx, orgID, expiresAt) ([]int, error)`.
- `MarkReminderSent(ctx, orgID, expiresAt, threshold int) error` — `INSERT … ON CONFLICT DO NOTHING`.

**Pure logic (`internal/auth/service`):**
- `dueReminder(daysLeft float64, sent []int) (emailThreshold int, markThresholds []int, ok bool)` — thresholds `{7,3,1}`; `applicable = daysLeft <= T`; `markThresholds = applicable`; `emailThreshold = min(applicable \ sent)`; `ok = unsent non-empty`.
- `renderRenewalEmail(plan string, daysLeft int, expiresAt time.Time, renewURL string) (subject, html string)` — Indonesian copy; subject e.g. `"Paket Pro Anda berakhir dalam 3 hari"`; body names plan, expiry date, days left, "perpanjang biar fitur gak keputus / gak turun ke Gratis", and a renew CTA linking `renewURL` (`APP_PUBLIC_URL` + billing/pricing path).

**Service orchestration (`internal/auth/service`):**
- `runRenewalReminders(ctx) error` — the loop above; uses `s.repo`, `s.email`, `s.CreateNotification`. Logs per-sub failures, continues.
- `StartRenewalReminderScheduler(ctx)` — `time.NewTicker(1 * time.Hour)` goroutine calling `runRenewalReminders`; also runs once shortly after start.

**Wiring (`cmd/auth-service/main.go`):** call `authService.StartRenewalReminderScheduler(ctx)` alongside `authRepo.StartCleanupScheduler(ctx)`. Needs `APP_PUBLIC_URL` (already configured) for the renew link.

## 5. Edge cases

- **Renewal mid-window:** new `expires_at` → tracking rows keyed by it are absent → fresh H-7/3/1 for the new cycle. Old rows are inert.
- **Scheduler downtime:** first run after crossing several windows emails once (most-urgent) and marks all applicable — no burst, no out-of-order.
- **Email fails:** not marked → retried next hour (bounded: once `expires_at <= now`, the sub leaves the query).
- **`notify_expiry=false`:** no email/in-app; thresholds marked so the org isn't re-processed hourly.
- **No owner / no email on file:** log and skip (no crash); not marked, so it self-heals if an owner/email appears.
- **Multiple owners:** `LIMIT 1` (first active owner) — one recipient.

## 6. Testing (TDD)

- **`dueReminder`**: at 7/3/1 days each fires once; between windows nothing new; all-sent → `ok=false`; downtime catch-up (daysLeft=2, none sent) → email `3`, mark `{3,7}` (not `1`); already-sent suppresses re-email.
- **`renderRenewalEmail`**: subject/body contain plan name, days left, formatted expiry date, and the renew URL; plural/singular day copy correct.
- Scheduler loop, repo SQL, owner resolution = build-verified + integration (no test DB in this repo).

## 7. Definition of Done

- Hourly scheduler emails the org owner at H-7/H-3/H-1 before a paid sub expires, once per threshold per cycle, respecting `notify_expiry` and skipping `cancel_at_period_end`.
- Each reminder also raises an in-app notification.
- Reminders are idempotent across ticks and survive renewals; transient email failures retry within the window.
- `go build ./...` + `go test ./...` green; pure logic covered by tests.
