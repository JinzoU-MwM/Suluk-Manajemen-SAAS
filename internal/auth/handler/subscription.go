package handler

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/auth/model"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

func (h *AuthHandler) GetSubscriptionStatus(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	result, err := h.svc.GetSubscriptionStatus(c.Context(), claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, result)
}

func (h *AuthHandler) UpgradeToPro(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.UpgradeRequest
	_ = c.BodyParser(&req)
	if err := h.svc.UpgradeToPro(c.Context(), claims.OrgID, req); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "upgrade successful"})
}

func (h *AuthHandler) GetTrialStatus(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	result, err := h.svc.GetTrialStatus(c.Context(), claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, result)
}

func (h *AuthHandler) ActivateTrial(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	if err := h.svc.ActivateTrial(c.Context(), claims.OrgID); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"trial_activated": true})
}

func (h *AuthHandler) GetPricing(c *fiber.Ctx) error {
	plans, err := h.svc.GetPricing(c.Context())
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"plans": plans})
}

// ActivatePlanInternal is a service-to-service endpoint (NOT behind AuthMiddleware)
// called by the invoice-service payment webhook after a verified, paid order.
// It is guarded by a shared INTERNAL_API_KEY in the X-Internal-Key header.
func (h *AuthHandler) ActivatePlanInternal(c *fiber.Ctx) error {
	want := os.Getenv("INTERNAL_API_KEY")
	if want == "" || c.Get("X-Internal-Key") != want {
		return response.Unauthorized(c, "invalid internal key")
	}
	var req model.ActivatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return response.BadRequest(c, "invalid org_id")
	}
	expiresAt, err := h.svc.ActivatePlan(c.Context(), orgID, req.Plan, req.Period)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	// Best-effort: email the buyer the confirmation + invoice. Never fail
	// activation if the email send errors.
	if err := h.svc.SendSubscriptionInvoice(c.Context(), req, expiresAt); err != nil {
		log.Printf("subscription invoice email failed (order %s): %v", req.OrderID, err)
	}
	return response.OK(c, fiber.Map{"activated": true, "plan": req.Plan})
}
