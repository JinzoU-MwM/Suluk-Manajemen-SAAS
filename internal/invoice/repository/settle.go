package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ErrJamaahMismatch means the credit source (e.g. a savings account) belongs
// to a different jamaah than the invoice being settled — never apply one
// jamaah's funds to another jamaah's invoice (finding T3).
var ErrJamaahMismatch = fmt.Errorf("credit source does not belong to this invoice's jamaah")

// SettleFromCredit applies a non-cash credit (e.g. savings conversion) to an
// invoice: it raises amount_paid by min(amount, remaining) and updates the
// status, WITHOUT inserting a payment row or emitting a payment event (the GL
// journal for the credit is posted by the originating module, e.g. tabungan's
// savings.converted). Returns the amount actually applied.
//
// jamaahID must match the invoice's own jamaah_id (ErrJamaahMismatch
// otherwise) — the caller's ownership check alone isn't a hard boundary,
// this is. idempotencyKey makes repeat calls for the same logical attempt
// safe: a second call with the same (invoiceID, idempotencyKey) returns the
// first call's applied amount without re-touching the invoice.
func (r *InvoiceRepo) SettleFromCredit(ctx context.Context, invoiceID, orgID, jamaahID uuid.UUID, amount int64, idempotencyKey string) (int64, error) {
	if amount <= 0 {
		return 0, nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var total, paid, remaining int64
	var status string
	var invJamaahID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT total_amount, amount_paid, amount_remaining, status, jamaah_id
		FROM invoices WHERE id=$1 AND org_id=$2 FOR UPDATE`, invoiceID, orgID).
		Scan(&total, &paid, &remaining, &status, &invJamaahID); err != nil {
		return 0, ErrInvoiceNotFound
	}
	if invJamaahID != jamaahID {
		return 0, ErrJamaahMismatch
	}
	if status == "batal" {
		return 0, ErrAlreadyCancelled
	}

	var existing int64
	if err := tx.QueryRow(ctx, `SELECT applied_amount FROM settle_applications WHERE invoice_id=$1 AND idempotency_key=$2`,
		invoiceID, idempotencyKey).Scan(&existing); err == nil {
		return existing, tx.Commit(ctx)
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
	if _, err := tx.Exec(ctx, `INSERT INTO settle_applications (org_id, invoice_id, idempotency_key, applied_amount) VALUES ($1,$2,$3,$4)`,
		orgID, invoiceID, idempotencyKey, applied); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return applied, nil
}
