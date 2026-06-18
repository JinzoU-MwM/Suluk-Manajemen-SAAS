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

	// Lock the invoice and recompute balances under the lock: prevents lost
	// updates on concurrent payments and rejects overpayment.
	var total, paid int64
	var curStatus string
	if err := tx.QueryRow(ctx, `SELECT total_amount, amount_paid, status FROM invoices WHERE id=$1 AND org_id=$2 FOR UPDATE`,
		inv.ID, inv.OrgID).Scan(&total, &paid, &curStatus); err != nil {
		return ErrInvoiceNotFound
	}
	if curStatus == "batal" {
		return ErrAlreadyCancelled
	}
	if curStatus == "lunas" {
		return ErrAlreadyLunas
	}
	if p.Amount > total-paid {
		return ErrOverpayment
	}
	newPaid := paid + p.Amount
	newRemaining := total - newPaid
	newStatus := "sebagian"
	if newRemaining <= 0 {
		newRemaining = 0
		newStatus = "lunas"
	}

	if err := tx.QueryRow(ctx, `INSERT INTO payments (id, org_id, invoice_id, amount, payment_method, bank_name, account_number, reference_number, proof_url, notes, received_by, paid_at, cash_session_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING created_at`,
		p.ID, p.OrgID, p.InvoiceID, p.Amount, p.PaymentMethod, p.BankName, p.AccountNumber,
		p.ReferenceNumber, p.ProofURL, p.Notes, p.ReceivedBy, p.PaidAt, p.CashSessionID).Scan(&p.CreatedAt); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `UPDATE invoices SET amount_paid=$3, amount_remaining=$4, status=$5, updated_at=NOW() WHERE id=$1 AND org_id=$2`,
		inv.ID, inv.OrgID, newPaid, newRemaining, newStatus); err != nil {
		return err
	}

	if err := outbox.Insert(ctx, tx, evt); err != nil {
		return err
	}
	// Reflect the locked computation back for the caller (schedule allocation + response).
	inv.AmountPaid = newPaid
	inv.AmountRemaining = newRemaining
	inv.Status = newStatus
	return tx.Commit(ctx)
}
