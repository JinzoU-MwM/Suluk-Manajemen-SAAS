package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

func (h *JamaahHandler) GetItinerary(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	items, err := h.svc.ListItineraries(c.Context(), claims.OrgID, groupID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"items": items})
}

func (h *JamaahHandler) CreateItinerary(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	var req model.CreateItineraryRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Title == "" {
		return response.BadRequest(c, "title is required")
	}
	it, err := h.svc.CreateItinerary(c.Context(), claims.OrgID, groupID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, it)
}

func (h *JamaahHandler) UpdateItinerary(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	itemID, err := uuid.Parse(c.Params("itemId"))
	if err != nil {
		return response.BadRequest(c, "invalid item id")
	}
	var req model.UpdateItineraryRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	it, err := h.svc.UpdateItinerary(c.Context(), claims.OrgID, itemID, req)
	if err != nil {
		return response.NotFound(c, err.Error())
	}
	return response.OK(c, it)
}

func (h *JamaahHandler) DeleteItinerary(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	itemID, err := uuid.Parse(c.Params("itemId"))
	if err != nil {
		return response.BadRequest(c, "invalid item id")
	}
	if err := h.svc.DeleteItinerary(c.Context(), claims.OrgID, itemID); err != nil {
		return response.NotFound(c, "itinerary not found")
	}
	return response.OK(c, fiber.Map{"message": "itinerary deleted"})
}

func (h *JamaahHandler) GetDocumentUrl(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{"url": ""})
}
