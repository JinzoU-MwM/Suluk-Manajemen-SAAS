package model

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleOwner   Role = "owner"
	RoleAdmin   Role = "admin"
	RoleFinance Role = "finance"
	RoleCS      Role = "cs"
	RoleViewer  Role = "viewer"
	// RoleAgent is an external B2B agent: they sign in to the agency portal and
	// see only their own subtree. Provisioned via agent-credentials, not normal
	// team invites.
	RoleAgent Role = "agent"
	// RoleJamaah is a pilgrim self-service portal user: they see only their own
	// profile, registrations, documents, visa, and payments.
	RoleJamaah Role = "jamaah"
)

func (r Role) String() string { return string(r) }

func ValidRoles() []string {
	return []string{"owner", "admin", "finance", "cs", "viewer", "agent", "jamaah"}
}

// IsAssignableRole reports whether a role may be granted via team management
// (invite / add-member / update-role). Excludes "owner" (use a dedicated
// ownership transfer) and "agent"/"jamaah" (provisioned via their own
// credential endpoints, never via team management).
func IsAssignableRole(role string) bool {
	switch role {
	case "admin", "editor", "finance", "cs", "viewer":
		return true
	default:
		return false
	}
}

type MemberStatus string

const (
	StatusActive  MemberStatus = "active"
	StatusPending MemberStatus = "pending"
	StatusRemoved MemberStatus = "removed"
)

type User struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	Email            string     `json:"email" db:"email"`
	Name             string     `json:"name" db:"name"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	EmailVerified    bool       `json:"email_verified" db:"email_verified"`
	Phone            *string    `json:"phone,omitempty" db:"phone"`
	PhoneVerified    bool       `json:"phone_verified" db:"phone_verified"`
	City             *string    `json:"city,omitempty" db:"city"`
	Bio              *string    `json:"bio,omitempty" db:"bio"`
	AvatarColor      string     `json:"avatar_color" db:"avatar_color"`
	NotifyUsageLimit bool       `json:"notify_usage_limit" db:"notify_usage_limit"`
	NotifyExpiry     bool       `json:"notify_expiry" db:"notify_expiry"`
	Role             string     `json:"role" db:"role"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	IsSuperAdmin     bool       `json:"is_super_admin" db:"is_super_admin"`
	AgentID          *uuid.UUID `json:"agent_id,omitempty" db:"agent_id"`
	JamaahID         *uuid.UUID `json:"jamaah_id,omitempty" db:"jamaah_id"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type Organization struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	LogoURL     *string   `json:"logo_url,omitempty" db:"logo_url"`
	Address     *string   `json:"address,omitempty" db:"address"`
	Phone       *string   `json:"phone,omitempty" db:"phone"`
	Email       *string   `json:"email,omitempty" db:"email"`
	NPWP        *string   `json:"npwp,omitempty" db:"npwp"`
	PPIUNumber  *string   `json:"ppiu_number,omitempty" db:"ppiu_number"`
	SKNumber    *string   `json:"sk_number,omitempty" db:"sk_number"`
	BankName    *string   `json:"bank_name,omitempty" db:"bank_name"`
	BankAccount *string   `json:"bank_account,omitempty" db:"bank_account"`
	BankHolder  *string   `json:"bank_holder,omitempty" db:"bank_holder"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type TeamMember struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	OrgID     uuid.UUID  `json:"org_id" db:"org_id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Role      string     `json:"role" db:"role"`
	Status    string     `json:"status" db:"status"`
	InvitedBy *uuid.UUID `json:"invited_by,omitempty" db:"invited_by"`
	JoinedAt  time.Time  `json:"joined_at" db:"joined_at"`
}

type TeamInvite struct {
	ID        uuid.UUID `json:"id" db:"id"`
	OrgID     uuid.UUID `json:"org_id" db:"org_id"`
	Email     string    `json:"email" db:"email"`
	Role      string    `json:"role" db:"role"`
	Token     string    `json:"token" db:"token"`
	InvitedBy uuid.UUID `json:"invited_by" db:"invited_by"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type RefreshToken struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	TokenHash  string    `json:"-" db:"token_hash"`
	DeviceInfo *string   `json:"device_info,omitempty" db:"device_info"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type AuditLog struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	OrgID     *uuid.UUID `json:"org_id,omitempty" db:"org_id"`
	UserID    *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Action    string     `json:"action" db:"action"`
	Entity    string     `json:"entity" db:"entity"`
	EntityID  *uuid.UUID `json:"entity_id,omitempty" db:"entity_id"`
	OldValue  any        `json:"old_value,omitempty" db:"old_value"`
	NewValue  any        `json:"new_value,omitempty" db:"new_value"`
	IPAddress string     `json:"ip_address" db:"ip_address"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type Notification struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	OrgID     uuid.UUID  `json:"org_id" db:"org_id"`
	UserID    *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Severity  string     `json:"severity" db:"severity"`
	Title     string     `json:"title" db:"title"`
	Message   string     `json:"message" db:"message"`
	GroupID   *string    `json:"group_id,omitempty" db:"group_id"`
	IsRead    bool       `json:"is_read" db:"is_read"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type Ticket struct {
	ID        uuid.UUID `json:"id" db:"id"`
	OrgID     uuid.UUID `json:"org_id" db:"org_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Subject   string    `json:"subject" db:"subject"`
	Message   string    `json:"message" db:"message"`
	Priority  string    `json:"priority" db:"priority"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type TicketMessage struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TicketID  uuid.UUID `json:"ticket_id" db:"ticket_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateTicketRequest struct {
	Subject  string `json:"subject" validate:"required"`
	Message  string `json:"message" validate:"required"`
	Priority string `json:"priority,omitempty"`
}

