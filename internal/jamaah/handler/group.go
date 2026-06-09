package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

func (h *JamaahHandler) ListGroups(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groups, err := h.svc.ListGroups(c.Context(), claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"groups": groups})
}

func (h *JamaahHandler) CreateGroup(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	group, err := h.svc.CreateGroup(c.Context(), claims.OrgID, c.Get("Authorization"), req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, group)
}

func (h *JamaahHandler) GetGroup(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	group, err := h.svc.GetGroup(c.Context(), groupID, claims.OrgID)
	if err != nil || group == nil {
		return response.NotFound(c, "group not found")
	}
	return response.OK(c, group)
}

func (h *JamaahHandler) UpdateGroup(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	var req model.UpdateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	group, err := h.svc.UpdateGroup(c.Context(), groupID, claims.OrgID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, group)
}

func (h *JamaahHandler) DeleteGroup(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	if err := h.svc.DeleteGroup(c.Context(), groupID, claims.OrgID); err != nil {
		return response.NotFound(c, "group not found")
	}
	return response.OK(c, fiber.Map{"message": "group deleted"})
}

func (h *JamaahHandler) AddGroupMembers(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	var req struct {
		Members []model.AddGroupMemberRequest `json:"members"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	added, err := h.svc.AddGroupMembers(c.Context(), groupID, claims.OrgID, req.Members)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"added": added})
}

func (h *JamaahHandler) ListGroupMembers(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	members, err := h.svc.ListGroupMembers(c.Context(), groupID, claims.OrgID)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"members": members})
}

func (h *JamaahHandler) UpdateGroupMember(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	memberID, err := uuid.Parse(c.Params("memberId"))
	if err != nil {
		return response.BadRequest(c, "invalid member id")
	}
	var req model.UpdateGroupMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.UpdateGroupMember(c.Context(), groupID, memberID, claims.OrgID, req); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "member updated"})
}

func (h *JamaahHandler) DeleteGroupMember(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	memberID, err := uuid.Parse(c.Params("memberId"))
	if err != nil {
		return response.BadRequest(c, "invalid member id")
	}
	if err := h.svc.DeleteGroupMember(c.Context(), groupID, memberID, claims.OrgID); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "member removed"})
}
