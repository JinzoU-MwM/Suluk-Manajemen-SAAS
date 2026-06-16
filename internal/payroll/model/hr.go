package model

import "time"

// WorkingDaysPerMonth is the denominator for the per-day rate used in
// attendance-based salary deductions.
const WorkingDaysPerMonth = 22

type Attendance struct {
	ID         string    `json:"id"`
	OrgID      string    `json:"org_id"`
	EmployeeID string    `json:"employee_id"`
	Date       string    `json:"date"` // YYYY-MM-DD
	Status     string    `json:"status"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	EmployeeName string `json:"employee_name,omitempty"`
}

// AttendanceSummary is the per-period rollup for one employee.
type AttendanceSummary struct {
	EmployeeID string         `json:"employee_id"`
	Period     string         `json:"period"` // YYYY-MM
	Counts     map[string]int `json:"counts"`
	UnpaidDays int            `json:"unpaid_days"` // absen + tanpa_gaji
}

type LeaveRequest struct {
	ID         string    `json:"id"`
	OrgID      string    `json:"org_id"`
	EmployeeID string    `json:"employee_id"`
	Type       string    `json:"type"`
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
	Days       int       `json:"days"`
	Reason     string    `json:"reason"`
	Status     string    `json:"status"`
	DecidedBy  *string   `json:"decided_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	EmployeeName string `json:"employee_name,omitempty"`
}

type RecordAttendanceRequest struct {
	EmployeeID string `json:"employee_id"`
	Date       string `json:"date"`
	Status     string `json:"status"`
	Notes      string `json:"notes"`
}

type CreateLeaveRequest struct {
	EmployeeID string `json:"employee_id"`
	Type       string `json:"type"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Reason     string `json:"reason"`
}

type DecideLeaveRequest struct {
	Status string `json:"status"` // approved|rejected
}
