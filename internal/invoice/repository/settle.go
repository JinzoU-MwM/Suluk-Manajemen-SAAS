package repository

import (
	"context"

	"github.com/google/uuid"
)

// SettleFromCredit applies a non-cash credit (e.g. savings conversion) to an
// invoice: it raises amount_paid by min(amount, remaining) and updates the
// status, WITHOUT inserting a payment row or emitting a payment event (the GL
// journal for the credit is posted by the originating module, e.g. tabungan's
// savings.converted). Returns the amount actually applied.
func (r *InvoiceRepo) SettleFromCredit(ctx context.Context, invoiceID, orgID uuid.UUID, amount int64) (int64, error) {
	if amount <= 0 {
		return 0, nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var total, paid, remaining int64
	var status string
	if err := tx.QueryRow(ctx, `SELECT total_amount, amount_paid, amount_remaining, status
		FROM invoices WHERE id=$1 AND org_id=$2 FOR UPDATE`, invoiceID, orgID).
		Scan(&total, &paid, &remaining, &status); err != nil {
		return 0, ErrInvoiceNotFound
	}
	if status == "batal" {
		return 0, ErrAlreadyCancelled
	}
	applied := amount
	if applied > remaining {
		applied = remaining
	}
	if applied <= 0 {
		return 0, nil // already fully paid
	}
	newPaid := paid + applied
	newRemaining := total - newPaid
	newStatus := "sebagian"
	if newRemaining <= 0 {
		newRemaining = 0
		newStatus = "lunas"
	}
	if _, err := tx.Exec(ctx, `UPDATE invoices SET amount_paid=$3, amount_remaining=$4, status=$5, updated_at=NOW()
		WHERE id=$1 AND org_id=$2`, invoiceID, orgID, newPaid, newRemaining, newStatus); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return applied, nil
}
