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
	ParentID          *string   `json:"parent_id,omitempty" db:"parent_id"`
	Level             int       `json:"level" db:"level"`
	Type              string    `json:"type" db:"type"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`

	ParentName       string `json:"parent_name,omitempty" db:"-"`
	TotalCommissions int64  `json:"total_commissions" db:"-"`
	TotalPaid        int64  `json:"total_paid" db:"-"`
	TotalOutstanding int64  `json:"total_outstanding" db:"-"`
	TotalJamaah      int    `json:"total_jamaah" db:"-"`
}

// CommissionTier is a per-org override rate for an upline level (distance from
// the seller). level 2 = direct upline, level 3 = next, etc.
type CommissionTier struct {
	Level   int     `json:"level" db:"level"`
	RatePct float64 `json:"rate_pct" db:"rate_pct"`
}

// DownlineNode is one agent in a hierarchy tree/list (with its depth relative to
// the queried root) plus its commission aggregates.
type DownlineNode struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	ParentID         *string `json:"parent_id,omitempty"`
	Level            int     `json:"level"`
	Depth            int     `json:"depth"` // distance below the queried root (root = 0)
	IsActive         bool    `json:"is_active"`
	TotalCommissions int64   `json:"total_commissions"`
	TotalJamaah      int     `json:"total_jamaah"`
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
	Notes              string     `json:"notes" db:"notes"`
	TierLevel          int        `json:"tier_level" db:"tier_level"`
	SourceCommissionID *string    `json:"source_commission_id,omitempty" db:"source_commission_id"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`

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
	ParentID          string  `json:"parent_id"`
	Type              string  `json:"type"`
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
	ParentID          *string  `json:"parent_id,omitempty"` // "" clears the parent
	Type              *string  `json:"type,omitempty"`
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

type SetTiersRequest struct {
	Tiers []CommissionTier `json:"tiers"`
}
