package repository

import (
	"context"

	"github.com/jamaah-in/v2/internal/invoice/model"
	"github.com/jamaah-in/v2/internal/shared/outbox"
)

// CreateInvoiceTx inserts an invoice and an invoice.issued outbox event in one
// transaction (no lost events on crash).
func (r *InvoiceRepo) CreateInvoiceTx(ctx context.Context, inv *model.Invoice, evt outbox.Event) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `INSERT INTO invoices (id, org_id, invoice_number, jamaah_id, package_id, registration_id,
		room_type, price_snapshot, discount_amount, surcharge_amount, total_amount, amount_paid, amount_remaining,
		payment_scheme, status, due_date, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		RETURNING issued_at, created_at, updated_at`,
		inv.ID, inv.OrgID, inv.InvoiceNumber, inv.JamaahID, inv.PackageID, inv.RegistrationID,
		inv.RoomType, inv.PriceSnapshot, inv.DiscountAmount, inv.SurchargeAmount, inv.TotalAmount,
		inv.AmountPaid, inv.AmountRemaining, inv.PaymentScheme, inv.Status, inv.DueDate, inv.Notes,
	).Scan(&inv.IssuedAt, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if isDuplicate(err) {
			return ErrDuplicateNumber
		}
		return err
	}
	if err := outbox.Insert(ctx, tx, evt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// RecordPaymentTx inserts a payment, updates the invoice totals/status, and
// writes a payment.received outbox event — all atomically.
func (r *InvoiceRepo) RecordPaymentTx(ctx context.Context, p *model.Payment, inv *model.Invoice, evt outbox.Event) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(ctx, `INSERT INTO payments (id, org_id, invoice_id, amount, payment_method, bank_name, account_number, reference_number, proof_url, notes, received_by, paid_at, cash_session_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING created_at`,
		p.ID, p.OrgID, p.InvoiceID, p.Amount, p.PaymentMethod, p.BankName, p.AccountNumber,
		p.ReferenceNumber, p.ProofURL, p.Notes, p.ReceivedBy, p.PaidAt, p.CashSessionID).Scan(&p.CreatedAt); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `UPDATE invoices SET discount_amount=$2, surcharge_amount=$3, total_amount=$4,
		amount_paid=$5, amount_remaining=$6, due_date=$7, notes=$8, status=$9, updated_at=NOW() WHERE id = $1 AND org_id = $10`,
		inv.ID, inv.DiscountAmount, inv.SurchargeAmount, inv.TotalAmount,
		inv.AmountPaid, inv.AmountRemaining, inv.DueDate, inv.Notes, inv.Status, inv.OrgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrInvoiceNotFound
	}

	if err := outbox.Insert(ctx, tx, evt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
