# Pricing Tier Restructure — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the acquisition-focused tier structure from `docs/superpowers/specs/2026-06-24-pricing-tier-structure-design.md` (Opsi A: Good-Better-Best, Pro hero, 3-card page, soft-cap scan quota with prepaid top-up, reverse trial).

**Architecture:** Tiers are defined once on the backend (`internal/shared/plan/plan.go`) and mirrored on the frontend (`frontend-svelte/src/lib/config/pricing.js`). Limits are enforced by `internal/jamaah/service/limits.go` reading `/subscription/status`. Billing is prepaid via Pakasir. We re-tune the catalog first (foundation everyone reads), then layer presentation, trial, metering, and reminders.

**Tech Stack:** Go (backend microservices, Fiber), Svelte (frontend), PostgreSQL (migrations), Pakasir (payments).

## Global Constraints

- Tier keys stored verbatim: `gratis`, `starter`, `pro`, `bisnis`, `enterprise` — do NOT rename.
- `Unlimited = -1` sentinel for "no cap" (backend); `UNLIMITED = -1` (frontend).
- Advanced modules (CRM, Keuangan, Kontrak) gate on `IsProOrHigher()` (rank ≥ Pro=2). Multi-cabang/multi-PT gate on `AtLeast(key, "bisnis")`.
- Prices held at current values (149k/299k/599k) — price tuning is a separate workstream, OUT OF SCOPE here.
- `plan.go` and `pricing.js` MUST stay in sync (they are mirror sources of truth).
- Commit messages: NO AI co-author line.

## Phase Roadmap

Each phase is its own PR / working increment. **This document fully details Phase 1**; Phases 2-5 are scoped here and get their own detailed plan when reached.

| Phase | Deliverable | Risk | Files |
|---|---|---|---|
| **1 (this plan)** | Re-tune tier catalog: add per-tier scan quota, fix Starter limits/copy (CRM→Pro+) | Low (data + tests) | `plan.go`, `plan_test.go`, `pricing.js` |
| 2 | Pricing page → 3 cards (Starter/Pro★/Bisnis main; Gratis/Enterprise secondary) | Low (presentation) | `LandingPage.svelte` |
| 3 | Reverse trial: auto Pro-14d on org creation (currently opt-in) | Medium (registration flow) | `internal/auth/service/subscription.go`, org-creation path |
| 4 | Scan metering + prepaid top-up SKU (100 scans = Rp49k via Pakasir) + fair-use cap | High (new billing subsystem) | new scan-usage table/migration, aiocr enforcement, invoice/Pakasir SKU, `ScannerPage.svelte` |
| 5 | Renewal reminders H-7/H-3/H-1 (email default) | Medium (scheduler + notify) | scheduler job, `internal/shared/email` |

---

## Phase 1: Re-tune the tier catalog

New target limits (from spec §3.2):

| | Gratis | Starter | Pro | Bisnis | Enterprise |
|---|---|---|---|---|---|
| MaxJamaah | 50 | **500** (was 250) | ∞ | ∞ | ∞ |
| MaxGroups | 2 | **10** (was 5) | ∞ | ∞ | ∞ |
| MaxUsers | 1 | 3 | 10 | 25 | ∞ |
| **MaxScansPerMonth** (new) | 5 | 100 | ∞ | ∞ | ∞ |

Copy fix: Starter currently advertises "CRM & pembayaran" but `IsProOrHigher` already locks CRM to Pro+. Starter copy must drop the CRM claim and state the 100/bln scan quota.

### Task 1: Backend — add scan quota field + re-tune catalog

**Files:**
- Modify: `internal/shared/plan/plan.go`
- Test: `internal/shared/plan/plan_test.go`

**Interfaces:**
- Produces: `Tier.MaxScansPerMonth int` (json `max_scans_per_month`); values via `Get(key).MaxScansPerMonth`. Consumed later by Phase 4 (scan metering).

- [ ] **Step 1: Write the failing tests**

Append to `internal/shared/plan/plan_test.go`:

```go
func TestScanQuota(t *testing.T) {
	cases := map[string]int{
		Gratis:     5,
		Starter:    100,
		Pro:        Unlimited,
		Bisnis:     Unlimited,
		Enterprise: Unlimited,
	}
	for key, want := range cases {
		if got := Get(key).MaxScansPerMonth; got != want {
			t.Errorf("%s MaxScansPerMonth = %d, want %d", key, got, want)
		}
	}
}

func TestStarterRetunedLimits(t *testing.T) {
	s := Get(Starter)
	if s.MaxJamaah != 500 {
		t.Errorf("Starter MaxJamaah = %d, want 500", s.MaxJamaah)
	}
	if s.MaxGroups != 10 {
		t.Errorf("Starter MaxGroups = %d, want 10", s.MaxGroups)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/shared/plan/ -run 'TestScanQuota|TestStarterRetunedLimits' -v`
Expected: FAIL — compile error `s.MaxScansPerMonth undefined` (field not yet added).

- [ ] **Step 3: Add the field to the Tier struct**

In `internal/shared/plan/plan.go`, add to `type Tier struct` after `MaxUsers`:

```go
	MaxUsers         int      `json:"max_users"`
	MaxScansPerMonth int      `json:"max_scans_per_month"`
```

- [ ] **Step 4: Set the field + re-tune limits in Catalog**

Update each entry in `var Catalog`:

