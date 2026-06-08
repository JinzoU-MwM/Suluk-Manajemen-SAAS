package model

import "time"

type RegistrationLink struct {
	ID        string    `json:"id" db:"id"`
	OrgID     string    `json:"org_id" db:"org_id"`
	GroupID   *string   `json:"group_id,omitempty" db:"group_id"`
	PackageID *string   `json:"package_id,omitempty" db:"package_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedBy *string   `json:"created_by,omitempty" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type PendingRegistration struct {
	ID                 string     `json:"id" db:"id"`
	OrgID              string     `json:"org_id" db:"org_id"`
	RegistrationLinkID *string    `json:"registration_link_id,omitempty" db:"registration_link_id"`
	PhoneNumber        string     `json:"phone_number" db:"phone_number"`
	Name               string     `json:"name" db:"name"`
	Email              string     `json:"email" db:"email"`
	KtpFileURL         string     `json:"ktp_file_url" db:"ktp_file_url"`
	PassportFileURL    string     `json:"passport_file_url" db:"passport_file_url"`
	VisaFileURL        string     `json:"visa_file_url" db:"visa_file_url"`
	Notes              string     `json:"notes" db:"notes"`
	Status             string     `json:"status" db:"status"`
	JamaahID           *string    `json:"jamaah_id,omitempty" db:"jamaah_id"`
	ReviewedBy         *string    `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt         *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

type GenerateLinkRequest struct {
	GroupID       string `json:"group_id"`
	PackageID     string `json:"package_id"`
	ExpiresInDays int    `json:"expires_in_days"`
}

type PublicRegistrationInfo struct {
	GroupName   string    `json:"group_name"`
	PackageName string    `json:"package_name"`
	OrgName     string    `json:"org_name"`
	ExpiresAt   time.Time `json:"expires_at"`
	IsExpired   bool      `json:"is_expired"`
}

type PublicRegistrationSubmit struct {
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
	Email       string `json:"email"`
}
