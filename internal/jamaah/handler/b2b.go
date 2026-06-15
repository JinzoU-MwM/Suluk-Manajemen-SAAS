package handler

import (
	"github.com/gofiber/fiber/v2"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

// B2BMyLeads returns the jamaah referred by the signed-in agent and their
// downline. RequireAgentScope guarantees claims.AgentID is set.
func (h *JamaahHandler) B2BMyLeads(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	if claims.AgentID == nil {
		return response.Forbidden(c, "akun bukan agen")
	}
	leads, err := h.svc.GetMyLeads(c.Context(), claims.OrgID, c.Get("Authorization"), *claims.AgentID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"leads": leads})
}
