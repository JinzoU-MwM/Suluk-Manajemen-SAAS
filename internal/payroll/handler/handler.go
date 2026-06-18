package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/jamaah-in/v2/internal/payroll/model"
	"github.com/jamaah-in/v2/internal/payroll/repository"
	"github.com/jamaah-in/v2/internal/payroll/service"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
)

// validateEmployeeMoney returns a non-empty message when any salary field is out
// of range (negatives, or a PPh21 rate outside 0–100). Garbage here produces
// net<0/gross≤0 slips whose payroll.posted event the accounting consumer drops.
func validateEmployeeMoney(base, allowance, bpjsTk, bpjsKes int64, pph21 float64) string {
	switch {
	case base < 0:
		return "base_salary tidak boleh negatif"
	case allowance < 0:
		return "allowance tidak boleh negatif"
	case bpjsTk < 0 || bpjsKes < 0:
		return "bpjs tidak boleh negatif"
	case pph21 < 0 || pph21 > 100:
		return "pph21_rate harus antara 0 dan 100"
	}
	return ""
}

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
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "name wajib diisi"})
	}
	if msg := validateEmployeeMoney(req.BaseSalary, req.Allowance, req.BpjsTk, req.BpjsKes, req.Pph21Rate); msg != "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": msg})
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
	if (req.BaseSalary != nil && *req.BaseSalary < 0) ||
		(req.Allowance != nil && *req.Allowance < 0) ||
		(req.BpjsTk != nil && *req.BpjsTk < 0) ||
		(req.BpjsKes != nil && *req.BpjsKes < 0) {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "nilai gaji/bpjs tidak boleh negatif"})
	}
	if req.Pph21Rate != nil && (*req.Pph21Rate < 0 || *req.Pph21Rate > 100) {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "pph21_rate harus antara 0 dan 100"})
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
	if req.EmployeeID == "" || req.Period == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "employee_id dan period wajib diisi"})
	}
	slip, err := h.svc.CreateSalarySlip(c.Context(), claims.OrgID.String(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlipExists):
			return c.Status(409).JSON(fiber.Map{"success": false, "error": "slip gaji untuk karyawan & periode ini sudah ada"})
		case errors.Is(err, repository.ErrEmployeeNotFound):
			return c.Status(404).JSON(fiber.Map{"success": false, "error": "karyawan tidak ditemukan"})
		}
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "gagal membuat slip gaji"})
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
	if req.EmployeeID == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "employee_id wajib diisi"})
	}
	a, err := h.svc.CreateAdvance(c.Context(), claims.OrgID.String(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAmount):
			return c.Status(400).JSON(fiber.Map{"success": false, "error": "amount harus lebih dari 0"})
		case errors.Is(err, repository.ErrEmployeeNotFound):
			return c.Status(404).JSON(fiber.Map{"success": false, "error": "karyawan tidak ditemukan"})
		}
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "gagal membuat kasbon"})
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
		switch {
		case errors.Is(err, repository.ErrAdvanceNotFound):
			return c.Status(404).JSON(fiber.Map{"success": false, "error": "kasbon tidak ditemukan"})
		case errors.Is(err, repository.ErrRepayInvalid):
			return c.Status(400).JSON(fiber.Map{"success": false, "error": "amount harus lebih dari 0"})
		case errors.Is(err, repository.ErrRepayTooMuch):
			return c.Status(409).JSON(fiber.Map{"success": false, "error": "pembayaran melebihi sisa kasbon"})
		}
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "gagal mencatat pembayaran kasbon"})
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
