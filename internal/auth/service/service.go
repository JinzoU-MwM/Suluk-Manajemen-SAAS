package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/jamaah-in/v2/internal/auth/model"
	"github.com/jamaah-in/v2/internal/auth/repository"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/email"
	"github.com/jamaah-in/v2/internal/shared/httpclient"
	"github.com/jamaah-in/v2/internal/shared/plan"
	sharedRedis "github.com/jamaah-in/v2/internal/shared/redis"
)

type AuthService struct {
	repo        *repository.AuthRepo
	jwt         *sharedAuth.JWTManager
	redis       *sharedRedis.Client
	jamaahAddr  string
	invoiceAddr string
	httpc       *httpclient.Client
	email       *email.Client
}

func NewAuthService(repo *repository.AuthRepo, jwt *sharedAuth.JWTManager, redis *sharedRedis.Client, jamaahAddr, invoiceAddr string) *AuthService {
	return &AuthService{
		repo:        repo,
		jwt:         jwt,
		redis:       redis,
		jamaahAddr:  jamaahAddr,
		invoiceAddr: invoiceAddr,
		httpc:       httpclient.New(),
	}
}

// WithEmail attaches a transactional-email client (for subscription invoices, etc.).
func (s *AuthService) WithEmail(c *email.Client) *AuthService {
	s.email = c
	return s
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, *model.Organization, *sharedAuth.TokenPair, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("hash password: %w", err)
	}

	userID := uuid.New()
	user := &model.User{
		ID:           userID,
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hashedPassword),
		Role:         "owner",
		IsActive:     true,
	}
	if req.Phone != "" {
		normalizedPhone := req.Phone
		user.Phone = &normalizedPhone
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, nil, nil, err
	}

	slug := repository.GenerateSlug(req.Name)
	slug, err = s.repo.GenerateUniqueSlug(ctx, slug)
	if err != nil {
		return nil, nil, nil, err
	}

	org := &model.Organization{
		ID:        uuid.New(),
		Name:      req.Name,
		Slug:      slug,
		CreatedBy: userID,
	}
	if err := s.repo.CreateOrganization(ctx, org); err != nil {
		return nil, nil, nil, err
	}

	member := &model.TeamMember{
		ID:     uuid.New(),
		OrgID:  org.ID,
		UserID: userID,
		Role:   "owner",
		Status: "active",
	}
	if err := s.repo.AddTeamMember(ctx, member); err != nil {
		return nil, nil, nil, err
	}

	tokens, err := s.jwt.GenerateTokenPair(userID, org.ID, "owner", user.Email, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := s.storeRefreshToken(ctx, tokens.RefreshToken, userID); err != nil {
		return nil, nil, nil, err
	}

	s.repo.CreateAuditLog(ctx, &model.AuditLog{
		ID:       uuid.New(),
		OrgID:    &org.ID,
		UserID:   &userID,
		Action:   "user.register",
		Entity:   "user",
		EntityID: &userID,
		NewValue: map[string]string{"email": user.Email, "name": user.Name},
	})

	return user, org, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.User, *model.Organization, *sharedAuth.TokenPair, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		return nil, nil, nil, fmt.Errorf("account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, nil, fmt.Errorf("invalid credentials")
	}

	org, member, err := s.getUserOrgAndRole(ctx, user.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	role := "viewer"
	if member != nil {
		role = member.Role
	}

	tokens, err := s.jwt.GenerateTokenPair(user.ID, org.ID, role, user.Email, user.AgentID)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := s.storeRefreshToken(ctx, tokens.RefreshToken, user.ID); err != nil {
		return nil, nil, nil, err
	}

	return user, org, tokens, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*sharedAuth.TokenPair, error) {
	hash := hashToken(refreshToken)
	rt, err := s.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if rt.ExpiresAt.Before(time.Now()) {
		s.repo.DeleteRefreshToken(ctx, hash)
		return nil, fmt.Errorf("refresh token expired")
	}

	user, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	org, member, err := s.getUserOrgAndRole(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	role := "viewer"
	if member != nil {
		role = member.Role
	}

	s.repo.DeleteRefreshToken(ctx, hash)

	tokens, err := s.jwt.GenerateTokenPair(user.ID, org.ID, role, user.Email, user.AgentID)
	if err != nil {
		return nil, err
	}

	if err := s.storeRefreshToken(ctx, tokens.RefreshToken, user.ID); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	hash := hashToken(refreshToken)
	return s.repo.DeleteRefreshToken(ctx, hash)
}

func (s *AuthService) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// ProvisionAgentCredential creates a B2B portal login for an existing agent: a
// user with role "agent" linked to agentID, plus an "agent" team membership in
// the admin's org. The agent record lives in agent-service, so agentID is taken
// on trust from the caller (an org admin) — portal queries still scope by
// org_id, keeping a wrong id within the same org's data.
func (s *AuthService) ProvisionAgentCredential(ctx context.Context, orgID, invitedBy uuid.UUID, req model.ProvisionAgentRequest) (*model.User, error) {
	if req.Email == "" || req.Password == "" || req.AgentID == "" {
		return nil, fmt.Errorf("email, password, and agent_id are required")
	}
	if len(req.Password) < 6 {
		return nil, fmt.Errorf("password minimal 6 karakter")
	}
	agentID, err := uuid.Parse(req.AgentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agent_id")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	name := req.Name
	if name == "" {
		name = req.Email
	}
	user := &model.User{
		ID:           uuid.New(),
		Email:        req.Email,
		Name:         name,
		PasswordHash: string(hashed),
		Role:         string(model.RoleAgent),
		IsActive:     true,
		AgentID:      &agentID,
	}
	member := &model.TeamMember{
		ID:        uuid.New(),
		OrgID:     orgID,
		UserID:    user.ID,
		Role:      string(model.RoleAgent),
		Status:    string(model.StatusActive),
		InvitedBy: &invitedBy,
	}
	if err := s.repo.CreateAgentUserTx(ctx, user, member); err != nil {
		return nil, err
	}
	s.repo.CreateAuditLog(ctx, &model.AuditLog{
		ID:       uuid.New(),
		OrgID:    &orgID,
		UserID:   &invitedBy,
		Action:   "agent.credential.provision",
		Entity:   "user",
		EntityID: &user.ID,
		NewValue: map[string]string{"email": user.Email, "agent_id": agentID.String()},
	})
	return user, nil
}

func (s *AuthService) UpdateUser(ctx context.Context, userID uuid.UUID, name, phone string) (*model.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Name = name
	if phone != "" {
		user.Phone = &phone
	}
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) GetOrganization(ctx context.Context, orgID uuid.UUID) (*model.Organization, error) {
	return s.repo.GetOrganizationByID(ctx, orgID)
}

// UpdateOrganization patches the org profile and returns the fresh record.
func (s *AuthService) UpdateOrganization(ctx context.Context, orgID uuid.UUID, req model.UpdateOrgRequest) (*model.Organization, error) {
	if err := s.repo.UpdateOrganization(ctx, orgID, req); err != nil {
		return nil, err
	}
	return s.repo.GetOrganizationByID(ctx, orgID)
}

func (s *AuthService) CreateOrganization(ctx context.Context, userID uuid.UUID, req model.CreateOrgRequest) (*model.Organization, error) {
	slug := repository.GenerateSlug(req.Name)
	slug, err := s.repo.GenerateUniqueSlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	org := &model.Organization{
		ID:          uuid.New(),
		Name:        req.Name,
		Slug:        slug,
		Address:     req.Address,
		Phone:       req.Phone,
		Email:       req.Email,
		BankName:    req.BankName,
		BankAccount: req.BankAccount,
		BankHolder:  req.BankHolder,
		CreatedBy:   userID,
	}
	if err := s.repo.CreateOrganization(ctx, org); err != nil {
		return nil, err
	}

	member := &model.TeamMember{
		ID:     uuid.New(),
		OrgID:  org.ID,
		UserID: userID,
		Role:   "owner",
		Status: "active",
	}
	s.repo.AddTeamMember(ctx, member)

	return org, nil
}

// checkSeatLimit returns ErrPlanLimit if the org has already reached its
// per-tier user cap (plan.MaxUsers). It uses GetSubscriptionStatus so expired
// subscriptions are evaluated at their downgraded (Gratis) limit, and so both
// the invite and direct-add paths enforce the SAME cap. Fails open on lookup
// errors (availability over strictness for an auth dependency).
func (s *AuthService) checkSeatLimit(ctx context.Context, orgID uuid.UUID) error {
	status, err := s.GetSubscriptionStatus(ctx, orgID)
	if err != nil {
		return nil
	}
	if status.MaxUsers == plan.Unlimited {
		return nil
	}
	users, err := s.repo.ListUsersByOrg(ctx, orgID)
	if err != nil {
		return nil
	}
	if len(users) >= status.MaxUsers {
		return fmt.Errorf("%w: batas pengguna pada paket Anda (%d) telah tercapai. Upgrade paket untuk menambah anggota tim", ErrPlanLimit, status.MaxUsers)
	}
	return nil
}

func (s *AuthService) AddTeamMember(ctx context.Context, orgID, userID, addedBy uuid.UUID, role string) (*model.TeamMember, error) {
	if err := s.checkSeatLimit(ctx, orgID); err != nil {
		return nil, err
	}

	member := &model.TeamMember{
		ID:        uuid.New(),
		OrgID:     orgID,
		UserID:    userID,
		Role:      role,
		Status:    "active",
		InvitedBy: &addedBy,
	}
	if err := s.repo.AddTeamMember(ctx, member); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *AuthService) RemoveTeamMember(ctx context.Context, orgID, userID uuid.UUID) error {
	member, err := s.repo.GetTeamMember(ctx, orgID, userID)
	if err != nil {
		return err
	}
	if member.Role == "owner" {
		return fmt.Errorf("cannot remove owner from organization")
	}
	return s.repo.RemoveTeamMember(ctx, orgID, userID)
}

func (s *AuthService) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	member, err := s.repo.GetTeamMember(ctx, orgID, userID)
	if err != nil {
		return err
	}
	if member.Role == "owner" && role != "owner" {
		return fmt.Errorf("cannot change owner role; transfer ownership first")
	}
	return s.repo.UpdateMemberRole(ctx, orgID, userID, role)
}

func (s *AuthService) ListTeamMembers(ctx context.Context, orgID uuid.UUID) ([]model.TeamMember, error) {
	return s.repo.ListTeamMembers(ctx, orgID)
}

func (s *AuthService) ListUsersByOrg(ctx context.Context, orgID uuid.UUID) ([]model.User, error) {
	return s.repo.ListUsersByOrg(ctx, orgID)
}

func (s *AuthService) InviteMember(ctx context.Context, orgID, invitedBy uuid.UUID, email, role string) (*model.TeamInvite, error) {
	// Enforce the per-tier seat limit before provisioning a new invite.
	if err := s.checkSeatLimit(ctx, orgID); err != nil {
		return nil, err
	}

	token := generateToken(32)
	invite := &model.TeamInvite{
		ID:        uuid.New(),
		OrgID:     orgID,
		Email:     email,
		Role:      role,
		Token:     token,
		InvitedBy: invitedBy,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Status:    "pending",
	}
	if err := s.repo.CreateTeamInvite(ctx, invite); err != nil {
		return nil, err
	}
	return invite, nil
}

func (s *AuthService) AcceptInvite(ctx context.Context, token string, userID uuid.UUID) (*model.TeamMember, error) {
	invite, err := s.repo.GetTeamInviteByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invite not found")
	}
	if invite.Status != "pending" {
		return nil, fmt.Errorf("invite is no longer valid")
	}
	if time.Now().After(invite.ExpiresAt) {
		s.repo.UpdateInviteStatus(ctx, token, "expired")
		return nil, fmt.Errorf("invite has expired")
	}

	member := &model.TeamMember{
		ID:     uuid.New(),
		OrgID:  invite.OrgID,
		UserID: userID,
		Role:   invite.Role,
		Status: "active",
	}
	if err := s.repo.AddTeamMember(ctx, member); err != nil {
		return nil, err
	}

	s.repo.UpdateInviteStatus(ctx, token, "accepted")
	return member, nil
}

func (s *AuthService) CancelInvite(ctx context.Context, inviteID, orgID uuid.UUID) error {
	return s.repo.CancelInvite(ctx, inviteID, orgID)
}

// ── Notifications ──────────────────────────────────────────

func (s *AuthService) GetNotifications(ctx context.Context, orgID, userID uuid.UUID) ([]model.Notification, int, error) {
	return s.repo.GetUserNotifications(ctx, orgID, userID, 50)
}

func (s *AuthService) MarkNotificationRead(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.MarkNotificationRead(ctx, id, userID)
}

func (s *AuthService) MarkAllNotificationsRead(ctx context.Context, orgID, userID uuid.UUID) error {
	return s.repo.MarkAllNotificationsRead(ctx, orgID, userID)
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (*sharedAuth.Claims, error) {
	claims, err := s.jwt.ValidateToken(tokenStr)
	if err != nil {
		return nil, err
	}

	tokenID := claims.ID
	blacklisted, err := s.redis.IsTokenBlacklisted(ctx, tokenID)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, fmt.Errorf("token has been revoked")
	}

	return claims, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("new password must be at least 8 characters")
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.repo.UpdatePassword(ctx, userID, string(hashed))
}

func (s *AuthService) GetActivity(ctx context.Context, userID uuid.UUID) ([]model.AuditLog, error) {
	return s.repo.GetAuditLogsByUser(ctx, userID, 50)
}

func (s *AuthService) DeleteAccount(ctx context.Context, userID uuid.UUID, password string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return fmt.Errorf("password is incorrect")
	}
	_ = s.repo.DeleteRefreshToken(ctx, hashToken(""))
	return s.repo.DeleteUser(ctx, userID)
}

