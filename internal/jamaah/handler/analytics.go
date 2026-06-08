package handler

import (
	"github.com/gofiber/fiber/v2"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
)

func (h *JamaahHandler) GetAnalyticsDashboard(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	orgID := claims.OrgID.String()

	var totalJamaah, jamaahThisMonth, maleCount, femaleCount, unknownCount int
	var equipmentRate float64

	h.svc.GetAnalyticsData(c.Context(), orgID,
		&totalJamaah, &jamaahThisMonth, &maleCount, &femaleCount, &unknownCount, &equipmentRate)

	passportExpiring := h.svc.GetPassportExpiringSoon(c.Context(), orgID)
	monthlyTrend := h.svc.GetMonthlyTrend(c.Context(), orgID)
	totalGroups := h.svc.GetTotalGroups(c.Context(), orgID)
	recentGroups := h.svc.GetRecentGroups(c.Context(), orgID)

	resp := fiber.Map{
		"total_jamaah":      totalJamaah,
		"jamaah_this_month": jamaahThisMonth,
		"equipment_rate":    equipmentRate,
		"gender_breakdown": fiber.Map{
			"male":    maleCount,
			"female":  femaleCount,
			"unknown": unknownCount,
		},
		"passport_expiring_soon": passportExpiring,
		"monthly_trend":          monthlyTrend,
		"total_groups":           totalGroups,
		"recent_groups":          recentGroups,
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}
