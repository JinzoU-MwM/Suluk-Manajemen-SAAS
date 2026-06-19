# Profile Core Implementation Plan (Sub-project 1)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the `/app` profile fully customizable (every editable field really persists), fix the super-admin nav + create a super-admin account, and show the subscription expiry date.

**Architecture:** Add five `users` columns in the auth DB; extend the user read path (GetUserByID SELECT + sanitizeUser) and the write path (PUT `/auth/me` → service → repo) to carry them, with a pure partial-update/validation helper that's unit-tested. Surface `expires_at` in the subscription-status response. Frontend `ProfilePage` makes name/phone/city/bio editable, persists avatar-color + notify prefs (now backed), removes dead placeholders, fixes the super-admin link, and shows the expiry.

**Tech Stack:** Go (Fiber, pgx v5, bcrypt), Postgres (`jamaah_auth` DB), SvelteKit 5 + Tailwind.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-06-19-profile-core-design.md`.
- Auth DB is `jamaah_auth`; migrations in `migrations/auth/` (next = `015`).
- Editable & persisted fields: `name, phone, city, bio, avatar_color, notify_usage_limit, notify_expiry`. **Email read-only; role system-controlled.**
- `avatar_color` ∈ {emerald, blue, purple, rose, amber, cyan, indigo, slate}; unknown → `blue`.
- Partial updates: a field omitted from the request is left unchanged (use pointers).
- **Any SELECT that builds a `*model.User` used for update or `sanitizeUser` must include the new columns** (else updates wipe them).
- No AI attribution in commits. Commit at the end of each task.
- Gates: `go build ./cmd/...`, `go vet ./...`, `go test ./...`; in `frontend-svelte/`: `npm run check`, `npm test`, `npm run build`.

---

### Task 1: Migration — user profile columns

**Files:**
- Create: `migrations/auth/015_profile_fields.up.sql`
- Create: `migrations/auth/015_profile_fields.down.sql`

**Interfaces:**
- Produces: `users.city`, `users.bio`, `users.avatar_color`, `users.notify_usage_limit`, `users.notify_expiry`.

- [ ] **Step 1: Write the up migration**

`migrations/auth/015_profile_fields.up.sql`:
```sql
-- Profile customization fields on users (phone already exists).
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS city               TEXT,
  ADD COLUMN IF NOT EXISTS bio                TEXT,
  ADD COLUMN IF NOT EXISTS avatar_color       TEXT    NOT NULL DEFAULT 'blue',
  ADD COLUMN IF NOT EXISTS notify_usage_limit BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS notify_expiry      BOOLEAN NOT NULL DEFAULT TRUE;
```

- [ ] **Step 2: Write the down migration**

`migrations/auth/015_profile_fields.down.sql`:
```sql
ALTER TABLE users
  DROP COLUMN IF EXISTS city,
  DROP COLUMN IF EXISTS bio,
  DROP COLUMN IF EXISTS avatar_color,
  DROP COLUMN IF EXISTS notify_usage_limit,
  DROP COLUMN IF EXISTS notify_expiry;
```

- [ ] **Step 3: Verify SQL by inspection**

No local Postgres — verify DDL is well-formed (the up adds 5 columns idempotently; the down reverses). Application is deferred to deploy.

- [ ] **Step 4: Commit**

```bash
git add migrations/auth/015_profile_fields.up.sql migrations/auth/015_profile_fields.down.sql
git commit -m "feat(auth): migration for user profile fields"
```

---

### Task 2: User model + read path (GetUserByID, sanitizeUser)

**Files:**
- Modify: `internal/auth/model/model.go` (User struct, ~line 54)
- Modify: `internal/auth/repository/repository.go` (`GetUserByID` SELECT/Scan)
- Modify: `internal/auth/handler/handler.go` (`sanitizeUserMap`, ~line 467)

**Interfaces:**
- Produces: `model.User` fields `City *string`, `Bio *string`, `AvatarColor string`, `NotifyUsageLimit bool`, `NotifyExpiry bool`; these flow through `GetUserByID` and `sanitizeUser` (consumed by Tasks 3 & 5).

- [ ] **Step 1: Add fields to the User struct**

In `internal/auth/model/model.go`, inside `type User struct` (after `Phone`/`PhoneVerified`), add:
```go
	City             *string `json:"city,omitempty" db:"city"`
	Bio              *string `json:"bio,omitempty" db:"bio"`
	AvatarColor      string  `json:"avatar_color" db:"avatar_color"`
	NotifyUsageLimit bool    `json:"notify_usage_limit" db:"notify_usage_limit"`
	NotifyExpiry     bool    `json:"notify_expiry" db:"notify_expiry"`
