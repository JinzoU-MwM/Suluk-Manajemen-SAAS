package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/accounting/model"
	"github.com/jamaah-in/v2/internal/accounting/service"
	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/response"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

// ---- COA ----

func (h *Handler) ListAccounts(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	accts, err := h.svc.ListAccounts(c.Context(), claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, accts)
}

type createAccountReq struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (h *Handler) CreateAccount(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	var req createAccountReq
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Code == "" || req.Name == "" {
		return response.BadRequest(c, "code dan name wajib diisi")
	}
	switch req.Type {
	case model.TypeAsset, model.TypeLiability, model.TypeEquity, model.TypeRevenue, model.TypeExpense:
	default:
		return response.BadRequest(c, "type harus salah satu: asset, liability, equity, revenue, expense")
	}
	a := &model.Account{OrgID: claims.OrgID, Code: req.Code, Name: req.Name, Type: req.Type}
	if err := h.svc.CreateAccount(c.Context(), a); err != nil {
		return response.Conflict(c, err.Error())
	}
	return response.Created(c, a)
}

// ---- Journals ----

func (h *Handler) ListJournals(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	journals, total, err := h.svc.ListJournals(c.Context(), claims.OrgID, page, limit)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Paginated(c, journals, int64(total), page, limit)
}

func (h *Handler) GetJournal(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	id, perr := uuid.Parse(c.Params("id"))
	if perr != nil {
		return response.BadRequest(c, "invalid journal id")
	}
	j, err := h.svc.GetJournal(c.Context(), claims.OrgID, id)
	if err != nil {
		return response.NotFound(c, "journal not found")
	}
	return response.OK(c, j)
}

func (h *Handler) GeneralLedger(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	accID, perr := uuid.Parse(c.Params("accountId"))
	if perr != nil {
		return response.BadRequest(c, "invalid account id")
	}
	from, to := dateRange(c)
	lines, err := h.svc.GeneralLedger(c.Context(), claims.OrgID, accID, from, to)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, lines)
}

// ---- Reports ----

func (h *Handler) TrialBalance(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	asOf := queryDate(c, "as_of", time.Now())
	rows, err := h.svc.TrialBalance(c.Context(), claims.OrgID, asOf)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, rows)
}

// Insights is the AI accounting copilot: rule-based financial findings over the
// GL, with an optional AI narrative.
func (h *Handler) Insights(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	asOf := queryDate(c, "as_of", time.Now())
	rep, err := h.svc.GenerateInsights(c.Context(), claims.OrgID, asOf)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, rep)
}

func (h *Handler) BalanceSheet(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	asOf := queryDate(c, "as_of", time.Now())
	bs, err := h.svc.BalanceSheet(c.Context(), claims.OrgID, asOf)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, bs)
}

func (h *Handler) IncomeStatement(c *fiber.Ctx) error {
	claims, err := sharedMW.RequireClaims(c)
	if err != nil {
		return err
	}
	from, to := dateRange(c)
	is, err := h.svc.IncomeStatement(c.Context(), claims.OrgID, from, to)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, is)
}

// dateRange reads ?from=&to= (YYYY-MM-DD); defaults to the current month.
func dateRange(c *fiber.Ctx) (time.Time, time.Time) {
	now := time.Now()
	defFrom := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	from := queryDate(c, "from", defFrom)
	to := queryDate(c, "to", now)
	return from, to
}

func queryDate(c *fiber.Ctx, key string, fallback time.Time) time.Time {
	v := c.Query(key)
	if v == "" {
		return fallback
	}
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t
	}
	return fallback
}
