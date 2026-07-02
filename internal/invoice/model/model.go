package model

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceStatusBelumBayar InvoiceStatus = "belum_bayar"
	InvoiceStatusSebagian   InvoiceStatus = "sebagian"
	InvoiceStatusLunas      InvoiceStatus = "lunas"
	InvoiceStatusBatal      InvoiceStatus = "batal"
)

type PaymentScheme string

const (
	SchemeDPAndLunas PaymentScheme = "dp_lunas"
	SchemeCicilan    PaymentScheme = "cicilan"
	SchemeFull       PaymentScheme = "full"
)

type PaymentMethod string

const (
	MethodTransferBank PaymentMethod = "transfer_bank"
	MethodTunai        PaymentMethod = "tunai"
	MethodQRIS         PaymentMethod = "qris"
	MethodKartuKredit  PaymentMethod = "kartu_kredit"
	MethodEWallet      PaymentMethod = "e_wallet"
)

func ValidInvoiceStatuses() []string {
	return []string{"belum_bayar", "sebagian", "lunas", "batal"}
}

func ValidPaymentSchemes() []string {
	return []string{"dp_lunas", "cicilan", "full"}
}

func ValidPaymentMethods() []string {
	return []string{"transfer_bank", "tunai", "qris", "kartu_kredit", "e_wallet"}
}

type Invoice struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	OrgID           uuid.UUID  `json:"org_id" db:"org_id"`
	InvoiceNumber   string     `json:"invoice_number" db:"invoice_number"`
	JamaahID        uuid.UUID  `json:"jamaah_id" db:"jamaah_id"`
	JamaahName      string     `json:"jamaah_name,omitempty" db:"jamaah_name"`
	PackageID       uuid.UUID  `json:"package_id" db:"package_id"`
	PackageName     string     `json:"package_name,omitempty" db:"package_name"`
	RegistrationID  uuid.UUID  `json:"registration_id" db:"registration_id"`
	RoomType        string     `json:"room_type" db:"room_type"`
	PriceSnapshot   int64      `json:"price_snapshot" db:"price_snapshot"`
	DiscountAmount  int64      `json:"discount_amount" db:"discount_amount"`
	SurchargeAmount int64      `json:"surcharge_amount" db:"surcharge_amount"`
	TotalAmount     int64      `json:"total_amount" db:"total_amount"`
	AmountPaid      int64      `json:"amount_paid" db:"amount_paid"`
	AmountRemaining int64      `json:"amount_remaining" db:"amount_remaining"`
	PaymentScheme   string     `json:"payment_scheme" db:"payment_scheme"`
	Status          string     `json:"status" db:"status"`
	IssuedAt        time.Time  `json:"issued_at" db:"issued_at"`
	DueDate         *time.Time `json:"due_date,omitempty" db:"due_date"`
	CancelledAt     *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CancelledReason *string    `json:"cancelled_reason,omitempty" db:"cancelled_reason"`
	Notes           string     `json:"notes,omitempty" db:"notes"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`

	PaymentSchedules []PaymentSchedule `json:"payment_schedules,omitempty" db:"-"`
	Payments         []Payment         `json:"payments,omitempty" db:"-"`
}

type PaymentSchedule struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	InvoiceID      uuid.UUID  `json:"invoice_id" db:"invoice_id"`
	InstallmentNum int        `json:"installment_num" db:"installment_num"`
	Amount         int64      `json:"amount" db:"amount"`
	DueDate        *time.Time `json:"due_date,omitempty" db:"due_date"`
	Description    *string    `json:"description,omitempty" db:"description"`
	IsPaid         bool       `json:"is_paid" db:"is_paid"`
	PaidAt         *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type Payment struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	OrgID           uuid.UUID  `json:"org_id" db:"org_id"`
	InvoiceID       uuid.UUID  `json:"invoice_id" db:"invoice_id"`
	Amount          int64      `json:"amount" db:"amount"`
	PaymentMethod   string     `json:"payment_method" db:"payment_method"`
	BankName        *string    `json:"bank_name,omitempty" db:"bank_name"`
	AccountNumber   *string    `json:"account_number,omitempty" db:"account_number"`
	ReferenceNumber *string    `json:"reference_number,omitempty" db:"reference_number"`
	ProofURL        *string    `json:"proof_url,omitempty" db:"proof_url"`
	Notes           string     `json:"notes,omitempty" db:"notes"`
	ReceivedBy      uuid.UUID  `json:"received_by" db:"received_by"`
	PaidAt          time.Time  `json:"paid_at" db:"paid_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	CashSessionID   *uuid.UUID `json:"cash_session_id,omitempty" db:"cash_session_id"`
}

// CashSession is a kasir (POS) cash-drawer session.
type CashSession struct {
	ID           uuid.UUID  `json:"id"`
	OrgID        uuid.UUID  `json:"org_id"`
	UserID       uuid.UUID  `json:"user_id"`
	OpeningFloat int64      `json:"opening_float"`
	ExpectedCash *int64     `json:"expected_cash,omitempty"`
	CountedCash  *int64     `json:"counted_cash,omitempty"`
	Difference   *int64     `json:"difference,omitempty"`
	Status       string     `json:"status"`
	OpenedAt     time.Time  `json:"opened_at"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
	Notes        string     `json:"notes"`
}

