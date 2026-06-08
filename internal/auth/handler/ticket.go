package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/auth/model"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

func (h *AuthHandler) CreateTicket(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)

	var req model.CreateTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	ticket, err := h.svc.CreateTicket(c.Context(), claims.OrgID, claims.UserID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, ticket)
}

func (h *AuthHandler) ListTickets(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	status := c.Query("status")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	tickets, total, err := h.svc.ListTickets(c.Context(), claims.OrgID, status, page, pageSize)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{
		"tickets": tickets,
		"total":   total,
		"page":    page,
	})
}

func (h *AuthHandler) GetTicketMessages(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	ticketID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid ticket id")
	}

	result, err := h.svc.GetTicketWithMessages(c.Context(), ticketID, claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	if result == nil {
		return response.NotFound(c, "ticket not found")
	}
	return response.OK(c, result)
}

func (h *AuthHandler) AddTicketMessage(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	ticketID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid ticket id")
	}

	var req model.AddTicketMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	msg, err := h.svc.AddTicketMessage(c.Context(), ticketID, claims.OrgID, claims.UserID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, msg)
}