```

- [ ] **Step 2: Extend `GetUserByID` to select + scan the new columns**

Find `func (r *AuthRepo) GetUserByID` in `internal/auth/repository/repository.go`. Add the five columns to its SELECT list and the matching `&u.City, &u.Bio, &u.AvatarColor, &u.NotifyUsageLimit, &u.NotifyExpiry` to the `.Scan(...)` call, in the same order. Use `COALESCE(avatar_color,'blue')` and `COALESCE(notify_usage_limit,TRUE)`/`COALESCE(notify_expiry,TRUE)` in the SELECT so pre-migration rows scan cleanly. Example shape (adapt to the existing query/scan):
```go
// SELECT id, email, name, password_hash, email_verified, phone, phone_verified,
//        role, is_active, is_super_admin, agent_id, jamaah_id,
//        city, bio, COALESCE(avatar_color,'blue'),
//        COALESCE(notify_usage_limit,TRUE), COALESCE(notify_expiry,TRUE),
//        created_at, updated_at
// ... Scan(..., &u.City, &u.Bio, &u.AvatarColor, &u.NotifyUsageLimit, &u.NotifyExpiry, ...)
```

- [ ] **Step 3: Add the fields to `sanitizeUserMap`**

In `internal/auth/handler/handler.go` `sanitizeUserMap`, add to the returned `fiber.Map` (after `"phone_verified"`):
```go
		"city":               u.City,
		"bio":                u.Bio,
		"avatar_color":       u.AvatarColor,
		"notify_usage_limit": u.NotifyUsageLimit,
		"notify_expiry":      u.NotifyExpiry,
```

- [ ] **Step 4: Build**

Run: `go build ./internal/auth/... ./cmd/auth-service/...`
Expected: success.

- [ ] **Step 5: Commit**

```bash
git add internal/auth/model/model.go internal/auth/repository/repository.go internal/auth/handler/handler.go
git commit -m "feat(auth): read profile fields (model, GetUserByID, sanitizeUser)"
```

---

### Task 3: Write path — persist profile fields (TDD on the pure helper)

**Files:**
- Modify: `internal/auth/model/model.go` (add `ProfileUpdate` + `applyProfileUpdate`)
- Test: `internal/auth/model/profile_update_test.go`
- Modify: `internal/auth/handler/handler.go` (`UpdateMe`, ~line 107)
- Modify: `internal/auth/service/service.go` (`UpdateUser`, ~line 330)
- Modify: `internal/auth/repository/repository.go` (`UpdateUser`, ~line 119)

**Interfaces:**
- Consumes: `model.User` (Task 2).
- Produces: `model.ProfileUpdate` (pointer fields), `model.ApplyProfileUpdate(u *model.User, in ProfileUpdate) error`, and `AuthService.UpdateUser(ctx, userID uuid.UUID, in model.ProfileUpdate) (*model.User, error)`.

- [ ] **Step 1: Write the failing test for the pure helper**

`internal/auth/model/profile_update_test.go`:
```go
package model

import "testing"

func sp(s string) *string { return &s }
func bp(b bool) *bool     { return &b }

