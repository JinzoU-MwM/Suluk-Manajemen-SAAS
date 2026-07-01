package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/invoice/model"
	"github.com/jamaah-in/v2/internal/shared/events"
	"github.com/jamaah-in/v2/internal/shared/outbox"
)

var (
	ErrRefundNotFound    = fmt.Errorf("refund not found")
	ErrRefundNotPending  = fmt.Errorf("refund not in pending status")
	ErrRefundExceedsPaid = fmt.Errorf("refund amount exceeds invoice paid amount")
	ErrRefundAlreadyOpen = fmt.Errorf("invoice already has an open refund")
	ErrPolicyNotFound    = fmt.Errorf("refund policy not found")
)

func (r *InvoiceRepo) CreateRefund(ctx context.Context, ref *model.Refund) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO refunds (org_id, invoice_id, amount, refund_pct, reason, notes, payment_method, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')
		RETURNING id, created_at, updated_at
	`, ref.OrgID, ref.InvoiceID, ref.Amount, ref.RefundPct, ref.Reason, ref.Notes, ref.PaymentMethod).Scan(&ref.ID, &ref.CreatedAt, &ref.UpdatedAt)
	if isDuplicate(err) {
		return ErrRefundAlreadyOpen
	}
	return err
}

func (r *InvoiceRepo) ListRefunds(ctx context.Context, orgID uuid.UUID, status string, page, limit int) ([]model.Refund, int64, error) {
	filterSQL, filterArgs := statusFilter(status)

	var total int64
	countQuery := "SELECT COUNT(*) FROM refunds WHERE org_id = $1" + filterSQL
	countArgs := append([]interface{}{orgID}, filterArgs...)
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	baseArgs := append([]interface{}{orgID}, filterArgs...)
	baseArgCount := len(baseArgs)

	selectQuery := fmt.Sprintf(`SELECT id, org_id, invoice_id, amount, refund_pct, payment_method, reason, status,
		approved_by, approved_at, processed_at, notes, created_at, updated_at
		FROM refunds WHERE org_id = $1%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		filterSQL, baseArgCount+1, baseArgCount+2)
	selectArgs := append(baseArgs, limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var refunds []model.Refund
	for rows.Next() {
		var ref model.Refund
		if err := rows.Scan(
			&ref.ID, &ref.OrgID, &ref.InvoiceID, &ref.Amount, &ref.RefundPct, &ref.PaymentMethod,
			&ref.Reason, &ref.Status, &ref.ApprovedBy, &ref.ApprovedAt,
			&ref.ProcessedAt, &ref.Notes, &ref.CreatedAt, &ref.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		refunds = append(refunds, ref)
	}
	return refunds, total, rows.Err()
}

func (r *InvoiceRepo) GetRefund(ctx context.Context, id, orgID uuid.UUID) (*model.Refund, error) {
	var ref model.Refund
	err := r.pool.QueryRow(ctx, `
		SELECT id, org_id, invoice_id, amount, refund_pct, payment_method, reason, status,
		       approved_by, approved_at, processed_at, notes, created_at, updated_at
		FROM refunds WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(
		&ref.ID, &ref.OrgID, &ref.InvoiceID, &ref.Amount, &ref.RefundPct, &ref.PaymentMethod,
		&ref.Reason, &ref.Status, &ref.ApprovedBy, &ref.ApprovedAt,
		&ref.ProcessedAt, &ref.Notes, &ref.CreatedAt, &ref.UpdatedAt,
	)
	if err != nil {
		return nil, ErrRefundNotFound
	}
	return &ref, nil
}

func (r *InvoiceRepo) ApproveRefund(ctx context.Context, id, orgID, approverID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE refunds SET status = 'approved', approved_by = $3, approved_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND org_id = $2 AND status = 'pending'
	`, id, orgID, approverID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrRefundNotPending
	}
	return nil
}

func (r *InvoiceRepo) ProcessRefund(ctx context.Context, id, orgID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE refunds SET status = 'processed', processed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND org_id = $2 AND status = 'approved'
	`, id, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("refund not in approved status")
	}
	return nil
}

func (r *InvoiceRepo) CompleteRefund(ctx context.Context, id, orgID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Confirm the refund is processed and capture its amount + invoice.
	var invoiceID uuid.UUID
	var amount int64
	var paymentMethod string
	if err := tx.QueryRow(ctx, `SELECT invoice_id, amount, payment_method FROM refunds WHERE id=$1 AND org_id=$2 AND status='processed' FOR UPDATE`,
		id, orgID).Scan(&invoiceID, &amount, &paymentMethod); err != nil {
		return fmt.Errorf("refund not in processed status")
	}
	if _, err := tx.Exec(ctx, `UPDATE refunds SET status='completed', updated_at=NOW() WHERE id=$1 AND org_id=$2`, id, orgID); err != nil {
		return err
	}

	// Reduce the invoice's paid amount so the subledger matches the GL.
	var total, paid int64
	var invStatus, invNumber string
	if err := tx.QueryRow(ctx, `SELECT total_amount, amount_paid, status, invoice_number FROM invoices WHERE id=$1 AND org_id=$2 FOR UPDATE`,
		invoiceID, orgID).Scan(&total, &paid, &invStatus, &invNumber); err != nil {
		return ErrInvoiceNotFound
	}
	newPaid := paid - amount
	if newPaid < 0 {
		newPaid = 0
	}
	newRemaining := total - newPaid
	if newRemaining < 0 {
		newRemaining = 0
	}
	if invStatus != "batal" {
		if newRemaining <= 0 {
			invStatus = "lunas"
		} else if newPaid > 0 {
			invStatus = "sebagian"
		} else {
			invStatus = "belum_bayar"
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE invoices SET amount_paid=$3, amount_remaining=$4, status=$5, updated_at=NOW() WHERE id=$1 AND org_id=$2`,
		invoiceID, orgID, newPaid, newRemaining, invStatus); err != nil {
		return err
	}

	// Emit refund.completed so accounting posts Dr Piutang / Cr Kas|Bank.
	payload, _ := json.Marshal(map[string]any{"amount": amount, "invoice_number": invNumber, "payment_method": paymentMethod})
	if err := outbox.Insert(ctx, tx, outbox.Event{
		OrgID:         orgID,
		AggregateType: "invoice",
		AggregateID:   invoiceID,
		EventType:     events.EventRefundCompleted,
		Payload:       payload,
		OccurredAt:    time.Now(),
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *InvoiceRepo) RejectRefund(ctx context.Context, id, orgID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE refunds SET status = 'rejected', updated_at = NOW()
		WHERE id = $1 AND org_id = $2 AND status = 'pending'
	`, id, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrRefundNotPending
	}
	return nil
}

func (r *InvoiceRepo) GetRefundsByInvoice(ctx context.Context, invoiceID, orgID uuid.UUID) ([]model.Refund, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, org_id, invoice_id, amount, refund_pct, payment_method, reason, status,
		       approved_by, approved_at, processed_at, notes, created_at, updated_at
		FROM refunds WHERE invoice_id = $1 AND org_id = $2 ORDER BY created_at DESC
	`, invoiceID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refunds []model.Refund
	for rows.Next() {
		var ref model.Refund
		if err := rows.Scan(
			&ref.ID, &ref.OrgID, &ref.InvoiceID, &ref.Amount, &ref.RefundPct, &ref.PaymentMethod,
			&ref.Reason, &ref.Status, &ref.ApprovedBy, &ref.ApprovedAt,
			&ref.ProcessedAt, &ref.Notes, &ref.CreatedAt, &ref.UpdatedAt,
		); err != nil {
			return nil, err
		}
		refunds = append(refunds, ref)
	}
	return refunds, rows.Err()
}

func (r *InvoiceRepo) CreateRefundPolicy(ctx context.Context, p *model.RefundPolicy) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO refund_policies (org_id, name, days_before, refund_pct, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, p.OrgID, p.Name, p.DaysBefore, p.RefundPct, p.Description).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *InvoiceRepo) ListRefundPolicies(ctx context.Context, orgID uuid.UUID) ([]model.RefundPolicy, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, org_id, name, days_before, refund_pct, description, is_active, created_at, updated_at
		FROM refund_policies WHERE org_id = $1 AND is_active = TRUE ORDER BY days_before DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []model.RefundPolicy
	for rows.Next() {
		var p model.RefundPolicy
		if err := rows.Scan(&p.ID, &p.OrgID, &p.Name, &p.DaysBefore, &p.RefundPct, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

func (r *InvoiceRepo) GetRefundPolicy(ctx context.Context, id, orgID uuid.UUID) (*model.RefundPolicy, error) {
	var p model.RefundPolicy
	err := r.pool.QueryRow(ctx, `
		SELECT id, org_id, name, days_before, refund_pct, description, is_active, created_at, updated_at
		FROM refund_policies WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(&p.ID, &p.OrgID, &p.Name, &p.DaysBefore, &p.RefundPct, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, ErrPolicyNotFound
	}
	return &p, nil
}

func (r *InvoiceRepo) UpdateRefundPolicy(ctx context.Context, id, orgID uuid.UUID, updates map[string]interface{}) error {
	query := `UPDATE refund_policies SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if v, ok := updates["name"]; ok {
		query += fmt.Sprintf(", name = $%d", argIdx)
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["days_before"]; ok {
		query += fmt.Sprintf(", days_before = $%d", argIdx)
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["refund_pct"]; ok {
		query += fmt.Sprintf(", refund_pct = $%d", argIdx)
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["description"]; ok {
		query += fmt.Sprintf(", description = $%d", argIdx)
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["is_active"]; ok {
		query += fmt.Sprintf(", is_active = $%d", argIdx)
		args = append(args, v)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d AND org_id = $%d", argIdx, argIdx+1)
	args = append(args, id, orgID)

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *InvoiceRepo) DeleteRefundPolicy(ctx context.Context, id, orgID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refund_policies WHERE id = $1 AND org_id = $2`, id, orgID)
	return err
}

func statusFilter(status string) (query string, args []interface{}) {
	if status == "" || status == "all" {
		return "", nil
	}
	return " AND status = $2", []interface{}{status}
}

func (r *InvoiceRepo) CancelInvoiceWithRefund(ctx context.Context, invoiceID, orgID uuid.UUID, refundAmount int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	result, err := tx.Exec(ctx, `
		UPDATE invoices SET status = 'batal', amount_paid = amount_paid - $3, amount_remaining = total_amount - (amount_paid - $3),
		cancelled_at = NOW(), cancelled_reason = 'Pembatalan dengan refund', updated_at = NOW()
		WHERE id = $1 AND org_id = $2 AND status != 'batal' AND status != 'lunas'
	`, invoiceID, orgID, refundAmount)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrAlreadyCancelled
	}

	return tx.Commit(ctx)
}
