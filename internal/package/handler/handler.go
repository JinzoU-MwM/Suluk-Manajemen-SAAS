package handler

import (
	"errors"
	"slices"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/package/model"
	"github.com/jamaah-in/v2/internal/package/repository"
	"github.com/jamaah-in/v2/internal/package/service"
	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/response"
)

type PackageHandler struct {
	svc *service.PackageService
}

func NewPackageHandler(svc *service.PackageService) *PackageHandler {
	return &PackageHandler{svc: svc}
}

// ReserveSeat reserves one seat (called by jamaah-service on registration).
func (h *PackageHandler) ReserveSeat(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid package id")
	}
	var body struct {
		RoomType string `json:"room_type"`
	}
	_ = c.BodyParser(&body) // optional; empty room_type reserves only the total
	if err := h.svc.ReserveSeat(c.Context(), id, claims.OrgID, body.RoomType); err != nil {
		if errors.Is(err, repository.ErrPackageFull) {
			return response.Conflict(c, "kuota paket sudah penuh")
		}
		if errors.Is(err, repository.ErrRoomTypeFull) {
			return response.Conflict(c, "kuota tipe kamar sudah penuh")
		}
		if errors.Is(err, repository.ErrPackageNotFound) {
			return response.NotFound(c, "paket tidak ditemukan")
		}
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"reserved": true})
}

// ReleaseSeat frees one previously-reserved seat (called on unregister/rollback).
func (h *PackageHandler) ReleaseSeat(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid package id")
	}
	var body struct {
		RoomType string `json:"room_type"`
	}
	_ = c.BodyParser(&body)
	if err := h.svc.ReleaseSeat(c.Context(), id, claims.OrgID, body.RoomType); err != nil {
		if errors.Is(err, repository.ErrPackageNotFound) {
			return response.NotFound(c, "paket tidak ditemukan")
		}
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"released": true})
}

func (h *PackageHandler) CreatePackage(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if !canEditPackages(claims.Role) {
		return response.Forbidden(c, "insufficient permissions to create package")
	}

	var req model.CreatePackageRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	if req.PackageType == "" {
		return response.BadRequest(c, "package_type is required")
	}
	if !slices.Contains(model.ValidPackageTypes(), req.PackageType) {
		return response.BadRequest(c, "package_type tidak valid")
	}
	if req.TotalSeats < 1 {
		return response.BadRequest(c, "total_seats must be at least 1")
	}

	pkg, err := h.svc.CreatePackage(c.Context(), claims.OrgID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Created(c, pkg)
}

func (h *PackageHandler) GetPackage(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid package id")
	}

	pkg, err := h.svc.GetPackage(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}
	return response.OK(c, pkg)
}

func (h *PackageHandler) ListPackages(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	status := c.Query("status")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("page_size", "20"))

	packages, total, err := h.svc.ListPackages(c.Context(), claims.OrgID, status, page, limit)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Paginated(c, packages, int64(total), page, limit)
}

func (h *PackageHandler) UpdatePackage(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if !canEditPackages(claims.Role) {
		return response.Forbidden(c, "insufficient permissions to update package")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid package id")
	}
	existing, err := h.svc.GetPackage(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}

	var req model.UpdatePackageRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.IsPublished != nil && !canPublishPackages(claims.Role) {
		return response.Forbidden(c, "only owner can publish package")
	}
	if req.TotalSeats != nil {
		if *req.TotalSeats < 1 {
			return response.BadRequest(c, "total_seats must be at least 1")
		}
		if *req.TotalSeats < existing.ReservedSeats {
			return response.BadRequest(c, "total_seats tidak boleh kurang dari kursi yang sudah dipesan")
		}
	}
	if req.PackageType != nil && !slices.Contains(model.ValidPackageTypes(), *req.PackageType) {
		return response.BadRequest(c, "package_type tidak valid")
	}

	pkg, err := h.svc.UpdatePackage(c.Context(), id, claims.OrgID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, pkg)
}

