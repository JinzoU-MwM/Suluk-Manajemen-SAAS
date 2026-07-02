package service

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/invoice/model"
	"github.com/jamaah-in/v2/internal/invoice/repository"
	"github.com/jamaah-in/v2/internal/shared/httpclient"
)

func NewRefundService(repo *repository.InvoiceRepo) *RefundService {
	return &RefundService{repo: repo, httpc: httpclient.New()}
}

// WithPackageAddr enables best-effort refund-policy enforcement by letting
// InitiateRefund ask package-service for a package's departure_date. Without
// it (empty string, e.g. in unit tests), policy enforcement is silently
// skipped — same "best-effort cross-service call" convention as
// InvoiceService.WithPayments and jamaah-service's releaseSeat.
func (s *RefundService) WithPackageAddr(addr string) *RefundService {
	s.packageAddr = addr
	return s
}

type RefundService struct {
	repo        *repository.InvoiceRepo
	packageAddr string
	httpc       *httpclient.Client
}

// packageDeparture is the subset of package-service's package fields this
// lookup needs.
type packageDeparture struct {
	DepartureDate *time.Time `json:"departure_date"`
}

// daysUntilDeparture best-effort resolves how many days remain until a
// package's departure. ok=false means "can't determine" (no package address
// configured, no auth token, the call failed, or the package has no
// departure_date) — callers must treat that as "skip policy enforcement,"
// not as an error, since not every org uses refund policies.
func (s *RefundService) daysUntilDeparture(ctx context.Context, packageID uuid.UUID, authToken string) (int, bool) {
	if s.packageAddr == "" || authToken == "" {
		return 0, false
	}
	var pkg packageDeparture
	if err := s.httpc.GetJSON(ctx, s.packageAddr, "/api/v1/packages/"+packageID.String(), authToken, &pkg); err != nil || pkg.DepartureDate == nil {
		return 0, false
	}
	days := int(time.Until(*pkg.DepartureDate).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return days, true
}

// applicablePolicy returns the refund policy that applies to a package right
// now, or (nil, nil) if none does (departure date unknown, or no policy
// configured for that many days out) — distinct from a real lookup error.
func (s *RefundService) applicablePolicy(ctx context.Context, orgID, packageID uuid.UUID, authToken string) (*model.RefundPolicy, error) {
	days, ok := s.daysUntilDeparture(ctx, packageID, authToken)
	if !ok {
		return nil, nil
	}
	policy, err := s.repo.GetApplicablePolicy(ctx, orgID, days)
	if err != nil {
		if err == repository.ErrPolicyNotFound {
			return nil, nil
		}
		return nil, err
	}
	return policy, nil
}

// GetApplicablePolicyForInvoice is the read-only preview the web "Ajukan
// Refund" form uses to auto-fill refund_pct before the staff submits —
// InitiateRefund independently re-derives and enforces this at submit time,
// so this endpoint is advisory only and never itself creates or blocks a refund.
func (s *RefundService) GetApplicablePolicyForInvoice(ctx context.Context, orgID, invoiceID uuid.UUID, authToken string) (*model.RefundPolicy, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, invoiceID, orgID)
	if err != nil {
		return nil, err
	}
	policy, err := s.applicablePolicy(ctx, orgID, inv.PackageID, authToken)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, repository.ErrPolicyNotFound
	}
	return policy, nil
}

func (s *RefundService) InitiateRefund(ctx context.Context, orgID uuid.UUID, invoiceID uuid.UUID, req model.InitiateRefundRequest, authToken string) (*model.Refund, error) {
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

	if policy, err := s.applicablePolicy(ctx, orgID, inv.PackageID, authToken); err != nil {
		return nil, err
	} else if policy != nil {
		maxAllowed := int64(math.Round(float64(inv.AmountPaid) * policy.RefundPct / 100))
		if req.Amount > maxAllowed {
			return nil, repository.ErrRefundExceedsPolicy
		}
		ref.PolicyID = &policy.ID
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
