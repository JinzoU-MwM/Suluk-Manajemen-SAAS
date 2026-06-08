package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jamaah-in/v2/internal/jamaah/model"
)

func (r *JamaahRepo) CreateRoom(ctx context.Context, room *model.Room) error {
	query := `INSERT INTO rooms (id, org_id, group_id, room_number, gender_type, room_type, capacity, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
		RETURNING created_at, updated_at`
	var id uuid.UUID
	id, _ = uuid.Parse(room.ID)
	var orgID uuid.UUID
	orgID, _ = uuid.Parse(room.OrgID)
	var groupID *uuid.UUID
	if room.GroupID != nil && *room.GroupID != "" {
		gid, _ := uuid.Parse(*room.GroupID)
		groupID = &gid
	}
	return r.pool.QueryRow(ctx, query, id, orgID, groupID, room.RoomNumber, room.GenderType, room.RoomType, room.Capacity).Scan(&room.CreatedAt, &room.UpdatedAt)
}

func (r *JamaahRepo) ListRooms(ctx context.Context, orgID uuid.UUID, groupID *uuid.UUID) ([]model.Room, error) {
	query := `SELECT id, org_id, group_id, room_number, gender_type, room_type, capacity, is_active, created_at, updated_at
		FROM rooms WHERE org_id = $1`
	args := []any{orgID}
	argIdx := 2
	if groupID != nil {
		query += fmt.Sprintf(" AND group_id = $%d", argIdx)
		args = append(args, *groupID)
		argIdx++
	}
	query += " ORDER BY room_number"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []model.Room
	for rows.Next() {
		var rm model.Room
		var id, oid uuid.UUID
		var gid *uuid.UUID
		if err := rows.Scan(&id, &oid, &gid, &rm.RoomNumber, &rm.GenderType, &rm.RoomType, &rm.Capacity, &rm.IsActive, &rm.CreatedAt, &rm.UpdatedAt); err != nil {
			return nil, err
		}
		rm.ID = id.String()
		rm.OrgID = oid.String()
		if gid != nil {
			gs := gid.String()
			rm.GroupID = &gs
		}
		rooms = append(rooms, rm)
	}
	return rooms, nil
}

func (r *JamaahRepo) DeleteRoom(ctx context.Context, roomID, orgID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM rooms WHERE id = $1 AND org_id = $2`, roomID, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("room not found")
	}
	return nil
}

func (r *JamaahRepo) DeleteRoomsByGroup(ctx context.Context, groupID, orgID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM rooms WHERE group_id = $1 AND org_id = $2`, groupID, orgID)
	return err
}

func (r *JamaahRepo) AssignMemberToRoom(ctx context.Context, orgID, roomID uuid.UUID, memberID string) error {
	query := `INSERT INTO room_assignments (id, org_id, room_id, member_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, uuid.New(), orgID, roomID, memberID)
	return err
}

func (r *JamaahRepo) UnassignMember(ctx context.Context, roomID uuid.UUID, memberID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM room_assignments WHERE room_id = $1 AND member_id = $2`, roomID, memberID)
	return err
}

func (r *JamaahRepo) GetRoomAssignments(ctx context.Context, roomID uuid.UUID) ([]model.RoomAssignment, error) {
	query := `SELECT id, org_id, room_id, member_id FROM room_assignments WHERE room_id = $1`
	rows, err := r.pool.Query(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []model.RoomAssignment
	for rows.Next() {
		var a model.RoomAssignment
		var id, oid, rid uuid.UUID
		if err := rows.Scan(&id, &oid, &rid, &a.MemberID); err != nil {
			return nil, err
		}
		a.ID = id.String()
		a.OrgID = oid.String()
		a.RoomID = rid.String()
		assignments = append(assignments, a)
	}
	return assignments, nil
}

func (r *JamaahRepo) CreateSharedManifest(ctx context.Context, sm *model.SharedManifest) error {
	query := `INSERT INTO shared_manifests (id, org_id, group_id, token, pin_hash, expires_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE)
		RETURNING created_at`
	var id, orgID uuid.UUID
	id, _ = uuid.Parse(sm.ID)
	orgID, _ = uuid.Parse(sm.OrgID)
	var groupID *uuid.UUID
	if sm.GroupID != nil && *sm.GroupID != "" {
		gid, _ := uuid.Parse(*sm.GroupID)
		groupID = &gid
	}
	return r.pool.QueryRow(ctx, query, id, orgID, groupID, sm.Token, sm.PinHash, sm.ExpiresAt).Scan(&sm.CreatedAt)
}

func (r *JamaahRepo) RevokeSharedManifest(ctx context.Context, groupID, orgID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE shared_manifests SET is_active = FALSE WHERE group_id = $1 AND org_id = $2`, groupID, orgID)
	return err
}

func (r *JamaahRepo) GetSharedManifestByToken(ctx context.Context, token string) (*model.SharedManifest, error) {
	query := `SELECT id, org_id, group_id, token, pin_hash, expires_at, is_active, created_at
		FROM shared_manifests WHERE token = $1 AND is_active = TRUE`
	var sm model.SharedManifest
	var id, orgID uuid.UUID
	var groupID *uuid.UUID
	err := r.pool.QueryRow(ctx, query, token).Scan(&id, &orgID, &groupID, &sm.Token, &sm.PinHash, &sm.ExpiresAt, &sm.IsActive, &sm.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sm.ID = id.String()
	sm.OrgID = orgID.String()
	if groupID != nil {
		gs := groupID.String()
		sm.GroupID = &gs
	}
	return &sm, nil
}

func (r *JamaahRepo) GetRoomingSummary(ctx context.Context, orgID uuid.UUID, groupID *uuid.UUID) (*model.RoomingSummary, error) {
	query := `SELECT COALESCE(SUM(capacity), 0), COUNT(*) FROM rooms WHERE org_id = $1 AND is_active = TRUE`
	args := []any{orgID}
	argIdx := 2
	if groupID != nil {
		query += fmt.Sprintf(" AND group_id = $%d", argIdx)
		args = append(args, *groupID)
		argIdx++
	}
	var totalCapacity, totalRooms int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&totalCapacity, &totalRooms)
	if err != nil {
		return nil, err
	}

	assignQuery := `SELECT COUNT(*) FROM room_assignments ra JOIN rooms r ON r.id = ra.room_id WHERE r.org_id = $1`
	args2 := []any{orgID}
	argIdx2 := 2
	if groupID != nil {
		assignQuery += fmt.Sprintf(" AND r.group_id = $%d", argIdx2)
		args2 = append(args2, *groupID)
	}
	var assignedCount int
	r.pool.QueryRow(ctx, assignQuery, args2...).Scan(&assignedCount)

	occupancyPct := 0
	if totalCapacity > 0 {
		occupancyPct = (assignedCount * 100) / totalCapacity
	}

	return &model.RoomingSummary{
		TotalRooms:    totalRooms,
		TotalCapacity: totalCapacity,
		AssignedCount: assignedCount,
		OccupancyPct:  occupancyPct,
	}, nil
}
