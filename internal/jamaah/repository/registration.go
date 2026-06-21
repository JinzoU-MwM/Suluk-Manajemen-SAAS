package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
)

func (r *JamaahRepo) CreateRegistrationLink(ctx context.Context, link *model.RegistrationLink) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO registration_links (org_id, group_id, package_id, token, expires_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, created_at, updated_at
	`, link.OrgID, link.GroupID, link.PackageID, link.Token, link.ExpiresAt, link.CreatedBy).
		Scan(&link.ID, &link.CreatedAt, &link.UpdatedAt)
}

func (r *JamaahRepo) GetRegistrationLinkByToken(ctx context.Context, token string) (*model.RegistrationLink, error) {
	var link model.RegistrationLink
	err := r.pool.QueryRow(ctx, `
		SELECT id, org_id, group_id, package_id, token, expires_at, is_active, created_by, created_at, updated_at
		FROM registration_links WHERE token = $1 AND is_active = TRUE
	`, token).Scan(&link.ID, &link.OrgID, &link.GroupID, &link.PackageID, &link.Token, &link.ExpiresAt, &link.IsActive, &link.CreatedBy, &link.CreatedAt, &link.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired registration link")
	}
	return &link, nil
}

func (r *JamaahRepo) GetActiveRegistrationLink(ctx context.Context, orgID, groupID string) (*model.RegistrationLink, error) {
	var link model.RegistrationLink
	err := r.pool.QueryRow(ctx, `
		SELECT id, org_id, group_id, package_id, token, expires_at, is_active, created_by, created_at, updated_at
		FROM registration_links WHERE org_id = $1 AND group_id = $2 AND is_active = TRUE AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT 1
	`, orgID, groupID).Scan(&link.ID, &link.OrgID, &link.GroupID, &link.PackageID, &link.Token, &link.ExpiresAt, &link.IsActive, &link.CreatedBy, &link.CreatedAt, &link.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("no active registration link")
	}
	return &link, nil
}

func (r *JamaahRepo) DeactivateRegistrationLink(ctx context.Context, orgID, groupID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE registration_links SET is_active = FALSE, updated_at = NOW()
		WHERE org_id = $1 AND group_id = $2 AND is_active = TRUE
	`, orgID, groupID)
	return err
}

func (r *JamaahRepo) CreatePendingRegistration(ctx context.Context, p *model.PendingRegistration) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO pending_registrations (org_id, registration_link_id, phone_number, name, email, ktp_file_url, passport_file_url, visa_file_url, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, created_at, updated_at
	`, p.OrgID, p.RegistrationLinkID, p.PhoneNumber, p.Name, p.Email, p.KtpFileURL, p.PassportFileURL, p.VisaFileURL, p.Notes).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *JamaahRepo) ListPendingRegistrations(ctx context.Context, orgID, groupID string) ([]model.PendingRegistration, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.org_id, p.registration_link_id, p.phone_number, p.name, p.email, p.ktp_file_url, p.passport_file_url, p.visa_file_url,
		       p.notes, p.status, p.jamaah_id, p.reviewed_by, p.reviewed_at, p.created_at, p.updated_at
		FROM pending_registrations p
		JOIN registration_links l ON p.registration_link_id = l.id
		WHERE p.org_id = $1 AND l.group_id = $2
		ORDER BY p.created_at DESC
	`, orgID, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []model.PendingRegistration
	for rows.Next() {
		var p model.PendingRegistration
		if err := rows.Scan(&p.ID, &p.OrgID, &p.RegistrationLinkID, &p.PhoneNumber, &p.Name, &p.Email, &p.KtpFileURL, &p.PassportFileURL, &p.VisaFileURL, &p.Notes, &p.Status, &p.JamaahID, &p.ReviewedBy, &p.ReviewedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *JamaahRepo) GetPendingRegistration(ctx context.Context, id, orgID string) (*model.PendingRegistration, error) {
	var p model.PendingRegistration
	err := r.pool.QueryRow(ctx, `
		SELECT id, org_id, registration_link_id, phone_number, name, email, ktp_file_url, passport_file_url, visa_file_url,
		       notes, status, jamaah_id, reviewed_by, reviewed_at, created_at, updated_at
		FROM pending_registrations WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(&p.ID, &p.OrgID, &p.RegistrationLinkID, &p.PhoneNumber, &p.Name, &p.Email, &p.KtpFileURL, &p.PassportFileURL, &p.VisaFileURL, &p.Notes, &p.Status, &p.JamaahID, &p.ReviewedBy, &p.ReviewedAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("pending registration not found")
	}
	return &p, nil
}

func (r *JamaahRepo) ApprovePendingRegistration(ctx context.Context, id, orgID, reviewerID, jamaahID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE pending_registrations SET status = 'approved', jamaah_id = $3, reviewed_by = $4, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND org_id = $2
	`, id, orgID, jamaahID, reviewerID)
	return err
}

// ApprovePendingTx creates the jamaah profile and marks the pending registration
// approved in one transaction, locking the pending row and re-checking its status
// under the lock. This makes approval atomic + idempotent: the old flow ran
// CreateProfile then the status UPDATE as two separate statements, so a failure
// or concurrent approval could create a duplicate profile (and the status guard
// was a TOCTOU check in the service).
func (r *JamaahRepo) ApprovePendingTx(ctx context.Context, pendingID, orgID, reviewerID uuid.UUID, profile *model.JamaahProfile) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM pending_registrations WHERE id = $1 AND org_id = $2 FOR UPDATE`, pendingID, orgID).Scan(&status); err != nil {
		return fmt.Errorf("pending registration not found")
	}
	if status != "pending" {
		return fmt.Errorf("registration already %s", status)
	}
	if err := insertProfile(ctx, tx, profile); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE pending_registrations SET status = 'approved', jamaah_id = $3, reviewed_by = $4, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND org_id = $2
	`, pendingID, orgID, profile.ID, reviewerID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *JamaahRepo) RejectPendingRegistration(ctx context.Context, id, orgID, reviewerID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE pending_registrations SET status = 'rejected', reviewed_by = $3, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND org_id = $2
	`, id, orgID, reviewerID)
	return err
}
