package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MonthlyTrend struct {
	Month string `json:"month"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type RecentGroup struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetAnalyticsStats fills the core jamaah aggregates in a single round-trip.
// equipmentRate is the share of jamaah (0-100) that have their seragam/baju size
// recorded — the only equipment field captured on the profile.
func (r *JamaahRepo) GetAnalyticsStats(ctx context.Context, orgID string,
	totalJamaah, jamaahThisMonth, maleCount, femaleCount, unknownCount *int, equipmentRate *float64) error {

	return r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())),
			COUNT(*) FILTER (WHERE gender = 'L'),
			COUNT(*) FILTER (WHERE gender = 'P'),
			COUNT(*) FILTER (WHERE gender NOT IN ('L','P')),
			COALESCE(ROUND(100.0 * COUNT(*) FILTER (WHERE COALESCE(baju_size, '') <> '') / NULLIF(COUNT(*), 0)), 0)::float8
		FROM jamaah_profiles WHERE org_id = $1`, orgID).
		Scan(totalJamaah, jamaahThisMonth, maleCount, femaleCount, unknownCount, equipmentRate)
}

// GetDocumentRate returns the share (0-100) of document records that have been
// received or completed (status diterima/selesai). Non-critical: returns 0 on error.
func (r *JamaahRepo) GetDocumentRate(ctx context.Context, orgID string) float64 {
	var rate float64
	r.pool.QueryRow(ctx, `
		SELECT COALESCE(ROUND(100.0 * COUNT(*) FILTER (WHERE status IN ('diterima','selesai')) / NULLIF(COUNT(*), 0)), 0)::float8
		FROM jamaah_documents WHERE org_id = $1`, orgID).Scan(&rate)
	return rate
}

func (r *JamaahRepo) GetPassportExpiringSoon(ctx context.Context, orgID string) int {
	var count int
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jamaah_profiles WHERE org_id = $1 AND tanggal_paspor IS NOT NULL AND tanggal_paspor <= NOW() + INTERVAL '90 days' AND tanggal_paspor > NOW()`, orgID).Scan(&count)
	return count
}

var idMonthAbbr = [12]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

// GetMonthlyTrend returns the registration count for each of the last 6 calendar
// months (oldest first), zero-filling months with no registrations so the chart
// always renders a consistent 6-bar window with human-readable labels.
func (r *JamaahRepo) GetMonthlyTrend(ctx context.Context, orgID string) []MonthlyTrend {
	counts := map[string]int{}
	rows, err := r.pool.Query(ctx, `
		SELECT TO_CHAR(DATE_TRUNC('month', created_at), 'YYYY-MM') as month, COUNT(*) as count
		FROM jamaah_profiles
		WHERE org_id = $1 AND created_at >= DATE_TRUNC('month', NOW()) - INTERVAL '5 months'
		GROUP BY 1
	`, orgID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var month string
			var count int
			if scanErr := rows.Scan(&month, &count); scanErr == nil {
				counts[month] = count
			}
		}
	}

	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	trends := make([]MonthlyTrend, 0, 6)
	for i := 5; i >= 0; i-- {
		m := firstOfMonth.AddDate(0, -i, 0)
		key := m.Format("2006-01")
		trends = append(trends, MonthlyTrend{
			Month: key,
			Label: idMonthAbbr[int(m.Month())-1],
			Count: counts[key],
		})
	}
	return trends
}

func (r *JamaahRepo) GetTotalGroups(ctx context.Context, orgID string) int {
	var count int
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM groups WHERE org_id = $1`, orgID).Scan(&count)
	return count
}

func (r *JamaahRepo) GetRecentGroups(ctx context.Context, orgID string, limit int) []RecentGroup {
	rows, err := r.pool.Query(ctx, `
		SELECT g.id, g.name, COALESCE(COUNT(gm.id), 0) as member_count, g.created_at
		FROM groups g
		LEFT JOIN group_members gm ON gm.group_id = g.id
		WHERE g.org_id = $1
		GROUP BY g.id, g.name, g.created_at
		ORDER BY g.created_at DESC
		LIMIT $2
	`, orgID, limit)
	if err != nil {
		return []RecentGroup{}
	}
	defer rows.Close()

	var groups []RecentGroup
	for rows.Next() {
		var g RecentGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.MemberCount, &g.CreatedAt); err == nil {
			groups = append(groups, g)
		}
	}
	return groups
}
