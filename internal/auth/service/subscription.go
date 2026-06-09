package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/auth/model"
	"github.com/jamaah-in/v2/internal/shared/plan"
)

// ErrPlanLimit is returned when an action would exceed the org's plan quota.
// Handlers map it to a 4xx (not a 500).
var ErrPlanLimit = errors.New("plan limit reached")

func statusResponse(p, status string, expiresAt *time.Time) *model.SubscriptionStatusResponse {
	t := plan.Get(p)
	return &model.SubscriptionStatusResponse{
		Plan:      t.Key,
		Status:    status,
		ExpiresAt: expiresAt,
		Rank:      t.Rank,
		MaxJamaah: t.MaxJamaah,
		MaxGroups: t.MaxGroups,
		MaxUsers:  t.MaxUsers,
	}
}

func (s *AuthService) GetSubscriptionStatus(ctx context.Context, orgID uuid.UUID) (*model.SubscriptionStatusResponse, error) {
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return statusResponse(plan.Gratis, "active", nil), nil
	}
	if sub.ExpiresAt != nil && sub.ExpiresAt.Before(time.Now()) && sub.Status != "cancelled" {
		sub.Status = "expired"
		_ = s.repo.UpdateSubscription(ctx, sub)
	}
	return statusResponse(sub.Plan, sub.Status, sub.ExpiresAt), nil
}

// ActivatePlan sets the org's subscription to the given paid tier with an expiry
// based on the billing period (monthly → +1 month, yearly → +1 year). It returns
// the new expiry so callers can include it in receipts.
func (s *AuthService) ActivatePlan(ctx context.Context, orgID uuid.UUID, planName, period string) (time.Time, error) {
	planName = plan.Normalize(planName)
	if !plan.Valid(planName) {
		return time.Time{}, fmt.Errorf("invalid plan: %s", planName)
	}
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return time.Time{}, err
	}
	now := time.Now()
	var expiresAt time.Time
	if period == plan.PeriodYearly || period == "annual" {
		expiresAt = now.AddDate(1, 0, 0)
	} else {
		expiresAt = now.AddDate(0, 1, 0)
	}
	if sub == nil {
		if err := s.repo.CreateSubscription(ctx, &model.Subscription{
			ID:        uuid.New(),
			OrgID:     orgID,
			Plan:      planName,
			Status:    "active",
			StartsAt:  now,
			ExpiresAt: &expiresAt,
		}); err != nil {
			return time.Time{}, err
		}
		return expiresAt, nil
	}
	sub.Plan = planName
	sub.Status = "active"
	sub.ExpiresAt = &expiresAt
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return time.Time{}, err
	}
	return expiresAt, nil
}

// UpgradeToPro is retained for the existing authenticated upgrade route; it now
// delegates to ActivatePlan with the Pro monthly tier.
func (s *AuthService) UpgradeToPro(ctx context.Context, orgID uuid.UUID, req model.UpgradeRequest) error {
	_, err := s.ActivatePlan(ctx, orgID, plan.Pro, plan.PeriodMonthly)
	return err
}

// SendSubscriptionInvoice emails the buyer a combined payment-confirmation +
// invoice/receipt. Best-effort: any error is returned for logging but must not
// fail the activation flow.
func (s *AuthService) SendSubscriptionInvoice(ctx context.Context, req model.ActivatePlanRequest, expiresAt time.Time) error {
	if s.email == nil || !s.email.Enabled() {
		return nil
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	orgID, _ := uuid.Parse(req.OrgID)
	orgName := ""
	if org, err := s.repo.GetOrganizationByID(ctx, orgID); err == nil && org != nil {
		orgName = org.Name
	}
	tier := plan.Get(req.Plan)
	appURL := strings.TrimSpace(os.Getenv("APP_PUBLIC_URL"))
	if appURL == "" {
		appURL = "https://suluk.site"
	}
	subject, html := buildInvoiceEmail(invoiceData{
		OrgName:       orgName,
		CustomerName:  user.Name,
		CustomerEmail: user.Email,
		PlanName:      tier.Name,
		Period:        req.Period,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		OrderID:       req.OrderID,
		StartsAt:      time.Now(),
		ExpiresAt:     expiresAt,
		AppURL:        appURL,
		Features:      tier.Features,
	})
	return s.email.Send(ctx, user.Email, subject, html)
}

// CreateNotification persists an in-app notification (id auto-filled).
func (s *AuthService) CreateNotification(ctx context.Context, n *model.Notification) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.Severity == "" {
		n.Severity = "info"
	}
	return s.repo.CreateNotification(ctx, n)
}

func (s *AuthService) GetTrialStatus(ctx context.Context, orgID uuid.UUID) (*model.TrialStatusResponse, error) {
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if sub != nil && sub.TrialUsed {
		return &model.TrialStatusResponse{
			TrialAvailable: false,
			TrialDays:      0,
		}, nil
	}
	return &model.TrialStatusResponse{
		TrialAvailable: true,
		TrialDays:      14,
	}, nil
}

func (s *AuthService) ActivateTrial(ctx context.Context, orgID uuid.UUID) error {
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return err
	}
	if sub != nil && sub.TrialUsed {
		return fmt.Errorf("trial already used")
	}
	now := time.Now()
	expiresAt := now.AddDate(0, 0, 14)
	if sub == nil {
		sub = &model.Subscription{
			ID:        uuid.New(),
			OrgID:     orgID,
			Plan:      "pro",
			Status:    "trial",
			StartsAt:  now,
			ExpiresAt: &expiresAt,
			TrialUsed: true,
		}
		return s.repo.CreateSubscription(ctx, sub)
	}
	sub.Plan = "pro"
	sub.Status = "trial"
	sub.ExpiresAt = &expiresAt
	sub.TrialUsed = true
	return s.repo.UpdateSubscription(ctx, sub)
}

func (s *AuthService) GetPricing(ctx context.Context) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(plan.Ordered))
	for _, t := range plan.Ordered {
		out = append(out, map[string]any{
			"key":           t.Key,
			"name":          t.Name,
			"rank":          t.Rank,
			"monthly_price": t.MonthlyPrice,
			"annual_price":  t.AnnualPrice,
			"max_jamaah":    t.MaxJamaah,
			"max_groups":    t.MaxGroups,
			"max_users":     t.MaxUsers,
			"purchasable":   t.Purchasable,
			"features":      t.Features,
		})
	}
	return out, nil
}
