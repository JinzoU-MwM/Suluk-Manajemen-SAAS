package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/auth/model"
)

func (s *AuthService) GetSubscriptionStatus(ctx context.Context, orgID uuid.UUID) (*model.SubscriptionStatusResponse, error) {
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return &model.SubscriptionStatusResponse{
			Plan:   "starter",
			Status: "active",
		}, nil
	}
	if sub.ExpiresAt != nil && sub.ExpiresAt.Before(time.Now()) && sub.Status != "cancelled" {
		sub.Status = "expired"
		_ = s.repo.UpdateSubscription(ctx, sub)
	}
	return &model.SubscriptionStatusResponse{
		Plan:      sub.Plan,
		Status:    sub.Status,
		ExpiresAt: sub.ExpiresAt,
	}, nil
}

func (s *AuthService) UpgradeToPro(ctx context.Context, orgID uuid.UUID, req model.UpgradeRequest) error {
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return err
	}
	now := time.Now()
	expiresAt := now.AddDate(0, 1, 0)
	if sub == nil {
		sub = &model.Subscription{
			ID:       uuid.New(),
			OrgID:    orgID,
			Plan:     "pro",
			Status:   "active",
			StartsAt: now,
			ExpiresAt: &expiresAt,
		}
		return s.repo.CreateSubscription(ctx, sub)
	}
	sub.Plan = "pro"
	sub.Status = "active"
	sub.ExpiresAt = &expiresAt
	return s.repo.UpdateSubscription(ctx, sub)
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
	return []map[string]any{
		{"name": "Starter", "price": 149000, "period": "monthly", "features": []string{"Up to 50 jamaah", "Basic CRM", "AI Scanner"}},
		{"name": "Pro", "price": 299000, "period": "monthly", "features": []string{"Unlimited jamaah", "Full CRM", "AI Scanner", "E-Kontrak", "Laporan Keuangan"}},
		{"name": "Business", "price": 599000, "period": "monthly", "features": []string{"Everything in Pro", "Multi-branch", "Priority support", "Custom integrations"}},
	}, nil
}