func TestApplyProfileUpdatePartial(t *testing.T) {
	u := &User{Name: "Old", AvatarColor: "blue", NotifyUsageLimit: true, NotifyExpiry: true}
	if err := ApplyProfileUpdate(u, ProfileUpdate{Name: sp("New"), City: sp("Bandung")}); err != nil {
		t.Fatalf("err = %v", err)
	}
	if u.Name != "New" {
		t.Fatalf("name = %q, want New", u.Name)
	}
	if u.City == nil || *u.City != "Bandung" {
		t.Fatalf("city = %v, want Bandung", u.City)
	}
	if u.AvatarColor != "blue" { // untouched
		t.Fatalf("avatar = %q, want blue (unchanged)", u.AvatarColor)
	}
}

func TestApplyProfileUpdateRejectsEmptyName(t *testing.T) {
	u := &User{Name: "Old"}
	if err := ApplyProfileUpdate(u, ProfileUpdate{Name: sp("  ")}); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestApplyProfileUpdateClampsAvatarColor(t *testing.T) {
	u := &User{AvatarColor: "blue"}
	if err := ApplyProfileUpdate(u, ProfileUpdate{AvatarColor: sp("neon")}); err != nil {
		t.Fatalf("err = %v", err)
	}
	if u.AvatarColor != "blue" {
		t.Fatalf("avatar = %q, want blue (clamped)", u.AvatarColor)
	}
}

func TestApplyProfileUpdateBoolPointers(t *testing.T) {
	u := &User{NotifyUsageLimit: true, NotifyExpiry: true}
	if err := ApplyProfileUpdate(u, ProfileUpdate{NotifyExpiry: bp(false)}); err != nil {
		t.Fatalf("err = %v", err)
	}
	if u.NotifyExpiry != false || u.NotifyUsageLimit != true {
		t.Fatalf("notify = (%v,%v), want (true,false)", u.NotifyUsageLimit, u.NotifyExpiry)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/auth/model/ -run TestApplyProfileUpdate -v`
Expected: FAIL — `undefined: ProfileUpdate` / `ApplyProfileUpdate`.

- [ ] **Step 3: Implement `ProfileUpdate` + `ApplyProfileUpdate`**

Append to `internal/auth/model/model.go`:
```go
import "errors" // ensure imported (add to the existing import block if absent)

// ProfileUpdate carries the user-editable profile fields. A nil pointer means
// "leave unchanged" (partial update).
type ProfileUpdate struct {
	Name             *string `json:"name"`
	Phone            *string `json:"phone"`
	City             *string `json:"city"`
	Bio              *string `json:"bio"`
	AvatarColor      *string `json:"avatar_color"`
	NotifyUsageLimit *bool   `json:"notify_usage_limit"`
	NotifyExpiry     *bool   `json:"notify_expiry"`
}

var avatarColors = map[string]bool{
	"emerald": true, "blue": true, "purple": true, "rose": true,
	"amber": true, "cyan": true, "indigo": true, "slate": true,
}

// ApplyProfileUpdate applies the non-nil fields of in onto u, with validation.
func ApplyProfileUpdate(u *User, in ProfileUpdate) error {
	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		if n == "" {
			return errors.New("name is required")
		}
		u.Name = n
	}
	if in.Phone != nil {
		p := strings.TrimSpace(*in.Phone)
		u.Phone = &p
	}
	if in.City != nil {
		c := strings.TrimSpace(*in.City)
		u.City = &c
	}
	if in.Bio != nil {
		b := strings.TrimSpace(*in.Bio)
		u.Bio = &b
	}
	if in.AvatarColor != nil {
		c := strings.TrimSpace(*in.AvatarColor)
		if !avatarColors[c] {
			c = "blue"
		}
		u.AvatarColor = c
	}
	if in.NotifyUsageLimit != nil {
		u.NotifyUsageLimit = *in.NotifyUsageLimit
	}
	if in.NotifyExpiry != nil {
		u.NotifyExpiry = *in.NotifyExpiry
	}
	return nil
}
```
*Note:* `model.go` already imports `"time"`; ensure `"errors"` and `"strings"` are in the import block.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/auth/model/ -run TestApplyProfileUpdate -v`
Expected: PASS (4 tests).

- [ ] **Step 5: Rewrite `AuthService.UpdateUser` to use the helper**

In `internal/auth/service/service.go`, replace the body of `UpdateUser` (change its signature to take `model.ProfileUpdate`):
```go
func (s *AuthService) UpdateUser(ctx context.Context, userID uuid.UUID, in model.ProfileUpdate) (*model.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := model.ApplyProfileUpdate(user, in); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
```

- [ ] **Step 6: Extend `repo.UpdateUser` to write all columns**

In `internal/auth/repository/repository.go` `UpdateUser`, change the query + args:
```go
	query := `UPDATE users SET name = $2, phone = $3, city = $4, bio = $5,
	          avatar_color = $6, notify_usage_limit = $7, notify_expiry = $8,
	          updated_at = NOW() WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, user.ID, user.Name, user.Phone,
		user.City, user.Bio, user.AvatarColor, user.NotifyUsageLimit, user.NotifyExpiry)
```
(Keep the existing `RowsAffected()==0 → ErrUserNotFound` check.)

- [ ] **Step 7: Extend the `UpdateMe` handler to parse all fields**

In `internal/auth/handler/handler.go` `UpdateMe`, replace the request struct + call:
```go
	var in model.ProfileUpdate
	if err := c.BodyParser(&in); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	user, err := h.svc.UpdateUser(c.Context(), claims.UserID, in)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, sanitizeUser(user))
