package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/response"
)

// The B2B handlers back the external agent portal. RequireAgentScope guarantees
// claims.AgentID is non-nil, so each reads its own agent id from the token — an
// agent can only ever see their own record, downline, and commissions.

func (h *AgentHandler) b2bAgent(c *fiber.Ctx) (orgID, agentID string, ok bool) {
	claims, found := middleware.GetClaims(c)
	if !found || claims.AgentID == nil {
		return "", "", false
	}
	return claims.OrgID.String(), claims.AgentID.String(), true
}

func (h *AgentHandler) B2BMe(c *fiber.Ctx) error {
	orgID, agentID, ok := h.b2bAgent(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	a, err := h.svc.GetAgent(c.Context(), agentID, orgID)
	if err != nil {
		return response.NotFound(c, "agent not found")
	}
	return response.OK(c, a)
}

func (h *AgentHandler) B2BDashboard(c *fiber.Ctx) error {
	orgID, agentID, ok := h.b2bAgent(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	dash, err := h.svc.GetAgentDashboard(c.Context(), agentID, orgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, dash)
}

func (h *AgentHandler) B2BDownline(c *fiber.Ctx) error {
	orgID, agentID, ok := h.b2bAgent(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	nodes, err := h.svc.GetDownline(c.Context(), agentID, orgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"downline": nodes})
}

func (h *AgentHandler) B2BCommissions(c *fiber.Ctx) error {
	orgID, agentID, ok := h.b2bAgent(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	comms, err := h.svc.GetAgentCommissions(c.Context(), agentID, orgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"commissions": comms})
}
