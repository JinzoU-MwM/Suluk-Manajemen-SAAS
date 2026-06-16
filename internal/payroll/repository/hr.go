package repository

import (
	"context"
	"fmt"

	"github.com/jamaah-in/v2/internal/payroll/model"
)

var ErrLeaveNotFound = fmt.Errorf("leave request not found or already decided")

// RecordAttendance upserts one employee/day attendance row.
func (r *PayrollRepo) RecordAttendance(ctx context.Context, a *model.Attendance) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO attendance (org_id, employee_id, date, status, notes)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (employee_id, date) DO UPDATE SET status = EXCLUDED.status, notes = EXCLUDED.notes, updated_at = NOW()
		RETURNING id, created_at, updated_at`,
		a.OrgID, a.EmployeeID, a.Date, a.Status, a.Notes,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

// ListAttendance returns one employee's attendance for a YYYY-MM period.
func (r *PayrollRepo) ListAttendance(ctx context.Context, orgID, employeeID, period string) ([]model.Attendance, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, org_id, employee_id, to_char(date, 'YYYY-MM-DD'), status, notes, created_at, updated_at
		FROM attendance
		WHERE org_id = $1 AND employee_id = $2 AND to_char(date, 'YYYY-MM') = $3
		ORDER BY date`, orgID, employeeID, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Attendance{}
	for rows.Next() {
		var a model.Attendance
		if err := rows.Scan(&a.ID, &a.OrgID, &a.EmployeeID, &a.Date, &a.Status, &a.Notes, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// AttendanceSummary rolls up status counts (and unpaid days) for one employee in
// a YYYY-MM period.
func (r *PayrollRepo) AttendanceSummary(ctx context.Context, orgID, employeeID, period string) (*model.AttendanceSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT status, COUNT(*) FROM attendance
		WHERE org_id = $1 AND employee_id = $2 AND to_char(date, 'YYYY-MM') = $3
		GROUP BY status`, orgID, employeeID, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sum := &model.AttendanceSummary{EmployeeID: employeeID, Period: period, Counts: map[string]int{}}
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return nil, err
		}
		sum.Counts[status] = n
		if status == "absen" || status == "tanpa_gaji" {
			sum.UnpaidDays += n
		}
	}
	return sum, rows.Err()
}

func (r *PayrollRepo) CreateLeave(ctx context.Context, l *model.LeaveRequest) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO leave_requests (org_id, employee_id, type, start_date, end_date, days, reason)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, status, created_at, updated_at`,
		l.OrgID, l.EmployeeID, l.Type, l.StartDate, l.EndDate, l.Days, l.Reason,
	).Scan(&l.ID, &l.Status, &l.CreatedAt, &l.UpdatedAt)
}

func (r *PayrollRepo) ListLeave(ctx context.Context, orgID, status string) ([]model.LeaveRequest, error) {
	q := `SELECT l.id, l.org_id, l.employee_id, l.type, to_char(l.start_date,'YYYY-MM-DD'), to_char(l.end_date,'YYYY-MM-DD'),
	             l.days, l.reason, l.status, l.decided_by, l.created_at, l.updated_at, e.name
	      FROM leave_requests l JOIN employees e ON e.id = l.employee_id WHERE l.org_id = $1`
	args := []any{orgID}
	if status != "" && status != "all" {
		args = append(args, status)
		q += " AND l.status = $2"
	}
	q += " ORDER BY l.created_at DESC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.LeaveRequest{}
	for rows.Next() {
		var l model.LeaveRequest
		if err := rows.Scan(&l.ID, &l.OrgID, &l.EmployeeID, &l.Type, &l.StartDate, &l.EndDate, &l.Days,
			&l.Reason, &l.Status, &l.DecidedBy, &l.CreatedAt, &l.UpdatedAt, &l.EmployeeName); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// DecideLeave approves/rejects a pending leave request.
func (r *PayrollRepo) DecideLeave(ctx context.Context, id, orgID, status, decidedBy string) error {
	ct, err := r.pool.Exec(ctx, `UPDATE leave_requests SET status = $3, decided_by = $4, updated_at = NOW()
		WHERE id = $1 AND org_id = $2 AND status = 'pending'`, id, orgID, status, decidedBy)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrLeaveNotFound
	}
	return nil
}
