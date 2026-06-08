package service

import (
	"context"

	"github.com/jamaah-in/v2/internal/jamaah/repository"
)

func (s *JamaahService) GetAnalyticsData(ctx context.Context, orgID string,
	totalJamaah, jamaahThisMonth, maleCount, femaleCount, unknownCount *int, equipmentRate *float64) {

	s.repo.GetAnalyticsStats(ctx, orgID, totalJamaah, jamaahThisMonth, maleCount, femaleCount, unknownCount, equipmentRate)
}

func (s *JamaahService) GetPassportExpiringSoon(ctx context.Context, orgID string) int {
	return s.repo.GetPassportExpiringSoon(ctx, orgID)
}

func (s *JamaahService) GetMonthlyTrend(ctx context.Context, orgID string) []repository.MonthlyTrend {
	trends := s.repo.GetMonthlyTrend(ctx, orgID)
	if trends == nil {
		return []repository.MonthlyTrend{}
	}
	return trends
}

func (s *JamaahService) GetTotalGroups(ctx context.Context, orgID string) int {
	return s.repo.GetTotalGroups(ctx, orgID)
}

func (s *JamaahService) GetRecentGroups(ctx context.Context, orgID string) []repository.RecentGroup {
	groups := s.repo.GetRecentGroups(ctx, orgID, 5)
	if groups == nil {
		return []repository.RecentGroup{}
	}
	return groups
}
