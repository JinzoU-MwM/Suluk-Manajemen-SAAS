package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/inventory/model"
	"github.com/jamaah-in/v2/internal/inventory/service"
)

type InventoryHandler struct {
	svc *service.InventoryService
}

func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

func (h *InventoryHandler) SyncMembers(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)

	var req model.SyncMembersRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	if req.PackageID == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "package_id is required"})
	}

	if err := h.svc.SyncMembers(c.Context(), claims.OrgID.String(), req); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"synced": len(req.Members)}})
}

func (h *InventoryHandler) GetForecast(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	packageID := c.Params("packageId")
	if _, err := uuid.Parse(packageID); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid package_id"})
	}

	resp, err := h.svc.GetForecast(c.Context(), claims.OrgID.String(), packageID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": resp})
}

func (h *InventoryHandler) GetFulfillment(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	packageID := c.Params("packageId")
	if _, err := uuid.Parse(packageID); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid package_id"})
	}

	resp, err := h.svc.GetFulfillment(c.Context(), claims.OrgID.String(), packageID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": resp})
}

func (h *InventoryHandler) MarkReceived(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	packageID := c.Params("packageId")
	var req model.MarkReceivedRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	if len(req.MemberIDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "member_ids is required"})
	}

	count, err := h.svc.MarkReceived(c.Context(), claims.OrgID.String(), packageID, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"marked": count}})
}

func (h *InventoryHandler) UpdateOperational(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	memberID := c.Params("memberId")
	var req model.UpdateOperationalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}

	if err := h.svc.UpdateOperational(c.Context(), claims.OrgID.String(), memberID, req); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"updated": true}})
}
