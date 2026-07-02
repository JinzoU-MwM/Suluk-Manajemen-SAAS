package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/invoice/model"
	"github.com/jamaah-in/v2/internal/shared/events"
	"github.com/jamaah-in/v2/internal/shared/outbox"
)

// CreateInvoiceTx inserts an invoice and an invoice.issued outbox event in one
// transaction (no lost events on crash).
func (r *InvoiceRepo) CreateInvoiceTx(ctx context.Context, inv *model.Invoice, evt outbox.Event) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx, `INSERT INTO invoices (id, org_id, invoice_number, jamaah_id, jamaah_name, package_id, package_name, registration_id,
		room_type, price_snapshot, discount_amount, surcharge_amount, total_amount, amount_paid, amount_remaining,
		payment_scheme, status, due_date, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		RETURNING issued_at, created_at, updated_at`,
		inv.ID, inv.OrgID, inv.InvoiceNumber, inv.JamaahID, inv.JamaahName, inv.PackageID, inv.PackageName, inv.RegistrationID,
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
func (r *InvoiceRepo) RecordPaymentTx(ctx context.Context, p *model.Payment, inv *model.Invoice) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Lock the invoice and recompute balances under the lock (prevents lost
	// updates on concurrent payments); overpayment is booked to titipan.
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

	// For cash payments, lock the stamped session and confirm it is still open, so
	// a payment cannot land in a session being concurrently closed (which would
	// drop it from the tutup-kas reconciliation). Detach from a missing/closed
	// session rather than corrupt its recon.
	if p.CashSessionID != nil {
		var sessStatus string
		if err := tx.QueryRow(ctx, `SELECT status FROM cash_sessions WHERE id=$1 AND org_id=$2 FOR UPDATE`, *p.CashSessionID, inv.OrgID).Scan(&sessStatus); err != nil || sessStatus != "open" {
			p.CashSessionID = nil
		}
	}

	remaining := total - paid
	applied := p.Amount
	if applied > remaining {
		applied = remaining
	}
	excess := p.Amount - applied
	newPaid := paid + applied
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

	if applied > 0 {
		pr, _ := json.Marshal(map[string]any{"amount": applied, "payment_method": p.PaymentMethod, "invoice_number": inv.InvoiceNumber, "jamaah_id": inv.JamaahID})
		if err := outbox.Insert(ctx, tx, outbox.Event{OrgID: inv.OrgID, AggregateType: "invoice", AggregateID: inv.ID, EventType: events.EventPaymentReceived, Payload: pr, OccurredAt: p.PaidAt}); err != nil {
			return err
		}
	}
	if excess > 0 {
		op, _ := json.Marshal(map[string]any{"amount": excess, "payment_method": p.PaymentMethod, "invoice_number": inv.InvoiceNumber})
		if err := outbox.Insert(ctx, tx, outbox.Event{OrgID: inv.OrgID, AggregateType: "invoice", AggregateID: inv.ID, EventType: events.EventOverpaymentReceived, Payload: op, OccurredAt: p.PaidAt}); err != nil {
			return err
		}
	}
	// Reflect the locked computation back for the caller (schedule allocation + response).
	inv.AmountPaid = newPaid
	inv.AmountRemaining = newRemaining
	inv.Status = newStatus
	return tx.Commit(ctx)
}

// CancelInvoiceTx marks an invoice cancelled and, when the invoice still has
// an uncollected balance, emits a GL reversal event in the same transaction.
// The invoice is locked and re-read here rather than trusted from the caller,
// so a concurrent payment landing between the caller's earlier read and this
// call can't make the reversal amount stale. Cancelling a fully-paid (lunas)
// invoice is rejected — collected money must go through the refund flow.
func (r *InvoiceRepo) CancelInvoiceTx(ctx context.Context, id, orgID uuid.UUID, reason string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status, invoiceNumber string
	var amountRemaining int64
	if err := tx.QueryRow(ctx, `SELECT status, amount_remaining, invoice_number FROM invoices WHERE id=$1 AND org_id=$2 FOR UPDATE`,
		id, orgID).Scan(&status, &amountRemaining, &invoiceNumber); err != nil {
		return ErrInvoiceNotFound
	}
	if status == "batal" {
		return ErrAlreadyCancelled
	}
	if status == "lunas" {
		return ErrAlreadyLunas
	}

	if _, err := tx.Exec(ctx, `UPDATE invoices SET status='batal', cancelled_at=NOW(), cancelled_reason=$3, updated_at=NOW() WHERE id=$1 AND org_id=$2`,
		id, orgID, reason); err != nil {
		return err
	}

	if amountRemaining > 0 {
		payload, _ := json.Marshal(map[string]any{"amount": amountRemaining, "invoice_number": invoiceNumber})
		if err := outbox.Insert(ctx, tx, outbox.Event{
			OrgID:         orgID,
			AggregateType: "invoice",
			AggregateID:   id,
			EventType:     events.EventInvoiceCancelled,
			Payload:       payload,
			OccurredAt:    time.Now(),
		}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
