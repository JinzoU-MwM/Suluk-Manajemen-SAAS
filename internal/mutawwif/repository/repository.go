package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jamaah-in/v2/internal/mutawwif/model"
)

type MutawwifRepo struct {
	pool *pgxpool.Pool
}

func NewMutawwifRepo(pool *pgxpool.Pool) *MutawwifRepo { return &MutawwifRepo{pool: pool} }

var ErrGuideNotFound = fmt.Errorf("guide not found")

const guideCols = `id, org_id, name, phone, email, type, license_no, license_expiry, is_active, notes, created_at, updated_at`

func scanGuide(row interface{ Scan(...any) error }) (*model.Guide, error) {
	var g model.Guide
	err := row.Scan(&g.ID, &g.OrgID, &g.Name, &g.Phone, &g.Email, &g.Type, &g.LicenseNo,
		&g.LicenseExpiry, &g.IsActive, &g.Notes, &g.CreatedAt, &g.UpdatedAt)
	return &g, err
}

func (r *MutawwifRepo) CreateGuide(ctx context.Context, g *model.Guide) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO guides (org_id, name, phone, email, type, license_no, license_expiry, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, is_active, created_at, updated_at`,
		g.OrgID, g.Name, g.Phone, g.Email, g.Type, g.LicenseNo, g.LicenseExpiry, g.Notes,
	).Scan(&g.ID, &g.IsActive, &g.CreatedAt, &g.UpdatedAt)
}

func (r *MutawwifRepo) GetGuide(ctx context.Context, id, orgID uuid.UUID) (*model.Guide, error) {
	g, err := scanGuide(r.pool.QueryRow(ctx, `SELECT `+guideCols+` FROM guides WHERE id = $1 AND org_id = $2`, id, orgID))
	if err == pgx.ErrNoRows {
		return nil, ErrGuideNotFound
	}
	return g, err
}

// ListGuides returns the org's guides with each guide's active assignment count.
func (r *MutawwifRepo) ListGuides(ctx context.Context, orgID uuid.UUID, search string) ([]model.Guide, error) {
	where := "WHERE g.org_id = $1"
	args := []any{orgID}
	if search != "" {
		args = append(args, "%"+search+"%")
		where += " AND (g.name ILIKE $2 OR g.phone ILIKE $2)"
	}
	rows, err := r.pool.Query(ctx, `
		SELECT g.id, g.org_id, g.name, g.phone, g.email, g.type, g.license_no, g.license_expiry,
		       g.is_active, g.notes, g.created_at, g.updated_at,
		       COALESCE(a.cnt, 0)
		FROM guides g
		LEFT JOIN (SELECT guide_id, COUNT(*) cnt FROM guide_assignments GROUP BY guide_id) a ON a.guide_id = g.id
		`+where+` ORDER BY g.name`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Guide{}
	for rows.Next() {
		var g model.Guide
		if err := rows.Scan(&g.ID, &g.OrgID, &g.Name, &g.Phone, &g.Email, &g.Type, &g.LicenseNo,
			&g.LicenseExpiry, &g.IsActive, &g.Notes, &g.CreatedAt, &g.UpdatedAt, &g.AssignmentCount); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *MutawwifRepo) UpdateGuide(ctx context.Context, id, orgID uuid.UUID, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	query := "UPDATE guides SET updated_at = NOW()"
	args := []any{}
	i := 1
	for k, v := range updates {
		query += fmt.Sprintf(", %s = $%d", k, i)
		args = append(args, v)
		i++
	}
	query += fmt.Sprintf(" WHERE id = $%d AND org_id = $%d", i, i+1)
	args = append(args, id, orgID)
	ct, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrGuideNotFound
	}
	return nil
}

func (r *MutawwifRepo) DeleteGuide(ctx context.Context, id, orgID uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM guides WHERE id = $1 AND org_id = $2`, id, orgID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrGuideNotFound
	}
	return nil
}

// Assign links a guide to a departure group (idempotent on the unique pair).
func (r *MutawwifRepo) Assign(ctx context.Context, a *model.GuideAssignment) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO guide_assignments (org_id, guide_id, group_id, role)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (group_id, guide_id) DO UPDATE SET role = EXCLUDED.role
		RETURNING id, assigned_at, created_at`,
		a.OrgID, a.GuideID, a.GroupID, a.Role,
	).Scan(&a.ID, &a.AssignedAt, &a.CreatedAt)
}

func (r *MutawwifRepo) Unassign(ctx context.Context, orgID, guideID, groupID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM guide_assignments WHERE org_id = $1 AND guide_id = $2 AND group_id = $3`, orgID, guideID, groupID)
	return err
}

// ListByGroup returns the guides assigned to a departure group (with names).
func (r *MutawwifRepo) ListByGroup(ctx context.Context, orgID, groupID uuid.UUID) ([]model.GuideAssignment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.org_id, a.guide_id, a.group_id, a.role, a.assigned_at, a.created_at,
		       g.name, g.type, g.phone
		FROM guide_assignments a JOIN guides g ON g.id = a.guide_id AND g.org_id = a.org_id
		WHERE a.org_id = $1 AND a.group_id = $2 ORDER BY a.role, g.name`, orgID, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAssignments(rows)
}

// ListByGuide returns a guide's assignments (which kloter they're on).
func (r *MutawwifRepo) ListByGuide(ctx context.Context, orgID, guideID uuid.UUID) ([]model.GuideAssignment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.org_id, a.guide_id, a.group_id, a.role, a.assigned_at, a.created_at,
		       g.name, g.type, g.phone
		FROM guide_assignments a JOIN guides g ON g.id = a.guide_id AND g.org_id = a.org_id
		WHERE a.org_id = $1 AND a.guide_id = $2 ORDER BY a.assigned_at DESC`, orgID, guideID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAssignments(rows)
}

func scanAssignments(rows pgx.Rows) ([]model.GuideAssignment, error) {
	out := []model.GuideAssignment{}
	for rows.Next() {
		var a model.GuideAssignment
		if err := rows.Scan(&a.ID, &a.OrgID, &a.GuideID, &a.GroupID, &a.Role, &a.AssignedAt, &a.CreatedAt,
			&a.GuideName, &a.GuideType, &a.GuidePhone); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// GuidesExpiringLicense lists active guides whose license expires within N days.
func (r *MutawwifRepo) GuidesExpiringLicense(ctx context.Context, orgID uuid.UUID, withinDays int) ([]model.Guide, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+guideCols+` FROM guides
		WHERE org_id = $1 AND is_active AND license_expiry IS NOT NULL
		  AND license_expiry <= (CURRENT_DATE + make_interval(days => $2)) AND license_expiry >= CURRENT_DATE
		ORDER BY license_expiry`, orgID, withinDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Guide{}
	for rows.Next() {
		g, err := scanGuide(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *g)
	}
	return out, rows.Err()
}
