package handler

import (
	"context"
	"crypto/subtle"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/auth/model"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

// validInternalKey authenticates a service-to-service call against the shared
// INTERNAL_API_KEY using a constant-time comparison (avoids the timing
// side-channel of a plain string ==). An unset key fails closed.
func validInternalKey(c *fiber.Ctx) bool {
	want := os.Getenv("INTERNAL_API_KEY")
	if want == "" {
		return false
	}
	got := c.Get("X-Internal-Key")
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

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
	if !validInternalKey(c) {
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
	// Best-effort: email the buyer the confirmation + invoice OFF the hot path.
	// A slow mail provider must not delay the activation response (which would
	// hold the cross-service call and risk a Pakasir webhook retry). Detached
	// goroutine with its own background context (c.Context() is reused after the
	// request returns, so it cannot be used here).
	go func(req model.ActivatePlanRequest, expiresAt time.Time) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := h.svc.SendSubscriptionInvoice(ctx, req, expiresAt); err != nil {
			log.Printf("subscription invoice email failed (order %s): %v", req.OrderID, err)
		}
	}(req, expiresAt)
	return response.OK(c, fiber.Map{"activated": true, "plan": req.Plan})
}

// CreateNotificationInternal is a service-to-service endpoint (NOT behind
// AuthMiddleware) used by other services to push in-app notifications on key
// events. Guarded by the shared INTERNAL_API_KEY in the X-Internal-Key header.
func (h *AuthHandler) CreateNotificationInternal(c *fiber.Ctx) error {
	if !validInternalKey(c) {
		return response.Unauthorized(c, "invalid internal key")
	}
	var req struct {
		OrgID    string `json:"org_id"`
		UserID   string `json:"user_id"`
		Severity string `json:"severity"`
		Title    string `json:"title"`
		Message  string `json:"message"`
		GroupID  string `json:"group_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return response.BadRequest(c, "invalid org_id")
	}
	n := &model.Notification{
		OrgID:    orgID,
		Severity: req.Severity,
		Title:    req.Title,
		Message:  req.Message,
	}
	if req.UserID != "" {
		if uid, err := uuid.Parse(req.UserID); err == nil {
			n.UserID = &uid
		}
	}
	if req.GroupID != "" {
		n.GroupID = &req.GroupID
	}
	if err := h.svc.CreateNotification(c.Context(), n); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"created": true})
}

// BillingInfoInternal returns the org + buyer display fields for an invoice,
// called by the invoice-service when rendering the subscription-invoice PDF.
// Guarded by the shared INTERNAL_API_KEY in the X-Internal-Key header.
func (h *AuthHandler) BillingInfoInternal(c *fiber.Ctx) error {
	if !validInternalKey(c) {
		return response.Unauthorized(c, "invalid internal key")
	}
	var req struct {
		OrgID  string `json:"org_id"`
		UserID string `json:"user_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return response.BadRequest(c, "invalid org_id")
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return response.BadRequest(c, "invalid user_id")
	}
	orgName, userName, userEmail, err := h.svc.GetBillingInfo(c.Context(), orgID, userID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{
		"org_name": orgName, "user_name": userName, "user_email": userEmail,
	})
}
