package handler

import (
	"crypto/subtle"
	"errors"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/invoice/model"
	"github.com/jamaah-in/v2/internal/invoice/repository"
	"github.com/jamaah-in/v2/internal/invoice/service"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

type RefundHandler struct {
	svc *service.RefundService
}

func NewRefundHandler(svc *service.RefundService) *RefundHandler {
	return &RefundHandler{svc: svc}
}

func (h *RefundHandler) ListRefunds(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	status := c.Query("status", "all")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	resp, err := h.svc.ListRefunds(c.Context(), claims.OrgID, status, page, limit)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, resp)
}

func (h *RefundHandler) GetRefund(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid refund id")
	}

	ref, err := h.svc.GetRefund(c.Context(), id, claims.OrgID)
	if err != nil {
		if errors.Is(err, repository.ErrRefundNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.Internal(c, err)
	}
	return response.OK(c, ref)
}

func (h *RefundHandler) InitiateRefund(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid invoice id")
	}

	var req model.InitiateRefundRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Amount <= 0 {
		return response.BadRequest(c, "amount must be greater than 0")
	}

	ref, err := h.svc.InitiateRefund(c.Context(), claims.OrgID, invoiceID, req, c.Get("Authorization"))
	if err != nil {
		if errors.Is(err, repository.ErrInvoiceNotFound) {
			return response.NotFound(c, "invoice not found")
		}
		if errors.Is(err, repository.ErrRefundExceedsPaid) || errors.Is(err, repository.ErrRefundAlreadyOpen) || errors.Is(err, repository.ErrRefundExceedsPolicy) {
			return response.BadRequest(c, err.Error())
		}
		return response.Internal(c, err)
	}
	return response.Created(c, ref)
}

type internalRefundRequest struct {
	InvoiceID string  `json:"invoice_id"`
	OrgID     string  `json:"org_id"`
	Amount    int64   `json:"amount"`
	RefundPct float64 `json:"refund_pct"`
	Reason    string  `json:"reason"`
	Notes     string  `json:"notes"`
}

// InitiateRefundInternal is for service-to-service cascades (jamaah gagal
// berangkat, kloter cancel, removed from package) that must return exactly
// what the customer paid — they aren't at fault, so the early-cancellation
// policy cap (which exists to stop a manual refund request from overriding
// the tiered penalty schedule) does not apply here. Service-to-service only:
// guarded by the shared INTERNAL_API_KEY in the X-Internal-Key header
// (constant-time compared), same pattern as SettleInternal. No JWT.
func (h *RefundHandler) InitiateRefundInternal(c *fiber.Ctx) error {
	want := os.Getenv("INTERNAL_API_KEY")
	got := c.Get("X-Internal-Key")
	if want == "" || subtle.ConstantTimeCompare([]byte(want), []byte(got)) != 1 {
		return response.Unauthorized(c, "invalid internal key")
	}
	var req internalRefundRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	invoiceID, err := uuid.Parse(req.InvoiceID)
	if err != nil {
		return response.BadRequest(c, "invalid invoice_id")
	}
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return response.BadRequest(c, "invalid org_id")
	}
	if req.Amount <= 0 {
		return response.BadRequest(c, "amount must be greater than 0")
	}

	ref, err := h.svc.InitiateRefundInternal(c.Context(), orgID, invoiceID, model.InitiateRefundRequest{
		Amount:    req.Amount,
		RefundPct: req.RefundPct,
		Reason:    req.Reason,
		Notes:     req.Notes,
	})
	if err != nil {
		if errors.Is(err, repository.ErrInvoiceNotFound) {
			return response.NotFound(c, "invoice not found")
		}
		if errors.Is(err, repository.ErrRefundExceedsPaid) || errors.Is(err, repository.ErrRefundAlreadyOpen) {
			return response.BadRequest(c, err.Error())
		}
		return response.Internal(c, err)
	}
	return response.Created(c, ref)
}

func (h *RefundHandler) GetApplicablePolicy(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid invoice id")
	}

	policy, err := h.svc.GetApplicablePolicyForInvoice(c.Context(), claims.OrgID, invoiceID, c.Get("Authorization"))
	if err != nil {
		if errors.Is(err, repository.ErrInvoiceNotFound) || errors.Is(err, repository.ErrPolicyNotFound) {
			return response.NotFound(c, "no applicable refund policy")
		}
		return response.Internal(c, err)
	}
	return response.OK(c, policy)
}

func (h *RefundHandler) ApproveRefund(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid refund id")
	}

	if err := h.svc.ApproveRefund(c.Context(), id, claims.OrgID, claims.UserID); err != nil {
		if errors.Is(err, repository.ErrRefundNotPending) {
			return response.BadRequest(c, err.Error())
		}
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "refund approved"})
}

func (h *RefundHandler) ProcessRefund(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid refund id")
	}

	if err := h.svc.ProcessRefund(c.Context(), id, claims.OrgID); err != nil {
		if errors.Is(err, repository.ErrRefundNotApproved) {
			return response.BadRequest(c, err.Error())
		}
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "refund processed"})
}

func (h *RefundHandler) CompleteRefund(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid refund id")
	}

	if err := h.svc.CompleteRefund(c.Context(), id, claims.OrgID); err != nil {
		if errors.Is(err, repository.ErrRefundNotProcessed) || errors.Is(err, repository.ErrInvoiceNotFound) {
			return response.BadRequest(c, err.Error())
		}
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "refund completed"})
}

func (h *RefundHandler) RejectRefund(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid refund id")
	}

	if err := h.svc.RejectRefund(c.Context(), id, claims.OrgID); err != nil {
		if errors.Is(err, repository.ErrRefundNotPending) {
			return response.BadRequest(c, err.Error())
		}
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "refund rejected"})
}

func (h *RefundHandler) GetRefundsByInvoice(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid invoice id")
	}

	refunds, err := h.svc.GetRefundsByInvoice(c.Context(), invoiceID, claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"refunds": refunds})
}

func (h *RefundHandler) ListPolicies(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	policies, err := h.svc.ListPolicies(c.Context(), claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"policies": policies})
}

func (h *RefundHandler) CreatePolicy(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.CreateRefundPolicyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	p, err := h.svc.CreatePolicy(c.Context(), claims.OrgID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Created(c, p)
}

func (h *RefundHandler) UpdatePolicy(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid policy id")
	}

	var req model.UpdateRefundPolicyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	p, err := h.svc.UpdatePolicy(c.Context(), id, claims.OrgID, req)
	if err != nil {
		if errors.Is(err, repository.ErrPolicyNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.Internal(c, err)
	}
	return response.OK(c, p)
}

func (h *RefundHandler) DeletePolicy(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid policy id")
	}

	if err := h.svc.DeletePolicy(c.Context(), id, claims.OrgID); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "policy deleted"})
}
