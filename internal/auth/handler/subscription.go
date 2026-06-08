package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/jamaah-in/v2/internal/auth/model"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

func (h *AuthHandler) GetSubscriptionStatus(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	result, err := h.svc.GetSubscriptionStatus(c.Context(), claims.OrgID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.OK(c, result)
}

func (h *AuthHandler) UpgradeToPro(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.UpgradeRequest
	_ = c.BodyParser(&req)
	if err := h.svc.UpgradeToPro(c.Context(), claims.OrgID, req); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "upgrade successful"})
}

func (h *AuthHandler) GetTrialStatus(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	result, err := h.svc.GetTrialStatus(c.Context(), claims.OrgID)
	if err != nil {
		return response.InternalError(c, err.Error())
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
		return response.InternalError(c, err.Error())
	}
	return response.OK(c, fiber.Map{"plans": plans})
}
