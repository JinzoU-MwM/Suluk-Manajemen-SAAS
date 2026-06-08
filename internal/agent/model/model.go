package model

import "time"

type Agent struct {
	ID                string    `json:"id" db:"id"`
	OrgID             string    `json:"org_id" db:"org_id"`
	Name              string    `json:"name" db:"name"`
	Phone             string    `json:"phone" db:"phone"`
	Email             string    `json:"email" db:"email"`
	Address           string    `json:"address" db:"address"`
	CommissionRate    float64   `json:"commission_rate" db:"commission_rate"`
	BankName          string    `json:"bank_name" db:"bank_name"`
	BankAccountNumber string    `json:"bank_account_number" db:"bank_account_number"`
	BankAccountName   string    `json:"bank_account_name" db:"bank_account_name"`
	Notes             string    `json:"notes" db:"notes"`
	IsActive          bool      `json:"is_active" db:"is_active"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`

	TotalCommissions int64 `json:"total_commissions" db:"-"`
	TotalPaid        int64 `json:"total_paid" db:"-"`
	TotalOutstanding int64 `json:"total_outstanding" db:"-"`
	TotalJamaah      int   `json:"total_jamaah" db:"-"`
}

type AgentCommission struct {
	ID               string     `json:"id" db:"id"`
	OrgID            string     `json:"org_id" db:"org_id"`
	AgentID          string     `json:"agent_id" db:"agent_id"`
	JamaahID         *string    `json:"jamaah_id,omitempty" db:"jamaah_id"`
	InvoiceID        *string    `json:"invoice_id,omitempty" db:"invoice_id"`
	PackageID        *string    `json:"package_id,omitempty" db:"package_id"`
	JamaahName       string     `json:"jamaah_name" db:"jamaah_name"`
	PackageName      string     `json:"package_name" db:"package_name"`
	CommissionAmount int64      `json:"commission_amount" db:"commission_amount"`
	CommissionRate   float64    `json:"commission_rate" db:"commission_rate"`
	Status           string     `json:"status" db:"status"`
	PaidAt           *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	Notes            string     `json:"notes" db:"notes"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`

	AgentName string `json:"agent_name,omitempty" db:"-"`
}

type CreateAgentRequest struct {
	Name              string  `json:"name"`
	Phone             string  `json:"phone"`
	Email             string  `json:"email"`
	Address           string  `json:"address"`
	CommissionRate    float64 `json:"commission_rate"`
	BankName          string  `json:"bank_name"`
	BankAccountNumber string  `json:"bank_account_number"`
	BankAccountName   string  `json:"bank_account_name"`
	Notes             string  `json:"notes"`
}

type UpdateAgentRequest struct {
	Name              *string  `json:"name,omitempty"`
	Phone             *string  `json:"phone,omitempty"`
	Email             *string  `json:"email,omitempty"`
	Address           *string  `json:"address,omitempty"`
	CommissionRate    *float64 `json:"commission_rate,omitempty"`
	BankName          *string  `json:"bank_name,omitempty"`
	BankAccountNumber *string  `json:"bank_account_number,omitempty"`
	BankAccountName   *string  `json:"bank_account_name,omitempty"`
	Notes             *string  `json:"notes,omitempty"`
	IsActive          *bool    `json:"is_active,omitempty"`
}

type CreateCommissionRequest struct {
	AgentID          string  `json:"agent_id"`
	JamaahID         string  `json:"jamaah_id"`
	InvoiceID        string  `json:"invoice_id"`
	PackageID        string  `json:"package_id"`
	JamaahName       string  `json:"jamaah_name"`
	PackageName      string  `json:"package_name"`
	CommissionAmount int64   `json:"commission_amount"`
	CommissionRate   float64 `json:"commission_rate"`
	Notes            string  `json:"notes"`
}

type PayCommissionRequest struct {
	Notes string `json:"notes"`
}

type AgentListResponse struct {
	Agents []Agent `json:"agents"`
	Total  int64   `json:"total"`
}

type CommissionListResponse struct {
	Commissions []AgentCommission `json:"commissions"`
	Total       int64             `json:"total"`
}
