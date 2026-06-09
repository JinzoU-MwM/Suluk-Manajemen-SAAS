package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

func (h *JamaahHandler) ListRooms(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var groupID *uuid.UUID
	if gid := c.Query("group_id"); gid != "" {
		if parsed, err := uuid.Parse(gid); err == nil {
			groupID = &parsed
		}
	}
	rooms, err := h.svc.ListRooms(c.Context(), claims.OrgID, groupID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"rooms": rooms})
}

func (h *JamaahHandler) CreateRoom(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	var groupID *uuid.UUID
	if gid := c.Query("group_id"); gid != "" {
		if parsed, err := uuid.Parse(gid); err == nil {
			groupID = &parsed
		}
	}
	room, err := h.svc.CreateRoom(c.Context(), claims.OrgID, groupID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, room)
}

func (h *JamaahHandler) DeleteRoom(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	roomID, err := uuid.Parse(c.Params("roomId"))
	if err != nil {
		return response.BadRequest(c, "invalid room id")
	}
	if err := h.svc.DeleteRoom(c.Context(), roomID, claims.OrgID); err != nil {
		return response.NotFound(c, "room not found")
	}
	return response.OK(c, fiber.Map{"message": "room deleted"})
}

func (h *JamaahHandler) AutoRooming(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	capacity := c.QueryInt("capacity", 4)
	rooms, err := h.svc.AutoRooming(c.Context(), claims.OrgID, groupID, capacity)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"rooms": rooms, "message": "auto-rooming completed"})
}

func (h *JamaahHandler) ClearAutoRooming(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	if err := h.svc.ClearAutoRooming(c.Context(), claims.OrgID, groupID); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "auto-rooming cleared"})
}

func (h *JamaahHandler) AssignMemberToRoom(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	roomID, err := uuid.Parse(c.Params("roomId"))
	if err != nil {
		return response.BadRequest(c, "invalid room id")
	}
	var req model.AssignMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	memberID := req.MemberID
	if memberID == "" {
		memberID = c.Params("memberId")
	}
	if err := h.svc.AssignMemberToRoom(c.Context(), claims.OrgID, roomID, memberID); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"assigned": true})
}

func (h *JamaahHandler) UnassignMember(c *fiber.Ctx) error {
	roomID, err := uuid.Parse(c.Params("roomId"))
	if err != nil {
		return response.BadRequest(c, "invalid room id")
	}
	memberID := c.Params("memberId")
	if err := h.svc.UnassignMember(c.Context(), roomID, memberID); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"unassigned": true})
}

func (h *JamaahHandler) ShareGroup(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	var req struct {
		Pin           string `json:"pin"`
		ExpiresInDays int    `json:"expires_in_days"`
	}
	_ = c.BodyParser(&req)
	if req.ExpiresInDays < 1 {
		req.ExpiresInDays = 30
	}
	sm, err := h.svc.ShareGroup(c.Context(), claims.OrgID, groupID, req.Pin, req.ExpiresInDays)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"token": sm.Token, "expires_at": sm.ExpiresAt})
}

func (h *JamaahHandler) RevokeShare(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	if err := h.svc.RevokeShare(c.Context(), claims.OrgID, groupID); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"revoked": true})
}

func (h *JamaahHandler) GetSharedManifest(c *fiber.Ctx) error {
	token := c.Params("token")
	sm, err := h.svc.GetSharedManifest(c.Context(), token)
	if err != nil || sm == nil {
		return response.NotFound(c, "manifest not found or expired")
	}
	return response.OK(c, fiber.Map{"manifest": sm})
}