```

- [ ] **Step 8: Build, vet, test**

Run: `go build ./cmd/... && go vet ./internal/auth/... && go test ./internal/auth/...`
Expected: success; the 4 `ApplyProfileUpdate` tests pass.

- [ ] **Step 9: Commit**

```bash
git add internal/auth/model/model.go internal/auth/model/profile_update_test.go internal/auth/handler/handler.go internal/auth/service/service.go internal/auth/repository/repository.go
git commit -m "feat(auth): persist editable profile fields with validation"
```

---

### Task 4: Surface subscription `expires_at`

**Files:**
- Modify: `internal/auth/service/service.go` (or wherever `GetSubscriptionStatus` builds its result)

**Interfaces:**
- Produces: `expires_at` key in the subscription-status response (consumed by Task 5).

- [ ] **Step 1: Locate the result construction**

Run: `grep -rn "func (s \*AuthService) GetSubscriptionStatus" internal/auth/service/`
Read that function; it builds a map or struct returned to the handler.

- [ ] **Step 2: Add `expires_at`**

Add the subscription's expiry to the returned result. If it returns a `fiber.Map`/map, add `"expires_at": sub.ExpiresAt` (use the subscription record's expiry field — confirm the exact field name from the struct, e.g. `ExpiresAt`). If it returns a typed struct, add an `ExpiresAt time.Time \`json:"expires_at"\`` field and populate it. When there is no active paid subscription (free), leave it zero/null.

- [ ] **Step 3: Build**

Run: `go build ./cmd/auth-service/... && go vet ./internal/auth/...`
Expected: success.

- [ ] **Step 4: Commit**

```bash
git add internal/auth/service/service.go
git commit -m "feat(auth): include expires_at in subscription status"
```

---

### Task 5: Frontend — editable profile, cleanup, super-admin link, expiry

**Files:**
- Modify: `frontend-svelte/src/lib/pages/ProfilePage.svelte`

**Interfaces:**
- Consumes: `getMe` now returns city/bio/avatar_color/notify_* (Task 2); `PUT /auth/me` persists them (Task 3); subscription status has `expires_at` (Task 4). `ApiService.updateProfile(updates)` already PUTs `/auth/me` — no change needed to the API method.

- [ ] **Step 1: Replace `editName` with an `editForm` object**

In `<script>`, replace `let editName = $state("");` with:
```js
let editForm = $state({ name: "", phone: "", city: "", bio: "" });
```
In `onMount`, replace `editName = me.name;` with:
```js
editForm = { name: me.name || "", phone: me.phone || "", city: me.city || "", bio: me.bio || "" };
```
In `cancelEdit`, replace `editName = profile?.name || "";` with:
```js
editForm = { name: profile?.name || "", phone: profile?.phone || "", city: profile?.city || "", bio: profile?.bio || "" };
```

