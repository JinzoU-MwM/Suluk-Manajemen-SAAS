package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jamaah-in/v2/internal/auth/model"
)

type AuthRepo struct {
	pool *pgxpool.Pool
}

func NewAuthRepo(pool *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{pool: pool}
}

func (r *AuthRepo) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, email, name, password_hash, phone, role, is_active, agent_id, jamaah_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`
	err := r.pool.QueryRow(ctx, query,
		user.ID, user.Email, user.Name, user.PasswordHash,
		user.Phone, user.Role, user.IsActive, user.AgentID, user.JamaahID,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			if strings.Contains(err.Error(), "email") {
				return ErrEmailExists
			}
			if strings.Contains(err.Error(), "phone") {
				return ErrPhoneExists
			}
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// CreateAgentUserTx provisions a B2B portal login: the user (role agent, linked
// to agent_id) and its team_member row are written in one transaction so a
// failed membership can't leave an orphan login.
func (r *AuthRepo) CreateAgentUserTx(ctx context.Context, user *model.User, member *model.TeamMember) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO users (id, email, name, password_hash, phone, role, is_active, agent_id, jamaah_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`,
		user.ID, user.Email, user.Name, user.PasswordHash, user.Phone, user.Role, user.IsActive, user.AgentID, user.JamaahID,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			if strings.Contains(err.Error(), "email") {
				return ErrEmailExists
			}
			if strings.Contains(err.Error(), "phone") {
				return ErrPhoneExists
			}
		}
		return fmt.Errorf("create agent user: %w", err)
	}

	if err := tx.QueryRow(ctx, `
		INSERT INTO team_members (id, org_id, user_id, role, status, invited_by, joined_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING joined_at`,
		member.ID, member.OrgID, member.UserID, member.Role, member.Status, member.InvitedBy,
	).Scan(&member.JoinedAt); err != nil {
		return fmt.Errorf("add agent team member: %w", err)
	}
	return tx.Commit(ctx)
}

func (r *AuthRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, email, name, password_hash, email_verified, phone, phone_verified,
		city, bio, COALESCE(avatar_color,'blue'), COALESCE(notify_usage_limit,TRUE), COALESCE(notify_expiry,TRUE),
		role, is_active, is_super_admin, agent_id, jamaah_id, created_at, updated_at FROM users WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.EmailVerified,
		&u.Phone, &u.PhoneVerified,
		&u.City, &u.Bio, &u.AvatarColor, &u.NotifyUsageLimit, &u.NotifyExpiry,
		&u.Role, &u.IsActive, &u.IsSuperAdmin, &u.AgentID, &u.JamaahID,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (r *AuthRepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, email, name, password_hash, email_verified, phone, phone_verified, role, is_active, is_super_admin, agent_id, jamaah_id, created_at, updated_at FROM users WHERE email = $1`
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.EmailVerified,
		&u.Phone, &u.PhoneVerified, &u.Role, &u.IsActive, &u.IsSuperAdmin, &u.AgentID, &u.JamaahID,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

func (r *AuthRepo) UpdateUser(ctx context.Context, user *model.User) error {
	query := `UPDATE users SET name = $2, phone = $3, updated_at = NOW() WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, user.ID, user.Name, user.Phone)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *AuthRepo) CreateOrganization(ctx context.Context, org *model.Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug, address, phone, email, bank_name, bank_account, bank_holder, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at`
	err := r.pool.QueryRow(ctx, query,
		org.ID, org.Name, org.Slug, org.Address, org.Phone, org.Email,
		org.BankName, org.BankAccount, org.BankHolder, org.CreatedBy,
	).Scan(&org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			if strings.Contains(err.Error(), "slug") {
				return ErrSlugExists
			}
		}
		return fmt.Errorf("create org: %w", err)
	}
	return nil
}

func (r *AuthRepo) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	o := &model.Organization{}
	query := `SELECT id, name, slug, logo_url, address, phone, email, npwp, ppiu_number, sk_number, bank_name, bank_account, bank_holder, created_by, created_at, updated_at FROM organizations WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.Name, &o.Slug, &o.LogoURL, &o.Address, &o.Phone, &o.Email,
		&o.NPWP, &o.PPIUNumber, &o.SKNumber,
		&o.BankName, &o.BankAccount, &o.BankHolder, &o.CreatedBy,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrOrgNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get org: %w", err)
	}
	return o, nil
}

