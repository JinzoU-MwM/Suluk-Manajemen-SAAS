# Profile Core — Design (Sub-project 1 of the Profile overhaul)

- **Date:** 2026-06-19
- **Status:** Approved — proceeding to implementation plan
- **Area:** `auth-service` (Go) + `frontend-svelte` `ProfilePage`
- **Part of:** a 5-part profile effort. This is **Sub-project 1**. Later: (2) subscription upgrade/downgrade, (3) avatar photo upload, (4) active-session management, (5) real 2FA.

## 1. Context & problem

The `/app` profile (`ProfilePage.svelte`) looks full-featured but most of it is non-functional:

- `UpdateMe` only accepts **name** and **phone**. The frontend also "saves" `avatar_color`, `notify_usage_limit`, `notify_expiry` — but the backend ignores them (no columns, not parsed), so they don't persist.
- `city` and `bio` are shown but have **no DB columns** and are always "—".
- The super-admin link uses a **broken legacy `/#/super-admin`** hash URL.
- The subscription's **expiry date is never shown**, even though subscriptions carry an `expires_at`.

The user wants the profile **fully customizable** (every field really saves), the super-admin nav working **plus a new super-admin account**, and the **subscription expiry visible**.

## 2. Goals

- Persist every editable profile field: name, phone, **city**, **bio**, **avatar_color**, **notify_usage_limit**, **notify_expiry**.
- Make those fields editable in the Profil tab; avatar-color + notification toggles actually save.
- Fix the super-admin entry on the profile (`-> /super-admin`, shown only to super-admins).
- Create a **new super-admin account** (generated credentials, shared with the user).
- Show the subscription **expiry date** on the profile.
- Remove the genuinely-broken bits in scope (the mis-wired "suspicious-login" toggle; the fake "change photo" camera button).

## 3. Non-goals (handled in later sub-projects)

- **Photo upload + storage** (SP3) — avatar stays color-based for now; remove the fake camera button.
- **Active-session management** (SP4) — leave the "Sesi Aktif" placeholder as-is (do not build, do not remove).
- **Real 2FA** (SP5) — leave the disabled "2FA — Segera hadir" toggle as-is.
- **Subscription upgrade/downgrade** (SP2) — keep the existing "Upgrade ke Pro" button (Pakasir flow) unchanged; this sub-project only *shows* the expiry.
- **Changing the login email** — email stays read-only (needs a verification flow).
- **Role/jabatan editing** — system-controlled, read-only.

## 4. Decisions

