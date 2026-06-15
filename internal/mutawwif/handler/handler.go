package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/mutawwif/model"
	"github.com/jamaah-in/v2/internal/mutawwif/repository"
	"github.com/jamaah-in/v2/internal/mutawwif/service"
	"github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/response"
)

type MutawwifHandler struct {
	svc *service.MutawwifService
}

func NewMutawwifHandler(svc *service.MutawwifService) *MutawwifHandler { return &MutawwifHandler{svc: svc} }

func (h *MutawwifHandler) CreateGuide(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	var req model.CreateGuideRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	g, err := h.svc.CreateGuide(c.Context(), claims.OrgID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, g)
}

func (h *MutawwifHandler) ListGuides(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	resp, err := h.svc.ListGuides(c.Context(), claims.OrgID, c.Query("search"))
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, resp)
}

func (h *MutawwifHandler) GetGuide(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid guide id")
	}
	g, err := h.svc.GetGuide(c.Context(), id, claims.OrgID)
	if errors.Is(err, repository.ErrGuideNotFound) {
		return response.NotFound(c, "guide not found")
	}
	if err != nil {
		return response.Internal(c, err)
	}
	asgs, _ := h.svc.ListByGuide(c.Context(), claims.OrgID, id)
	return response.OK(c, fiber.Map{"guide": g, "assignments": asgs})
}

func (h *MutawwifHandler) UpdateGuide(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid guide id")
	}
	var req model.UpdateGuideRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	g, err := h.svc.UpdateGuide(c.Context(), id, claims.OrgID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, g)
}

func (h *MutawwifHandler) DeleteGuide(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid guide id")
	}
	if err := h.svc.DeleteGuide(c.Context(), id, claims.OrgID); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "guide deleted"})
}

func (h *MutawwifHandler) Assign(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	var req model.AssignGuideRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	a, err := h.svc.Assign(c.Context(), claims.OrgID, req)
	if errors.Is(err, repository.ErrGuideNotFound) {
		return response.NotFound(c, "guide not found")
	}
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, a)
}

func (h *MutawwifHandler) Unassign(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	guideID, err := uuid.Parse(c.Params("guideId"))
	if err != nil {
		return response.BadRequest(c, "invalid guide id")
	}
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	if err := h.svc.Unassign(c.Context(), claims.OrgID, guideID, groupID); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "unassigned"})
}

func (h *MutawwifHandler) ListByGroup(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	asgs, err := h.svc.ListByGroup(c.Context(), claims.OrgID, groupID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"assignments": asgs})
}
