package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jamaah-in/v2/internal/payroll/model"
)

var (
	ErrEmployeeNotFound = errors.New("employee not found")
	ErrAdvanceNotFound  = errors.New("advance not found")
	ErrRepayInvalid     = errors.New("repayment amount must be greater than 0")
	ErrRepayTooMuch     = errors.New("repayment exceeds remaining balance")
)

type PayrollRepo struct {
	pool *pgxpool.Pool
}

func NewPayrollRepo(pool *pgxpool.Pool) *PayrollRepo {
	return &PayrollRepo{pool: pool}
}

func (r *PayrollRepo) CreateEmployee(ctx context.Context, e *model.Employee) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO employees (org_id, name, position, type, base_salary, allowance, bpjs_tk, bpjs_kes, pph21_rate, phone, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, created_at, updated_at
	`, e.OrgID, e.Name, e.Position, e.Type, e.BaseSalary, e.Allowance, e.BpjsTk, e.BpjsKes, e.Pph21Rate, e.Phone, e.Email).
		Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

func (r *PayrollRepo) ListEmployees(ctx context.Context, orgID string) ([]model.Employee, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, org_id, name, position, type, base_salary, allowance, bpjs_tk, bpjs_kes, pph21_rate, phone, email, is_active, created_at, updated_at
		FROM employees WHERE org_id = $1 ORDER BY name
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []model.Employee
	for rows.Next() {
		var e model.Employee
		if err := rows.Scan(&e.ID, &e.OrgID, &e.Name, &e.Position, &e.Type, &e.BaseSalary, &e.Allowance, &e.BpjsTk, &e.BpjsKes, &e.Pph21Rate, &e.Phone, &e.Email, &e.IsActive, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func (r *PayrollRepo) GetEmployee(ctx context.Context, id, orgID string) (*model.Employee, error) {
	var e model.Employee
	err := r.pool.QueryRow(ctx, `
		SELECT id, org_id, name, position, type, base_salary, allowance, bpjs_tk, bpjs_kes, pph21_rate, phone, email, is_active, created_at, updated_at
		FROM employees WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(&e.ID, &e.OrgID, &e.Name, &e.Position, &e.Type, &e.BaseSalary, &e.Allowance, &e.BpjsTk, &e.BpjsKes, &e.Pph21Rate, &e.Phone, &e.Email, &e.IsActive, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, ErrEmployeeNotFound
	}
	return &e, nil
}

