package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/auth/model"
	"github.com/jamaah-in/v2/internal/auth/service"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		return response.BadRequest(c, "name, email, and password are required")
	}
	if len(req.Password) < 8 {
		return response.BadRequest(c, "password must be at least 8 characters")
	}

	user, org, tokens, err := h.svc.Register(c.Context(), req)
	if err != nil {
		return response.Internal(c, err)
	}

	return response.Created(c, fiber.Map{
		"user":          sanitizeUser(user),
		"organization":  org,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if req.Email == "" || req.Password == "" {
		return response.BadRequest(c, "email and password are required")
	}

	user, org, tokens, err := h.svc.Login(c.Context(), req)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	return response.OK(c, fiber.Map{
		"user":          sanitizeUser(user),
		"organization":  org,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt,
	})
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	tokens, err := h.svc.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	return response.OK(c, fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Get("X-Refresh-Token", "")
	h.svc.Logout(c.Context(), refreshToken)
	return response.OK(c, fiber.Map{"message": "logged out"})
}

func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	user, err := h.svc.GetUser(c.Context(), claims.UserID)
	if err != nil {
		return response.NotFound(c, "user not found")
	}
	return response.OK(c, sanitizeUser(user))
}

func (h *AuthHandler) UpdateMe(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)

	var req struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	user, err := h.svc.UpdateUser(c.Context(), claims.UserID, req.Name, req.Phone)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, sanitizeUser(user))
}

func (h *AuthHandler) CreateOrganization(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)

	var req model.CreateOrgRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "organization name is required")
	}

	org, err := h.svc.CreateOrganization(c.Context(), claims.UserID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Created(c, org)
}

func (h *AuthHandler) GetOrganization(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	org, err := h.svc.GetOrganization(c.Context(), claims.OrgID)
	if err != nil {
		return response.NotFound(c, "organization not found")
	}
	return response.OK(c, org)
}

func (h *AuthHandler) UpdateOrganization(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.UpdateOrgRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	org, err := h.svc.UpdateOrganization(c.Context(), claims.OrgID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, org)
}

func (h *AuthHandler) AddTeamMember(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)

	var req model.AddTeamMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	member, err := h.svc.AddTeamMember(c.Context(), claims.OrgID, req.UserID, claims.UserID, req.Role)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Created(c, member)
}

// ProvisionAgentCredential creates a B2B portal login bound to an agent record
// (owner/admin only).
func (h *AuthHandler) ProvisionAgentCredential(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.ProvisionAgentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	user, err := h.svc.ProvisionAgentCredential(c.Context(), claims.OrgID, claims.UserID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, sanitizeUser(user))
}

func (h *AuthHandler) RemoveTeamMember(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return response.BadRequest(c, "invalid user id")
	}

	if err := h.svc.RemoveTeamMember(c.Context(), claims.OrgID, userID); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "member removed"})
}

func (h *AuthHandler) UpdateMemberRole(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return response.BadRequest(c, "invalid user id")
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if err := h.svc.UpdateMemberRole(c.Context(), claims.OrgID, userID, req.Role); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "role updated"})
}

func (h *AuthHandler) ListTeamMembers(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	members, err := h.svc.ListTeamMembers(c.Context(), claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, members)
}

func (h *AuthHandler) ListUsersByOrg(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	users, err := h.svc.ListUsersByOrg(c.Context(), claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	sanitized := make([]fiber.Map, len(users))
	for i, u := range users {
		sanitized[i] = sanitizeUserMap(&u)
	}
	return response.OK(c, sanitized)
}

func (h *AuthHandler) InviteMember(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)

	var req model.InviteMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	invite, err := h.svc.InviteMember(c.Context(), claims.OrgID, claims.UserID, req.Email, req.Role)
	if err != nil {
		if errors.Is(err, service.ErrPlanLimit) {
			return response.BadRequest(c, err.Error())
		}
		return response.Internal(c, err)
	}
	return response.Created(c, invite)
}

func (h *AuthHandler) AcceptInvite(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)

	var req model.AcceptInviteRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	member, err := h.svc.AcceptInvite(c.Context(), req.Token, claims.UserID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Created(c, member)
}

func (h *AuthHandler) CancelInvite(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	inviteID, err := uuid.Parse(c.Params("inviteId"))
	if err != nil {
		return response.BadRequest(c, "invalid invite id")
	}
	if err := h.svc.CancelInvite(c.Context(), inviteID, claims.OrgID); err != nil {
		return response.NotFound(c, "invite not found or already processed")
	}
	return response.OK(c, fiber.Map{"message": "invite cancelled"})
}

// ── Notifications ─────────────────────────────────────────

func (h *AuthHandler) ListNotifications(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	notifications, count, err := h.svc.GetNotifications(c.Context(), claims.OrgID, claims.UserID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, model.NotificationsResponse{
		Notifications: notifications,
		Count:         count,
	})
}

func (h *AuthHandler) MarkNotificationRead(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid notification id")
	}
	if err := h.svc.MarkNotificationRead(c.Context(), id, claims.UserID); err != nil {
		return response.NotFound(c, "notification not found")
	}
	return response.OK(c, fiber.Map{"message": "notification marked as read"})
}

func (h *AuthHandler) MarkAllNotificationsRead(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	if err := h.svc.MarkAllNotificationsRead(c.Context(), claims.OrgID, claims.UserID); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "all notifications marked as read"})
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.ChangePassword(c.Context(), claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "password changed"})
}

func (h *AuthHandler) GetActivity(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	logs, err := h.svc.GetActivity(c.Context(), claims.UserID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"activities": logs})
}

func (h *AuthHandler) DeleteAccount(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req struct {
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.DeleteAccount(c.Context(), claims.UserID, req.Password); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "account deleted"})
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
		Otp   string `json:"otp"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.VerifyEmail(c.Context(), req.Email, req.Otp); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "email verified"})
}

func (h *AuthHandler) ResendOtp(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.ResendOtp(c.Context(), req.Email); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "otp sent"})
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.ForgotPassword(c.Context(), req.Email); err != nil {
		return response.OK(c, fiber.Map{"message": "if the email exists, a reset link has been sent"})
	}
	return response.OK(c, fiber.Map{"message": "if the email exists, a reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.ResetPassword(c.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "password reset successful"})
}

func (h *AuthHandler) SendPhoneOtp(c *fiber.Ctx) error {
	var req struct {
		PhoneNumber string `json:"phone_number"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.SendPhoneOtp(c.Context(), req.PhoneNumber); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "otp sent"})
}

func (h *AuthHandler) VerifyPhone(c *fiber.Ctx) error {
	var req struct {
		PhoneNumber string `json:"phone_number"`
		Otp         string `json:"otp"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.VerifyPhone(c.Context(), req.PhoneNumber, req.Otp); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "phone verified"})
}

func sanitizeUser(u *model.User) fiber.Map {
	return sanitizeUserMap(u)
}

func sanitizeUserMap(u *model.User) fiber.Map {
	m := fiber.Map{
		"id":             u.ID,
		"email":          u.Email,
		"name":           u.Name,
		"email_verified": u.EmailVerified,
		"phone":          u.Phone,
		"phone_verified": u.PhoneVerified,
		"role":           u.Role,
		"is_active":      u.IsActive,
		"is_super_admin": u.IsSuperAdmin,
		"agent_id":       u.AgentID,
		"created_at":     u.CreatedAt,
		"updated_at":     u.UpdatedAt,
	}
	return m
}
