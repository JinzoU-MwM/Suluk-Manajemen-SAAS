package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/payroll/model"
	"github.com/jamaah-in/v2/internal/shared/outbox"
)

// CreateSalarySlipTx inserts a salary slip and a payroll.posted outbox event in
// one transaction. The slip id is DB-generated, so the outbox aggregate_id is
// set from the returned id inside the same tx.
func (r *PayrollRepo) CreateSalarySlipTx(ctx context.Context, s *model.SalarySlip, eventType string, payload []byte) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := tx.QueryRow(ctx, `
		INSERT INTO salary_slips (org_id, employee_id, period, base_salary, allowance, deductions, pph21_amount, bpjs_amount, advance_deduction, net_salary, package_id, status, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'draft',$12)
		RETURNING id, created_at, updated_at
	`, s.OrgID, s.EmployeeID, s.Period, s.BaseSalary, s.Allowance, s.Deductions, s.Pph21Amount, s.BpjsAmount, s.AdvanceDeduction, s.NetSalary, s.PackageID, s.Notes).
		Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return err
	}

	orgID, _ := uuid.Parse(s.OrgID)
	aggID, _ := uuid.Parse(s.ID)
	if err := outbox.Insert(ctx, tx, outbox.Event{
		OrgID:         orgID,
		AggregateType: "salary_slip",
		AggregateID:   aggID,
		EventType:     eventType,
		Payload:       payload,
		OccurredAt:    s.CreatedAt,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
