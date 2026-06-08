package model

import (
	"time"

	"github.com/google/uuid"
)

type Itinerary struct {
	ID          uuid.UUID `json:"id" db:"id"`
	OrgID       uuid.UUID `json:"org_id" db:"org_id"`
	GroupID     uuid.UUID `json:"group_id" db:"group_id"`
	DayNumber   int       `json:"day_number" db:"day_number"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Location    string    `json:"location" db:"location"`
	StartTime   *string   `json:"start_time,omitempty" db:"start_time"`
	EndTime     *string   `json:"end_time,omitempty" db:"end_time"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateItineraryRequest struct {
	DayNumber   int     `json:"day_number"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	StartTime   *string `json:"start_time,omitempty"`
	EndTime     *string `json:"end_time,omitempty"`
	SortOrder   int     `json:"sort_order"`
}

type UpdateItineraryRequest struct {
	DayNumber   *int    `json:"day_number,omitempty"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Location    *string `json:"location,omitempty"`
	StartTime   *string `json:"start_time,omitempty"`
	EndTime     *string `json:"end_time,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}