// UpdateOrganization patches editable org fields; nil request fields are kept.
func (r *AuthRepo) UpdateOrganization(ctx context.Context, orgID uuid.UUID, req model.UpdateOrgRequest) error {
	query := `UPDATE organizations SET
		name = COALESCE($2, name),
		address = COALESCE($3, address),
		phone = COALESCE($4, phone),
		email = COALESCE($5, email),
		npwp = COALESCE($6, npwp),
		ppiu_number = COALESCE($7, ppiu_number),
		sk_number = COALESCE($8, sk_number),
		bank_name = COALESCE($9, bank_name),
		bank_account = COALESCE($10, bank_account),
		bank_holder = COALESCE($11, bank_holder),
		updated_at = NOW()
		WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, orgID, req.Name, req.Address, req.Phone, req.Email,
		req.NPWP, req.PPIUNumber, req.SKNumber, req.BankName, req.BankAccount, req.BankHolder)
	return err
}

func (r *AuthRepo) AddTeamMember(ctx context.Context, member *model.TeamMember) error {
	query := `
		INSERT INTO team_members (id, org_id, user_id, role, status, invited_by, joined_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING joined_at`
	err := r.pool.QueryRow(ctx, query,
		member.ID, member.OrgID, member.UserID, member.Role, member.Status, member.InvitedBy,
	).Scan(&member.JoinedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return ErrMemberExists
		}
		return fmt.Errorf("add team member: %w", err)
	}
	return nil
}

func (r *AuthRepo) GetTeamMember(ctx context.Context, orgID, userID uuid.UUID) (*model.TeamMember, error) {
	m := &model.TeamMember{}
	query := `SELECT id, org_id, user_id, role, status, invited_by, joined_at FROM team_members WHERE org_id = $1 AND user_id = $2`
	err := r.pool.QueryRow(ctx, query, orgID, userID).Scan(
		&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.Status, &m.InvitedBy, &m.JoinedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrMemberNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get team member: %w", err)
	}
	return m, nil
}

func (r *AuthRepo) RemoveTeamMember(ctx context.Context, orgID, userID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM team_members WHERE org_id = $1 AND user_id = $2`, orgID, userID)
	if err != nil {
		return fmt.Errorf("remove team member: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (r *AuthRepo) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	result, err := r.pool.Exec(ctx, `UPDATE team_members SET role = $3 WHERE org_id = $1 AND user_id = $2`, orgID, userID, role)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (r *AuthRepo) ListTeamMembers(ctx context.Context, orgID uuid.UUID) ([]model.TeamMember, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, org_id, user_id, role, status, invited_by, joined_at FROM team_members WHERE org_id = $1 ORDER BY joined_at`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	defer rows.Close()

	members := []model.TeamMember{}
	for rows.Next() {
		var m model.TeamMember
		if err := rows.Scan(&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.Status, &m.InvitedBy, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *AuthRepo) GetOrgByUserID(ctx context.Context, userID uuid.UUID) (*model.Organization, error) {
	o := &model.Organization{}
	query := `
		SELECT o.id, o.name, o.slug, o.logo_url, o.address, o.phone, o.email, o.bank_name, o.bank_account, o.bank_holder, o.created_by, o.created_at, o.updated_at
		FROM organizations o
		JOIN team_members tm ON o.id = tm.org_id
		WHERE tm.user_id = $1 AND tm.status = 'active'
		ORDER BY o.created_at ASC
		LIMIT 1`
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&o.ID, &o.Name, &o.Slug, &o.LogoURL, &o.Address, &o.Phone, &o.Email,
		&o.BankName, &o.BankAccount, &o.BankHolder, &o.CreatedBy,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrOrgNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get org by user: %w", err)
	}
	return o, nil
}

func (r *AuthRepo) CreateRefreshToken(ctx context.Context, rt *model.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token_hash, device_info, expires_at) VALUES ($1, $2, $3, $4, $5) RETURNING created_at`
	return r.pool.QueryRow(ctx, query, rt.ID, rt.UserID, rt.TokenHash, rt.DeviceInfo, rt.ExpiresAt).Scan(&rt.CreatedAt)
}

func (r *AuthRepo) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	rt := &model.RefreshToken{}
	query := `SELECT id, user_id, token_hash, device_info, expires_at, created_at FROM refresh_tokens WHERE token_hash = $1`
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.DeviceInfo, &rt.ExpiresAt, &rt.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	return rt, nil
}

func (r *AuthRepo) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *AuthRepo) DeleteRefreshTokensByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

func (r *AuthRepo) CreateTeamInvite(ctx context.Context, invite *model.TeamInvite) error {
	query := `INSERT INTO team_invites (id, org_id, email, role, token, invited_by, expires_at, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at`
	return r.pool.QueryRow(ctx, query, invite.ID, invite.OrgID, invite.Email, invite.Role, invite.Token, invite.InvitedBy, invite.ExpiresAt, invite.Status).Scan(&invite.CreatedAt)
}

func (r *AuthRepo) GetTeamInviteByToken(ctx context.Context, token string) (*model.TeamInvite, error) {
	inv := &model.TeamInvite{}
	query := `SELECT id, org_id, email, role, token, invited_by, expires_at, status, created_at FROM team_invites WHERE token = $1`
	err := r.pool.QueryRow(ctx, query, token).Scan(&inv.ID, &inv.OrgID, &inv.Email, &inv.Role, &inv.Token, &inv.InvitedBy, &inv.ExpiresAt, &inv.Status, &inv.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrInviteNotFound
	}
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *AuthRepo) UpdateInviteStatus(ctx context.Context, token string, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE team_invites SET status = $2 WHERE token = $1`, token, status)
	return err
}

func (r *AuthRepo) CancelInvite(ctx context.Context, inviteID, orgID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM team_invites WHERE id = $1 AND org_id = $2 AND status = 'pending'`, inviteID, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrInviteNotFound
	}
	return nil
}

func (r *AuthRepo) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	query := `INSERT INTO audit_logs (id, org_id, user_id, action, entity, entity_id, old_value, new_value, ip_address) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(ctx, query, log.ID, log.OrgID, log.UserID, log.Action, log.Entity, log.EntityID, log.OldValue, log.NewValue, log.IPAddress)
	return err
}

// ── Notifications ──────────────────────────────────────────

func (r *AuthRepo) GetUserNotifications(ctx context.Context, orgID, userID uuid.UUID, limit int) ([]model.Notification, int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE org_id = $1 AND (user_id IS NULL OR user_id = $2) AND is_read = FALSE`,
		orgID, userID,
	).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, org_id, user_id, severity, title, message, group_id, is_read, created_at
		FROM notifications
		WHERE org_id = $1 AND (user_id IS NULL OR user_id = $2)
		ORDER BY created_at DESC LIMIT $3`,
		orgID, userID, limit,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	notifications := []model.Notification{}
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.OrgID, &n.UserID, &n.Severity, &n.Title, &n.Message, &n.GroupID, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, 0, err
		}
		notifications = append(notifications, n)
	}
	return notifications, count, nil
}