func (h *PackageHandler) DeletePackage(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if !canDeletePackages(claims.Role) {
		return response.Forbidden(c, "only owner can delete package")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid package id")
	}
	_, err = h.svc.GetPackage(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}
	if err := h.svc.DeletePackage(c.Context(), id, claims.OrgID); err != nil {
		return response.NotFound(c, "package not found")
	}
	return response.OK(c, fiber.Map{"message": "package deleted"})
}

func (h *PackageHandler) UpdatePackageStatus(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if !canEditPackages(claims.Role) {
		return response.Forbidden(c, "insufficient permissions to update package status")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid package id")
	}
	_, err = h.svc.GetPackage(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}

	var req model.UpdatePackageStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if !slices.Contains(model.ValidPackageStatuses(), req.Status) {
		return response.BadRequest(c, "status tidak valid")
	}

	pkg, err := h.svc.UpdatePackageStatus(c.Context(), id, claims.OrgID, req.Status)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, pkg)
}

func (h *PackageHandler) GetPackageQuota(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid package id")
	}
	_, err = h.svc.GetPackage(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}
	quota, err := h.svc.GetPackageQuota(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}
	return response.OK(c, quota)
}

func (h *PackageHandler) GetProfitProjection(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid package id")
	}
	_, err = h.svc.GetPackage(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}
	proj, err := h.svc.GetProfitProjection(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}
	return response.OK(c, proj)
}

func (h *PackageHandler) CreatePricingTier(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if !canEditPackages(claims.Role) {
		return response.Forbidden(c, "insufficient permissions to manage pricing tiers")
	}

	packageID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid package id")
	}
	_, err = h.svc.GetPackage(c.Context(), packageID, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}

	var req model.CreatePricingTierRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.RoomType == "" {
		return response.BadRequest(c, "room_type is required")
	}
	if !slices.Contains(model.ValidRoomTypes(), req.RoomType) {
		return response.BadRequest(c, "room_type tidak valid")
	}
	if req.Price < 1 {
		return response.BadRequest(c, "price must be at least 1")
	}

	tier, err := h.svc.CreatePricingTier(c.Context(), packageID, claims.OrgID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Created(c, tier)
}

func (h *PackageHandler) UpdatePricingTier(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if !canEditPackages(claims.Role) {
		return response.Forbidden(c, "insufficient permissions to manage pricing tiers")
	}

	tierID, err := uuid.Parse(c.Params("tid"))
	if err != nil {
		return response.BadRequest(c, "invalid tier id")
	}
	tier, err := h.svc.GetPricingTier(c.Context(), tierID)
	if err != nil || tier == nil {
		return response.NotFound(c, "tier not found")
	}
	_, err = h.svc.GetPackage(c.Context(), tier.PackageID, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}

	var req model.CreatePricingTierRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.RoomType == "" || !slices.Contains(model.ValidRoomTypes(), req.RoomType) {
		return response.BadRequest(c, "room_type tidak valid")
	}
	if req.Price < 1 {
		return response.BadRequest(c, "price must be at least 1")
	}

	tier, err = h.svc.UpdatePricingTier(c.Context(), tierID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, tier)
}

func (h *PackageHandler) DeletePricingTier(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if !canEditPackages(claims.Role) {
		return response.Forbidden(c, "insufficient permissions to manage pricing tiers")
	}

	tierID, err := uuid.Parse(c.Params("tid"))
	if err != nil {
		return response.BadRequest(c, "invalid tier id")
	}
	tier, err := h.svc.GetPricingTier(c.Context(), tierID)
	if err != nil || tier == nil {
		return response.NotFound(c, "tier not found")
	}
	_, err = h.svc.GetPackage(c.Context(), tier.PackageID, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}
	if err := h.svc.DeletePricingTier(c.Context(), tierID); err != nil {
		return response.NotFound(c, "tier not found")
	}
	return response.OK(c, fiber.Map{"message": "tier deleted"})
}

