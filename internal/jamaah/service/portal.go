package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
	"github.com/jamaah-in/v2/internal/jamaah/repository"
)

// Portal* methods back the jemaah self-service portal. Every call is scoped to
// the signed-in jamaah's own id (taken from the JWT, never the client).

func (s *JamaahService) PortalProfile(ctx context.Context, orgID, jamaahID uuid.UUID) (*model.JamaahProfile, error) {
	return s.repo.GetProfileByID(ctx, jamaahID, orgID)
}

func (s *JamaahService) PortalRegistrations(ctx context.Context, orgID, jamaahID uuid.UUID) ([]model.JamaahPackageRegistration, error) {
	return s.repo.ListRegistrationsByJamaah(ctx, orgID, jamaahID)
}

func (s *JamaahService) PortalDocuments(ctx context.Context, orgID, jamaahID uuid.UUID) ([]model.JamaahDocument, error) {
	return s.repo.ListDocuments(ctx, orgID, jamaahID)
}

func (s *JamaahService) PortalVisa(ctx context.Context, orgID, jamaahID uuid.UUID) (*model.VisaApplication, error) {
	v, err := s.repo.GetVisaByJamaah(ctx, orgID, jamaahID)
	if errors.Is(err, repository.ErrVisaNotFound) {
		return nil, nil
	}
	return v, err
}

// PortalPayments returns the jamaah's outstanding invoice balance (best-effort
// from invoice-service).
func (s *JamaahService) PortalPayments(ctx context.Context, authToken string, jamaahID uuid.UUID) *invoiceBalance {
	balances := s.fetchBalances(ctx, authToken)
	if balances == nil {
		return nil
	}
	if b, ok := balances[jamaahID]; ok {
		return &b
	}
	return nil
}
