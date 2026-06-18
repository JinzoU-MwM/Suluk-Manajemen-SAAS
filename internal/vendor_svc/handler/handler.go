package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/response"
	"github.com/jamaah-in/v2/internal/vendor_svc/model"
	"github.com/jamaah-in/v2/internal/vendor_svc/repository"
	"github.com/jamaah-in/v2/internal/vendor_svc/service"
)

type VendorHandler struct {
	svc *service.VendorService
}

func NewVendorHandler(svc *service.VendorService) *VendorHandler {
	return &VendorHandler{svc: svc}
}

// --- Vendor Master ---

func (h *VendorHandler) CreateVendor(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}

	var req model.CreateVendorRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	if req.Type == "" {
		return response.BadRequest(c, "type is required")
	}
	if !isValidVendorType(req.Type) {
		return response.BadRequest(c, "type must be one of: maskapai, hotel, transport, perlengkapan, katering, lainnya")
	}

	vendor, err := h.svc.CreateVendor(c.Context(), claims.OrgID, req)
	if err != nil {
		return writeError(c, err)
	}
	return response.Created(c, vendor)
}

func (h *VendorHandler) GetVendor(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid vendor id")
	}

	vendor, err := h.svc.GetVendor(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "vendor not found")
	}
	return response.OK(c, vendor)
}

func (h *VendorHandler) UpdateVendor(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid vendor id")
	}

	var req model.UpdateVendorRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Type != nil && !isValidVendorType(*req.Type) {
		return response.BadRequest(c, "type must be one of: maskapai, hotel, transport, perlengkapan, katering, lainnya")
	}

	vendor, err := h.svc.UpdateVendor(c.Context(), id, claims.OrgID, req)
	if err != nil {
		if errors.Is(err, repository.ErrVendorNotFound) {
			return response.NotFound(c, "vendor not found")
		}
		return writeError(c, err)
	}
	return response.OK(c, vendor)
}

func (h *VendorHandler) DeleteVendor(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid vendor id")
	}
	if err := h.svc.DeleteVendor(c.Context(), id, claims.OrgID); err != nil {
		if errors.Is(err, service.ErrVendorNotFound) {
			return response.NotFound(c, "vendor not found")
		}
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "violates") {
			return response.Conflict(c, "cannot delete vendor: has associated bills or payments")
		}
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "vendor deleted"})
}

func (h *VendorHandler) ListVendors(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	vendorType := c.Query("type")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("page_size", "20"))

	vendors, total, err := h.svc.ListVendors(c.Context(), claims.OrgID, vendorType, search, page, limit)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Paginated(c, vendors, int64(total), page, limit)
}

// --- Vendor Bills ---

func (h *VendorHandler) CreateBill(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}

	var req model.CreateBillRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.VendorID == uuid.Nil {
		return response.BadRequest(c, "vendor_id is required")
	}
	if req.PackageID == uuid.Nil {
		return response.BadRequest(c, "package_id is required")
	}
	if req.Description == "" {
		return response.BadRequest(c, "description is required")
	}
	if req.Amount < 1 {
		return response.BadRequest(c, "amount must be at least 1")
	}
	if req.Currency != "" {
		req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
		if !isValidCurrency(req.Currency) {
			return response.BadRequest(c, "currency harus salah satu dari: IDR, USD")
		}
	}
	if req.ExchangeRate < 0 {
		return response.BadRequest(c, "exchange_rate tidak boleh negatif")
	}
	if req.DueDate != "" {
		if _, err := repository.ParseDate(req.DueDate); err != nil {
			return response.BadRequest(c, "format due_date tidak valid (gunakan YYYY-MM-DD)")
		}
	}

	bill, err := h.svc.CreateBill(c.Context(), claims.OrgID, req, c.Get("Authorization"))
	if err != nil {
		if errors.Is(err, repository.ErrVendorNotFound) {
			return response.BadRequest(c, "vendor_id tidak ditemukan untuk organisasi ini")
		}
		if errors.Is(err, service.ErrPackageNotFound) {
			return response.BadRequest(c, "package_id tidak ditemukan untuk organisasi ini")
		}
		return writeError(c, err)
	}
	return response.Created(c, bill)
}

func (h *VendorHandler) GetBill(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid bill id")
	}

	bill, err := h.svc.GetBill(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "vendor bill not found")
	}
	return response.OK(c, bill)
}

func (h *VendorHandler) UpdateBill(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid bill id")
	}

	var req model.UpdateBillRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Amount != nil && *req.Amount < 1 {
		return response.BadRequest(c, "amount must be at least 1")
	}
	if req.Currency != nil {
		*req.Currency = strings.ToUpper(strings.TrimSpace(*req.Currency))
		if !isValidCurrency(*req.Currency) {
			return response.BadRequest(c, "currency harus salah satu dari: IDR, USD")
		}
	}
	if req.ExchangeRate != nil && *req.ExchangeRate <= 0 {
		return response.BadRequest(c, "exchange_rate harus lebih dari 0")
	}
	if req.DueDate != nil && *req.DueDate != "" {
		if _, err := repository.ParseDate(*req.DueDate); err != nil {
			return response.BadRequest(c, "format due_date tidak valid (gunakan YYYY-MM-DD)")
		}
	}

	bill, err := h.svc.UpdateBill(c.Context(), id, claims.OrgID, req)
	if err != nil {
		if errors.Is(err, repository.ErrBillNotFound) {
			return response.NotFound(c, "vendor bill not found")
		}
		return writeError(c, err)
	}
	return response.OK(c, bill)
}