func (r *AuthRepo) MarkNotificationRead(ctx context.Context, id, orgID, userID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE id = $1 AND org_id = $2 AND (user_id IS NULL OR user_id = $3)`,
		id, orgID, userID,
	)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}
	return nil
}

func (r *AuthRepo) MarkAllNotificationsRead(ctx context.Context, orgID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE org_id = $1 AND (user_id IS NULL OR user_id = $2) AND is_read = FALSE`,
		orgID, userID,
	)
	return err
}

func (r *AuthRepo) CreateNotification(ctx context.Context, n *model.Notification) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO notifications (id, org_id, user_id, severity, title, message, group_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at`,
		n.ID, n.OrgID, n.UserID, n.Severity, n.Title, n.Message, n.GroupID,
	).Scan(&n.CreatedAt)
}

// Repository-level errors
var (
	ErrUserNotFound   = fmt.Errorf("user not found")
	ErrEmailExists    = fmt.Errorf("email already exists")
	ErrPhoneExists    = fmt.Errorf("phone already exists")
	ErrOrgNotFound    = fmt.Errorf("organization not found")
	ErrSlugExists     = fmt.Errorf("organization slug already exists")
	ErrMemberExists   = fmt.Errorf("user is already a team member")
	ErrMemberNotFound = fmt.Errorf("team member not found")
	ErrTokenNotFound  = fmt.Errorf("refresh token not found")
	ErrInviteNotFound = fmt.Errorf("invite not found")
)

// Helper to generate slug from name
func GenerateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "--", "-")
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	slug = strings.Trim(slug, "-")
	if len(slug) > 100 {
		slug = slug[:100]
	}
	if len(slug) == 0 {
		slug = fmt.Sprintf("org-%s", uuid.New().String()[:8])
	}
	return slug
}

func (r *AuthRepo) IsSlugTaken(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organizations WHERE slug = $1)`, slug).Scan(&exists)
	return exists, err
}

