package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jamaah-in/v2/internal/invoice/model"
	"github.com/jamaah-in/v2/internal/shared/plan"
)

// BadWebhookError marks a webhook payload as permanently invalid (a client
// error → HTTP 4xx): the gateway should NOT retry it. Any other error returned
// from webhook handling is treated as transient (HTTP 5xx) so Pakasir retries.
type BadWebhookError struct{ msg string }

func (e *BadWebhookError) Error() string { return e.msg }

func errBadWebhook(format string, a ...any) error {
	return &BadWebhookError{msg: fmt.Sprintf(format, a...)}
}

// CreatePaymentOrder creates a pending order for a purchasable tier + period,
// prices it from the shared catalog, and returns the Pakasir checkout URL.
func (s *InvoiceService) CreatePaymentOrder(ctx context.Context, orgID, userID uuid.UUID, req model.CreatePaymentOrderRequest) (*model.PaymentOrderResponse, error) {
	planName := plan.Normalize(req.Plan)
	// Normalize the period (the frontend sends "annual" for yearly) so the
	// canonical "monthly"/"yearly" is stored and priced consistently — a raw
	// equality check here rejected every annual purchase outright.
	period, ok := plan.NormalizePeriod(req.PlanType)
	if !ok {
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

// callerPlan fetches the caller org's tier from auth (forwarding the caller's
// bearer token), used to gate Starter-only purchases.
func (s *InvoiceService) callerPlan(ctx context.Context, authToken string) (string, error) {
	var out struct {
		Plan string `json:"plan"`
	}
	if err := s.httpc.GetJSON(ctx, s.authAddr, "/api/v1/subscription/status", authToken, &out); err != nil {
		return "", fmt.Errorf("check plan: %w", err)
	}
	return out.Plan, nil
}

// CreateTopupOrder creates a pending scan-topup order (server-priced) and returns
// the Pakasir checkout URL. Only Starter orgs may buy top-ups.
func (s *InvoiceService) CreateTopupOrder(ctx context.Context, orgID, userID uuid.UUID, authToken string) (*model.PaymentOrderResponse, error) {
	p, err := s.callerPlan(ctx, authToken)
	if err != nil {
		return nil, err
	}
	if plan.Normalize(p) != plan.Starter {
		return nil, fmt.Errorf("top-up tersedia hanya untuk paket Starter")
	}

	orderID := uuid.New()
	order := &model.PaymentOrder{
		ID:      orderID,
		OrgID:   orgID,
		UserID:  userID,
		Plan:    plan.Starter,
		Amount:  plan.ScanTopupPrice,
		Status:  "pending",
		Purpose: "scan_topup",
	}
	if err := s.repo.CreatePaymentOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("create topup order: %w", err)
	}

	payURL := s.pakasirPayURL(orderID.String(), plan.ScanTopupPrice)
	order.RedirectURL = &payURL
	return &model.PaymentOrderResponse{
		OrderID:    orderID.String(),
		PaymentURL: payURL,
		Status:     "pending",
		Amount:     plan.ScanTopupPrice,
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
		return errBadWebhook("invalid order_id")
	}
	// Defense-in-depth: reject callbacks for a different project or a non-completed
	// status before doing any work. The Transaction Detail call below is the
	// authoritative guard, but these cheap checks drop obvious spoofs/noise.
	if s.pakasir.ProjectSlug != "" && p.Project != "" && p.Project != s.pakasir.ProjectSlug {
		return errBadWebhook("project mismatch: %s", p.Project)
	}
	if p.Status != "" && p.Status != "completed" {
		return errBadWebhook("webhook status not completed: %s", p.Status)
	}

	order, err := s.repo.GetPaymentOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errBadWebhook("unknown order_id")
		}
		return err // transient DB error → 5xx → retry
	}
	if order.Status == "paid" {
		return nil // already fully processed (mark-paid is the final step below)
	}
	if p.Amount != order.Amount {
		return errBadWebhook("amount mismatch: webhook %d vs order %d", p.Amount, order.Amount)
	}

	// Authoritative confirmation against Pakasir (webhooks are unsigned). Confirm
	// both the completed status AND the actual paid amount — the amount in the
	// webhook body is attacker-controlled, so it cannot be the only amount check.
	status, paidAmount, err := s.VerifyTransaction(ctx, order.ID.String(), order.Amount)
	if err != nil {
		return fmt.Errorf("verify transaction: %w", err) // transient → 5xx → retry
	}
	if status != "completed" {
		return errBadWebhook("transaction not completed: %s", status)
	}
	if paidAmount != order.Amount {
		return errBadWebhook("verified amount mismatch: paid %d vs order %d", paidAmount, order.Amount)
	}

	// Atomically claim the order (single winner under concurrent/duplicate
	// delivery). A false result means another delivery already processed it.
	claimed, err := s.repo.MarkPaymentOrderPaid(ctx, order.ID, p.PaymentMethod)
	if err != nil {
		return fmt.Errorf("mark paid: %w", err)
	}
	if !claimed {
		return nil
	}

	// Activate AFTER claiming. If activation fails, roll the order back to
	// pending so a later Pakasir retry re-attempts it — this avoids the
	// paid-but-never-activated stuck state (customer charged, org left on free).
	if err := s.activateSubscription(ctx, order, p.PaymentMethod); err != nil {
		if rerr := s.repo.RevertPaymentOrderToPending(ctx, order.ID); rerr != nil {
			log.Printf("CRITICAL: order %s paid but activation AND revert failed — needs reconciliation: activate=%v revert=%v", order.ID, err, rerr)
		}
		return fmt.Errorf("activate subscription: %w", err) // transient → 5xx → retry
	}
	return nil
}

// activateSubscription calls the auth-service internal endpoint to flip the org
// onto the paid tier (which also sends the invoice/confirmation email),
// authenticated with the shared internal key.
func (s *InvoiceService) activateSubscription(ctx context.Context, order *model.PaymentOrder, paymentMethod string) error {
	if s.authAddr == "" {
		return fmt.Errorf("auth service address not configured")
	}
	body := model.ActivatePlanBody{
		OrgID:         order.OrgID.String(),
		UserID:        order.UserID.String(),
		Plan:          order.Plan,
		Period:        order.PlanType,
		Amount:        order.Amount,
		OrderID:       order.ID.String(),
		PaymentMethod: paymentMethod,
	}
	headers := map[string]string{"X-Internal-Key": s.internalKey}
	return s.httpc.PostJSON(ctx, s.authAddr, "/api/v1/internal/subscription/activate", headers, body, nil)
}
