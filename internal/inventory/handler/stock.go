package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/inventory/model"
	"github.com/jamaah-in/v2/internal/inventory/repository"
	"github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/response"
)

func (h *InventoryHandler) ListItems(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	items, err := h.svc.ListStockItems(c.Context(), claims.OrgID.String())
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"items": items})
}

func (h *InventoryHandler) CreateItem(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	var req model.CreateItemRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	item, err := h.svc.CreateStockItem(c.Context(), claims.OrgID.String(), req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, item)
}

func (h *InventoryHandler) UpdateItem(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return response.BadRequest(c, "invalid id")
	}
	var req model.UpdateItemRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.UpdateStockItem(c.Context(), claims.OrgID.String(), id, req); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"updated": true})
}

func (h *InventoryHandler) RestockItem(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id := c.Params("id")
	var req model.RestockRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Qty <= 0 {
		return response.BadRequest(c, "qty must be positive")
	}
	if err := h.svc.RestockItem(c.Context(), claims.OrgID.String(), id, req.Qty, req.Note, claims.UserID.String()); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"ok": true})
}

func (h *InventoryHandler) AdjustItem(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id := c.Params("id")
	var req model.AdjustRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Delta == 0 {
		return response.BadRequest(c, "delta must be non-zero")
	}
	if err := h.svc.AdjustItem(c.Context(), claims.OrgID.String(), id, req.Delta, req.Note, claims.UserID.String()); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"ok": true})
}

func (h *InventoryHandler) ListItemMovements(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id := c.Params("id")
	moves, err := h.svc.ListMovements(c.Context(), claims.OrgID.String(), id, c.QueryInt("limit", 50))
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"movements": moves})
}

func (h *InventoryHandler) DeleteItem(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id := c.Params("id")
	if err := h.svc.DeleteStockItem(c.Context(), claims.OrgID.String(), id); err != nil {
		if errors.Is(err, repository.ErrItemInKit) {
			return response.Conflict(c, "item dipakai di kit paket; hapus dari kit dulu")
		}
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"deleted": true})
}

func (h *InventoryHandler) GetKit(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	pkg := c.Params("packageId")
	if _, err := uuid.Parse(pkg); err != nil {
		return response.BadRequest(c, "invalid package_id")
	}
	kit, err := h.svc.GetPackageKit(c.Context(), claims.OrgID.String(), pkg)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"items": kit})
}

func (h *InventoryHandler) SetKit(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	pkg := c.Params("packageId")
	if _, err := uuid.Parse(pkg); err != nil {
		return response.BadRequest(c, "invalid package_id")
	}
	var req model.SetKitRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.SetPackageKit(c.Context(), claims.OrgID.String(), pkg, req.Items); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"saved": len(req.Items)})
}
