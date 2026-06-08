package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/agent/model"
	"github.com/jamaah-in/v2/internal/agent/service"
)

type AgentHandler struct {
	svc *service.AgentService
}

func NewAgentHandler(svc *service.AgentService) *AgentHandler { return &AgentHandler{svc: svc} }

func (h *AgentHandler) CreateAgent(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.CreateAgentRequest
	if err := c.BodyParser(&req); err != nil { return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"}) }
	if req.Name == "" { return c.Status(400).JSON(fiber.Map{"success": false, "error": "name is required"}) }
	a, err := h.svc.CreateAgent(c.Context(), claims.OrgID.String(), req)
	if err != nil { return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()}) }
	return c.Status(201).JSON(fiber.Map{"success": true, "data": a})
}

func (h *AgentHandler) ListAgents(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	search := c.Query("search", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	resp, err := h.svc.ListAgents(c.Context(), claims.OrgID.String(), search, page, limit)
	if err != nil { return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()}) }
	return c.JSON(fiber.Map{"success": true, "data": resp})
}

func (h *AgentHandler) GetAgent(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	a, err := h.svc.GetAgent(c.Context(), c.Params("id"), claims.OrgID.String())
	if err != nil { return c.Status(404).JSON(fiber.Map{"success": false, "error": err.Error()}) }
	return c.JSON(fiber.Map{"success": true, "data": a})
}

func (h *AgentHandler) UpdateAgent(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.UpdateAgentRequest
	if err := c.BodyParser(&req); err != nil { return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"}) }
	a, err := h.svc.UpdateAgent(c.Context(), c.Params("id"), claims.OrgID.String(), req)
	if err != nil { return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()}) }
	return c.JSON(fiber.Map{"success": true, "data": a})
}

func (h *AgentHandler) CreateCommission(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.CreateCommissionRequest
	if err := c.BodyParser(&req); err != nil { return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"}) }
	if req.AgentID == "" { return c.Status(400).JSON(fiber.Map{"success": false, "error": "agent_id is required"}) }
	comm, err := h.svc.CreateCommission(c.Context(), claims.OrgID.String(), req)
	if err != nil { return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()}) }
	return c.Status(201).JSON(fiber.Map{"success": true, "data": comm})
}

func (h *AgentHandler) ListCommissions(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	agentID := c.Query("agent_id", "")
	status := c.Query("status", "all")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	resp, err := h.svc.ListCommissions(c.Context(), claims.OrgID.String(), agentID, status, page, limit)
	if err != nil { return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()}) }
	return c.JSON(fiber.Map{"success": true, "data": resp})
}

func (h *AgentHandler) PayCommission(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	if err := h.svc.PayCommission(c.Context(), c.Params("id"), claims.OrgID.String()); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "commission paid"}})
}

func (h *AgentHandler) GetAgentCommissions(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	comms, err := h.svc.GetAgentCommissions(c.Context(), c.Params("id"), claims.OrgID.String())
	if err != nil { return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()}) }
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"commissions": comms}})
}
