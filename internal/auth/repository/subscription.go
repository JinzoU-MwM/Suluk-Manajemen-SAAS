package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jamaah-in/v2/internal/auth/model"
)

func (r *AuthRepo) GetSubscription(ctx context.Context, orgID uuid.UUID) (*model.Subscription, error) {
	query := `SELECT id, org_id, plan, status, starts_at, expires_at, trial_used, created_at, updated_at
		FROM subscriptions WHERE org_id = $1`
	var sub model.Subscription
	err := r.pool.QueryRow(ctx, query, orgID).Scan(
		&sub.ID, &sub.OrgID, &sub.Plan, &sub.Status,
		&sub.StartsAt, &sub.ExpiresAt, &sub.TrialUsed,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	return &sub, nil
}

func (r *AuthRepo) CreateSubscription(ctx context.Context, sub *model.Subscription) error {
	query := `INSERT INTO subscriptions (id, org_id, plan, status, starts_at, expires_at, trial_used)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		sub.ID, sub.OrgID, sub.Plan, sub.Status,
		sub.StartsAt, sub.ExpiresAt, sub.TrialUsed,
	).Scan(&sub.CreatedAt, &sub.UpdatedAt)
}

func (r *AuthRepo) UpdateSubscription(ctx context.Context, sub *model.Subscription) error {
	query := `UPDATE subscriptions SET plan = $2, status = $3, expires_at = $4, trial_used = $5, updated_at = NOW()
		WHERE org_id = $1`
	_, err := r.pool.Exec(ctx, query, sub.OrgID, sub.Plan, sub.Status, sub.ExpiresAt, sub.TrialUsed)
	return err
}