- [ ] **Step 2: Update `saveProfile` to send all editable fields**

Replace the body's validation + call:
```js
async function saveProfile() {
    if (!editForm.name.trim()) return;
    savingProfile = true;
    profileMsg = { type: "", text: "" };
    try {
        const updated = await ApiService.updateProfile({
            name: editForm.name.trim(),
            phone: editForm.phone.trim(),
            city: editForm.city.trim(),
            bio: editForm.bio.trim(),
        });
        profile = updated;
        const stored = JSON.parse(localStorage.getItem("user") || "{}");
        stored.name = updated.name;
        localStorage.setItem("user", JSON.stringify(stored));
        profileMsg = { type: "success", text: "Profil berhasil diperbarui!" };
        editing = false;
    } catch (e) {
        profileMsg = { type: "error", text: e.message };
    } finally {
        savingProfile = false;
    }
}
```
And update the Save button `disabled={savingProfile || !editForm.name.trim()}` (was `!editName.trim()`).

- [ ] **Step 3: Make name/phone/city/bio editable in the Profil tab**

In the Profil tab's `fields-grid`:
- **Nama:** change the input `bind:value={editName}` → `bind:value={editForm.name}`.
- **Telepon (No. Telepon):** wrap in an editing branch:
```svelte
{#if editing}
    <input class="field-input" type="tel" bind:value={editForm.phone} placeholder="cth. 0812-3456-7890" />
{:else}
    <div class="field-view"><Phone size={16} class="field-ic" />{profile.phone || "—"}</div>
{/if}
```
- **Kota (City):**
```svelte
{#if editing}
    <input class="field-input" bind:value={editForm.city} placeholder="cth. Bandung" />
{:else}
    <div class="field-view"><MapPin size={16} class="field-ic" />{profile.city || "—"}</div>
{/if}
```
- **Bio (field-full):**
```svelte
{#if editing}
    <input class="field-input" bind:value={editForm.bio} placeholder="Ceritakan sedikit tentang Anda" />
{:else}
    <div class="field-bio">{profile.bio || "—"}</div>
{/if}
```
Leave **Email** and **Jabatan** as read-only `field-view` (no editing branch).

- [ ] **Step 4: Remove the fake camera button + fix super-admin link**

Delete the `<button class="summary-avatar-cam" ...>…</button>` block (the "Ubah foto" camera). Change the super-admin link `href="/#/super-admin"` to `href="/super-admin"`.

- [ ] **Step 5: Remove the two mis-wired toggle rows**

Delete the **Keamanan** tab's "Notifikasi Login Mencurigakan" `setting-row` (the one whose toggle is `onclick={() => (notifyExpiry = !notifyExpiry)}` under "Keamanan Akun"), and the **Notifikasi** tab's "Email" channel `setting-row` (also bound to `notifyExpiry`). Keep the two genuine preference toggles in the Notifikasi tab (Peringatan batas kuota → `notifyUsageLimit`, Peringatan masa berlaku → `notifyExpiry`) and the Mode Gelap rows.

- [ ] **Step 6: Show the subscription expiry in the plan block**

In the `plan-block` (after `plan-block-desc`), add, when an expiry exists:
```svelte
{#if subscription?.expires_at && subscription?.status !== "expired"}
    <div class="plan-expiry">Berlaku hingga {formatDate(subscription.expires_at)}</div>
{:else if subscription?.status === "expired"}
    <div class="plan-expiry plan-expiry-danger">Langganan kedaluwarsa</div>
{/if}
```
Add styles:
```css
.plan-expiry { font-size: 12px; color: var(--c-muted); font-weight: 600; }
.plan-expiry-danger { color: var(--c-danger); }
```

- [ ] **Step 7: Check + build**

Run (in `frontend-svelte/`): `npm run check` (0 errors) then `npm run build` (success).

- [ ] **Step 8: Commit**

```bash
git add frontend-svelte/src/lib/pages/ProfilePage.svelte
git commit -m "feat(profile): editable fields, super-admin link, subscription expiry, cleanup"
```