func (s *AuthService) VerifyEmail(ctx context.Context, email, otp string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if user.EmailVerified {
		return nil
	}
	valid, err := s.repo.ConsumeOtp(ctx, email, otp)
	if err != nil {
		return fmt.Errorf("verify otp: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid or expired OTP")
	}
	return s.repo.UpdateEmailVerified(ctx, user.ID, true)
}

func (s *AuthService) ResendOtp(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if user.EmailVerified {
		return nil
	}
	code := generateNumericCode(6)
	if err := s.repo.StoreOtp(ctx, email, code, 15*time.Minute); err != nil {
		return fmt.Errorf("store otp: %w", err)
	}
	// TODO: send OTP via email service
	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal whether the email exists
		return nil
	}
	code := generateToken(32)
	if err := s.repo.StorePasswordResetCode(ctx, user.Email, code, 15*time.Minute); err != nil {
		return fmt.Errorf("store reset code: %w", err)
	}
	// TODO: send reset code via email service
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	valid, err := s.repo.ConsumePasswordResetCode(ctx, email, code)
	if err != nil {
		return fmt.Errorf("validate reset code: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid or expired reset code")
	}
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.repo.UpdatePassword(ctx, user.ID, string(hashed))
}

func (s *AuthService) SendPhoneOtp(ctx context.Context, phoneNumber string) error {
	return nil
}

func (s *AuthService) VerifyPhone(ctx context.Context, phoneNumber, otp string) error {
	return nil
}

func (s *AuthService) getUserOrgAndRole(ctx context.Context, userID uuid.UUID) (*model.Organization, *model.TeamMember, error) {
	org, err := s.repo.GetOrgByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	member, err := s.repo.GetTeamMember(ctx, org.ID, userID)
	if err != nil {
		return org, nil, nil
	}
	return org, member, nil
}

func (s *AuthService) storeRefreshToken(ctx context.Context, refreshToken string, userID uuid.UUID) error {
	hash := hashToken(refreshToken)
	rt := &model.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	return s.repo.CreateRefreshToken(ctx, rt)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func generateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

func generateNumericCode(digits int) string {
	b := make([]byte, digits)
	rand.Read(b)
	for i := range b {
		b[i] = b[i]%10 + '0'
	}
	return string(b)
}