func (h *VendorHandler) DeleteBill(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid bill id")
	}
	if err := h.svc.DeleteBill(c.Context(), id, claims.OrgID); err != nil {
		if errors.Is(err, repository.ErrBillNotFound) {
			return response.NotFound(c, "vendor bill not found")
		}
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "violates") {
			return response.Conflict(c, "cannot delete bill: has associated payments")
		}
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "vendor bill deleted"})
}

func (h *VendorHandler) ListBills(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	var vendorID *uuid.UUID
	if vid := c.Query("vendor_id"); vid != "" {
		parsed, err := uuid.Parse(vid)
		if err == nil {
			vendorID = &parsed
		}
	}
	var packageID *uuid.UUID
	if pid := c.Query("package_id"); pid != "" {
		parsed, err := uuid.Parse(pid)
		if err == nil {
			packageID = &parsed
		}
	}
	status := c.Query("status")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("page_size", "20"))

	bills, total, err := h.svc.ListBills(c.Context(), claims.OrgID, vendorID, packageID, status, page, limit)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Paginated(c, bills, int64(total), page, limit)
}

func (h *VendorHandler) GetOverdueBills(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	bills, err := h.svc.GetOverdueBills(c.Context(), claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, bills)
}

func (h *VendorHandler) GetBillsDueSoon(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	days, _ := strconv.Atoi(c.Query("days", "7"))
	bills, err := h.svc.GetBillsDueSoon(c.Context(), claims.OrgID, days)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, bills)
}

func (h *VendorHandler) GetDebtSummary(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	var vendorID *uuid.UUID
	if vid := c.Query("vendor_id"); vid != "" {
		parsed, err := uuid.Parse(vid)
		if err == nil {
			vendorID = &parsed
		}
	}

	summary, err := h.svc.GetDebtSummary(c.Context(), claims.OrgID, vendorID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, summary)
}

func (h *VendorHandler) GetPackageBillSummary(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	packageID, err := uuid.Parse(c.Params("pkgId"))
	if err != nil {
		return response.BadRequest(c, "invalid package id")
	}
	summary, err := h.svc.GetPackageBillSummary(c.Context(), claims.OrgID, packageID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, summary)
}

// --- Vendor Payments ---

func (h *VendorHandler) CreatePayment(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}

	var req model.CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.VendorBillID == uuid.Nil {
		return response.BadRequest(c, "vendor_bill_id is required")
	}
	if req.PaymentDate == "" {
		return response.BadRequest(c, "payment_date is required")
	}
	if _, err := repository.ParseDate(req.PaymentDate); err != nil {
		return response.BadRequest(c, "format payment_date tidak valid (gunakan YYYY-MM-DD)")
	}
	if req.Amount < 1 {
		return response.BadRequest(c, "amount must be at least 1")
	}
	if req.Currency != "" {
		req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
		if !isValidCurrency(req.Currency) {
			return response.BadRequest(c, "currency harus salah satu dari: IDR, USD")
		}
	}
	if req.ExchangeRate < 0 {
		return response.BadRequest(c, "exchange_rate tidak boleh negatif")
	}

	payment, err := h.svc.CreatePayment(c.Context(), claims.OrgID, req)
	if err != nil {
		if errors.Is(err, repository.ErrBillNotFound) {
			return response.NotFound(c, "vendor bill not found")
		}
		return writeError(c, err)
	}
	return response.Created(c, payment)
}

func (h *VendorHandler) GetPayment(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid payment id")
	}

	payment, err := h.svc.GetPayment(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "vendor payment not found")
	}
	return response.OK(c, payment)
}

func (h *VendorHandler) DeletePayment(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid payment id")
	}
	if err := h.svc.DeletePayment(c.Context(), id, claims.OrgID); err != nil {
		return response.NotFound(c, "vendor payment not found")
	}
	return response.OK(c, fiber.Map{"message": "vendor payment deleted"})
}

func (h *VendorHandler) ListPaymentsByBill(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	billID, err := uuid.Parse(c.Params("billId"))
	if err != nil {
		return response.BadRequest(c, "invalid bill id")
	}

	payments, err := h.svc.ListPaymentsByBill(c.Context(), billID, claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, payments)
}

func (h *VendorHandler) ListPaymentsByVendor(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	vendorID, err := uuid.Parse(c.Params("vendorId"))
	if err != nil {
		return response.BadRequest(c, "invalid vendor id")
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("page_size", "20"))

	payments, total, err := h.svc.ListPaymentsByVendor(c.Context(), vendorID, claims.OrgID, page, limit)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Paginated(c, payments, int64(total), page, limit)
}

func isValidVendorType(s string) bool {
	for _, v := range model.ValidVendorTypes() {
		if s == v {
			return true
		}
	}
	return false
}

func isValidCurrency(s string) bool {
	for _, v := range model.ValidCurrencies() {
		if s == v {
			return true
		}
	}
	return false
}

// writeError maps known Postgres constraint violations from a write to a precise
// 4xx instead of a generic 500: a unique/duplicate violation is a 409, and a
// length-overflow or CHECK violation is bad input (400). Anything else is a real
// server error. Keeps create/update paths from surfacing raw DB failures as 500.
func writeError(c *fiber.Ctx, err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint"):
		return response.Conflict(c, "data sudah ada (duplikat)")
	case strings.Contains(msg, "value too long"):
		return response.BadRequest(c, "salah satu nilai melebihi batas panjang yang diizinkan")
	case strings.Contains(msg, "violates check constraint"):
		return response.BadRequest(c, "nilai tidak valid")
	default:
		return response.Internal(c, err)
	}
}
