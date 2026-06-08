package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/inventory/model"
	"github.com/jamaah-in/v2/internal/inventory/service"
	"github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/response"
)

type InventoryHandler struct {
	svc *service.InventoryService
}

func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

func (h *InventoryHandler) SyncMembers(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	var req model.SyncMembersRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.PackageID == "" {
		return response.BadRequest(c, "package_id is required")
	}
	if err := h.svc.SyncMembers(c.Context(), claims.OrgID.String(), req); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"synced": len(req.Members)})
}

func (h *InventoryHandler) GetForecast(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	packageID := c.Params("packageId")
	if _, err := uuid.Parse(packageID); err != nil {
		return response.BadRequest(c, "invalid package_id")
	}
	resp, err := h.svc.GetForecast(c.Context(), claims.OrgID.String(), packageID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, resp)
}

func (h *InventoryHandler) GetFulfillment(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	packageID := c.Params("packageId")
	if _, err := uuid.Parse(packageID); err != nil {
		return response.BadRequest(c, "invalid package_id")
	}
	resp, err := h.svc.GetFulfillment(c.Context(), claims.OrgID.String(), packageID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, resp)
}

func (h *InventoryHandler) MarkReceived(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	packageID := c.Params("packageId")
	var req model.MarkReceivedRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if len(req.MemberIDs) == 0 {
		return response.BadRequest(c, "member_ids is required")
	}
	count, err := h.svc.MarkReceived(c.Context(), claims.OrgID.String(), packageID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"marked": count})
}

func (h *InventoryHandler) UpdateOperational(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	memberID := c.Params("memberId")
	var req model.UpdateOperationalRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.UpdateOperational(c.Context(), claims.OrgID.String(), memberID, req); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"updated": true})
}