```go
	Gratis: {
		Key: Gratis, Name: "Gratis", Rank: 0,
		MonthlyPrice: 0, AnnualPrice: 0,
		MaxJamaah: 50, MaxGroups: 2, MaxUsers: 1, MaxScansPerMonth: 5,
		Purchasable: false,
		Features:    []string{"Hingga 50 jamaah", "Data jamaah & grup", "Manajemen paket dasar", "1 pengguna"},
	},
	Starter: {
		Key: Starter, Name: "Starter", Rank: 1,
		MonthlyPrice: 149000, AnnualPrice: 1490000,
		MaxJamaah: 500, MaxGroups: 10, MaxUsers: 3, MaxScansPerMonth: 100,
		Purchasable: true,
		Features:    []string{"Hingga 500 jamaah", "Pembayaran & cicilan jamaah", "AI Scanner 100 dok/bulan", "Hingga 3 pengguna", "Laporan dasar"},
	},
	Pro: {
		Key: Pro, Name: "Pro", Rank: 2,
		MonthlyPrice: 299000, AnnualPrice: 2990000,
		MaxJamaah: Unlimited, MaxGroups: Unlimited, MaxUsers: 10, MaxScansPerMonth: Unlimited,
		Purchasable: true,
		Features:    []string{"Jamaah tak terbatas", "Semua modul (CRM, Keuangan, Kontrak)", "AI Scanner tanpa batas", "Hingga 10 pengguna", "Laporan & ekspor lanjutan"},
	},
	Bisnis: {
		Key: Bisnis, Name: "Bisnis", Rank: 3,
		MonthlyPrice: 599000, AnnualPrice: 5990000,
		MaxJamaah: Unlimited, MaxGroups: Unlimited, MaxUsers: 25, MaxScansPerMonth: Unlimited,
		Purchasable: true,
		Features:    []string{"Semua fitur Pro", "Multi-cabang & multi-PT", "Hingga 25 pengguna", "Dukungan prioritas"},
	},
	Enterprise: {
		Key: Enterprise, Name: "Enterprise", Rank: 4,
		MonthlyPrice: 0, AnnualPrice: 0,
		MaxJamaah: Unlimited, MaxGroups: Unlimited, MaxUsers: Unlimited, MaxScansPerMonth: Unlimited,
		Purchasable: false,
		Features:    []string{"Semua fitur Bisnis", "Akses API & integrasi", "Pengguna tak terbatas", "Dukungan prioritas 24/7"},
	},
```

- [ ] **Step 5: Run all plan tests to verify they pass**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" test ./internal/shared/plan/ -v`
Expected: PASS (all, including the two new tests).

- [ ] **Step 6: Verify the whole module still builds**

Run: `go -C "D:\Codding\Project\Suluk-Manajemen-SAAS" build ./...`
Expected: no output (success).

- [ ] **Step 7: Commit**

```bash
git add internal/shared/plan/plan.go internal/shared/plan/plan_test.go
git commit -m "feat(plan): add per-tier scan quota and re-tune Starter limits"
```

### Task 2: Frontend — mirror catalog in pricing.js

**Files:**
- Modify: `frontend-svelte/src/lib/config/pricing.js`

**Interfaces:**
- Consumes: the new limit values from Task 1 (must match exactly).
- Produces: `maxScansPerMonth` on each plan object; consumed later by Phase 2 (page) and Phase 4 (scanner quota UI).

- [ ] **Step 1: Update each plan object to mirror plan.go**

In `frontend-svelte/src/lib/config/pricing.js`, for each entry in `PLANS` add `maxScansPerMonth` and update Starter's limits + features + copy:

- `gratis`: add `maxScansPerMonth: 5,`
- `starter`: set `maxJamaah: 500,` `maxGroups: 10,` add `maxScansPerMonth: 100,`; replace `features` with `['Hingga 500 jamaah', 'Pembayaran & cicilan jamaah', 'AI Scanner 100 dok/bulan', 'Hingga 3 pengguna', 'Laporan dasar']`; update `desc` to `'Untuk travel kecil yang mau mulai rapi.'`
- `pro`: add `maxScansPerMonth: UNLIMITED,`
- `bisnis`: add `maxScansPerMonth: UNLIMITED,`
- `enterprise`: add `maxScansPerMonth: UNLIMITED,`

- [ ] **Step 2: Verify frontend type/lint check passes**

Run: `npm --prefix "D:\Codding\Project\Suluk-Manajemen-SAAS\frontend-svelte" run check`
Expected: 0 errors (svelte-check). (If `node_modules` missing, run `npm --prefix ... install` first.)

- [ ] **Step 3: Manual sync check**

Confirm by eye that every `maxJamaah`, `maxGroups`, `maxUsers`, `maxScansPerMonth` in `pricing.js` matches `plan.go` for all five tiers. They are mirror sources of truth.

- [ ] **Step 4: Commit**

```bash
git add frontend-svelte/src/lib/config/pricing.js
git commit -m "feat(pricing): mirror scan quota and re-tuned Starter limits on frontend"
```

---

## Self-Review (Phase 1)

- **Spec coverage:** §3.2 matrix limits (jamaah/groups/users/scan) → Task 1 + Task 2. Copy fix (Starter CRM) → Task 1 Step 4 + Task 2 Step 1. ✓ Phases 2-5 cover the rest of the spec (presentation, trial, metering, reminders).
- **Placeholder scan:** No TBD/TODO in Phase 1 tasks; all code shown. ✓
- **Type consistency:** `MaxScansPerMonth` (Go) / `maxScansPerMonth` (JS), `Unlimited`/`UNLIMITED = -1` used consistently. `Get(key)` returns `Tier`. ✓
