package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/invoice/repository"
	"github.com/jamaah-in/v2/internal/shared/response"
)

type openSessionReq struct {
	OpeningFloat int64  `json:"opening_float"`
	Notes        string `json:"notes"`
}
type closeSessionReq struct {
	CountedCash int64 `json:"counted_cash"`
}

func (h *InvoiceHandler) OpenCashSession(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	var req openSessionReq
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.OpeningFloat < 0 {
		return response.BadRequest(c, "opening_float tidak boleh negatif")
	}
	s, err := h.svc.OpenCashSession(c.Context(), claims.OrgID, claims.UserID, req.OpeningFloat, req.Notes)
	if err != nil {
		if err == repository.ErrSessionExists {
			return response.Conflict(c, "masih ada sesi kas terbuka — tutup dulu sebelum buka yang baru")
		}
		return response.Internal(c, err)
	}
	return response.Created(c, s)
}

func (h *InvoiceHandler) GetActiveCashSession(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	s, err := h.svc.GetActiveCashSession(c.Context(), claims.OrgID, claims.UserID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, s) // s may be null when no session open
}

func (h *InvoiceHandler) ListCashSessions(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	items, err := h.svc.ListCashSessions(c.Context(), claims.OrgID, c.QueryInt("limit", 30))
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, items)
}

func (h *InvoiceHandler) CloseCashSession(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	id, perr := uuid.Parse(c.Params("id"))
	if perr != nil {
		return response.BadRequest(c, "invalid session id")
	}
	var req closeSessionReq
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.CountedCash < 0 {
		return response.BadRequest(c, "counted_cash tidak boleh negatif")
	}
	s, err := h.svc.CloseCashSession(c.Context(), claims.OrgID, id, req.CountedCash)
	if err != nil {
		switch err {
		case repository.ErrSessionNotFound:
			return response.NotFound(c, "sesi kas tidak ditemukan")
		case repository.ErrSessionClosed:
			return response.Conflict(c, "sesi kas sudah ditutup")
		}
		return response.Internal(c, err)
	}
	return response.OK(c, s)
}
