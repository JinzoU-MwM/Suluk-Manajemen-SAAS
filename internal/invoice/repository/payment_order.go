package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/invoice/model"
)

func (r *InvoiceRepo) CreatePaymentOrder(ctx context.Context, order *model.PaymentOrder) error {
	query := `
		INSERT INTO payment_orders (id, org_id, user_id, plan, plan_type, amount, status, redirect_url, gateway_ref)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		order.ID, order.OrgID, order.UserID, order.Plan, order.PlanType, order.Amount,
		order.Status, order.RedirectURL, order.GatewayRef,
	).Scan(&order.CreatedAt, &order.UpdatedAt)
}

func (r *InvoiceRepo) GetPaymentOrder(ctx context.Context, orderID, orgID uuid.UUID) (*model.PaymentOrder, error) {
	query := `SELECT id, org_id, user_id, plan, plan_type, amount, status, redirect_url, gateway_ref, payment_method, completed_at, created_at, updated_at
		FROM payment_orders WHERE id = $1 AND org_id = $2`
	return scanPaymentOrder(r.pool.QueryRow(ctx, query, orderID, orgID))
}

// GetPaymentOrderByID looks up an order without org scoping — used by the
// payment webhook, which only receives the order_id from Pakasir.
func (r *InvoiceRepo) GetPaymentOrderByID(ctx context.Context, orderID uuid.UUID) (*model.PaymentOrder, error) {
	query := `SELECT id, org_id, user_id, plan, plan_type, amount, status, redirect_url, gateway_ref, payment_method, completed_at, created_at, updated_at
		FROM payment_orders WHERE id = $1`
	return scanPaymentOrder(r.pool.QueryRow(ctx, query, orderID))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPaymentOrder(row rowScanner) (*model.PaymentOrder, error) {
	var o model.PaymentOrder
	err := row.Scan(
		&o.ID, &o.OrgID, &o.UserID, &o.Plan, &o.PlanType, &o.Amount,
		&o.Status, &o.RedirectURL, &o.GatewayRef, &o.PaymentMethod, &o.CompletedAt,
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

// MarkPaymentOrderPaid marks an order paid and records the gateway payment method.
func (r *InvoiceRepo) MarkPaymentOrderPaid(ctx context.Context, orderID uuid.UUID, paymentMethod string) error {
	query := `UPDATE payment_orders
		SET status = 'paid', payment_method = $2, completed_at = NOW(), updated_at = NOW()
		WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, orderID, paymentMethod)
	return err
}
