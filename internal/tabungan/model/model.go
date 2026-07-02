// Package model holds the savings (tabungan) domain types.
package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusAktif     = "aktif"
	StatusConverted = "converted"
	StatusClosed    = "closed"
)

type SavingsAccount struct {
	ID              uuid.UUID  `json:"id"`
	OrgID           uuid.UUID  `json:"org_id"`
	JamaahID        uuid.UUID  `json:"jamaah_id"`
	JamaahName      string     `json:"jamaah_name"`
	TargetPackageID *uuid.UUID `json:"target_package_id,omitempty"`
	TargetAmount    int64      `json:"target_amount"`
	Balance         int64      `json:"balance"`
	Status          string     `json:"status"`
	Notes           string     `json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Deposits        []Deposit  `json:"deposits,omitempty"`
}

type Deposit struct {
	ID             uuid.UUID  `json:"id"`
	AccountID      uuid.UUID  `json:"account_id"`
	OrgID          uuid.UUID  `json:"org_id"`
	Amount         int64      `json:"amount"`
	Direction      string     `json:"direction"` // in | out
	Type           string     `json:"type"`      // setor | konversi | tarik
	Method         string     `json:"method"`
	Reference      string     `json:"reference"`
	Notes          string     `json:"notes"`
	IdempotencyKey string     `json:"idempotency_key,omitempty"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CreateAccountRequest struct {
	JamaahID        string `json:"jamaah_id" validate:"required"`
	JamaahName      string `json:"jamaah_name"`
	TargetPackageID string `json:"target_package_id"`
	TargetAmount    int64  `json:"target_amount"`
	Notes           string `json:"notes"`
}

type DepositRequest struct {
	Amount         int64  `json:"amount" validate:"min=1"`
	Method         string `json:"method"`
	Reference      string `json:"reference"`
	Notes          string `json:"notes"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type ConvertRequest struct {
	InvoiceID string `json:"invoice_id" validate:"required"`
	Amount    int64  `json:"amount"` // optional; default = min(balance, requested)
}
