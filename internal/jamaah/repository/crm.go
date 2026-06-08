package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
)

// ListCRM returns jamaah profiles joined with their most-recent package
// registration (status/package/room/price), paginated and searchable. Invoice
// balances are attached by the service layer.
func (r *JamaahRepo) ListCRM(ctx context.Context, orgID uuid.UUID, search string, offset, limit int) ([]model.CRMJamaahRow, int, error) {
	where := `WHERE p.org_id = $1`
	args := []any{orgID}
	if search != "" {
		where += ` AND (p.nama ILIKE $2 OR p.no_identitas ILIKE $2 OR p.no_paspor ILIKE $2 OR p.no_hp ILIKE $2 OR p.email ILIKE $2)`
		args = append(args, "%"+search+"%")
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jamaah_profiles p `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `
		SELECT p.id, p.nama, COALESCE(p.no_hp, ''), COALESCE(p.no_identitas, ''), COALESCE(p.no_paspor, ''),
			COALESCE(p.email, ''), COALESCE(p.gender, ''),
			r.package_id, COALESCE(r.room_type, ''), COALESCE(r.pipeline_status, ''),
			COALESCE(r.price_snapshot, 0), COALESCE(r.discount_amount, 0)
		FROM jamaah_profiles p
		LEFT JOIN LATERAL (
			SELECT package_id, room_type, pipeline_status, price_snapshot, discount_amount
			FROM jamaah_package_registrations
			WHERE jamaah_id = p.id
			ORDER BY registered_at DESC
			LIMIT 1
		) r ON TRUE
		` + where + fmt.Sprintf(` ORDER BY p.created_at DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]model.CRMJamaahRow, 0)
	for rows.Next() {
		var row model.CRMJamaahRow
		if err := rows.Scan(&row.ID, &row.Nama, &row.NoHP, &row.NoIdentitas, &row.NoPaspor,
			&row.Email, &row.Gender, &row.PackageID, &row.RoomType, &row.PipelineStatus,
			&row.PriceSnapshot, &row.DiscountAmount); err != nil {
			return nil, 0, err
		}
		out = append(out, row)
	}
	return out, total, rows.Err()
}
