package repository

import (
	"context"

	"github.com/google/uuid"
)

func (r *AuthRepo) CreateBranch(ctx context.Context, parentOrgID uuid.UUID, name, slug string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO organizations (name, slug, created_by, parent_org_id, is_branch, branch_name)
		VALUES ($1, $2, (SELECT created_by FROM organizations WHERE id = $3), $3, TRUE, $1)
		RETURNING id
	`, name, slug, parentOrgID).Scan(&id)
	return id, err
}

func (r *AuthRepo) ListBranches(ctx context.Context, parentOrgID uuid.UUID) ([]struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	CreatedAt interface{}
}, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, branch_name, slug, created_at
		FROM organizations WHERE parent_org_id = $1 AND is_branch = TRUE
		ORDER BY branch_name
	`, parentOrgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type Branch struct {
		ID        uuid.UUID   `json:"id"`
		Name      string      `json:"name"`
		Slug      string      `json:"slug"`
		CreatedAt interface{} `json:"created_at"`
	}
	var result []struct {
		ID        uuid.UUID
		Name      string
		Slug      string
		CreatedAt interface{}
	}
	for rows.Next() {
		var b struct {
			ID        uuid.UUID
			Name      string
			Slug      string
			CreatedAt interface{}
		}
		rows.Scan(&b.ID, &b.Name, &b.Slug, &b.CreatedAt)
		result = append(result, b)
	}
	return result, rows.Err()
}

func (r *AuthRepo) GetOrganization(ctx context.Context, id uuid.UUID) (interface{}, error) {
	var orgID uuid.UUID
	var name, slug string
	err := r.pool.QueryRow(ctx, `SELECT id, name, slug FROM organizations WHERE id = $1`, id).Scan(&orgID, &name, &slug)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"id": orgID, "name": name, "slug": slug}, nil
}

// Consolidated multi-branch totals are aggregated over HTTP in the service layer
// (AuthService.GetConsolidatedStats) because each service owns a separate
// database — jamaah_profiles and invoices are not reachable from this pool.
