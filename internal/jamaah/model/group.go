package model

import (
	"time"

	"github.com/google/uuid"
)

type Group struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	OrgID               uuid.UUID  `json:"org_id" db:"org_id"`
	Name                string     `json:"name" db:"name"`
	Description         string     `json:"description" db:"description"`
	MemberCount         int        `json:"member_count" db:"member_count"`
	IsActive            bool       `json:"is_active" db:"is_active"`
	PackageID           *uuid.UUID `json:"package_id,omitempty" db:"package_id"`
	DepartureDate       *time.Time `json:"departure_date,omitempty" db:"departure_date"`
	DepartureStatus     string     `json:"departure_status" db:"departure_status"`
	ManifestFinalizedAt *time.Time `json:"manifest_finalized_at,omitempty" db:"manifest_finalized_at"`
	DepartedAt          *time.Time `json:"departed_at,omitempty" db:"departed_at"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// Kloter / departure status workflow.
type DepartureStatus string

const (
	DepartureDraft     DepartureStatus = "draft"     // assembling members
	DepartureSiap      DepartureStatus = "siap"      // manifest finalized, ready
	DepartureBerangkat DepartureStatus = "berangkat" // departed
	DepartureSelesai   DepartureStatus = "selesai"   // trip completed
	DepartureBatal     DepartureStatus = "batal"     // cancelled
)

var departureTransitions = map[DepartureStatus]map[DepartureStatus]bool{
	DepartureDraft:     {DepartureSiap: true, DepartureBatal: true},
	DepartureSiap:      {DepartureBerangkat: true, DepartureDraft: true, DepartureBatal: true}, // reopen or depart
	DepartureBerangkat: {DepartureSelesai: true},
	DepartureSelesai:   {},
	DepartureBatal:     {DepartureDraft: true}, // reactivate a cancelled kloter
}

// CanTransitionDeparture reports whether from→to is a legal departure transition.
func CanTransitionDeparture(from, to DepartureStatus) bool {
	if from == to {
		return false
	}
	return departureTransitions[from][to]
}

// DepartureManifest is the boarding view for a finalized/active kloter.
type DepartureManifest struct {
	Group   Group         `json:"group"`
	Members []GroupMember `json:"members"`
}

type SetDepartureRequest struct {
	PackageID     string `json:"package_id,omitempty"`
	DepartureDate string `json:"departure_date,omitempty"`
}

type DepartureTransitionRequest struct {
	Status string `json:"status" validate:"required,oneof=draft siap berangkat selesai batal"`
}

type GroupMember struct {
	ID        uuid.UUID `json:"id" db:"id"`
	OrgID     uuid.UUID `json:"org_id" db:"org_id"`
	GroupID   uuid.UUID `json:"group_id" db:"group_id"`
	MemberID  uuid.UUID `json:"member_id" db:"member_id"`
	Name      string    `json:"name" db:"name"`
	Phone     string    `json:"phone" db:"phone"`
	Notes     string    `json:"notes" db:"notes"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateGroupRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type AddGroupMemberRequest struct {
	MemberID uuid.UUID `json:"member_id"`
	Name     string    `json:"name"`
	Phone    string    `json:"phone"`
	Notes    string    `json:"notes"`
}

type UpdateGroupMemberRequest struct {
	Name  *string `json:"name,omitempty"`
	Phone *string `json:"phone,omitempty"`
	Notes *string `json:"notes,omitempty"`
}
