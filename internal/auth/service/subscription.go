package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/auth/model"
	"github.com/jamaah-in/v2/internal/shared/plan"
	"github.com/jamaah-in/v2/internal/shared/sign"
)

// ErrPlanLimit is returned when an action would exceed the org's plan quota.
// Handlers map it to a 4xx (not a 500).
var ErrPlanLimit = errors.New("plan limit reached")

func statusResponse(p, status string, expiresAt *time.Time) *model.SubscriptionStatusResponse {
	t := plan.Get(p)
	return &model.SubscriptionStatusResponse{
		Plan:       t.Key,
		Status:     status,
		ExpiresAt:  expiresAt,
		Rank:       t.Rank,
		MaxJamaah:  t.MaxJamaah,
		MaxGroups:  t.MaxGroups,
		MaxUsers:   t.MaxUsers,
		UsageLimit: t.MaxScansPerMonth,
	}
}

type scanUsageCacheEntry struct {
	count  int
	expiry time.Time
}

// scanUsageThisMonth returns the org's AI-scan count for the current calendar
// month, fetched from the ai-ocr service (which owns the scan_usage table in its
// own DB). It FAILS OPEN (returns 0) when unconfigured or on any error, so a
// flaky ai-ocr never breaks subscription status. Results are cached briefly per
// org so frequent status polls don't hammer ai-ocr (mirrors jamaah's limit cache).
func (s *AuthService) scanUsageThisMonth(ctx context.Context, orgID uuid.UUID) int {
	if s.aiocrAddr == "" || s.internalKey == "" {
		return 0
	}
	if v, ok := s.scanUsageCache.Load(orgID); ok {
		if e := v.(scanUsageCacheEntry); e.expiry.After(time.Now()) {
			return e.count
		}
	}
	var out struct {
		DocumentsScanned int `json:"documents_scanned"`
	}
	if err := s.httpc.PostJSON(ctx, s.aiocrAddr, "/api/v1/internal/scan-usage",
		map[string]string{"X-Internal-Key": s.internalKey},
		map[string]string{"org_id": orgID.String()}, &out); err != nil {
		return 0
	}
	s.scanUsageCache.Store(orgID, scanUsageCacheEntry{count: out.DocumentsScanned, expiry: time.Now().Add(45 * time.Second)})
	return out.DocumentsScanned
}

func (s *AuthService) GetSubscriptionStatus(ctx context.Context, orgID uuid.UUID) (*model.SubscriptionStatusResponse, error) {
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, err
	}
	// Org's scans this month (from ai-ocr); fail-open 0. Computed once and set on
	// whichever response path runs so the frontend quota bar always has a value.
	usage := s.scanUsageThisMonth(ctx, orgID)
	if sub == nil {
		resp := statusResponse(plan.Gratis, "active", nil)
		resp.UsageCount = usage
		return resp, nil
	}
	if sub.ExpiresAt != nil && sub.ExpiresAt.Before(time.Now()) && sub.Status != "cancelled" {
		sub.Status = "expired"
		if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
			// Non-fatal, but log so a persistently failing write is visible.
			log.Printf("auto-expire subscription (org %s): %v", orgID, err)
		}
	}
	// An expired (or cancelled) subscription must not keep paid-tier limits —
	// report Gratis limits/rank so quota enforcement downgrades the org. The
	// original plan name is still surfaced so the UI can prompt a renewal.
	if sub.Status == "expired" || sub.Status == "cancelled" {
		resp := statusResponse(plan.Gratis, sub.Status, sub.ExpiresAt)
		resp.Plan = plan.Normalize(sub.Plan)
		resp.UsageCount = usage
		return resp, nil
	}
	resp := statusResponse(sub.Plan, sub.Status, sub.ExpiresAt)
	resp.CancelAtPeriodEnd = sub.CancelAtPeriodEnd
	resp.UsageCount = usage
	return resp, nil
}

// ActivatePlan sets the org's subscription to the given paid tier with an expiry
// based on the billing period (monthly → +1 month, yearly → +1 year). It returns
// the new expiry so callers can include it in receipts.
func (s *AuthService) ActivatePlan(ctx context.Context, orgID uuid.UUID, planName, period string) (time.Time, error) {
	planName = plan.Normalize(planName)
	if !plan.Valid(planName) {
		return time.Time{}, fmt.Errorf("invalid plan: %s", planName)
	}
	canonicalPeriod, ok := plan.NormalizePeriod(period)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid period: %s", period)
	}
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return time.Time{}, err
	}
	now := time.Now()
	var expiresAt time.Time
	if canonicalPeriod == plan.PeriodYearly {
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
	sub.CancelAtPeriodEnd = false // re-committing/renewing clears a pending cancel
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
	// Signed public link to the subscription-invoice PDF (served by invoice-service);
	// the HMAC uses the shared INTERNAL_API_KEY so invoice-service can verify it.
	pdfURL := ""
	if key := strings.TrimSpace(os.Getenv("INTERNAL_API_KEY")); key != "" && req.OrderID != "" {
		pdfURL = appURL + "/api/payment/invoice/" + req.OrderID + "?sig=" + sign.Token(req.OrderID, key)
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
		PDFURL:        pdfURL,
		Features:      tier.Features,
	})
	return s.email.Send(ctx, user.Email, subject, html)
}

// GetBillingInfo returns the org + buyer display fields used on the subscription
// invoice (PDF), looked up by the invoice-service when rendering the receipt.
func (s *AuthService) GetBillingInfo(ctx context.Context, orgID, userID uuid.UUID) (orgName, userName, userEmail string, err error) {
	user, userErr := s.repo.GetUserByID(ctx, userID)
	if userErr == nil && user != nil {
		userName, userEmail = user.Name, user.Email
	}
	org, orgErr := s.repo.GetOrganizationByID(ctx, orgID)
	if orgErr == nil && org != nil {
		orgName = org.Name
	}
	// If neither record resolved, surface the failure instead of returning a
	// blank-name invoice silently.
	if userName == "" && orgName == "" {
		return "", "", "", fmt.Errorf("billing info not found for org %s / user %s", orgID, userID)
	}
	return orgName, userName, userEmail, nil
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

// CancelSubscription flags a paid, active subscription to not renew. The org
// keeps its tier and limits until expires_at, after which the existing
// auto-expire drops it to Gratis. Returns the refreshed status.
func (s *AuthService) CancelSubscription(ctx context.Context, orgID uuid.UUID) (*model.SubscriptionStatusResponse, error) {
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if err := model.CanCancelSubscription(sub, time.Now()); err != nil {
		return nil, err
	}
	sub.CancelAtPeriodEnd = true
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return nil, err
	}
	return s.GetSubscriptionStatus(ctx, orgID)
}

// ResumeSubscription undoes a pending cancel-at-period-end (before expiry).
func (s *AuthService) ResumeSubscription(ctx context.Context, orgID uuid.UUID) (*model.SubscriptionStatusResponse, error) {
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if err := model.CanResumeSubscription(sub, time.Now()); err != nil {
		return nil, err
	}
	sub.CancelAtPeriodEnd = false
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return nil, err
	}
	return s.GetSubscriptionStatus(ctx, orgID)
}
