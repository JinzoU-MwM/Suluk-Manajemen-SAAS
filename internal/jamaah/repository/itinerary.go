package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jamaah-in/v2/internal/jamaah/model"
)

func (r *JamaahRepo) ListItineraries(ctx context.Context, groupID, orgID uuid.UUID) ([]model.Itinerary, error) {
	query := `SELECT id, org_id, group_id, day_number, title, description, location, start_time, end_time, sort_order, created_at, updated_at
		FROM itineraries WHERE group_id = $1 AND org_id = $2 ORDER BY day_number, sort_order`
	rows, err := r.pool.Query(ctx, query, groupID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Itinerary
	for rows.Next() {
		var it model.Itinerary
		if err := rows.Scan(&it.ID, &it.OrgID, &it.GroupID, &it.DayNumber, &it.Title, &it.Description, &it.Location, &it.StartTime, &it.EndTime, &it.SortOrder, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (r *JamaahRepo) CreateItinerary(ctx context.Context, it *model.Itinerary) error {
	query := `INSERT INTO itineraries (id, org_id, group_id, day_number, title, description, location, start_time, end_time, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at`
	return r.pool.QueryRow(ctx, query, it.ID, it.OrgID, it.GroupID, it.DayNumber, it.Title, it.Description, it.Location, it.StartTime, it.EndTime, it.SortOrder).Scan(&it.CreatedAt, &it.UpdatedAt)
}

func (r *JamaahRepo) GetItinerary(ctx context.Context, id, orgID uuid.UUID) (*model.Itinerary, error) {
	query := `SELECT id, org_id, group_id, day_number, title, description, location, start_time, end_time, sort_order, created_at, updated_at
		FROM itineraries WHERE id = $1 AND org_id = $2`
	var it model.Itinerary
	err := r.pool.QueryRow(ctx, query, id, orgID).Scan(&it.ID, &it.OrgID, &it.GroupID, &it.DayNumber, &it.Title, &it.Description, &it.Location, &it.StartTime, &it.EndTime, &it.SortOrder, &it.CreatedAt, &it.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get itinerary: %w", err)
	}
	return &it, nil
}

func (r *JamaahRepo) UpdateItinerary(ctx context.Context, it *model.Itinerary) error {
	query := `UPDATE itineraries SET day_number = $2, title = $3, description = $4, location = $5, start_time = $6, end_time = $7, sort_order = $8, updated_at = NOW()
		WHERE id = $1 AND org_id = $9`
	_, err := r.pool.Exec(ctx, query, it.ID, it.DayNumber, it.Title, it.Description, it.Location, it.StartTime, it.EndTime, it.SortOrder, it.OrgID)
	return err
}

func (r *JamaahRepo) DeleteItinerary(ctx context.Context, id, orgID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM itineraries WHERE id = $1 AND org_id = $2`, id, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("itinerary not found")
	}
	return nil
}
