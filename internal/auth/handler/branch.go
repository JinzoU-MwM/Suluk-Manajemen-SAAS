package handler

import (
	"github.com/gofiber/fiber/v2"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

func (h *AuthHandler) ListBranches(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	branches, err := h.svc.ListBranches(c.Context(), claims.OrgID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.OK(c, fiber.Map{"branches": branches})
}

func (h *AuthHandler) CreateBranch(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return response.BadRequest(c, "name is required")
	}

	branch, err := h.svc.CreateBranch(c.Context(), claims.OrgID.String(), req.Name)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Created(c, branch)
}

func (h *AuthHandler) GetConsolidatedDashboard(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	stats, err := h.svc.GetConsolidatedStats(c.Context(), claims.OrgID.String())
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.OK(c, stats)
}
