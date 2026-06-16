package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
)

// ListRegistrationsByJamaah returns all of a jamaah's package registrations
// (for the pilgrim self-service portal).
func (r *JamaahRepo) ListRegistrationsByJamaah(ctx context.Context, orgID, jamaahID uuid.UUID) ([]model.JamaahPackageRegistration, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, org_id, jamaah_id, package_id, room_type, price_snapshot, discount_amount, custom_price,
			pipeline_status, registered_at, dp_date, lunas_date, berangkat_date, mahram_id, internal_notes, created_at, updated_at
		FROM jamaah_package_registrations
		WHERE org_id = $1 AND jamaah_id = $2
		ORDER BY registered_at DESC`, orgID, jamaahID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.JamaahPackageRegistration{}
	for rows.Next() {
		var reg model.JamaahPackageRegistration
		if err := rows.Scan(&reg.ID, &reg.OrgID, &reg.JamaahID, &reg.PackageID, &reg.RoomType, &reg.PriceSnapshot,
			&reg.DiscountAmount, &reg.CustomPrice, &reg.PipelineStatus, &reg.RegisteredAt,
			&reg.DPDate, &reg.LunasDate, &reg.BerangkatDate, &reg.MahramID, &reg.InternalNotes, &reg.CreatedAt, &reg.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, reg)
	}
	return out, rows.Err()
}
