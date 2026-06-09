package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jamaah-in/v2/internal/jamaah/model"
)

func (r *JamaahRepo) CreateGroup(ctx context.Context, g *model.Group) error {
	query := `INSERT INTO groups (id, org_id, name, description, member_count, is_active)
		VALUES ($1, $2, $3, $4, 0, TRUE)
		RETURNING created_at, updated_at`
	return r.pool.QueryRow(ctx, query, g.ID, g.OrgID, g.Name, g.Description).Scan(&g.CreatedAt, &g.UpdatedAt)
}

// ListGroupMembersWithGender returns a group's members joined with their jamaah
// profile gender (empty when no matching profile), ordered so same-gender members
// — and families added together — stay adjacent for auto-rooming.
func (r *JamaahRepo) ListGroupMembersWithGender(ctx context.Context, groupID uuid.UUID) ([]model.RoomCandidate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT gm.member_id, gm.name, COALESCE(p.gender, '')
		FROM group_members gm
		LEFT JOIN jamaah_profiles p ON p.id = gm.member_id
		WHERE gm.group_id = $1
		ORDER BY COALESCE(p.gender, ''), gm.created_at`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.RoomCandidate
	for rows.Next() {
		var c model.RoomCandidate
		if err := rows.Scan(&c.MemberID, &c.Name, &c.Gender); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// CountGroups returns the number of groups owned by an org (for plan-limit checks).
func (r *JamaahRepo) CountGroups(ctx context.Context, orgID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM groups WHERE org_id = $1`, orgID).Scan(&n)
	return n, err
}

// CountProfiles returns the number of jamaah profiles owned by an org (for plan-limit checks).
func (r *JamaahRepo) CountProfiles(ctx context.Context, orgID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jamaah_profiles WHERE org_id = $1`, orgID).Scan(&n)
	return n, err
}

func (r *JamaahRepo) ListGroups(ctx context.Context, orgID uuid.UUID) ([]model.Group, error) {
	query := `SELECT id, org_id, name, description, member_count, is_active, created_at, updated_at
		FROM groups WHERE org_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []model.Group
	for rows.Next() {
		var g model.Group
		if err := rows.Scan(&g.ID, &g.OrgID, &g.Name, &g.Description, &g.MemberCount, &g.IsActive, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *JamaahRepo) GetGroup(ctx context.Context, groupID, orgID uuid.UUID) (*model.Group, error) {
	query := `SELECT id, org_id, name, description, member_count, is_active, created_at, updated_at
		FROM groups WHERE id = $1 AND org_id = $2`
	var g model.Group
	err := r.pool.QueryRow(ctx, query, groupID, orgID).Scan(&g.ID, &g.OrgID, &g.Name, &g.Description, &g.MemberCount, &g.IsActive, &g.CreatedAt, &g.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *JamaahRepo) UpdateGroup(ctx context.Context, g *model.Group) error {
	query := `UPDATE groups SET name = $2, description = $3, updated_at = NOW() WHERE id = $1 AND org_id = $4`
	_, err := r.pool.Exec(ctx, query, g.ID, g.Name, g.Description, g.OrgID)
	return err
}

func (r *JamaahRepo) DeleteGroup(ctx context.Context, groupID, orgID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM groups WHERE id = $1 AND org_id = $2`, groupID, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("group not found")
	}
	return nil
}

func (r *JamaahRepo) AddGroupMember(ctx context.Context, gm *model.GroupMember) error {
	query := `INSERT INTO group_members (id, org_id, group_id, member_id, name, phone, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (group_id, member_id) DO NOTHING
		RETURNING created_at`
	err := r.pool.QueryRow(ctx, query, gm.ID, gm.OrgID, gm.GroupID, gm.MemberID, gm.Name, gm.Phone, gm.Notes).Scan(&gm.CreatedAt)
	if err != nil {
		return err
	}
	r.pool.Exec(ctx, `UPDATE groups SET member_count = member_count + 1, updated_at = NOW() WHERE id = $1`, gm.GroupID)
	return nil
}

func (r *JamaahRepo) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]model.GroupMember, error) {
	query := `SELECT id, org_id, group_id, member_id, name, phone, notes, created_at
		FROM group_members WHERE group_id = $1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.GroupMember
	for rows.Next() {
		var m model.GroupMember
		if err := rows.Scan(&m.ID, &m.OrgID, &m.GroupID, &m.MemberID, &m.Name, &m.Phone, &m.Notes, &m.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *JamaahRepo) UpdateGroupMember(ctx context.Context, groupID, memberID uuid.UUID, name, phone, notes string) error {
	query := `UPDATE group_members SET name = $3, phone = $4, notes = $5 WHERE group_id = $1 AND member_id = $2`
	_, err := r.pool.Exec(ctx, query, groupID, memberID, name, phone, notes)
	return err
}

func (r *JamaahRepo) DeleteGroupMember(ctx context.Context, groupID, memberID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM group_members WHERE group_id = $1 AND member_id = $2`, groupID, memberID)
	if err != nil {
		return err
	}
	if result.RowsAffected() > 0 {
		r.pool.Exec(ctx, `UPDATE groups SET member_count = GREATEST(member_count - 1, 0), updated_at = NOW() WHERE id = $1`, groupID)
	}
	return nil
}
