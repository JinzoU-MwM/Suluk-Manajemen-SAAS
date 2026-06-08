package handler

import (
	"github.com/gofiber/fiber/v2"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/payroll/model"
	"github.com/jamaah-in/v2/internal/payroll/service"
)

type PayrollHandler struct {
	svc *service.PayrollService
}

func NewPayrollHandler(svc *service.PayrollService) *PayrollHandler {
	return &PayrollHandler{svc: svc}
}

func (h *PayrollHandler) CreateEmployee(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.CreateEmployeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	e, err := h.svc.CreateEmployee(c.Context(), claims.OrgID.String(), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": e})
}

func (h *PayrollHandler) ListEmployees(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	employees, err := h.svc.ListEmployees(c.Context(), claims.OrgID.String())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"employees": employees}})
}

func (h *PayrollHandler) GetEmployee(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	e, err := h.svc.GetEmployee(c.Context(), c.Params("id"), claims.OrgID.String())
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": e})
}

func (h *PayrollHandler) UpdateEmployee(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.UpdateEmployeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	e, err := h.svc.UpdateEmployee(c.Context(), c.Params("id"), claims.OrgID.String(), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": e})
}

func (h *PayrollHandler) CreateSalarySlip(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.CreateSalarySlipRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	slip, err := h.svc.CreateSalarySlip(c.Context(), claims.OrgID.String(), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": slip})
}

func (h *PayrollHandler) ListSalarySlips(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	period := c.Query("period", "")
	slips, err := h.svc.ListSalarySlips(c.Context(), claims.OrgID.String(), period)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"slips": slips}})
}

func (h *PayrollHandler) FinalizeSlip(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	if err := h.svc.FinalizeSlip(c.Context(), c.Params("id"), claims.OrgID.String()); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "salary slip finalized"}})
}

func (h *PayrollHandler) CreateAdvance(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.CreateAdvanceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	a, err := h.svc.CreateAdvance(c.Context(), claims.OrgID.String(), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": a})
}

func (h *PayrollHandler) ListAdvances(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	advances, err := h.svc.ListAdvances(c.Context(), claims.OrgID.String())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"advances": advances}})
}

func (h *PayrollHandler) RepayAdvance(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.RepayAdvanceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	if err := h.svc.RepayAdvance(c.Context(), c.Params("id"), claims.OrgID.String(), req); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "advance repaid"}})
}

func (h *PayrollHandler) GetSummary(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	s, err := h.svc.GetPayrollSummary(c.Context(), claims.OrgID.String())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": s})
}