type AddTicketMessageRequest struct {
	Content string `json:"content" validate:"required"`
}

type TicketWithMessages struct {
	Ticket   Ticket          `json:"ticket"`
	Messages []TicketMessage `json:"messages"`
}

type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
	Count         int            `json:"count"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Phone    string `json:"phone,omitempty" validate:"omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Subscription struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	OrgID             uuid.UUID  `json:"org_id" db:"org_id"`
	Plan              string     `json:"plan" db:"plan"`
	Status            string     `json:"status" db:"status"`
	StartsAt          time.Time  `json:"starts_at" db:"starts_at"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	TrialUsed         bool       `json:"trial_used" db:"trial_used"`
	CancelAtPeriodEnd bool       `json:"cancel_at_period_end" db:"cancel_at_period_end"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// ExpiringSub is the slice of a subscription the renewal-reminder job needs.
type ExpiringSub struct {
	OrgID     uuid.UUID
	Plan      string
	ExpiresAt time.Time
}

type SubscriptionStatusResponse struct {
	Plan              string     `json:"plan"`
	Status            string     `json:"status"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	Rank              int        `json:"rank"`
	MaxJamaah         int        `json:"max_jamaah"`
	MaxGroups         int        `json:"max_groups"`
	MaxUsers          int        `json:"max_users"`
	CancelAtPeriodEnd bool       `json:"cancel_at_period_end"`
	// UsageCount is the org's AI scans this calendar month (sourced from the
	// ai-ocr service); UsageLimit is the tier's monthly quota (Unlimited = -1).
	// The frontend quota bar reads both off the subscription-status object.
	UsageCount int `json:"usage_count"`
	UsageLimit int `json:"usage_limit"`
}

// ActivatePlanRequest is the body of the internal service-to-service
// activation endpoint called by the payment webhook after a paid order.
type ActivatePlanRequest struct {
	OrgID         string `json:"org_id"`
	UserID        string `json:"user_id"`
	Plan          string `json:"plan"`
	Period        string `json:"period"`
	Amount        int64  `json:"amount"`
	OrderID       string `json:"order_id"`
	PaymentMethod string `json:"payment_method"`
}

type TrialStatusResponse struct {
	TrialAvailable bool `json:"trial_available"`
	TrialDays      int  `json:"trial_days"`
}

type UpgradeRequest struct {
	PaymentRef *string `json:"payment_ref,omitempty"`
}

type CreateOrgRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=255"`
	Address     *string `json:"address,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Email       *string `json:"email,omitempty"`
	BankName    *string `json:"bank_name,omitempty"`
	BankAccount *string `json:"bank_account,omitempty"`
	BankHolder  *string `json:"bank_holder,omitempty"`
}

// UpdateOrgRequest patches editable organization profile fields. Nil fields are
// left unchanged (COALESCE).
type UpdateOrgRequest struct {
	Name        *string `json:"name,omitempty"`
	Address     *string `json:"address,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Email       *string `json:"email,omitempty"`
	NPWP        *string `json:"npwp,omitempty"`
	PPIUNumber  *string `json:"ppiu_number,omitempty"`
	SKNumber    *string `json:"sk_number,omitempty"`
	BankName    *string `json:"bank_name,omitempty"`
	BankAccount *string `json:"bank_account,omitempty"`
	BankHolder  *string `json:"bank_holder,omitempty"`
}

