package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
	"github.com/jamaah-in/v2/internal/jamaah/repository"
	"github.com/jamaah-in/v2/internal/jamaah/service"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

// ListVisas returns the visa board (filterable by status/search).
func (h *JamaahHandler) ListVisas(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("page_size", 50)
	rows, total, err := h.svc.ListVisas(c.Context(), claims.OrgID, c.Query("status"), c.Query("search"), page, limit)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Paginated(c, rows, int64(total), page, limit)
}

// GetVisa returns a jamaah's visa application (null if none yet).
func (h *JamaahHandler) GetVisa(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	jamaahID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid jamaah id")
	}
	v, err := h.svc.GetVisa(c.Context(), claims.OrgID, jamaahID)
	if errors.Is(err, repository.ErrVisaNotFound) {
		return response.OK(c, nil)
	}
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, v)
}

// UpsertVisa creates or edits the draft visa application.
func (h *JamaahHandler) UpsertVisa(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	jamaahID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid jamaah id")
	}
	var req model.UpsertVisaRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	v, err := h.svc.UpsertVisa(c.Context(), claims.OrgID, jamaahID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, v)
}

// TransitionVisa moves the visa application to a new status.
func (h *JamaahHandler) TransitionVisa(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	jamaahID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid jamaah id")
	}
	var req model.VisaTransitionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	v, err := h.svc.TransitionVisa(c.Context(), claims.OrgID, claims.UserID, jamaahID, req)
	if errors.Is(err, service.ErrVisaGate) {
		return response.BadRequest(c, err.Error())
	}
	if errors.Is(err, repository.ErrVisaNotFound) {
		return response.NotFound(c, "belum ada pengajuan visa untuk jamaah ini")
	}
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, v)
}

// GetVisaHistory returns the audit trail of a jamaah's visa transitions.
func (h *JamaahHandler) GetVisaHistory(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	jamaahID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid jamaah id")
	}
	hist, err := h.svc.GetVisaHistory(c.Context(), claims.OrgID, jamaahID)
	if errors.Is(err, repository.ErrVisaNotFound) {
		return response.OK(c, []any{})
	}
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, hist)
}
