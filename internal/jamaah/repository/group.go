package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/jamaah-in/v2/internal/jamaah/model"
	"github.com/jamaah-in/v2/internal/shared/outbox"
)

func (r *JamaahRepo) CreateGroup(ctx context.Context, g *model.Group) error {
	return insertGroup(ctx, r.pool, g)
}

// insertGroup runs the group INSERT on any querier (the pool or a tx), shared by
// the plain create and the transactional, limit-enforced CreateGroupTx.
func insertGroup(ctx context.Context, q querier, g *model.Group) error {
	query := `INSERT INTO groups (id, org_id, name, description, member_count, is_active)
		VALUES ($1, $2, $3, $4, 0, TRUE)
		RETURNING created_at, updated_at`
	return q.QueryRow(ctx, query, g.ID, g.OrgID, g.Name, g.Description).Scan(&g.CreatedAt, &g.UpdatedAt)
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

const groupCols = `id, org_id, name, description, member_count, is_active,
	package_id, departure_date, departure_status, manifest_finalized_at, departed_at, created_at, updated_at`

func scanGroup(row rowScanner) (*model.Group, error) {
	var g model.Group
	err := row.Scan(&g.ID, &g.OrgID, &g.Name, &g.Description, &g.MemberCount, &g.IsActive,
		&g.PackageID, &g.DepartureDate, &g.DepartureStatus, &g.ManifestFinalizedAt, &g.DepartedAt,
		&g.CreatedAt, &g.UpdatedAt)
	return &g, err
}

func (r *JamaahRepo) ListGroups(ctx context.Context, orgID uuid.UUID) ([]model.Group, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+groupCols+` FROM groups WHERE org_id = $1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []model.Group
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, *g)
	}
	return groups, nil
}

func (r *JamaahRepo) GetGroup(ctx context.Context, groupID, orgID uuid.UUID) (*model.Group, error) {
	g, err := scanGroup(r.pool.QueryRow(ctx, `SELECT `+groupCols+` FROM groups WHERE id = $1 AND org_id = $2`, groupID, orgID))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return g, nil
}

// SetDeparture links a group to its package + departure date (editable only
// while the kloter is still in draft).
func (r *JamaahRepo) SetDeparture(ctx context.Context, groupID, orgID uuid.UUID, packageID *uuid.UUID, departureDate *time.Time) error {
	ct, err := r.pool.Exec(ctx, `UPDATE groups
		SET package_id = $3, departure_date = $4, updated_at = NOW()
		WHERE id = $1 AND org_id = $2 AND departure_status = 'draft'`,
		groupID, orgID, packageID, departureDate)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("group not found or not in draft")
	}
	return nil
}

// TransitionDeparture moves the kloter to a new status, stamping the relevant
// timestamps and enqueueing the outbox event, in one transaction.
func (r *JamaahRepo) TransitionDeparture(ctx context.Context, groupID, orgID uuid.UUID, to string, manifestFinalizedAt, departedAt *time.Time, eventType string, payload []byte) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	ct, err := tx.Exec(ctx, `UPDATE groups
		SET departure_status = $3,
		    manifest_finalized_at = COALESCE($4, manifest_finalized_at),
		    departed_at = COALESCE($5, departed_at),
		    updated_at = NOW()
		WHERE id = $1 AND org_id = $2`, groupID, orgID, to, manifestFinalizedAt, departedAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("group not found")
	}

	if eventType != "" {
		if err := outbox.Insert(ctx, tx, outbox.Event{
			OrgID:         orgID,
			AggregateType: "group_departure",
			AggregateID:   groupID,
			EventType:     eventType,
			Payload:       payload,
		}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
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
	_, _ = r.pool.Exec(ctx, `UPDATE groups SET member_count = member_count + 1, updated_at = NOW() WHERE id = $1`, gm.GroupID)
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
		_, _ = r.pool.Exec(ctx, `UPDATE groups SET member_count = GREATEST(member_count - 1, 0), updated_at = NOW() WHERE id = $1`, groupID)
	}
	return nil
}