func (h *PackageHandler) CreateCostComponent(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if !canEditPackages(claims.Role) {
		return response.Forbidden(c, "insufficient permissions to manage cost components")
	}

	packageID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid package id")
	}
	_, err = h.svc.GetPackage(c.Context(), packageID, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}

	var req model.CreateCostComponentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	if req.Category == "" || !slices.Contains(model.ValidCostCategories(), req.Category) {
		return response.BadRequest(c, "category tidak valid")
	}
	if req.Quantity < 1 {
		return response.BadRequest(c, "quantity must be at least 1")
	}
	if req.AmountPerPerson < 0 {
		return response.BadRequest(c, "amount_per_person tidak boleh negatif")
	}

	cc, err := h.svc.CreateCostComponent(c.Context(), packageID, claims.OrgID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Created(c, cc)
}

func (h *PackageHandler) UpdateCostComponent(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if !canEditPackages(claims.Role) {
		return response.Forbidden(c, "insufficient permissions to manage cost components")
	}

	ccID, err := uuid.Parse(c.Params("cid"))
	if err != nil {
		return response.BadRequest(c, "invalid cost component id")
	}
	cc, err := h.svc.GetCostComponent(c.Context(), ccID)
	if err != nil || cc == nil {
		return response.NotFound(c, "cost component not found")
	}
	_, err = h.svc.GetPackage(c.Context(), cc.PackageID, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}

	var req model.CreateCostComponentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	if req.Category == "" || !slices.Contains(model.ValidCostCategories(), req.Category) {
		return response.BadRequest(c, "category tidak valid")
	}
	if req.Quantity < 1 {
		return response.BadRequest(c, "quantity must be at least 1")
	}
	if req.AmountPerPerson < 0 {
		return response.BadRequest(c, "amount_per_person tidak boleh negatif")
	}

	cc, err = h.svc.UpdateCostComponent(c.Context(), ccID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, cc)
}

func (h *PackageHandler) DeleteCostComponent(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if !canEditPackages(claims.Role) {
		return response.Forbidden(c, "insufficient permissions to manage cost components")
	}

	ccID, err := uuid.Parse(c.Params("cid"))
	if err != nil {
		return response.BadRequest(c, "invalid cost component id")
	}
	cc, err := h.svc.GetCostComponent(c.Context(), ccID)
	if err != nil || cc == nil {
		return response.NotFound(c, "cost component not found")
	}
	_, err = h.svc.GetPackage(c.Context(), cc.PackageID, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "package not found")
	}
	if err := h.svc.DeleteCostComponent(c.Context(), ccID); err != nil {
		return response.NotFound(c, "cost component not found")
	}
	return response.OK(c, fiber.Map{"message": "cost component deleted"})
}

func (h *PackageHandler) GetPublicPackage(c *fiber.Ctx) error {
	slug := c.Params("slug")
	pkg, err := h.svc.GetPackageBySlugPublic(c.Context(), slug)
	if err != nil {
		return response.NotFound(c, "package not found")
	}
	if !pkg.IsPublished {
		return response.NotFound(c, "package not found")
	}
	public := fiber.Map{
		"id":                   pkg.ID,
		"name":                 pkg.Name,
		"slug":                 pkg.Slug,
		"description":          pkg.Description,
		"package_type":         pkg.PackageType,
		"departure_date":       pkg.DepartureDate,
		"return_date":          pkg.ReturnDate,
		"duration_days":        pkg.DurationDays,
		"total_seats":          pkg.TotalSeats,
		"reserved_seats":       pkg.ReservedSeats,
		"available_seats":      max(0, pkg.TotalSeats-pkg.ReservedSeats),
		"airline":              pkg.Airline,
		"flight_number_go":     pkg.FlightNumberGo,
		"flight_number_return": pkg.FlightNumberReturn,
		"hotel_makkah_name":    pkg.HotelMakkahName,
		"hotel_madinah_name":   pkg.HotelMadinahName,
		"pricing_tiers":        pkg.PricingTiers,
	}
	return response.OK(c, public)
}

func canEditPackages(role string) bool {
	return role == "owner" || role == "admin"
}

func canPublishPackages(role string) bool {
	return role == "owner"
}

func canDeletePackages(role string) bool {
	return role == "owner"
}
