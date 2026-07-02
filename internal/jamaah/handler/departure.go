package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
	"github.com/jamaah-in/v2/internal/jamaah/service"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

// SetDeparture links a group to its package + departure date.
func (h *JamaahHandler) SetDeparture(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	var req model.SetDepartureRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	g, err := h.svc.SetDeparture(c.Context(), groupID, claims.OrgID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, g)
}

// TransitionDeparture advances the kloter status.
func (h *JamaahHandler) TransitionDeparture(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	var req model.DepartureTransitionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	g, cascade, err := h.svc.TransitionDeparture(c.Context(), groupID, claims.OrgID, req.Status, c.Get("Authorization"))
	if errors.Is(err, service.ErrDepartureGate) {
		return response.BadRequest(c, err.Error())
	}
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"group": g, "cascade": cascade})
}

// GetDepartureManifest returns the kloter + members for boarding.
func (h *JamaahHandler) GetDepartureManifest(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	m, err := h.svc.GetDepartureManifest(c.Context(), groupID, claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, m)
}
