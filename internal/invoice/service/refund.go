package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/invoice/model"
	"github.com/jamaah-in/v2/internal/invoice/repository"
)

func NewRefundService(repo *repository.InvoiceRepo) *RefundService {
	return &RefundService{repo: repo}
}

type RefundService struct {
	repo *repository.InvoiceRepo
}

func (s *RefundService) InitiateRefund(ctx context.Context, orgID uuid.UUID, invoiceID uuid.UUID, req model.InitiateRefundRequest) (*model.Refund, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, invoiceID, orgID)
	if err != nil {
		return nil, err
	}
	if req.Amount > inv.AmountPaid {
		return nil, repository.ErrRefundExceedsPaid
	}
	// GetPayments orders by paid_at DESC, so [0] is the most recent payment —
	// that's the account the refund should come back out of.
	paymentMethod := "transfer_bank"
	if payments, err := s.repo.GetPayments(ctx, invoiceID, orgID); err == nil && len(payments) > 0 {
		paymentMethod = payments[0].PaymentMethod
	}
	ref := &model.Refund{
		OrgID:         orgID,
		InvoiceID:     invoiceID,
		Amount:        req.Amount,
		RefundPct:     req.RefundPct,
		Reason:        req.Reason,
		Notes:         req.Notes,
		PaymentMethod: paymentMethod,
		Status:        "pending",
	}
	if err := s.repo.CreateRefund(ctx, ref); err != nil {
		return nil, err
	}
	return ref, nil
}

func (s *RefundService) ListRefunds(ctx context.Context, orgID uuid.UUID, status string, page, limit int) (*model.RefundListResponse, error) {
	refunds, total, err := s.repo.ListRefunds(ctx, orgID, status, page, limit)
	if err != nil {
		return nil, err
	}
	return &model.RefundListResponse{Refunds: refunds, Total: total}, nil
}

func (s *RefundService) GetRefund(ctx context.Context, id, orgID uuid.UUID) (*model.Refund, error) {
	return s.repo.GetRefund(ctx, id, orgID)
}

func (s *RefundService) ApproveRefund(ctx context.Context, id, orgID, approverID uuid.UUID) error {
	return s.repo.ApproveRefund(ctx, id, orgID, approverID)
}

func (s *RefundService) ProcessRefund(ctx context.Context, id, orgID uuid.UUID) error {
	return s.repo.ProcessRefund(ctx, id, orgID)
}

func (s *RefundService) CompleteRefund(ctx context.Context, id, orgID uuid.UUID) error {
	return s.repo.CompleteRefund(ctx, id, orgID)
}

func (s *RefundService) RejectRefund(ctx context.Context, id, orgID uuid.UUID) error {
	return s.repo.RejectRefund(ctx, id, orgID)
}

func (s *RefundService) GetRefundsByInvoice(ctx context.Context, invoiceID, orgID uuid.UUID) ([]model.Refund, error) {
	return s.repo.GetRefundsByInvoice(ctx, invoiceID, orgID)
}

func (s *RefundService) CreatePolicy(ctx context.Context, orgID uuid.UUID, req model.CreateRefundPolicyRequest) (*model.RefundPolicy, error) {
	p := &model.RefundPolicy{
		OrgID:       orgID,
		Name:        req.Name,
		DaysBefore:  req.DaysBefore,
		RefundPct:   req.RefundPct,
		Description: req.Description,
		IsActive:    true,
	}
	if err := s.repo.CreateRefundPolicy(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *RefundService) ListPolicies(ctx context.Context, orgID uuid.UUID) ([]model.RefundPolicy, error) {
	return s.repo.ListRefundPolicies(ctx, orgID)
}

func (s *RefundService) UpdatePolicy(ctx context.Context, id, orgID uuid.UUID, req model.UpdateRefundPolicyRequest) (*model.RefundPolicy, error) {
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.DaysBefore != nil {
		updates["days_before"] = *req.DaysBefore
	}
	if req.RefundPct != nil {
		updates["refund_pct"] = *req.RefundPct
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if err := s.repo.UpdateRefundPolicy(ctx, id, orgID, updates); err != nil {
		return nil, err
	}
	return s.repo.GetRefundPolicy(ctx, id, orgID)
}

func (s *RefundService) DeletePolicy(ctx context.Context, id, orgID uuid.UUID) error {
	return s.repo.DeleteRefundPolicy(ctx, id, orgID)
}