type OtpRecord struct {
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PasswordResetRecord struct {
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

type AddTeamMemberRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Role   string    `json:"role" validate:"required,oneof=owner admin finance cs viewer"`
}

type InviteMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=owner admin finance cs viewer"`
}

// ProvisionAgentRequest creates a B2B portal login bound to an agent record.
type ProvisionAgentRequest struct {
	AgentID  string `json:"agent_id" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name"`
	Password string `json:"password" validate:"required,min=6"`
}

// ProvisionJamaahRequest creates a pilgrim portal login bound to a jamaah profile.
type ProvisionJamaahRequest struct {
	JamaahID string `json:"jamaah_id" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name"`
	Password string `json:"password" validate:"required,min=6"`
}

type AcceptInviteRequest struct {
	Token string `json:"token" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresAt    int64         `json:"expires_at"`
	User         User          `json:"user"`
	Organization *Organization `json:"organization,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// ProfileUpdate carries the user-editable profile fields. A nil pointer means
// "leave unchanged" (partial update).
type ProfileUpdate struct {
	Name             *string `json:"name"`
	Phone            *string `json:"phone"`
	City             *string `json:"city"`
	Bio              *string `json:"bio"`
	AvatarColor      *string `json:"avatar_color"`
	NotifyUsageLimit *bool   `json:"notify_usage_limit"`
	NotifyExpiry     *bool   `json:"notify_expiry"`
}

var ErrNameRequired = errors.New("name is required")

var avatarColors = map[string]bool{
	"emerald": true, "blue": true, "purple": true, "rose": true,
	"amber": true, "cyan": true, "indigo": true, "slate": true,
}

// ApplyProfileUpdate applies the non-nil fields of in onto u, with validation.
func ApplyProfileUpdate(u *User, in ProfileUpdate) error {
	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		if n == "" {
			return ErrNameRequired
		}
		u.Name = n
	}
	if in.Phone != nil {
		p := strings.TrimSpace(*in.Phone)
		u.Phone = &p
	}
	if in.City != nil {
		c := strings.TrimSpace(*in.City)
		u.City = &c
	}
	if in.Bio != nil {
		b := strings.TrimSpace(*in.Bio)
		u.Bio = &b
	}
	if in.AvatarColor != nil {
		c := strings.TrimSpace(*in.AvatarColor)
		if !avatarColors[c] {
			c = "blue"
		}
		u.AvatarColor = c
	}
	if in.NotifyUsageLimit != nil {
		u.NotifyUsageLimit = *in.NotifyUsageLimit
	}
	if in.NotifyExpiry != nil {
		u.NotifyExpiry = *in.NotifyExpiry
	}
	return nil
}
