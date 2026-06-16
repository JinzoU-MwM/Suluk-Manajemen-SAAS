package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/jamaah-in/v2/internal/payroll/model"
	"github.com/jamaah-in/v2/internal/payroll/repository"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
)

func (h *PayrollHandler) RecordAttendance(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.RecordAttendanceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	a, err := h.svc.RecordAttendance(c.Context(), claims.OrgID.String(), req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": a})
}

func (h *PayrollHandler) ListAttendance(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	empID := c.Query("employee_id")
	period := c.Query("period")
	if empID == "" || period == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "employee_id and period required"})
	}
	list, err := h.svc.ListAttendance(c.Context(), claims.OrgID.String(), empID, period)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	sum, _ := h.svc.AttendanceSummary(c.Context(), claims.OrgID.String(), empID, period)
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"attendance": list, "summary": sum}})
}

func (h *PayrollHandler) CreateLeave(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.CreateLeaveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	l, err := h.svc.CreateLeave(c.Context(), claims.OrgID.String(), req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": l})
}

func (h *PayrollHandler) ListLeave(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	list, err := h.svc.ListLeave(c.Context(), claims.OrgID.String(), c.Query("status", "all"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"leave": list}})
}

func (h *PayrollHandler) DecideLeave(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.DecideLeaveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	err := h.svc.DecideLeave(c.Context(), c.Params("id"), claims.OrgID.String(), claims.UserID.String(), req)
	if errors.Is(err, repository.ErrLeaveNotFound) {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "ok"}})
}