| Topic | Decision |
|-------|----------|
| Persistence | **Full** — add columns + extend the API (Approach A). |
| Editable fields | name, phone, city, bio, avatar_color, notify_usage_limit, notify_expiry. |
| Email | Read-only (out of scope). |
| Super-admin account | Generate `superadmin@suluk.site` + strong password, `is_super_admin=true`; share creds. |
| Photo | Defer to SP3; remove the fake camera button now. |
| Placeholders for SP4/SP5 | Leave in place (they're planned), don't remove. |

## 5. Backend changes (`auth-service`)

### 5.1 Migration `migrations/auth/015_profile_fields.{up,down}.sql`
Add nullable columns to `users` (`phone` already exists):
```sql
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS city               TEXT,
  ADD COLUMN IF NOT EXISTS bio                TEXT,
  ADD COLUMN IF NOT EXISTS avatar_color       TEXT    NOT NULL DEFAULT 'blue',
  ADD COLUMN IF NOT EXISTS notify_usage_limit BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS notify_expiry      BOOLEAN NOT NULL DEFAULT TRUE;
```
Down: `ALTER TABLE users DROP COLUMN IF EXISTS ...` for the five.

### 5.2 User model (`internal/auth/model/model.go`)
Add fields to `User`: `City *string`, `Bio *string`, `AvatarColor string`, `NotifyUsageLimit bool`, `NotifyExpiry bool` (with `json`/`db` tags).

### 5.3 `UpdateMe` (`internal/auth/handler/handler.go:107`)
Extend the request struct + the `UpdateUser` service/repo to accept and persist all seven fields:
`name, phone, city, bio, avatar_color, notify_usage_limit, notify_expiry`. Validate: name required (non-empty); avatar_color must be one of the known palette keys (else default 'blue'); strings trimmed. Booleans default to their current value when omitted (partial update).

### 5.4 `sanitizeUser` + `getMe` (`handler.go:~468`)
Include the new fields in the returned map so the frontend reads real values (city, bio, avatar_color, notify_usage_limit, notify_expiry).

### 5.5 Subscription status (`internal/auth/service` GetSubscriptionStatus)
Add `expires_at` (the subscription's `ExpiresAt`) to the returned result if not already present, so the frontend can show it. (The value already exists on the subscription record; this only surfaces it.)

## 6. Frontend changes (`ProfilePage.svelte`)

- **Profil tab edit mode:** name, phone, city, bio become inputs (currently only name). Email + Jabatan stay read-only display. On Save, send all editable fields via `ApiService.updateProfile(...)`.
- **Avatar color:** already calls `updateProfile({avatar_color})`; now it persists. No UI change needed beyond removing the fake **camera button** (`.summary-avatar-cam`).
- **Notification toggles:** the two real preference toggles in the Notifikasi tab (usage-limit, expiry) now persist via the backend. Clean up the **mis-wired toggles** bound to the wrong state — the Keamanan tab's "Notifikasi Login Mencurigakan" and the Notifikasi tab's "Email" channel are both wired to `notifyExpiry`: remove these placeholder rows (their real behavior — login alerts, an email channel — is future work).
- **Super-admin link:** change `href="/#/super-admin"` → `href="/super-admin"` (still gated by `user?.is_super_admin`).
- **Subscription expiry:** in the plan block, show `Berlaku hingga {formatDate(subscription.expires_at)}` when present (and a clear "Kedaluwarsa" state when `status === 'expired'`).
- Keep SP4 (Sesi Aktif) and SP5 (2FA disabled) placeholders untouched.

## 7. Super-admin account creation

A one-off data operation on the server's `jamaah_auth` DB (not app UI):
- Email `superadmin@suluk.site` (confirm/adjust at run time), a strong generated password, `name` e.g. "Super Admin", `is_super_admin = true`, `is_active = true`, `email_verified = true`.
- Created with a **properly bcrypt-hashed** password (reuse the auth service's hashing — e.g. register via the auth flow then promote, or a small seed that uses the same hash cost). Exact mechanism decided in the plan.
- Credentials shared with the user after creation; the user changes the password on first login.

## 8. Edge cases

- Partial updates: omitting a field leaves it unchanged (don't null out city/bio on a name-only save).
- `avatar_color` constrained to the known palette; unknown → 'blue'.
- Old users (pre-migration) get defaults (avatar 'blue', notify true) via the column defaults.
- Email read-only — never sent in the update payload.
- Subscription with no `expires_at` (e.g. free plan) → hide the expiry line.

## 9. Testing

- **Go:** unit/service test that `UpdateUser` persists and returns the new fields (name/phone/city/bio/avatar_color/notify flags); validation (name required, avatar_color clamp).
- **Frontend:** `npm run check` (0 errors), `npm run build`.
- **Manual:** edit each field → reload → persists; avatar color + notifications persist; super-admin link reaches `/super-admin`; expiry shows.
- Gates: `go build ./cmd/...`, `go vet`, `go test ./...`, `npm run check`, `npm test`.

## 10. Rollout

- Migration `015` on `jamaah_auth`; rebuild + restart **auth-service** + **frontend** (surgical).
- Create the super-admin account on the server after deploy.
- Additive and backward-compatible (new nullable/defaulted columns; existing behavior unchanged).