type CreateInvoiceRequest struct {
	JamaahID        uuid.UUID `json:"jamaah_id" validate:"required"`
	JamaahName      string    `json:"jamaah_name,omitempty"`
	PackageID       uuid.UUID `json:"package_id" validate:"required"`
	PackageName     string    `json:"package_name,omitempty"`
	RegistrationID  uuid.UUID `json:"registration_id" validate:"required"`
	RoomType        string    `json:"room_type" validate:"required,oneof=quad triple double single"`
	PriceSnapshot   int64     `json:"price_snapshot" validate:"min=1"`
	DiscountAmount  int64     `json:"discount_amount,omitempty"`
	SurchargeAmount int64     `json:"surcharge_amount,omitempty"`
	PaymentScheme   string    `json:"payment_scheme" validate:"required,oneof=dp_lunas cicilan full"`
	DueDate         string    `json:"due_date,omitempty"`
	Notes           string    `json:"notes,omitempty"`
}

type UpdateInvoiceRequest struct {
	DiscountAmount  *int64  `json:"discount_amount,omitempty"`
	SurchargeAmount *int64  `json:"surcharge_amount,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	DueDate         *string `json:"due_date,omitempty"`
}

type CancelInvoiceRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type CreatePaymentScheduleRequest struct {
	Installments []ScheduleInstallment `json:"installments" validate:"required,dive"`
}

type ScheduleInstallment struct {
	Amount      int64  `json:"amount" validate:"min=1"`
	DueDate     string `json:"due_date,omitempty"`
	Description string `json:"description,omitempty"`
}

type RecordPaymentRequest struct {
	Amount          int64  `json:"amount" validate:"min=1"`
	PaymentMethod   string `json:"payment_method" validate:"required,oneof=transfer_bank tunai qris kartu_kredit e_wallet"`
	BankName        string `json:"bank_name,omitempty"`
	AccountNumber   string `json:"account_number,omitempty"`
	ReferenceNumber string `json:"reference_number,omitempty"`
	ProofURL        string `json:"proof_url,omitempty"`
	Notes           string `json:"notes,omitempty"`
	PaidAt          string `json:"paid_at,omitempty"`
}

type InvoiceSummary struct {
	TotalInvoices    int64 `json:"total_invoices"`
	TotalAmount      int64 `json:"total_amount"`
	TotalPaid        int64 `json:"total_paid"`
	TotalRemaining   int64 `json:"total_remaining"`
	OutstandingCount int64 `json:"outstanding_count"`
	OverdueCount     int64 `json:"overdue_count"`
}

// JamaahBalance is the per-jamaah invoice aggregate used by the CRM list to show
// each jamaah's outstanding balance without an N+1 of per-jamaah requests.
type JamaahBalance struct {
	JamaahID       uuid.UUID `json:"jamaah_id"`
	TotalAmount    int64     `json:"total_amount"`
	TotalPaid      int64     `json:"total_paid"`
	TotalRemaining int64     `json:"total_remaining"`
}

// MonthlyRevenuePoint is one bar of the owner dashboard revenue trend chart.
// Total is the sum of payments received in that calendar month.
type MonthlyRevenuePoint struct {
	Month string `json:"month"`
	Year  int    `json:"year"`
	Total int64  `json:"total"`
}

type PackageRevenueSummary struct {
	PackageID       uuid.UUID `json:"package_id"`
	TotalInvoices   int       `json:"total_invoices"`
	TotalAmount     int64     `json:"total_amount"`
	TotalPaid       int64     `json:"total_paid"`
	TotalRemaining  int64     `json:"total_remaining"`
	LunasCount      int       `json:"lunas_count"`
	SebagianCount   int       `json:"sebagian_count"`
	BelumBayarCount int       `json:"belum_bayar_count"`
	BatalCount      int       `json:"batal_count"`
}

type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusApproved  RefundStatus = "approved"
	RefundStatusProcessed RefundStatus = "processed"
	RefundStatusCompleted RefundStatus = "completed"
	RefundStatusRejected  RefundStatus = "rejected"
)

type RefundPolicy struct {
	ID          uuid.UUID `json:"id" db:"id"`
	OrgID       uuid.UUID `json:"org_id" db:"org_id"`
	Name        string    `json:"name" db:"name"`
	DaysBefore  int       `json:"days_before" db:"days_before"`
	RefundPct   float64   `json:"refund_pct" db:"refund_pct"`
	Description string    `json:"description" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Refund struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrgID         uuid.UUID  `json:"org_id" db:"org_id"`
	InvoiceID     uuid.UUID  `json:"invoice_id" db:"invoice_id"`
	InvoiceNumber string     `json:"invoice_number,omitempty" db:"-"`
	JamaahName    string     `json:"jamaah_name,omitempty" db:"-"`
	PackageName   string     `json:"package_name,omitempty" db:"-"`
	Amount        int64      `json:"amount" db:"amount"`
	RefundPct     float64    `json:"refund_pct" db:"refund_pct"`
	PolicyID      *uuid.UUID `json:"policy_id,omitempty" db:"policy_id"`
	PaymentMethod string     `json:"payment_method" db:"payment_method"`
	Reason        string     `json:"reason" db:"reason"`
	Status        string     `json:"status" db:"status"`
	ApprovedBy    *uuid.UUID `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty" db:"approved_at"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	Notes         string     `json:"notes" db:"notes"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateRefundPolicyRequest struct {
	Name        string  `json:"name"`
	DaysBefore  int     `json:"days_before"`
	RefundPct   float64 `json:"refund_pct"`
	Description string  `json:"description"`
}

type UpdateRefundPolicyRequest struct {
	Name        *string  `json:"name,omitempty"`
	DaysBefore  *int     `json:"days_before,omitempty"`
	RefundPct   *float64 `json:"refund_pct,omitempty"`
	Description *string  `json:"description,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

type InitiateRefundRequest struct {
	Amount    int64   `json:"amount"`
	RefundPct float64 `json:"refund_pct"`
	Reason    string  `json:"reason"`
	Notes     string  `json:"notes"`
}

type RefundActionRequest struct {
	Notes string `json:"notes"`
}

type RefundListResponse struct {
	Refunds []Refund `json:"refunds"`
	Total   int64    `json:"total"`
}

type PaymentOrder struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrgID         uuid.UUID  `json:"org_id" db:"org_id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	Plan          string     `json:"plan" db:"plan"`           // tier: starter/pro/bisnis
	PlanType      string     `json:"plan_type" db:"plan_type"` // period: monthly/yearly
	Amount        int64      `json:"amount" db:"amount"`
	Status        string     `json:"status" db:"status"`
	Purpose       string     `json:"purpose" db:"purpose"` // "subscription" | "scan_topup"
	RedirectURL   *string    `json:"redirect_url,omitempty" db:"redirect_url"`
	GatewayRef    *string    `json:"gateway_ref,omitempty" db:"gateway_ref"`
	PaymentMethod *string    `json:"payment_method,omitempty" db:"payment_method"`
	CompletedAt   *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type CreatePaymentOrderRequest struct {
	Plan     string `json:"plan"`      // tier: starter/pro/bisnis
	PlanType string `json:"plan_type"` // period: monthly/yearly
}

type PaymentOrderResponse struct {
	OrderID    string `json:"order_id"`
	PaymentURL string `json:"payment_url"`
	Status     string `json:"status"`
	Amount     int64  `json:"amount"`
}

// ActivatePlanBody is the payload sent to the auth-service internal activation endpoint.
type ActivatePlanBody struct {
	OrgID         string `json:"org_id"`
	UserID        string `json:"user_id"`
	Plan          string `json:"plan"`
	Period        string `json:"period"`
	Amount        int64  `json:"amount"`
	OrderID       string `json:"order_id"`
	PaymentMethod string `json:"payment_method"`
}

// PakasirWebhookPayload is the POST body Pakasir sends on a completed payment.
type PakasirWebhookPayload struct {
	Amount        int64  `json:"amount"`
	OrderID       string `json:"order_id"`
	Project       string `json:"project"`
	Status        string `json:"status"`
	PaymentMethod string `json:"payment_method"`
	CompletedAt   string `json:"completed_at"`
}

type PaymentStatusResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Amount  int64  `json:"amount"`
}
