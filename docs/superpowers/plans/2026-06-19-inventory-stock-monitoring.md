# Inventory Stock Monitoring + Auto-Deduct — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add stock-level monitoring to the Inventaris module (add items, restock, adjust, history, low/negative alerts, per-package departure kit) and auto-deduct stock when a group departs.

**Architecture:** Build out the dormant `inventory_items` stock side in `inventory-service` (model → repository → service → handler), add `package_kit_items` and `stock_movements` tables, and subscribe inventory-service to the existing `group.departed` NATS event to deduct `qty_per_jamaah × member_count` idempotently. Frontend: the Inventaris page gains a second tab (**Stok**) alongside the unchanged Distribusi/QR workflow.

**Tech Stack:** Go (Fiber, pgx v5, NATS JetStream via `internal/shared/events`), Postgres, SvelteKit 5 + Tailwind.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-06-19-inventory-stock-monitoring-design.md`.
- inventory-service DB is `jamaah_inventory`; migrations live in `migrations/inventory/`.
- All queries are **org-scoped** (`org_id` from JWT claims for HTTP; from `env.OrgID` for the event consumer).
- New HTTP routes go under `/api/v1/inventory` with `authMW, sharedMW.RequireStaff` (gateway already proxies `/inventory/*` — **no gateway change**).
- Negative stock is **allowed**; never block. Departure deduction must be **idempotent** (re-delivered event = no-op).
- No AI attribution in commits (per repo AGENTS.md). Commit only at the end of each task.
- Gates: `go build ./cmd/...`, `go vet ./...`, `go test ./...`, and in `frontend-svelte/`: `npm run check`, `npm test`.
- Frontend UI uses the established Tailwind idiom: `rounded-xl bg-white p-4 shadow-sm ring-1 ring-slate-200/60`, inputs `rounded-xl border-slate-200 … focus:ring-primary-100`, `font-serif … text-slate-800`, `PageHeader`, `FilterTabs`, `EmptyState`, `StatCard`.

---

### Task 1: DB migration — kit + movement tables

**Files:**
- Create: `migrations/inventory/003_stock_monitoring.up.sql`
- Create: `migrations/inventory/003_stock_monitoring.down.sql`

**Interfaces:**
- Produces: tables `package_kit_items`, `stock_movements`; reused `inventory_items`.

- [ ] **Step 1: Write the up migration**

`migrations/inventory/003_stock_monitoring.up.sql`:
```sql
-- Phase 6: stock monitoring + auto-deduct on departure.

CREATE TABLE IF NOT EXISTS package_kit_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL,
    package_id     UUID NOT NULL,
    item_id        UUID NOT NULL,
    qty_per_jamaah INTEGER NOT NULL CHECK (qty_per_jamaah > 0),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, package_id, item_id)
);
CREATE INDEX IF NOT EXISTS idx_package_kit_items_pkg ON package_kit_items(org_id, package_id);

CREATE TABLE IF NOT EXISTS stock_movements (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL,
    item_id     UUID NOT NULL,
    delta       INTEGER NOT NULL,
    reason      TEXT NOT NULL,         -- initial | restock | adjustment | departure
    note        TEXT NOT NULL DEFAULT '',
    group_id    UUID,
    package_id  UUID,
    created_by  UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_stock_movements_item ON stock_movements(org_id, item_id, created_at DESC);

-- One departure deduction per (group, item) ⇒ a re-delivered group.departed is a no-op.
CREATE UNIQUE INDEX IF NOT EXISTS uniq_departure_movement
    ON stock_movements(group_id, item_id) WHERE reason = 'departure';
```

- [ ] **Step 2: Write the down migration**

`migrations/inventory/003_stock_monitoring.down.sql`:
```sql
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS package_kit_items;
```

- [ ] **Step 3: Apply and verify**

Run: `go run ./cmd/migration/main.go -service inventory -direction up`
Expected: applies `003` with no error. (Then `-direction down` once to confirm it reverts, then `up` again.)

- [ ] **Step 4: Commit**

```bash
git add migrations/inventory/003_stock_monitoring.up.sql migrations/inventory/003_stock_monitoring.down.sql
git commit -m "feat(inventory): migration for stock kits + movements"
```

---

### Task 2: Models — structs & DTOs

**Files:**
- Modify: `internal/inventory/model/model.go` (append)

**Interfaces:**
- Produces: `StockItem`, `StockMovement`, `PackageKitItem`, `Deduction`, `DepartedPayload`, and request DTOs used by every later backend task.

- [ ] **Step 1: Append the structs**

Append to `internal/inventory/model/model.go`:
```go
// --- Stock monitoring (Phase 6) ---

type StockItem struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Unit      string    `json:"unit"`
	Stock     int       `json:"stock"`
	MinStock  int       `json:"min_stock"`
	InKit     bool      `json:"in_kit"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StockMovement struct {
	ID        string    `json:"id"`
	ItemID    string    `json:"item_id"`
	Delta     int       `json:"delta"`
	Reason    string    `json:"reason"`
	Note      string    `json:"note"`
	GroupID   *string   `json:"group_id,omitempty"`
	PackageID *string   `json:"package_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type PackageKitItem struct {
	ItemID       string `json:"item_id"`
	ItemName     string `json:"item_name"`
	Unit         string `json:"unit"`
	QtyPerJamaah int    `json:"qty_per_jamaah"`
}

// Deduction is one computed (item, quantity) to subtract — pure logic output.
type Deduction struct {
	ItemID string
	Qty    int
}

// DepartedPayload is the group.departed event payload.
type DepartedPayload struct {
	GroupID     string `json:"group_id"`
	PackageID   string `json:"package_id"`
	MemberCount int    `json:"member_count"`
	Status      string `json:"status"`
}

type CreateItemRequest struct {
	Name         string `json:"name"`
	Category     string `json:"category"`
	Unit         string `json:"unit"`
	MinStock     int    `json:"min_stock"`
	InitialStock int    `json:"initial_stock"`
}

type UpdateItemRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Unit     string `json:"unit"`
	MinStock int    `json:"min_stock"`
}

type RestockRequest struct {
	Qty  int    `json:"qty"`
	Note string `json:"note"`
}

type AdjustRequest struct {
	Delta int    `json:"delta"`
	Note  string `json:"note"`
}

type KitLine struct {
	ItemID       string `json:"item_id"`
	QtyPerJamaah int    `json:"qty_per_jamaah"`
}

type SetKitRequest struct {
	Items []KitLine `json:"items"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/inventory/...`
Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add internal/inventory/model/model.go
git commit -m "feat(inventory): stock monitoring models"
```

---

### Task 3: Pure deduction logic (TDD)

**Files:**
- Create: `internal/inventory/service/deduct.go`
- Test: `internal/inventory/service/deduct_test.go`

**Interfaces:**
- Produces: `func ComputeDeductions(kit []model.PackageKitItem, memberCount int) []model.Deduction` — used by Task 5.

- [ ] **Step 1: Write the failing test**

`internal/inventory/service/deduct_test.go`:
```go
package service

import (
	"testing"

	"github.com/jamaah-in/v2/internal/inventory/model"
)

func TestComputeDeductionsMultipliesByHeadcount(t *testing.T) {
	kit := []model.PackageKitItem{
		{ItemID: "koper", QtyPerJamaah: 1},
		{ItemID: "ihram", QtyPerJamaah: 2},
	}
	got := ComputeDeductions(kit, 30)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ItemID != "koper" || got[0].Qty != 30 {
		t.Fatalf("koper = %+v, want qty 30", got[0])
	}
	if got[1].ItemID != "ihram" || got[1].Qty != 60 {
		t.Fatalf("ihram = %+v, want qty 60", got[1])
	}
}

func TestComputeDeductionsZeroMembersIsEmpty(t *testing.T) {
	kit := []model.PackageKitItem{{ItemID: "koper", QtyPerJamaah: 1}}
	if got := ComputeDeductions(kit, 0); len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestComputeDeductionsEmptyKitIsEmpty(t *testing.T) {
	if got := ComputeDeductions(nil, 30); len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/inventory/service/ -run TestComputeDeductions -v`
Expected: FAIL — `undefined: ComputeDeductions`.

- [ ] **Step 3: Write minimal implementation**

`internal/inventory/service/deduct.go`:
```go
package service

import "github.com/jamaah-in/v2/internal/inventory/model"

// ComputeDeductions returns the (item, qty) lines to subtract when a group of
// memberCount jamaah departs. Pure: no DB. Empty when there is nothing to do.
func ComputeDeductions(kit []model.PackageKitItem, memberCount int) []model.Deduction {
	if memberCount < 1 || len(kit) == 0 {
		return nil
	}
	out := make([]model.Deduction, 0, len(kit))
	for _, k := range kit {
		qty := k.QtyPerJamaah * memberCount
		if qty <= 0 {
			continue
		}
		out = append(out, model.Deduction{ItemID: k.ItemID, Qty: qty})
	}
	return out
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/inventory/service/ -run TestComputeDeductions -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/inventory/service/deduct.go internal/inventory/service/deduct_test.go
git commit -m "feat(inventory): pure departure-deduction logic + tests"
```

---

### Task 4: Repository — stock items, kits, movements, idempotent deduction

**Files:**
- Create: `internal/inventory/repository/stock.go`

**Interfaces:**
- Consumes: models from Task 2.
- Produces (used by Task 5):
  - `ListStockItems(ctx, orgID string) ([]model.StockItem, error)`
  - `CreateStockItem(ctx, orgID string, req model.CreateItemRequest) (model.StockItem, error)`
  - `UpdateStockItem(ctx, orgID, itemID string, req model.UpdateItemRequest) error`
  - `RestockItem(ctx, orgID, itemID string, qty int, note, userID string) error`
  - `AdjustItem(ctx, orgID, itemID string, delta int, note, userID string) error`
  - `ListMovements(ctx, orgID, itemID string, limit int) ([]model.StockMovement, error)`
  - `DeleteStockItem(ctx, orgID, itemID string) error` (returns `ErrItemInKit`)
  - `GetPackageKit(ctx, orgID, packageID string) ([]model.PackageKitItem, error)`
  - `SetPackageKit(ctx, orgID, packageID string, items []model.KitLine) error`
  - `ApplyDepartureDeduction(ctx, orgID, groupID, packageID string, deductions []model.Deduction) error`
  - `var ErrItemInKit error`

- [ ] **Step 1: Write the file**

`internal/inventory/repository/stock.go`:
```go
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/jamaah-in/v2/internal/inventory/model"
)

// ErrItemInKit blocks deleting an item that a package kit still references.
var ErrItemInKit = errors.New("item is used in a package kit")

func nullStr(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

func (r *InventoryRepo) ListStockItems(ctx context.Context, orgID string) ([]model.StockItem, error) {
	const q = `
		SELECT i.id, i.org_id, i.name, i.category, COALESCE(i.unit,'pcs'), i.stock, i.min_stock,
		       EXISTS(SELECT 1 FROM package_kit_items k WHERE k.item_id = i.id) AS in_kit,
		       i.created_at, i.updated_at
		FROM inventory_items i
		WHERE i.org_id = $1
		ORDER BY i.name`
	rows, err := r.pool.Query(ctx, q, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.StockItem
	for rows.Next() {
		var it model.StockItem
		if err := rows.Scan(&it.ID, &it.OrgID, &it.Name, &it.Category, &it.Unit,
			&it.Stock, &it.MinStock, &it.InKit, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *InventoryRepo) CreateStockItem(ctx context.Context, orgID string, req model.CreateItemRequest) (model.StockItem, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.StockItem{}, err
	}
	defer tx.Rollback(ctx)

	var it model.StockItem
	const ins = `
		INSERT INTO inventory_items (org_id, name, category, unit, stock, min_stock)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, org_id, name, category, COALESCE(unit,'pcs'), stock, min_stock, created_at, updated_at`
	unit := req.Unit
	if unit == "" {
		unit = "pcs"
	}
	cat := req.Category
	if cat == "" {
		cat = "perlengkapan"
	}
	if err := tx.QueryRow(ctx, ins, orgID, req.Name, cat, unit, req.InitialStock, req.MinStock).
		Scan(&it.ID, &it.OrgID, &it.Name, &it.Category, &it.Unit, &it.Stock, &it.MinStock, &it.CreatedAt, &it.UpdatedAt); err != nil {
		return model.StockItem{}, err
	}
	if req.InitialStock > 0 {
		if _, err := tx.Exec(ctx,
			`INSERT INTO stock_movements (org_id, item_id, delta, reason, note) VALUES ($1,$2,$3,'initial','')`,
			orgID, it.ID, req.InitialStock); err != nil {
			return model.StockItem{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return model.StockItem{}, err
	}
	return it, nil
}

func (r *InventoryRepo) UpdateStockItem(ctx context.Context, orgID, itemID string, req model.UpdateItemRequest) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE inventory_items SET name=$3, category=$4, unit=$5, min_stock=$6, updated_at=NOW()
		 WHERE org_id=$1 AND id=$2`,
		orgID, itemID, req.Name, req.Category, req.Unit, req.MinStock)
	return err
}

// changeStock applies a signed delta and records a movement in one transaction.
func (r *InventoryRepo) changeStock(ctx context.Context, orgID, itemID string, delta int, reason, note, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx,
		`UPDATE inventory_items SET stock = stock + $3, updated_at=NOW() WHERE org_id=$1 AND id=$2`,
		orgID, itemID, delta)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	var by any
	if userID != "" {
		by = userID
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO stock_movements (org_id, item_id, delta, reason, note, created_by) VALUES ($1,$2,$3,$4,$5,$6)`,
		orgID, itemID, delta, reason, note, by); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *InventoryRepo) RestockItem(ctx context.Context, orgID, itemID string, qty int, note, userID string) error {
	return r.changeStock(ctx, orgID, itemID, qty, "restock", note, userID)
}

func (r *InventoryRepo) AdjustItem(ctx context.Context, orgID, itemID string, delta int, note, userID string) error {
	return r.changeStock(ctx, orgID, itemID, delta, "adjustment", note, userID)
}

func (r *InventoryRepo) ListMovements(ctx context.Context, orgID, itemID string, limit int) ([]model.StockMovement, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `
		SELECT id, item_id, delta, reason, note, group_id, package_id, created_at
		FROM stock_movements WHERE org_id=$1 AND item_id=$2
		ORDER BY created_at DESC LIMIT $3`
	rows, err := r.pool.Query(ctx, q, orgID, itemID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.StockMovement
	for rows.Next() {
		var m model.StockMovement
		if err := rows.Scan(&m.ID, &m.ItemID, &m.Delta, &m.Reason, &m.Note, &m.GroupID, &m.PackageID, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *InventoryRepo) DeleteStockItem(ctx context.Context, orgID, itemID string) error {
	var inKit bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM package_kit_items WHERE org_id=$1 AND item_id=$2)`,
		orgID, itemID).Scan(&inKit); err != nil {
		return err
	}
	if inKit {
		return ErrItemInKit
	}
	_, err := r.pool.Exec(ctx, `DELETE FROM inventory_items WHERE org_id=$1 AND id=$2`, orgID, itemID)
	return err
}

func (r *InventoryRepo) GetPackageKit(ctx context.Context, orgID, packageID string) ([]model.PackageKitItem, error) {
	const q = `
		SELECT k.item_id, i.name, COALESCE(i.unit,'pcs'), k.qty_per_jamaah
		FROM package_kit_items k JOIN inventory_items i ON i.id = k.item_id
		WHERE k.org_id=$1 AND k.package_id=$2
		ORDER BY i.name`
	rows, err := r.pool.Query(ctx, q, orgID, packageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.PackageKitItem
	for rows.Next() {
		var k model.PackageKitItem
		if err := rows.Scan(&k.ItemID, &k.ItemName, &k.Unit, &k.QtyPerJamaah); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// SetPackageKit replaces the package's kit with the given lines (qty>0 only).
func (r *InventoryRepo) SetPackageKit(ctx context.Context, orgID, packageID string, items []model.KitLine) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM package_kit_items WHERE org_id=$1 AND package_id=$2`, orgID, packageID); err != nil {
		return err
	}
	for _, ln := range items {
		if ln.QtyPerJamaah <= 0 {
			continue
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO package_kit_items (org_id, package_id, item_id, qty_per_jamaah) VALUES ($1,$2,$3,$4)`,
			orgID, packageID, ln.ItemID, ln.QtyPerJamaah); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ApplyDepartureDeduction subtracts each deduction once per (group,item). The
// unique partial index makes a re-delivered group.departed a no-op: the movement
// INSERT conflicts (DO NOTHING), and the stock UPDATE runs only when a row was
// actually inserted. Negative stock is allowed.
func (r *InventoryRepo) ApplyDepartureDeduction(ctx context.Context, orgID, groupID, packageID string, deductions []model.Deduction) error {
	if len(deductions) == 0 {
		return nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, d := range deductions {
		var movementID string
		err := tx.QueryRow(ctx,
			`INSERT INTO stock_movements (org_id, item_id, delta, reason, group_id, package_id)
			 VALUES ($1,$2,$3,'departure',$4,$5)
			 ON CONFLICT (group_id, item_id) WHERE reason='departure' DO NOTHING
			 RETURNING id`,
			orgID, d.ItemID, -d.Qty, groupID, packageID).Scan(&movementID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue // already deducted for this (group,item)
		}
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`UPDATE inventory_items SET stock = stock - $3, updated_at=NOW() WHERE org_id=$1 AND id=$2`,
			orgID, d.ItemID, d.Qty); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/inventory/...`
Expected: success. (`nullStr` is a helper kept for symmetry; if `go vet` flags it as unused, delete it.)

- [ ] **Step 3: Commit**

```bash
git add internal/inventory/repository/stock.go
git commit -m "feat(inventory): stock/kit/movement repository + idempotent deduction"
```

---

### Task 5: Service — stock methods + departure consumer logic

**Files:**
- Create: `internal/inventory/service/stock.go`

**Interfaces:**
- Consumes: repo methods (Task 4), `ComputeDeductions` (Task 3), `events.Envelope`/`events.Bus`.
- Produces (used by Tasks 6 & 7):
  - thin pass-throughs: `ListStockItems`, `CreateStockItem`, `UpdateStockItem`, `RestockItem`, `AdjustItem`, `ListMovements`, `DeleteStockItem`, `GetPackageKit`, `SetPackageKit`
  - `DeductForDeparture(ctx, env *events.Envelope) error`
  - `StartConsumer(ctx, bus *events.Bus) error`

- [ ] **Step 1: Write the file**

`internal/inventory/service/stock.go`:
```go
package service

import (
	"context"
	"encoding/json"

	"github.com/jamaah-in/v2/internal/inventory/model"
	"github.com/jamaah-in/v2/internal/shared/events"
)

func (s *InventoryService) ListStockItems(ctx context.Context, orgID string) ([]model.StockItem, error) {
	return s.repo.ListStockItems(ctx, orgID)
}
func (s *InventoryService) CreateStockItem(ctx context.Context, orgID string, req model.CreateItemRequest) (model.StockItem, error) {
	return s.repo.CreateStockItem(ctx, orgID, req)
}
func (s *InventoryService) UpdateStockItem(ctx context.Context, orgID, itemID string, req model.UpdateItemRequest) error {
	return s.repo.UpdateStockItem(ctx, orgID, itemID, req)
}
func (s *InventoryService) RestockItem(ctx context.Context, orgID, itemID string, qty int, note, userID string) error {
	return s.repo.RestockItem(ctx, orgID, itemID, qty, note, userID)
}
func (s *InventoryService) AdjustItem(ctx context.Context, orgID, itemID string, delta int, note, userID string) error {
	return s.repo.AdjustItem(ctx, orgID, itemID, delta, note, userID)
}
func (s *InventoryService) ListMovements(ctx context.Context, orgID, itemID string, limit int) ([]model.StockMovement, error) {
	return s.repo.ListMovements(ctx, orgID, itemID, limit)
}
func (s *InventoryService) DeleteStockItem(ctx context.Context, orgID, itemID string) error {
	return s.repo.DeleteStockItem(ctx, orgID, itemID)
}
func (s *InventoryService) GetPackageKit(ctx context.Context, orgID, packageID string) ([]model.PackageKitItem, error) {
	return s.repo.GetPackageKit(ctx, orgID, packageID)
}
func (s *InventoryService) SetPackageKit(ctx context.Context, orgID, packageID string, items []model.KitLine) error {
	return s.repo.SetPackageKit(ctx, orgID, packageID, items)
}

// DeductForDeparture deducts the package kit × member_count when a group departs.
// Idempotent (repo-level). No kit / no members ⇒ no-op.
func (s *InventoryService) DeductForDeparture(ctx context.Context, env *events.Envelope) error {
	var p model.DepartedPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return err
	}
	if p.PackageID == "" || p.MemberCount < 1 || env.OrgID == "" {
		return nil
	}
	kit, err := s.repo.GetPackageKit(ctx, env.OrgID, p.PackageID)
	if err != nil {
		return err
	}
	deductions := ComputeDeductions(kit, p.MemberCount)
	if len(deductions) == 0 {
		return nil
	}
	return s.repo.ApplyDepartureDeduction(ctx, env.OrgID, p.GroupID, p.PackageID, deductions)
}

// StartConsumer subscribes inventory-service to the bus and deducts on departure.
// Non-group.departed events are ACKed and ignored.
func (s *InventoryService) StartConsumer(ctx context.Context, bus *events.Bus) error {
	_, err := bus.Subscribe(ctx, "inventory-deduct", func(ctx context.Context, env *events.Envelope) error {
		if env.EventType != events.EventGroupDeparted {
			return nil
		}
		return s.DeductForDeparture(ctx, env)
	})
	return err
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/inventory/... && go test ./internal/inventory/service/ -run TestComputeDeductions`
Expected: success, tests still PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/inventory/service/stock.go
git commit -m "feat(inventory): stock service methods + departure consumer logic"
```

---

### Task 6: HTTP handlers

**Files:**
- Create: `internal/inventory/handler/stock.go`

**Interfaces:**
- Consumes: service methods (Task 5), `middleware.GetClaims`, `response` helpers, `repository.ErrItemInKit`.
- Produces (used by Task 7 route registration): handler methods `ListItems, CreateItem, UpdateItem, RestockItem, AdjustItem, ListItemMovements, DeleteItem, GetKit, SetKit`.

- [ ] **Step 1: Write the file**

`internal/inventory/handler/stock.go`:
```go
package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/inventory/model"
	"github.com/jamaah-in/v2/internal/inventory/repository"
	"github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/response"
)

func (h *InventoryHandler) ListItems(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	items, err := h.svc.ListStockItems(c.Context(), claims.OrgID.String())
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"items": items})
}

func (h *InventoryHandler) CreateItem(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	var req model.CreateItemRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	item, err := h.svc.CreateStockItem(c.Context(), claims.OrgID.String(), req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, item)
}

func (h *InventoryHandler) UpdateItem(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return response.BadRequest(c, "invalid id")
	}
	var req model.UpdateItemRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.UpdateStockItem(c.Context(), claims.OrgID.String(), id, req); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"updated": true})
}

func (h *InventoryHandler) RestockItem(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id := c.Params("id")
	var req model.RestockRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Qty <= 0 {
		return response.BadRequest(c, "qty must be positive")
	}
	if err := h.svc.RestockItem(c.Context(), claims.OrgID.String(), id, req.Qty, req.Note, claims.UserID.String()); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"ok": true})
}

func (h *InventoryHandler) AdjustItem(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id := c.Params("id")
	var req model.AdjustRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Delta == 0 {
		return response.BadRequest(c, "delta must be non-zero")
	}
	if err := h.svc.AdjustItem(c.Context(), claims.OrgID.String(), id, req.Delta, req.Note, claims.UserID.String()); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"ok": true})
}

func (h *InventoryHandler) ListItemMovements(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id := c.Params("id")
	moves, err := h.svc.ListMovements(c.Context(), claims.OrgID.String(), id, c.QueryInt("limit", 50))
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"movements": moves})
}

func (h *InventoryHandler) DeleteItem(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id := c.Params("id")
	if err := h.svc.DeleteStockItem(c.Context(), claims.OrgID.String(), id); err != nil {
		if errors.Is(err, repository.ErrItemInKit) {
			return response.Conflict(c, "item dipakai di kit paket; hapus dari kit dulu")
		}
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"deleted": true})
}

func (h *InventoryHandler) GetKit(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	pkg := c.Params("packageId")
	if _, err := uuid.Parse(pkg); err != nil {
		return response.BadRequest(c, "invalid package_id")
	}
	kit, err := h.svc.GetPackageKit(c.Context(), claims.OrgID.String(), pkg)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"items": kit})
}

func (h *InventoryHandler) SetKit(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	pkg := c.Params("packageId")
	if _, err := uuid.Parse(pkg); err != nil {
		return response.BadRequest(c, "invalid package_id")
	}
	var req model.SetKitRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.SetPackageKit(c.Context(), claims.OrgID.String(), pkg, req.Items); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"saved": len(req.Items)})
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/inventory/...`
Expected: success.
*Note:* confirm `response.Conflict` and `claims.UserID` exist (grep `internal/shared/response` and the claims struct). If `response.Conflict` is absent, use `return c.Status(409).JSON(fiber.Map{"error": ...})`. If the claims field is named differently (e.g. `claims.Sub`), use that for `created_by`.

- [ ] **Step 3: Commit**

```bash
git add internal/inventory/handler/stock.go
git commit -m "feat(inventory): stock + kit HTTP handlers"
```

---

### Task 7: Wire routes + event consumer in main.go

**Files:**
- Modify: `cmd/inventory-service/main.go`

**Interfaces:**
- Consumes: handler methods (Task 6), `svc.StartConsumer` (Task 5), `events.Connect` / `cfg.NATS.Addr`.

- [ ] **Step 1: Add the events import**

In `cmd/inventory-service/main.go`, add to the import block:
```go
	sharedEvents "github.com/jamaah-in/v2/internal/shared/events"
```

- [ ] **Step 2: Connect the bus + start the consumer**

After `inventoryHandler := handler.NewInventoryHandler(inventorySvc)` (around line 65), insert:
```go
	// Auto-deduct stock when a group departs (subscribe to group.departed).
	var bus *sharedEvents.Bus
	if b, berr := sharedEvents.Connect(cfg.NATS.Addr, logger); berr != nil {
		logger.Errorf("event bus unavailable (auto-deduct disabled): %v", berr)
	} else {
		bus = b
		defer bus.Close()
		if serr := inventorySvc.StartConsumer(ctx, bus); serr != nil {
			logger.Errorf("start inventory consumer: %v", serr)
		} else {
			logger.Info("inventory departure-deduct consumer started")
		}
	}
```

- [ ] **Step 3: Register the new routes**

After the existing `api.Get("/checkpoints/:packageId", ...)` line (~line 92), add:
```go
	// Stock monitoring (Phase 6)
	api.Get("/items", inventoryHandler.ListItems)
	api.Post("/items", inventoryHandler.CreateItem)
	api.Patch("/items/:id", inventoryHandler.UpdateItem)
	api.Delete("/items/:id", inventoryHandler.DeleteItem)
	api.Post("/items/:id/restock", inventoryHandler.RestockItem)
	api.Post("/items/:id/adjust", inventoryHandler.AdjustItem)
	api.Get("/items/:id/movements", inventoryHandler.ListItemMovements)
	api.Get("/kits/:packageId", inventoryHandler.GetKit)
	api.Put("/kits/:packageId", inventoryHandler.SetKit)
```

- [ ] **Step 4: Build & vet the whole backend**

Run: `go build ./cmd/... && go vet ./internal/inventory/... && go test ./internal/inventory/...`
Expected: success; deduction tests PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/inventory-service/main.go
git commit -m "feat(inventory): wire stock routes + departure-deduct consumer"
```

---

### Task 8: Frontend API methods

**Files:**
- Modify: `frontend-svelte/src/lib/services/apiDomains/groupOpsApi.js`

**Interfaces:**
- Produces (used by Task 9): `listStockItems()`, `createStockItem(body)`, `updateStockItem(id, body)`, `deleteStockItem(id)`, `restockItem(id, body)`, `adjustItem(id, body)`, `getItemMovements(id)`, `getPackageKit(packageId)`, `setPackageKit(packageId, items)`.

- [ ] **Step 1: Add the methods**

Inside the object returned by `groupOpsApi.js` (next to `getInventoryForecast`), add — matching the existing `apiFetch` style:
```js
        async listStockItems() {
            const response = await apiFetch(`${API_URL}/inventory/items`, { headers: authHeaders() });
            return handle(response);
        },
        async createStockItem(body) {
            const response = await apiFetch(`${API_URL}/inventory/items`, {
                method: 'POST', headers: authHeaders(), body: JSON.stringify(body),
            });
            return handle(response);
        },
        async updateStockItem(id, body) {
            const response = await apiFetch(`${API_URL}/inventory/items/${id}`, {
                method: 'PATCH', headers: authHeaders(), body: JSON.stringify(body),
            });
            return handle(response);
        },
        async deleteStockItem(id) {
            const response = await apiFetch(`${API_URL}/inventory/items/${id}`, {
                method: 'DELETE', headers: authHeaders(),
            });
            return handle(response);
        },
        async restockItem(id, body) {
            const response = await apiFetch(`${API_URL}/inventory/items/${id}/restock`, {
                method: 'POST', headers: authHeaders(), body: JSON.stringify(body),
            });
            return handle(response);
        },
        async adjustItem(id, body) {
            const response = await apiFetch(`${API_URL}/inventory/items/${id}/adjust`, {
                method: 'POST', headers: authHeaders(), body: JSON.stringify(body),
            });
            return handle(response);
        },
        async getItemMovements(id) {
            const response = await apiFetch(`${API_URL}/inventory/items/${id}/movements`, { headers: authHeaders() });
            return handle(response);
        },
        async getPackageKit(packageId) {
            const response = await apiFetch(`${API_URL}/inventory/kits/${packageId}`, { headers: authHeaders() });
            return handle(response);
        },
        async setPackageKit(packageId, items) {
            const response = await apiFetch(`${API_URL}/inventory/kits/${packageId}`, {
                method: 'PUT', headers: authHeaders(), body: JSON.stringify({ items }),
            });
            return handle(response);
        },
```
*Note:* match the file's actual helpers — use the same response-unwrap (`handle(...)` / `response.json()`) and header helper (`authHeaders()` or inline) that `getInventoryForecast` uses. Read lines 120–185 of the file first and copy that exact idiom.

- [ ] **Step 2: Verify**

Run (in `frontend-svelte/`): `npm run check`
Expected: 0 errors.

- [ ] **Step 3: Commit**

```bash
git add frontend-svelte/src/lib/services/apiDomains/groupOpsApi.js
git commit -m "feat(inventory): frontend stock API methods"
```

---

### Task 9: Frontend Stok tab

**Files:**
- Create: `frontend-svelte/src/lib/pages/inventory/StockTab.svelte` (the Stok view: summary, alert, item table, add/restock/adjust, history drawer, kit config)
- Modify: `frontend-svelte/src/lib/pages/InventoryPage.svelte` (add `FilterTabs`; wrap existing content as the **Distribusi** tab; render `StockTab` as the **Stok** tab)

**Interfaces:**
- Consumes: API methods (Task 8), `groups` prop (already passed to InventoryPage), `ApiService`.

- [ ] **Step 1: Add tab state + switcher to InventoryPage**

In `InventoryPage.svelte` `<script>`, add `import FilterTabs from "../components/ui/FilterTabs.svelte";`, `import StockTab from "./inventory/StockTab.svelte";`, and `let tab = $state("distribusi");`. Just below `<PageHeader …>`, add:
```svelte
<FilterTabs
  tabs={[{ id: "distribusi", label: "Distribusi" }, { id: "stok", label: "Stok" }]}
  active={tab}
  onChange={(id) => (tab = id)}
/>
```
Wrap the existing forecast/table/QR markup in `{#if tab === "distribusi"} … {/if}`, then add `{#if tab === "stok"}<StockTab {groups} />{/if}`.
*Note:* match `FilterTabs`'s real prop names — open `components/ui/FilterTabs.svelte` and mirror how `AgentsPage.svelte` (line ~186) passes them.

- [ ] **Step 2: Build the Stok view**

Create `StockTab.svelte` using the established Tailwind idiom. Structure (full handlers; data via `ApiService`):
```svelte
<script>
  import { onMount } from "svelte";
  import { Plus, Search, AlertTriangle } from "lucide-svelte";
  import StatCard from "../../components/StatCard.svelte";
  import EmptyState from "../../components/EmptyState.svelte";
  import { ApiService } from "../../services/api.js";
  import { showToast } from "../../services/toast.svelte.js";

  let { groups = [] } = $props();

  let items = $state([]);
  let loading = $state(true);
  let q = $state("");

  let lowCount = $derived(items.filter((i) => i.stock < i.min_stock || i.stock <= 0).length);
  let filtered = $derived(
    items.filter((i) => !q.trim() || i.name.toLowerCase().includes(q.trim().toLowerCase())),
  );

  onMount(load);
  async function load() {
    loading = true;
    try {
      const res = await ApiService.listStockItems();
      items = res?.items ?? [];
    } catch (e) {
      showToast(e.message || "Gagal memuat stok", "error");
    } finally {
      loading = false;
    }
  }
  // addItem(), restock(id), adjust(id), openHistory(id), saveKit(packageId, lines)
  // each call the matching ApiService method then await load(); show a toast.
</script>

<div class="flex flex-col gap-5">
  {#if lowCount > 0}
    <div class="flex items-center gap-2 rounded-xl bg-amber-50 p-3 text-[13.5px] text-amber-900">
      <AlertTriangle class="h-[18px] w-[18px] text-amber-600" />
      {lowCount} item stok menipis atau habis — periksa dan tambah stok.
    </div>
  {/if}

  <div class="grid grid-cols-2 gap-4 lg:grid-cols-3">
    <StatCard label="Jenis Item" value={items.length} />
    <StatCard label="Stok Menipis" value={lowCount} />
  </div>

  <div class="relative max-w-md">
    <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
    <input
      bind:value={q}
      placeholder="Cari item…"
      class="w-full rounded-xl border border-slate-200 bg-white py-2.5 pl-9 pr-3 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
    />
  </div>

  {#if loading}
    <div class="h-24 animate-pulse rounded-xl bg-slate-100"></div>
  {:else if filtered.length === 0}
    <EmptyState icon={Plus} title="Belum ada item" text="Tambahkan stok perlengkapan untuk mulai memantau." />
  {:else}
    <div class="overflow-hidden rounded-xl bg-white shadow-sm ring-1 ring-slate-200/60">
      <table class="w-full text-sm">
        <thead class="border-b border-slate-100 text-left text-[12px] uppercase tracking-wide text-slate-500">
          <tr><th class="p-3">Item</th><th class="p-3">Kategori</th><th class="p-3 text-right">Stok</th><th class="p-3 text-right">Min</th><th class="p-3"></th></tr>
        </thead>
        <tbody>
          {#each filtered as it (it.id)}
            <tr class="border-b border-slate-50">
              <td class="p-3 font-medium text-slate-800">{it.name}</td>
              <td class="p-3 text-slate-500">{it.category}</td>
              <td class="p-3 text-right font-semibold {it.stock <= 0 || it.stock < it.min_stock ? 'text-red-600' : 'text-slate-800'}">{it.stock} {it.unit}</td>
              <td class="p-3 text-right text-slate-400">{it.min_stock}</td>
              <td class="p-3 text-right"><!-- Restock / Adjust / Riwayat buttons --></td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}

  <!-- Add-item form, Restock/Adjust modals, History drawer (SlideDrawer), and
       "Kit per Paket" section (group/package picker + item rows + qty/jamaah,
       saved via ApiService.setPackageKit) follow the same idiom. -->
</div>
```

- [ ] **Step 3: Verify**

Run (in `frontend-svelte/`): `npm run check`
Expected: 0 errors. Confirm `StatCard`/`FilterTabs`/`SlideDrawer` prop names against their components.

- [ ] **Step 4: Commit**

```bash
git add frontend-svelte/src/lib/pages/InventoryPage.svelte frontend-svelte/src/lib/pages/inventory/StockTab.svelte
git commit -m "feat(inventory): Stok tab (levels, alerts, add/restock/adjust, kit per paket)"
```

---

### Task 10: End-to-end verification

**Files:** none (verification only)

- [ ] **Step 1: Backend gates**

Run: `go build ./cmd/... && go vet ./... && go test ./...`
Expected: all pass.

- [ ] **Step 2: Frontend gates**

Run (in `frontend-svelte/`): `npm run check && npm test && npm run build`
Expected: 0 errors; build succeeds.

- [ ] **Step 3: Manual smoke (local compose)**

Bring up infra + inventory-service + jamaah-service + gateway + frontend. In the app: Inventaris → **Stok** → add an item with initial stock; set a **Kit per Paket** for a package; in **Grup**, transition that package's group to **Berangkat**; confirm stock dropped by `qty × member_count` and a **departure** movement appears in the item's history. Re-emitting departure must NOT double-deduct.

- [ ] **Step 4: Commit (if any verification fixups were needed)**

```bash
git add -A && git commit -m "test(inventory): verification fixups"
```

---

## Self-Review

- **Spec coverage:** data model (T1–T2), item CRUD + restock/adjust/history (T4–T6), kit get/set (T4–T6), event consumer + idempotent deduction + negative-allowed (T3–T5, T7), frontend two-tab + Stok view + kit config + alerts (T8–T9), testing (T3, T10). All spec sections map to a task.
- **Idempotency:** the partial unique index (T1) + conditional INSERT…RETURNING/UPDATE (T4) + EventType filter (T5) implement "re-delivery is a no-op."
- **Type consistency:** model names (`StockItem`, `PackageKitItem`, `Deduction`, `DepartedPayload`, `KitLine`) are defined in T2 and used unchanged in T3–T6; repo method names in T4 match the service calls in T5 and the handler calls in T6; API method names in T8 match the StockTab calls in T9.
- **Verify-before-claim notes:** T6/T8/T9 flag the few names to confirm against the codebase (`response.Conflict`, `claims.UserID`, the `apiFetch` unwrap helper, `FilterTabs`/`StatCard` props) rather than assuming.
