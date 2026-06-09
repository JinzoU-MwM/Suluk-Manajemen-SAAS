package service

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/invoice/model"
	"github.com/jamaah-in/v2/internal/shared/plan"
)

// CreatePaymentOrder creates a pending order for a purchasable tier + period,
// prices it from the shared catalog, and returns the Pakasir checkout URL.
func (s *InvoiceService) CreatePaymentOrder(ctx context.Context, orgID, userID uuid.UUID, req model.CreatePaymentOrderRequest) (*model.PaymentOrderResponse, error) {
	planName := plan.Normalize(req.Plan)
	period := req.PlanType
	if period == "" {
		period = plan.PeriodMonthly
	}
	if period != plan.PeriodMonthly && period != plan.PeriodYearly {
		return nil, fmt.Errorf("invalid period: must be monthly or yearly")
	}
	amount, err := plan.PriceFor(planName, period)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("plan %q is not purchasable", planName)
	}

	orderID := uuid.New()

	order := &model.PaymentOrder{
		ID:       orderID,
		OrgID:    orgID,
		UserID:   userID,
		Plan:     planName,
		PlanType: period,
		Amount:   amount,
		Status:   "pending",
	}
	if err := s.repo.CreatePaymentOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("create payment order: %w", err)
	}

	payURL := s.pakasirPayURL(orderID.String(), amount)
	order.RedirectURL = &payURL

	return &model.PaymentOrderResponse{
		OrderID:    orderID.String(),
		PaymentURL: payURL,
		Status:     "pending",
		Amount:     amount,
	}, nil
}

// pakasirPayURL builds the Pakasir hosted-checkout URL (redirect method):
// {base}/pay/{slug}/{amount}?order_id={id}&redirect={publicURL}/?payment=success
func (s *InvoiceService) pakasirPayURL(orderID string, amount int64) string {
	base := s.pakasir.BaseURL
	if base == "" {
		base = "https://app.pakasir.com"
	}
	redirect := s.publicURL + "/?payment=success"
	return fmt.Sprintf("%s/pay/%s/%d?order_id=%s&redirect=%s",
		base, s.pakasir.ProjectSlug, amount, url.QueryEscape(orderID), url.QueryEscape(redirect))
}

func (s *InvoiceService) CheckPaymentStatus(ctx context.Context, orderIDStr string, orgID uuid.UUID) (*model.PaymentStatusResponse, error) {
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid order id")
	}

	order, err := s.repo.GetPaymentOrder(ctx, orderID, orgID)
	if err != nil {
		return nil, err
	}

	return &model.PaymentStatusResponse{
		OrderID: order.ID.String(),
		Status:  order.Status,
		Amount:  order.Amount,
	}, nil
}

// HandlePakasirWebhook processes a Pakasir completed-payment callback. Because
// Pakasir webhooks are unsigned, we (1) match the order_id + amount against our
// stored pending order and (2) independently confirm via the Transaction Detail
// API before marking the order paid and activating the subscription. Idempotent.
func (s *InvoiceService) HandlePakasirWebhook(ctx context.Context, p model.PakasirWebhookPayload) error {
	orderID, err := uuid.Parse(p.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order_id")
	}
	order, err := s.repo.GetPaymentOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status == "paid" {
		return nil // already processed
	}
	if p.Amount != order.Amount {
		return fmt.Errorf("amount mismatch: webhook %d vs order %d", p.Amount, order.Amount)
	}

	// Independent confirmation against Pakasir (defends against spoofed webhooks).
	status, err := s.VerifyTransaction(ctx, order.ID.String(), order.Amount)
	if err != nil {
		return fmt.Errorf("verify transaction: %w", err)
	}
	if status != "completed" {
		return fmt.Errorf("transaction not completed: %s", status)
	}

	if err := s.repo.MarkPaymentOrderPaid(ctx, order.ID, p.PaymentMethod); err != nil {
		return fmt.Errorf("mark paid: %w", err)
	}
	if err := s.activateSubscription(ctx, order.OrgID.String(), order.Plan, order.PlanType); err != nil {
		return fmt.Errorf("activate subscription: %w", err)
	}
	return nil
}

// activateSubscription calls the auth-service internal endpoint to flip the org
// onto the paid tier, authenticated with the shared internal key.
func (s *InvoiceService) activateSubscription(ctx context.Context, orgID, planName, period string) error {
	if s.authAddr == "" {
		return fmt.Errorf("auth service address not configured")
	}
	body := model.ActivatePlanBody{OrgID: orgID, Plan: planName, Period: period}
	headers := map[string]string{"X-Internal-Key": s.internalKey}
	return s.httpc.PostJSON(ctx, s.authAddr, "/api/v1/internal/subscription/activate", headers, body, nil)
}
