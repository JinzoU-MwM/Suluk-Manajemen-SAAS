package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/jamaah-in/v2/internal/jamaah/model"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

// ListCRM returns the paginated CRM list (profiles + latest registration +
// balance + lead score), filterable by stage/temperature/min-score and sortable
// by score.
func (h *JamaahHandler) ListCRM(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 25)
	authToken := c.Get("Authorization")

	f := model.CRMFilter{
		Search:   c.Query("search"),
		Stage:    c.Query("stage"),
		Temp:     c.Query("temp"),
		MinScore: c.QueryInt("min_score", 0),
		Sort:     c.Query("sort"),
	}

	rows, total, err := h.svc.ListCRM(c.Context(), claims.OrgID, authToken, f, page, pageSize)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Paginated(c, rows, int64(total), page, pageSize)
}

// GetPipelineFunnel returns CRM funnel analytics (per-stage counts/value/avg
// time + lead-source breakdown).
func (h *JamaahHandler) GetPipelineFunnel(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	funnel, err := h.svc.GetPipelineFunnel(c.Context(), claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, funnel)
}

// RecomputeScores triggers a full org-wide lead-score recompute (admin tool /
// backfill after the scoring rules change).
func (h *JamaahHandler) RecomputeScores(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	if err := h.svc.RecomputeOrgActive(c.Context(), claims.OrgID); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"status": "ok"})
}