func (r *AuthRepo) GenerateUniqueSlug(ctx context.Context, baseSlug string) (string, error) {
	slug := baseSlug
	taken, err := r.IsSlugTaken(ctx, slug)
	if err != nil {
		return "", err
	}
	if !taken {
		return slug, nil
	}
	for i := 1; i < 100; i++ {
		slug = fmt.Sprintf("%s-%d", baseSlug, i)
		taken, err = r.IsSlugTaken(ctx, slug)
		if err != nil {
			return "", err
		}
		if !taken {
			return slug, nil
		}
	}
	return fmt.Sprintf("%s-%s", baseSlug, uuid.New().String()[:8]), nil
}

// Clean up expired refresh tokens
func (r *AuthRepo) CleanExpiredTokens(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	return err
}

// Clean up expired invites
func (r *AuthRepo) CleanExpiredInvites(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `UPDATE team_invites SET status = 'expired' WHERE expires_at < NOW() AND status = 'pending'`)
	return err
}

// Count team members
func (r *AuthRepo) CountTeamMembers(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM team_members WHERE org_id = $1 AND status = 'active'`, orgID).Scan(&count)
	return count, err
}

// Get users by org
func (r *AuthRepo) ListUsersByOrg(ctx context.Context, orgID uuid.UUID) ([]model.User, error) {
	query := `
		SELECT u.id, u.email, u.name, u.email_verified, u.phone, u.phone_verified, u.role, u.is_active, u.created_at, u.updated_at
		FROM users u
		JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.org_id = $1 AND tm.status = 'active'
		ORDER BY u.name`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.EmailVerified, &u.Phone, &u.PhoneVerified, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *AuthRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, hash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`, userID, hash)
	return err
}

func (r *AuthRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	return err
}

func (r *AuthRepo) GetAuditLogsByUser(ctx context.Context, userID uuid.UUID, limit int) ([]model.AuditLog, error) {
	if limit < 1 {
		limit = 50
	}
	query := `SELECT id, org_id, user_id, action, entity, entity_id, old_value, new_value, ip_address, created_at FROM audit_logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		if err := rows.Scan(&l.ID, &l.OrgID, &l.UserID, &l.Action, &l.Entity, &l.EntityID, &l.OldValue, &l.NewValue, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *AuthRepo) UpdateEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET email_verified = $2, updated_at = NOW() WHERE id = $1`, userID, verified)
	return err
}

func (r *AuthRepo) UpdatePhoneVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET phone_verified = $2, updated_at = NOW() WHERE id = $1`, userID, verified)
	return err
}

// StoreOtp saves an OTP code with a TTL in PostgreSQL.
func (r *AuthRepo) StoreOtp(ctx context.Context, email, code string, ttl time.Duration) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO email_otps (email, code, expires_at) VALUES ($1, $2, NOW() + $3::interval)
		 ON CONFLICT (email) DO UPDATE SET code = EXCLUDED.code, expires_at = EXCLUDED.expires_at`,
		email, code, fmt.Sprintf("%d seconds", int(ttl.Seconds())),
	)
	return err
}

// ConsumeOtp validates an OTP and deletes it if valid.
func (r *AuthRepo) ConsumeOtp(ctx context.Context, email, code string) (bool, error) {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM email_otps WHERE email = $1 AND code = $2 AND expires_at > NOW()`,
		email, code,
	)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

// StorePasswordResetCode saves a reset code with a 15-min TTL.
func (r *AuthRepo) StorePasswordResetCode(ctx context.Context, email, code string, ttl time.Duration) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO password_resets (email, code, expires_at) VALUES ($1, $2, NOW() + $3::interval)
		 ON CONFLICT (email) DO UPDATE SET code = EXCLUDED.code, expires_at = EXCLUDED.expires_at`,
		email, code, fmt.Sprintf("%d seconds", int(ttl.Seconds())),
	)
	return err
}

// ConsumePasswordResetCode validates a reset code and deletes it if valid.
func (r *AuthRepo) ConsumePasswordResetCode(ctx context.Context, email, code string) (bool, error) {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM password_resets WHERE email = $1 AND code = $2 AND expires_at > NOW()`,
		email, code,
	)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

// Startup cleanup
func (r *AuthRepo) RunStartupCleanup(ctx context.Context) {
	r.CleanExpiredTokens(ctx)
	r.CleanExpiredInvites(ctx)
}

// Schedule periodic cleanup
func (r *AuthRepo) StartCleanupScheduler(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			select {
			case <-ticker.C:
				r.RunStartupCleanup(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	r.RunStartupCleanup(ctx)
}
