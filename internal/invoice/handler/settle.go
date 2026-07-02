package handler

import (
	"crypto/subtle"
	"errors"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/invoice/repository"
	"github.com/jamaah-in/v2/internal/shared/response"
)

type settleRequest struct {
	InvoiceID      string `json:"invoice_id"`
	OrgID          string `json:"org_id"`
	JamaahID       string `json:"jamaah_id"`
	Amount         int64  `json:"amount"`
	IdempotencyKey string `json:"idempotency_key"`
}

// SettleInternal applies a non-cash credit to an invoice (used by tabungan
// conversion). Service-to-service only: guarded by the shared INTERNAL_API_KEY
// in the X-Internal-Key header (constant-time compared). No JWT.
func (h *InvoiceHandler) SettleInternal(c *fiber.Ctx) error {
	want := os.Getenv("INTERNAL_API_KEY")
	got := c.Get("X-Internal-Key")
	if want == "" || subtle.ConstantTimeCompare([]byte(want), []byte(got)) != 1 {
		return response.Unauthorized(c, "invalid internal key")
	}
	var req settleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	invID, err := uuid.Parse(req.InvoiceID)
	if err != nil {
		return response.BadRequest(c, "invalid invoice_id")
	}
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return response.BadRequest(c, "invalid org_id")
	}
	jamaahID, err := uuid.Parse(req.JamaahID)
	if err != nil {
		return response.BadRequest(c, "invalid jamaah_id")
	}
	if req.IdempotencyKey == "" {
		return response.BadRequest(c, "idempotency_key is required")
	}
	applied, err := h.svc.SettleFromCredit(c.Context(), invID, orgID, jamaahID, req.Amount, req.IdempotencyKey)
	if err != nil {
		if errors.Is(err, repository.ErrJamaahMismatch) || errors.Is(err, repository.ErrAlreadyCancelled) {
			return response.BadRequest(c, err.Error())
		}
		if errors.Is(err, repository.ErrInvoiceNotFound) {
			return response.NotFound(c, "invoice not found")
		}
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"applied": applied})
}
