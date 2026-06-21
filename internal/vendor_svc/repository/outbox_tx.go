package repository

import (
	"context"
	"fmt"

	"github.com/jamaah-in/v2/internal/shared/outbox"
	"github.com/jamaah-in/v2/internal/vendor_svc/model"
)

// CreateBillTx inserts a vendor bill and a vendor.bill.created outbox event in
// one transaction.
func (r *VendorRepo) CreateBillTx(ctx context.Context, b *model.VendorBill, evt outbox.Event) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := fmt.Sprintf(`INSERT INTO vendor_bills (%s) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING amount_idr, paid_amount, created_at, updated_at`, billInsertCols)
	if err := tx.QueryRow(ctx, query,
		b.ID, b.OrgID, b.VendorID, b.PackageID, b.Description,
		b.Amount, b.Currency, b.ExchangeRate, b.DueDate, b.Status,
	).Scan(&b.AmountIDR, &b.PaidAmount, &b.CreatedAt, &b.UpdatedAt); err != nil {
		return err
	}
	if err := outbox.Insert(ctx, tx, evt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