// SlipExistsForPeriod reports whether a salary slip already exists for this
// employee in the given period — used to block double-running payroll (which
// would emit a second payroll.posted event and double-book the GL).
func (r *PayrollRepo) SlipExistsForPeriod(ctx context.Context, orgID, employeeID, period string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM salary_slips WHERE org_id=$1 AND employee_id=$2 AND period=$3)`,
		orgID, employeeID, period).Scan(&exists)
	return exists, err
}

func (r *PayrollRepo) UpdateEmployee(ctx context.Context, id, orgID string, updates map[string]interface{}) error {
	query := "UPDATE employees SET updated_at = NOW()"
	args := []interface{}{}
	idx := 1
	for k, v := range updates {
		query += fmt.Sprintf(", %s = $%d", k, idx)
		args = append(args, v)
		idx++
	}
	query += fmt.Sprintf(" WHERE id = $%d AND org_id = $%d", idx, idx+1)
	args = append(args, id, orgID)
	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *PayrollRepo) CreateSalarySlip(ctx context.Context, s *model.SalarySlip) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO salary_slips (org_id, employee_id, period, base_salary, allowance, deductions, pph21_amount, bpjs_amount, advance_deduction, net_salary, package_id, status, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'draft',$12)
		RETURNING id, created_at, updated_at
	`, s.OrgID, s.EmployeeID, s.Period, s.BaseSalary, s.Allowance, s.Deductions, s.Pph21Amount, s.BpjsAmount, s.AdvanceDeduction, s.NetSalary, s.PackageID, s.Notes).
		Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *PayrollRepo) ListSalarySlips(ctx context.Context, orgID, period string) ([]model.SalarySlip, error) {
	query := `
		SELECT s.id, s.org_id, s.employee_id, s.period, s.base_salary, s.allowance, s.deductions,
		       s.pph21_amount, s.bpjs_amount, s.advance_deduction, s.net_salary, s.package_id, s.status, s.notes, s.created_at, s.updated_at,
		       e.name
		FROM salary_slips s JOIN employees e ON s.employee_id = e.id AND e.org_id = s.org_id
		WHERE s.org_id = $1
	`
	args := []interface{}{orgID}
	if period != "" {
		query += " AND s.period = $2 ORDER BY e.name"
		args = append(args, period)
	} else {
		query += " ORDER BY s.period DESC, e.name"
	}
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []model.SalarySlip
	for rows.Next() {
		var s model.SalarySlip
		if err := rows.Scan(&s.ID, &s.OrgID, &s.EmployeeID, &s.Period, &s.BaseSalary, &s.Allowance, &s.Deductions, &s.Pph21Amount, &s.BpjsAmount, &s.AdvanceDeduction, &s.NetSalary, &s.PackageID, &s.Status, &s.Notes, &s.CreatedAt, &s.UpdatedAt, &s.EmployeeName); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (r *PayrollRepo) UpdateSalarySlipStatus(ctx context.Context, id, orgID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE salary_slips SET status = $1, updated_at = NOW() WHERE id = $2 AND org_id = $3`, status, id, orgID)
	return err
}

func (r *PayrollRepo) CreateAdvance(ctx context.Context, a *model.Advance) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO advances (org_id, employee_id, amount, remaining, reason)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, created_at, updated_at
	`, a.OrgID, a.EmployeeID, a.Amount, a.Amount, a.Reason).
		Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (r *PayrollRepo) ListAdvances(ctx context.Context, orgID string) ([]model.Advance, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.org_id, a.employee_id, a.amount, a.remaining, a.reason, a.status, a.created_at, a.updated_at, e.name
		FROM advances a JOIN employees e ON a.employee_id = e.id AND e.org_id = a.org_id
		WHERE a.org_id = $1 ORDER BY a.created_at DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []model.Advance
	for rows.Next() {
		var a model.Advance
		if err := rows.Scan(&a.ID, &a.OrgID, &a.EmployeeID, &a.Amount, &a.Remaining, &a.Reason, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.EmployeeName); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (r *PayrollRepo) RepayAdvance(ctx context.Context, advanceID, orgID string, amount int64, slipID *string) error {
	if amount <= 0 {
		return ErrRepayInvalid
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Lock + org-scope the advance, then validate the balance BEFORE inserting the
	// repayment. The old code had no org filter (cross-org IDOR) and inserted the
	// repayment row before the remaining>=amount guard (orphan row on over-repay).
	var remaining int64
	if err := tx.QueryRow(ctx, `SELECT remaining FROM advances WHERE id=$1 AND org_id=$2 FOR UPDATE`, advanceID, orgID).Scan(&remaining); err != nil {
		return ErrAdvanceNotFound
	}
	if amount > remaining {
		return ErrRepayTooMuch
	}

	if _, err := tx.Exec(ctx, `INSERT INTO advance_repayments (advance_id, amount, salary_slip_id) VALUES ($1,$2,$3)`, advanceID, amount, slipID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE advances SET remaining = remaining - $1, status = CASE WHEN remaining - $1 <= 0 THEN 'settled' ELSE status END, updated_at = NOW()
		WHERE id = $2 AND org_id = $3
	`, amount, advanceID, orgID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PayrollRepo) GetAdvance(ctx context.Context, id, orgID string) (*model.Advance, error) {
	var a model.Advance
	err := r.pool.QueryRow(ctx, `
		SELECT a.id, a.org_id, a.employee_id, a.amount, a.remaining, a.reason, a.status, a.created_at, a.updated_at, e.name
		FROM advances a JOIN employees e ON a.employee_id = e.id AND e.org_id = a.org_id
		WHERE a.id = $1 AND a.org_id = $2
	`, id, orgID).Scan(&a.ID, &a.OrgID, &a.EmployeeID, &a.Amount, &a.Remaining, &a.Reason, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.EmployeeName)
	if err != nil {
		return nil, ErrAdvanceNotFound
	}
	return &a, nil
}

func (r *PayrollRepo) GetPayrollSummary(ctx context.Context, orgID string) (*model.PayrollSummary, error) {
	s := &model.PayrollSummary{}
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE is_active = TRUE)
		FROM employees WHERE org_id = $1`, orgID).Scan(&s.TotalEmployees, &s.ActiveEmployees); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0), COALESCE(SUM(remaining) FILTER (WHERE status = 'active'), 0)
		FROM advances WHERE org_id = $1`, orgID).Scan(&s.TotalAdvances, &s.OutstandingAdvances); err != nil {
		return nil, err
	}
	currentPeriod := time.Now().Format("2006-01")
	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(net_salary), 0)
		FROM salary_slips WHERE org_id = $1 AND period = $2`, orgID, currentPeriod).Scan(&s.MonthlyPayroll); err != nil {
		return nil, err
	}
	return s, nil
}
