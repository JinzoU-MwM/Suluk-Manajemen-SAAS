package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/mutawwif/model"
	"github.com/jamaah-in/v2/internal/mutawwif/repository"
)

type MutawwifService struct {
	repo *repository.MutawwifRepo
}

func NewMutawwifService(repo *repository.MutawwifRepo) *MutawwifService {
	return &MutawwifService{repo: repo}
}

func parseDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	for _, f := range []string{"2006-01-02", "2006-01-02T15:04:05Z", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(f, s); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("invalid date: %s", s)
}

func (s *MutawwifService) CreateGuide(ctx context.Context, orgID uuid.UUID, req model.CreateGuideRequest) (*model.Guide, error) {
	g := &model.Guide{
		OrgID:     orgID,
		Name:      req.Name,
		Phone:     req.Phone,
		Email:     req.Email,
		Type:      req.Type,
		LicenseNo: req.LicenseNo,
		Notes:     req.Notes,
	}
	if g.Type == "" {
		g.Type = "mutawwif"
	}
	exp, err := parseDate(req.LicenseExpiry)
	if err != nil {
		return nil, fmt.Errorf("license_expiry: %w", err)
	}
	g.LicenseExpiry = exp
	if err := s.repo.CreateGuide(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *MutawwifService) ListGuides(ctx context.Context, orgID uuid.UUID, search string) (*model.GuideListResponse, error) {
	guides, err := s.repo.ListGuides(ctx, orgID, search)
	if err != nil {
		return nil, err
	}
	return &model.GuideListResponse{Guides: guides, Total: len(guides)}, nil
}

func (s *MutawwifService) GetGuide(ctx context.Context, id, orgID uuid.UUID) (*model.Guide, error) {
	return s.repo.GetGuide(ctx, id, orgID)
}

func (s *MutawwifService) UpdateGuide(ctx context.Context, id, orgID uuid.UUID, req model.UpdateGuideRequest) (*model.Guide, error) {
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.LicenseNo != nil {
		updates["license_no"] = *req.LicenseNo
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if req.LicenseExpiry != nil {
		exp, err := parseDate(*req.LicenseExpiry)
		if err != nil {
			return nil, fmt.Errorf("license_expiry: %w", err)
		}
		updates["license_expiry"] = exp
	}
	if err := s.repo.UpdateGuide(ctx, id, orgID, updates); err != nil {
		return nil, err
	}
	return s.repo.GetGuide(ctx, id, orgID)
}

func (s *MutawwifService) DeleteGuide(ctx context.Context, id, orgID uuid.UUID) error {
	return s.repo.DeleteGuide(ctx, id, orgID)
}

func (s *MutawwifService) Assign(ctx context.Context, orgID uuid.UUID, req model.AssignGuideRequest) (*model.GuideAssignment, error) {
	guideID, err := uuid.Parse(req.GuideID)
	if err != nil {
		return nil, fmt.Errorf("invalid guide_id")
	}
	groupID, err := uuid.Parse(req.GroupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group_id")
	}
	// The guide must belong to this org (the group is cross-service, trusted).
	if _, err := s.repo.GetGuide(ctx, guideID, orgID); err != nil {
		return nil, err
	}
	role := req.Role
	if role == "" {
		role = "leader"
	}
	a := &model.GuideAssignment{OrgID: orgID, GuideID: guideID, GroupID: groupID, Role: role}
	if err := s.repo.Assign(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *MutawwifService) Unassign(ctx context.Context, orgID, guideID, groupID uuid.UUID) error {
	return s.repo.Unassign(ctx, orgID, guideID, groupID)
}

func (s *MutawwifService) ListByGroup(ctx context.Context, orgID, groupID uuid.UUID) ([]model.GuideAssignment, error) {
	return s.repo.ListByGroup(ctx, orgID, groupID)
}

func (s *MutawwifService) ListByGuide(ctx context.Context, orgID, guideID uuid.UUID) ([]model.GuideAssignment, error) {
	return s.repo.ListByGuide(ctx, orgID, guideID)
}

func (s *MutawwifService) GuidesExpiringLicense(ctx context.Context, orgID uuid.UUID, withinDays int) ([]model.Guide, error) {
	return s.repo.GuidesExpiringLicense(ctx, orgID, withinDays)
}
