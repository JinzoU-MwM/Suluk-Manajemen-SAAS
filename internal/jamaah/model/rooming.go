package model

import (
	"time"

	"github.com/google/uuid"
)

// RoomCandidate is a group member enriched with the gender from their jamaah
// profile, used by auto-rooming to split rooms by gender.
type RoomCandidate struct {
	MemberID uuid.UUID `json:"member_id"`
	Name     string    `json:"name"`
	Gender   string    `json:"gender"`
}

type Room struct {
	ID         string    `json:"id"`
	OrgID      string    `json:"org_id"`
	GroupID    *string   `json:"group_id,omitempty"`
	RoomNumber string    `json:"room_number"`
	GenderType string    `json:"gender_type"`
	RoomType   string    `json:"room_type"`
	Capacity   int       `json:"capacity"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type RoomAssignment struct {
	ID       string `json:"id"`
	OrgID    string `json:"org_id"`
	RoomID   string `json:"room_id"`
	MemberID string `json:"member_id"`
}

type SharedManifest struct {
	ID        string     `json:"id"`
	OrgID     string     `json:"org_id"`
	GroupID   *string    `json:"group_id,omitempty"`
	Token     string     `json:"token"`
	PinHash   *string    `json:"-"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateRoomRequest struct {
	RoomNumber string `json:"room_number"`
	GenderType string `json:"gender_type"`
	RoomType   string `json:"room_type"`
	Capacity   int    `json:"capacity"`
}

type AssignMemberRequest struct {
	MemberID string `json:"member_id"`
	RoomID   string `json:"room_id"`
}

type RoomingSummary struct {
	TotalRooms    int `json:"total_rooms"`
	TotalCapacity int `json:"total_capacity"`
	AssignedCount int `json:"assigned_count"`
	OccupancyPct  int `json:"occupancy_pct"`
}
