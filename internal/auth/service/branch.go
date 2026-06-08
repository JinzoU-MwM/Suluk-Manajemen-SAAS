package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (s *AuthService) CreateBranch(ctx context.Context, parentOrgID, name string) (map[string]interface{}, error) {
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	id, err := s.repo.CreateBranch(ctx, uuid.MustParse(parentOrgID), name, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}
	return map[string]interface{}{"id": id.String(), "name": name, "slug": slug}, nil
}

func (s *AuthService) ListBranches(ctx context.Context, parentOrgID uuid.UUID) ([]interface{}, error) {
	branches, err := s.repo.ListBranches(ctx, parentOrgID)
	if err != nil {
		return nil, err
	}
	var result []interface{}
	for _, b := range branches {
		result = append(result, b)
	}
	return result, nil
}

func (s *AuthService) GetConsolidatedStats(ctx context.Context, parentOrgID string) (map[string]interface{}, error) {
	return s.repo.GetConsolidatedStats(ctx, uuid.MustParse(parentOrgID))
}
