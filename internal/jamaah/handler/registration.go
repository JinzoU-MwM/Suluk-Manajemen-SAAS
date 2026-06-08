package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/jamaah/model"
)

func (h *JamaahHandler) PublicRegistrationInfo(c *fiber.Ctx) error {
	token := c.Params("token")
	info, err := h.svc.GetPublicRegistrationInfo(c.Context(), token)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": info})
}

func (h *JamaahHandler) PublicRegistrationSubmit(c *fiber.Ctx) error {
	token := c.Params("token")

	phoneNumber := c.FormValue("phone_number")
	name := c.FormValue("name", "")
	email := c.FormValue("email", "")

	if phoneNumber == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "phone_number is required"})
	}

	filePaths := map[string]string{}
	fields := []string{"ktp", "passport", "visa"}
	for _, field := range fields {
		file, err := c.FormFile(field)
		if err == nil {
			fileName := fmt.Sprintf("uploads/registrations/%s_%s", field, uuid.New().String())
			dest := fmt.Sprintf("./%s", fileName)
			if err := c.SaveFile(file, dest); err == nil {
				filePaths[field] = fileName
			}
		}
	}

	req := model.PublicRegistrationSubmit{
		PhoneNumber: phoneNumber,
		Name:        name,
		Email:       email,
	}

	p, err := h.svc.SubmitPublicRegistration(c.Context(), token, req, filePaths)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": fiber.Map{"id": p.ID, "message": "Registration submitted successfully"}})
}

func (h *JamaahHandler) GenerateLink(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.GenerateLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	link, err := h.svc.GenerateRegistrationLink(c.Context(), claims.OrgID.String(), claims.UserID.String(), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": link})
}

func (h *JamaahHandler) GetActiveLink(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID := c.Params("groupId")
	link, err := h.svc.GetActiveLink(c.Context(), claims.OrgID.String(), groupID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": link})
}

func (h *JamaahHandler) RevokeLink(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID := c.Params("groupId")
	if err := h.svc.RevokeLink(c.Context(), claims.OrgID.String(), groupID); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "link revoked"}})
}

func (h *JamaahHandler) ListPendingRegistrations(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	groupID := c.Params("groupId")
	pending, err := h.svc.ListPending(c.Context(), claims.OrgID.String(), groupID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"registrations": pending}})
}

func (h *JamaahHandler) ApproveRegistration(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	pendingID := c.Params("pendingId")
	pid, err := uuid.Parse(pendingID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid pending id"})
	}
	if err := h.svc.ApprovePending(c.Context(), claims.OrgID.String(), pid, claims.UserID); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "registration approved"}})
}

func (h *JamaahHandler) RejectRegistration(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	pendingID := c.Params("pendingId")
	pid, err := uuid.Parse(pendingID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid pending id"})
	}
	if err := h.svc.RejectPending(c.Context(), claims.OrgID.String(), pid, claims.UserID); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "registration rejected"}})
}
