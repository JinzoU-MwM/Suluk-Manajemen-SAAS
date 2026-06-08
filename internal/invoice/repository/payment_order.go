package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/invoice/model"
)

func (r *InvoiceRepo) CreatePaymentOrder(ctx context.Context, order *model.PaymentOrder) error {
	query := `
		INSERT INTO payment_orders (id, org_id, user_id, plan_type, amount, status, redirect_url, gateway_ref)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		order.ID, order.OrgID, order.UserID, order.PlanType, order.Amount,
		order.Status, order.RedirectURL, order.GatewayRef,
	).Scan(&order.CreatedAt, &order.UpdatedAt)
}

func (r *InvoiceRepo) GetPaymentOrder(ctx context.Context, orderID, orgID uuid.UUID) (*model.PaymentOrder, error) {
	query := `SELECT id, org_id, user_id, plan_type, amount, status, redirect_url, gateway_ref, created_at, updated_at
		FROM payment_orders WHERE id = $1 AND org_id = $2`
	var o model.PaymentOrder
	err := r.pool.QueryRow(ctx, query, orderID, orgID).Scan(
		&o.ID, &o.OrgID, &o.UserID, &o.PlanType, &o.Amount,
		&o.Status, &o.RedirectURL, &o.GatewayRef,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get payment order: %w", err)
	}
	return &o, nil
}

func (r *InvoiceRepo) UpdatePaymentOrderStatus(ctx context.Context, orderID, orgID uuid.UUID, status string) error {
	query := `UPDATE payment_orders SET status = $1, updated_at = NOW() WHERE id = $2 AND org_id = $3`
	_, err := r.pool.Exec(ctx, query, status, orderID, orgID)
	return err
}
