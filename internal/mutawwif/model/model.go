package model

import (
	"time"

	"github.com/google/uuid"
)

// Guide is a tour guide / mutawwif on the roster.
type Guide struct {
	ID            uuid.UUID  `json:"id"`
	OrgID         uuid.UUID  `json:"org_id"`
	Name          string     `json:"name"`
	Phone         string     `json:"phone"`
	Email         string     `json:"email"`
	Type          string     `json:"type"` // mutawwif|tour_leader|kesehatan
	LicenseNo     string     `json:"license_no"`
	LicenseExpiry *time.Time `json:"license_expiry,omitempty"`
	IsActive      bool       `json:"is_active"`
	Notes         string     `json:"notes"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// AssignmentCount is hydrated on list (how many kloter this guide leads).
	AssignmentCount int `json:"assignment_count"`
}

// GuideAssignment links a guide to a departure group (kloter).
type GuideAssignment struct {
	ID         uuid.UUID `json:"id"`
	OrgID      uuid.UUID `json:"org_id"`
	GuideID    uuid.UUID `json:"guide_id"`
	GroupID    uuid.UUID `json:"group_id"`
	Role       string    `json:"role"` // leader|co_leader|kesehatan
	AssignedAt time.Time `json:"assigned_at"`
	CreatedAt  time.Time `json:"created_at"`

	// GuideName/GuideType/GuidePhone hydrated when listing a group's guides.
	GuideName  string `json:"guide_name,omitempty"`
	GuideType  string `json:"guide_type,omitempty"`
	GuidePhone string `json:"guide_phone,omitempty"`
}

type CreateGuideRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=255"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Type          string `json:"type"`
	LicenseNo     string `json:"license_no"`
	LicenseExpiry string `json:"license_expiry"`
	Notes         string `json:"notes"`
}

type UpdateGuideRequest struct {
	Name          *string `json:"name,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Email         *string `json:"email,omitempty"`
	Type          *string `json:"type,omitempty"`
	LicenseNo     *string `json:"license_no,omitempty"`
	LicenseExpiry *string `json:"license_expiry,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}

type AssignGuideRequest struct {
	GuideID string `json:"guide_id" validate:"required"`
	GroupID string `json:"group_id" validate:"required"`
	Role    string `json:"role"`
}

type GuideListResponse struct {
	Guides []Guide `json:"guides"`
	Total  int     `json:"total"`
}
