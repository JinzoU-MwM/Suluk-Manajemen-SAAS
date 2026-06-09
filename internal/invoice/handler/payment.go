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
	_ = c.BodyParser(&req)
	// Allow query-param fallbacks (and back-compat with the old period-only call).
	if req.PlanType == "" {
		req.PlanType = c.Query("plan_type", "monthly")
	}
	if req.Plan == "" {
		req.Plan = c.Query("plan", "pro")
	}

	result, err := h.svc.CreatePaymentOrder(c.Context(), claims.OrgID, claims.UserID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, result)
}

// PakasirWebhook is the public callback Pakasir hits when a payment completes.
// It is NOT behind AuthMiddleware (called server-to-server by Pakasir); the
// service layer independently verifies the transaction before activating.
func (h *InvoiceHandler) PakasirWebhook(c *fiber.Ctx) error {
	var payload model.PakasirWebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.BadRequest(c, "invalid webhook body")
	}
	if err := h.svc.HandlePakasirWebhook(c.Context(), payload); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"received": true})
}

// SubscriptionInvoicePDF streams the signed subscription-invoice PDF. Public
// (linked from the confirmation email) — protected by the HMAC sig query param,
// not a JWT.
func (h *InvoiceHandler) SubscriptionInvoicePDF(c *fiber.Ctx) error {
	pdf, filename, err := h.svc.SubscriptionInvoicePDF(c.Context(), c.Params("orderID"), c.Query("sig"))
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	return c.Send(pdf)
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