---

### Task 6: Create the super-admin account (deploy-time procedure)

**Files:** none (runtime data operation; run after the auth-service is deployed with migration 015).

**Interfaces:** Consumes the deployed `jamaah_auth` schema.

- [ ] **Step 1: Confirm the auth password hashing**

Run: `grep -rn "bcrypt" internal/auth/service/ | head`
Confirm the cost used at registration (e.g. `bcrypt.DefaultCost`). Use the SAME cost when hashing the super-admin password.

- [ ] **Step 2: Generate a strong password + bcrypt hash**

Use a tiny Go program (Go is on the server; bcrypt is already a dependency). Create `/tmp/hash.go` on the server:
```go
package main
import ( "fmt"; "golang.org/x/crypto/bcrypt" )
func main() {
    pw := "REPLACE_WITH_GENERATED_PASSWORD" // a 20+ char random password
    h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
    fmt.Println(string(h))
}
```
Run it from the repo dir (for module resolution): `cd /data/docker/suluk && go run /tmp/hash.go`. Capture the hash. Pick the password with a secure generator (e.g. `openssl rand -base64 18`).

- [ ] **Step 3: Insert the super-admin user**

Into `jamaah_auth` (via `docker exec -i suluk-postgres psql -U jamaah -d jamaah_auth`):
```sql
INSERT INTO users (id, email, name, password_hash, email_verified, role,
                   is_active, is_super_admin, avatar_color, notify_usage_limit, notify_expiry,
                   created_at, updated_at)
VALUES (gen_random_uuid(), 'superadmin@suluk.site', 'Super Admin', '<BCRYPT_HASH>',
        TRUE, 'owner', TRUE, TRUE, 'blue', TRUE, TRUE, NOW(), NOW());
```
*Note:* confirm the actual NOT-NULL columns on `users` first (`\d users`) and include any other required ones; a super-admin is platform-level (no org membership row needed to reach `/super-admin`, which only checks `is_super_admin`).

- [ ] **Step 4: Verify + clean up**

```sql
SELECT email, is_super_admin, is_active FROM users WHERE email='superadmin@suluk.site';
```
Remove `/tmp/hash.go`. Share the email + generated password with the user (advise changing it on first login). **No commit** (data operation).

---

### Task 7: End-to-end verification

**Files:** none.

- [ ] **Step 1: Backend gates** — `go build ./cmd/... && go vet ./... && go test ./...` → all pass.
- [ ] **Step 2: Frontend gates** — in `frontend-svelte/`: `npm run check` (0 errors) && `npm test` && `npm run build`.
- [ ] **Step 3: Manual (after deploy)** — edit name/phone/city/bio → reload → persists; change avatar color + notification toggles → reload → persists; super-admin link reaches `/super-admin`; subscription expiry shows; log in as the new super-admin account and confirm the dashboard.
- [ ] **Step 4: Commit any fixups** — `git add -A && git commit -m "test(profile): verification fixups"` (only if needed).

---

## Self-Review

- **Spec coverage:** persist all fields (T1 migration, T2 read, T3 write); editable UI (T5); avatar/notify now backed (T2/T3/T5); super-admin link fix (T5) + account (T6); subscription expiry (T4 + T5); cleanup of mis-wired toggles + fake camera (T5); email read-only / role read-only (T5 leaves them as display). All spec sections map to a task.
- **Correctness guard:** the GetUserByID-must-select-new-columns risk is handled in T2 Step 2 (the fetch-mutate-save path can't wipe fields).
- **Type consistency:** `model.ProfileUpdate` + `ApplyProfileUpdate` defined in T3 and used by the service/handler in the same task; `User` fields defined in T2 and used in T3/T5; `avatar_color` palette identical in the spec, the Go validator (T3), and the existing ProfilePage picker.
- **Verify-before-claim notes:** T4 (exact subscription result shape + expiry field name) and T6 (bcrypt cost, NOT-NULL columns on `users`) are flagged to confirm against the codebase/DB rather than assumed.
