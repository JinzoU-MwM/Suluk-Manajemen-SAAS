package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/jamaah-in/v2/internal/agent/model"
	"github.com/jamaah-in/v2/internal/agent/service"
	"github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/response"
)

type AgentHandler struct {
	svc *service.AgentService
}

func NewAgentHandler(svc *service.AgentService) *AgentHandler { return &AgentHandler{svc: svc} }

func (h *AgentHandler) CreateAgent(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	var req model.CreateAgentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	a, err := h.svc.CreateAgent(c.Context(), claims.OrgID.String(), req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Created(c, a)
}

func (h *AgentHandler) ListAgents(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	search := c.Query("search", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	resp, err := h.svc.ListAgents(c.Context(), claims.OrgID.String(), search, page, limit)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, resp)
}

func (h *AgentHandler) GetAgent(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	a, err := h.svc.GetAgent(c.Context(), c.Params("id"), claims.OrgID.String())
	if err != nil {
		return response.NotFound(c, "agent not found")
	}
	return response.OK(c, a)
}

func (h *AgentHandler) UpdateAgent(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	var req model.UpdateAgentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	a, err := h.svc.UpdateAgent(c.Context(), c.Params("id"), claims.OrgID.String(), req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, a)
}

func (h *AgentHandler) CreateCommission(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	var req model.CreateCommissionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.AgentID == "" {
		return response.BadRequest(c, "agent_id is required")
	}
	comm, err := h.svc.CreateCommission(c.Context(), claims.OrgID.String(), req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Created(c, comm)
}

func (h *AgentHandler) ListCommissions(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	agentID := c.Query("agent_id", "")
	status := c.Query("status", "all")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	resp, err := h.svc.ListCommissions(c.Context(), claims.OrgID.String(), agentID, status, page, limit)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, resp)
}

func (h *AgentHandler) PayCommission(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if err := h.svc.PayCommission(c.Context(), c.Params("id"), claims.OrgID.String()); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "commission paid"})
}

func (h *AgentHandler) GetAgentCommissions(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	comms, err := h.svc.GetAgentCommissions(c.Context(), c.Params("id"), claims.OrgID.String())
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"commissions": comms})
}
