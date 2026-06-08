package model

import "time"

type Employee struct {
	ID         string    `json:"id" db:"id"`
	OrgID      string    `json:"org_id" db:"org_id"`
	Name       string    `json:"name" db:"name"`
	Position   string    `json:"position" db:"position"`
	Type       string    `json:"type" db:"type"`
	BaseSalary int64     `json:"base_salary" db:"base_salary"`
	Allowance  int64     `json:"allowance" db:"allowance"`
	BpjsTk     int64     `json:"bpjs_tk" db:"bpjs_tk"`
	BpjsKes    int64     `json:"bpjs_kes" db:"bpjs_kes"`
	Pph21Rate  float64   `json:"pph21_rate" db:"pph21_rate"`
	Phone      string    `json:"phone" db:"phone"`
	Email      string    `json:"email" db:"email"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type SalarySlip struct {
	ID               string    `json:"id" db:"id"`
	OrgID            string    `json:"org_id" db:"org_id"`
	EmployeeID       string    `json:"employee_id" db:"employee_id"`
	Period           string    `json:"period" db:"period"`
	BaseSalary       int64     `json:"base_salary" db:"base_salary"`
	Allowance        int64     `json:"allowance" db:"allowance"`
	Deductions       int64     `json:"deductions" db:"deductions"`
	Pph21Amount      int64     `json:"pph21_amount" db:"pph21_amount"`
	BpjsAmount       int64     `json:"bpjs_amount" db:"bpjs_amount"`
	AdvanceDeduction int64     `json:"advance_deduction" db:"advance_deduction"`
	NetSalary        int64     `json:"net_salary" db:"net_salary"`
	PackageID        *string   `json:"package_id,omitempty" db:"package_id"`
	Status           string    `json:"status" db:"status"`
	Notes            string    `json:"notes" db:"notes"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`

	EmployeeName string `json:"employee_name,omitempty" db:"-"`
}

type Advance struct {
	ID         string    `json:"id" db:"id"`
	OrgID      string    `json:"org_id" db:"org_id"`
	EmployeeID string    `json:"employee_id" db:"employee_id"`
	Amount     int64     `json:"amount" db:"amount"`
	Remaining  int64     `json:"remaining" db:"remaining"`
	Reason     string    `json:"reason" db:"reason"`
	Status     string    `json:"status" db:"status"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`

	EmployeeName string             `json:"employee_name,omitempty" db:"-"`
	Repayments   []AdvanceRepayment `json:"repayments,omitempty" db:"-"`
}

type AdvanceRepayment struct {
	ID           string    `json:"id" db:"id"`
	AdvanceID    string    `json:"advance_id" db:"advance_id"`
	Amount       int64     `json:"amount" db:"amount"`
	SalarySlipID *string   `json:"salary_slip_id,omitempty" db:"salary_slip_id"`
	PaidAt       time.Time `json:"paid_at" db:"paid_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type CreateEmployeeRequest struct {
	Name       string  `json:"name"`
	Position   string  `json:"position"`
	Type       string  `json:"type"`
	BaseSalary int64   `json:"base_salary"`
	Allowance  int64   `json:"allowance"`
	BpjsTk     int64   `json:"bpjs_tk"`
	BpjsKes    int64   `json:"bpjs_kes"`
	Pph21Rate  float64 `json:"pph21_rate"`
	Phone      string  `json:"phone"`
	Email      string  `json:"email"`
}

type UpdateEmployeeRequest struct {
	Name       *string  `json:"name,omitempty"`
	Position   *string  `json:"position,omitempty"`
	Type       *string  `json:"type,omitempty"`
	BaseSalary *int64   `json:"base_salary,omitempty"`
	Allowance  *int64   `json:"allowance,omitempty"`
	BpjsTk     *int64   `json:"bpjs_tk,omitempty"`
	BpjsKes    *int64   `json:"bpjs_kes,omitempty"`
	Pph21Rate  *float64 `json:"pph21_rate,omitempty"`
	Phone      *string  `json:"phone,omitempty"`
	Email      *string  `json:"email,omitempty"`
	IsActive   *bool    `json:"is_active,omitempty"`
}

type CreateSalarySlipRequest struct {
	EmployeeID string `json:"employee_id"`
	Period     string `json:"period"`
	PackageID  string `json:"package_id"`
	Notes      string `json:"notes"`
}

type CreateAdvanceRequest struct {
	EmployeeID string `json:"employee_id"`
	Amount     int64  `json:"amount"`
	Reason     string `json:"reason"`
}

type RepayAdvanceRequest struct {
	Amount       int64  `json:"amount"`
	SalarySlipID string `json:"salary_slip_id"`
}

type PayrollSummary struct {
	TotalEmployees      int   `json:"total_employees"`
	ActiveEmployees     int   `json:"active_employees"`
	TotalAdvances       int64 `json:"total_advances"`
	OutstandingAdvances int64 `json:"outstanding_advances"`
	MonthlyPayroll      int64 `json:"monthly_payroll"`
}
