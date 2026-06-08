package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/jamaah-in/v2/internal/invoice/model"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

func (h *InvoiceHandler) CreatePaymentOrder(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)

	var req model.CreatePaymentOrderRequest
	if err := c.BodyParser(&req); err != nil {
		req.PlanType = c.Query("plan_type", "monthly")
	}

	result, err := h.svc.CreatePaymentOrder(c.Context(), claims.OrgID, claims.UserID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, result)
}

func (h *InvoiceHandler) CheckPaymentStatus(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	orderID := c.Params("id")

	result, err := h.svc.CheckPaymentStatus(c.Context(), orderID, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "payment order not found")
	}
	return response.OK(c, result)
}
