package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jamaah-in/v2/internal/inventory/model"
)

type InventoryRepo struct {
	pool *pgxpool.Pool
}

func NewInventoryRepo(pool *pgxpool.Pool) *InventoryRepo {
	return &InventoryRepo{pool: pool}
}

func (r *InventoryRepo) UpsertMembers(ctx context.Context, orgID string, req model.SyncMembersRequest) error {
	query := `
		INSERT INTO member_equipment (org_id, package_id, member_id, nama, gender, baju_size, family_id, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (package_id, member_id)
		DO UPDATE SET nama = $4, gender = $5, baju_size = $6, family_id = $7, updated_at = NOW()
	`
	batch := &pgx.Batch{}
	for _, m := range req.Members {
		batch.Queue(query, orgID, req.PackageID, m.MemberID, m.Nama, m.Gender, m.BajuSize, m.FamilyID)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range req.Members {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *InventoryRepo) ListByPackage(ctx context.Context, orgID, packageID string) ([]model.MemberEquipment, error) {
	query := `
		SELECT id, org_id, package_id, member_id, nama, gender, baju_size, family_id,
		       is_equipment_received, COALESCE(received_items, '{}'), received_at, created_at, updated_at
		FROM member_equipment
		WHERE org_id = $1 AND package_id = $2
		ORDER BY nama
	`
	rows, err := r.pool.Query(ctx, query, orgID, packageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.MemberEquipment
	for rows.Next() {
		var m model.MemberEquipment
		if err := rows.Scan(
			&m.ID, &m.OrgID, &m.PackageID, &m.MemberID, &m.Nama, &m.Gender,
			&m.BajuSize, &m.FamilyID, &m.IsEquipmentReceived, &m.ReceivedItems,
			&m.ReceivedAt, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}

func (r *InventoryRepo) ListReceived(ctx context.Context, orgID, packageID string) ([]model.MemberEquipment, error) {
	query := `
		SELECT id, org_id, package_id, member_id, nama, gender, baju_size, family_id,
		       is_equipment_received, COALESCE(received_items, '{}'), received_at, created_at, updated_at
		FROM member_equipment
		WHERE org_id = $1 AND package_id = $2 AND is_equipment_received = TRUE
		ORDER BY nama
	`
	rows, err := r.pool.Query(ctx, query, orgID, packageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.MemberEquipment
	for rows.Next() {
		var m model.MemberEquipment
		if err := rows.Scan(
			&m.ID, &m.OrgID, &m.PackageID, &m.MemberID, &m.Nama, &m.Gender,
			&m.BajuSize, &m.FamilyID, &m.IsEquipmentReceived, &m.ReceivedItems,
			&m.ReceivedAt, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}

func (r *InventoryRepo) ListPending(ctx context.Context, orgID, packageID string) ([]model.MemberEquipment, error) {
	query := `
		SELECT id, org_id, package_id, member_id, nama, gender, baju_size, family_id,
		       is_equipment_received, COALESCE(received_items, '{}'), received_at, created_at, updated_at
		FROM member_equipment
		WHERE org_id = $1 AND package_id = $2 AND is_equipment_received = FALSE
		ORDER BY nama
	`
	rows, err := r.pool.Query(ctx, query, orgID, packageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.MemberEquipment
	for rows.Next() {
		var m model.MemberEquipment
		if err := rows.Scan(
			&m.ID, &m.OrgID, &m.PackageID, &m.MemberID, &m.Nama, &m.Gender,
			&m.BajuSize, &m.FamilyID, &m.IsEquipmentReceived, &m.ReceivedItems,
			&m.ReceivedAt, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}

func (r *InventoryRepo) MarkReceived(ctx context.Context, orgID, packageID string, memberIDs []string, itemsReceived []string) (int64, error) {
	now := time.Now()
	query := `
		UPDATE member_equipment
		SET is_equipment_received = TRUE, received_items = $1, received_at = $2, updated_at = $2
		WHERE org_id = $3 AND package_id = $4 AND member_id = ANY($5) AND is_equipment_received = FALSE
	`
	tag, err := r.pool.Exec(ctx, query, itemsReceived, now, orgID, packageID, memberIDs)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *InventoryRepo) UpdateOperational(ctx context.Context, orgID, memberID string, bajuSize string, familyID string) error {
	query := `
		UPDATE member_equipment
		SET baju_size = $1, family_id = $2, updated_at = NOW()
		WHERE org_id = $3 AND member_id = $4
	`
	_, err := r.pool.Exec(ctx, query, bajuSize, familyID, orgID, memberID)
	return err
}

func (r *InventoryRepo) GetByMemberID(ctx context.Context, orgID, memberID string) (*model.MemberEquipment, error) {
	query := `
		SELECT id, org_id, package_id, member_id, nama, gender, baju_size, family_id,
		       is_equipment_received, COALESCE(received_items, '{}'), received_at, created_at, updated_at
		FROM member_equipment
		WHERE org_id = $1 AND member_id = $2
		LIMIT 1
	`
	var m model.MemberEquipment
	err := r.pool.QueryRow(ctx, query, orgID, memberID).Scan(
		&m.ID, &m.OrgID, &m.PackageID, &m.MemberID, &m.Nama, &m.Gender,
		&m.BajuSize, &m.FamilyID, &m.IsEquipmentReceived, &m.ReceivedItems,
		&m.ReceivedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
