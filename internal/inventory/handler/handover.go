package handler

import (
	"github.com/gofiber/fiber/v2"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/jamaah-in/v2/internal/inventory/model"
	"github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/response"
)

// GetMemberQR renders the member's handover token as a PNG QR code.
func (h *InventoryHandler) GetMemberQR(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	m, err := h.svc.GetMember(c.Context(), claims.OrgID.String(), c.Params("memberId"))
	if err != nil || m == nil || m.HandoverToken == "" {
		return response.NotFound(c, "anggota tidak ditemukan")
	}
	png, err := qrcode.Encode(m.HandoverToken, qrcode.Medium, 256)
	if err != nil {
		return response.Internal(c, err)
	}
	c.Set("Content-Type", "image/png")
	c.Set("Cache-Control", "private, max-age=3600")
	return c.Send(png)
}

// Scan records a QR handover scan (equipment or luggage) by member token.
func (h *InventoryHandler) Scan(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	var req model.ScanRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	userID := claims.UserID
	m, err := h.svc.Scan(c.Context(), claims.OrgID.String(), &userID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, m)
}

// GetCheckpoints returns per-member handover progress for a package.
func (h *InventoryHandler) GetCheckpoints(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	members, err := h.svc.ListCheckpoints(c.Context(), claims.OrgID.String(), c.Params("packageId"))
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"members": members})
}
