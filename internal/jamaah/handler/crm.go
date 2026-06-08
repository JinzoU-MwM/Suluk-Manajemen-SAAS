package handler

import (
	"github.com/gofiber/fiber/v2"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

// ListCRM returns the paginated CRM list (profiles + latest registration + balance).
func (h *JamaahHandler) ListCRM(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 25)
	search := c.Query("search")
	authToken := c.Get("Authorization")

	rows, total, err := h.svc.ListCRM(c.Context(), claims.OrgID, authToken, search, page, pageSize)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Paginated(c, rows, int64(total), page, pageSize)
}
