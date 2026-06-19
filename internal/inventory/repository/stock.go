package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/jamaah-in/v2/internal/inventory/model"
)

// ErrItemInKit blocks deleting an item that a package kit still references.
var ErrItemInKit = errors.New("item is used in a package kit")

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
