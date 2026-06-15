package model

import (
	"time"

	"github.com/google/uuid"
)

// Visa application lifecycle (Phase 4B).
type VisaStatus string

const (
	VisaDraft     VisaStatus = "draft"
	VisaSubmitted VisaStatus = "submitted"
	VisaApproved  VisaStatus = "approved"
	VisaRejected  VisaStatus = "rejected"
	VisaExpired   VisaStatus = "expired"
)

func ValidVisaStatuses() []string {
	return []string{"draft", "submitted", "approved", "rejected", "expired"}
}

// visaTransitions is the allowed state machine: from → set of valid next states.
var visaTransitions = map[VisaStatus]map[VisaStatus]bool{
	VisaDraft:     {VisaSubmitted: true},
	VisaSubmitted: {VisaApproved: true, VisaRejected: true},
	VisaRejected:  {VisaSubmitted: true}, // resubmit after fixing
	VisaApproved:  {VisaExpired: true},   // by date (or manual)
	VisaExpired:   {VisaSubmitted: true}, // renew
}

// CanTransitionVisa reports whether from→to is a legal visa transition.
func CanTransitionVisa(from, to VisaStatus) bool {
	if from == to {
		return false
	}
	return visaTransitions[from][to]
}

type VisaApplication struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OrgID        uuid.UUID  `json:"org_id" db:"org_id"`
	JamaahID     uuid.UUID  `json:"jamaah_id" db:"jamaah_id"`
	PackageID    *uuid.UUID `json:"package_id,omitempty" db:"package_id"`
	Status       string     `json:"status" db:"status"`
	Provider     string     `json:"provider" db:"provider"`
	ReferenceNo  string     `json:"reference_no" db:"reference_no"`
	SubmittedAt  *time.Time `json:"submitted_at,omitempty" db:"submitted_at"`
	DecidedAt    *time.Time `json:"decided_at,omitempty" db:"decided_at"`
	ExpiryDate   *time.Time `json:"expiry_date,omitempty" db:"expiry_date"`
	RejectReason string     `json:"reject_reason" db:"reject_reason"`
	Notes        string     `json:"notes" db:"notes"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`

	// JamaahName is hydrated for the board/list view (join), not stored.
	JamaahName string `json:"jamaah_name,omitempty" db:"-"`
}

type VisaHistory struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	OrgID      uuid.UUID  `json:"org_id" db:"org_id"`
	VisaID     uuid.UUID  `json:"visa_id" db:"visa_id"`
	JamaahID   uuid.UUID  `json:"jamaah_id" db:"jamaah_id"`
	FromStatus *string    `json:"from_status,omitempty" db:"from_status"`
	ToStatus   string     `json:"to_status" db:"to_status"`
	Reason     *string    `json:"reason,omitempty" db:"reason"`
	ChangedBy  *uuid.UUID `json:"changed_by,omitempty" db:"changed_by"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// UpsertVisaRequest creates or edits the draft application's editable fields.
type UpsertVisaRequest struct {
	PackageID   string `json:"package_id,omitempty"`
	Provider    string `json:"provider,omitempty"`
	ReferenceNo string `json:"reference_no,omitempty"`
	ExpiryDate  string `json:"expiry_date,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// VisaTransitionRequest moves the application to a new status.
type VisaTransitionRequest struct {
	Status      string `json:"status" validate:"required,oneof=draft submitted approved rejected expired"`
	Reason      string `json:"reason,omitempty"`
	ReferenceNo string `json:"reference_no,omitempty"` // set on approve
	ExpiryDate  string `json:"expiry_date,omitempty"`  // set on approve
}
