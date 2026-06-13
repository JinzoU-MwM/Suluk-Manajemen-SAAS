package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/response"
	"github.com/jamaah-in/v2/internal/tabungan/model"
	"github.com/jamaah-in/v2/internal/tabungan/repository"
	"github.com/jamaah-in/v2/internal/tabungan/service"
)

type Handler struct{ svc *service.Service }

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	items, total, err := h.svc.ListAccounts(c.Context(), claims.OrgID, page, limit)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Paginated(c, items, int64(total), page, limit)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	id, perr := uuid.Parse(c.Params("id"))
	if perr != nil {
		return response.BadRequest(c, "invalid id")
	}
	a, err := h.svc.GetAccount(c.Context(), claims.OrgID, id)
	if err != nil {
		return response.NotFound(c, "tabungan tidak ditemukan")
	}
	return response.OK(c, a)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	var req model.CreateAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.JamaahID == "" {
		return response.BadRequest(c, "jamaah_id wajib diisi")
	}
	a, err := h.svc.CreateAccount(c.Context(), claims.OrgID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, a)
}

func (h *Handler) Deposit(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	id, perr := uuid.Parse(c.Params("id"))
	if perr != nil {
		return response.BadRequest(c, "invalid id")
	}
	var req model.DepositRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Amount < 1 {
		return response.BadRequest(c, "amount minimal 1")
	}
	a, err := h.svc.Deposit(c.Context(), claims.OrgID, claims.UserID, id, req)
	if err != nil {
		if err == repository.ErrNotActive {
			return response.Conflict(c, "tabungan tidak aktif")
		}
		if err == repository.ErrNotFound {
			return response.NotFound(c, "tabungan tidak ditemukan")
		}
		return response.Internal(c, err)
	}
	return response.OK(c, a)
}

func (h *Handler) Convert(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	id, perr := uuid.Parse(c.Params("id"))
	if perr != nil {
		return response.BadRequest(c, "invalid id")
	}
	var req model.ConvertRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.InvoiceID == "" {
		return response.BadRequest(c, "invoice_id wajib diisi")
	}
	a, err := h.svc.Convert(c.Context(), claims.OrgID, claims.UserID, id, req)
	if err != nil {
		switch err {
		case repository.ErrInsufficient:
			return response.Conflict(c, "saldo tabungan tidak cukup")
		case repository.ErrNotActive:
			return response.Conflict(c, "tabungan tidak aktif")
		case repository.ErrNotFound:
			return response.NotFound(c, "tabungan tidak ditemukan")
		}
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, a)
}
