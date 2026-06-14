package handler

import (
	"crypto/subtle"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/shared/response"
)

type settleRequest struct {
	InvoiceID string `json:"invoice_id"`
	OrgID     string `json:"org_id"`
	Amount    int64  `json:"amount"`
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
	applied, err := h.svc.SettleFromCredit(c.Context(), invID, orgID, req.Amount)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"applied": applied})
}
