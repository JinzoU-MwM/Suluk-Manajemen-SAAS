package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MonthlyTrend struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type RecentGroup struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	MemberCount int     `json:"member_count"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *JamaahRepo) GetAnalyticsStats(ctx context.Context, orgID string,
	totalJamaah, jamaahThisMonth, maleCount, femaleCount, unknownCount *int, equipmentRate *float64) {

	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jamaah_profiles WHERE org_id = $1`, orgID).Scan(totalJamaah)
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jamaah_profiles WHERE org_id = $1 AND DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())`, orgID).Scan(jamaahThisMonth)
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jamaah_profiles WHERE org_id = $1 AND gender = 'L'`, orgID).Scan(maleCount)
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jamaah_profiles WHERE org_id = $1 AND gender = 'P'`, orgID).Scan(femaleCount)
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jamaah_profiles WHERE org_id = $1 AND gender NOT IN ('L','P')`, orgID).Scan(unknownCount)
	*equipmentRate = 0.0
}

func (r *JamaahRepo) GetPassportExpiringSoon(ctx context.Context, orgID string) int {
	var count int
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jamaah_profiles WHERE org_id = $1 AND passport_expiry IS NOT NULL AND passport_expiry <= NOW() + INTERVAL '90 days' AND passport_expiry > NOW()`, orgID).Scan(&count)
	return count
}

func (r *JamaahRepo) GetMonthlyTrend(ctx context.Context, orgID string) []MonthlyTrend {
	rows, err := r.pool.Query(ctx, `
		SELECT TO_CHAR(DATE_TRUNC('month', created_at), 'YYYY-MM') as month, COUNT(*) as count
		FROM jamaah_profiles
		WHERE org_id = $1 AND created_at >= NOW() - INTERVAL '12 months'
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month
	`, orgID)
	if err != nil {
		return []MonthlyTrend{}
	}
	defer rows.Close()

	var trends []MonthlyTrend
	for rows.Next() {
		var t MonthlyTrend
		if err := rows.Scan(&t.Month, &t.Count); err == nil {
			trends = append(trends, t)
		}
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
