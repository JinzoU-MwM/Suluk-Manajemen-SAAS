package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/jamaah/model"
	"github.com/jamaah-in/v2/internal/jamaah/repository"
)

func (s *JamaahService) CreateGroup(ctx context.Context, orgID uuid.UUID, authToken string, req model.CreateGroupRequest) (*model.Group, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	lim := s.fetchLimits(ctx, orgID, authToken)
	g := &model.Group{
		ID:          uuid.New(),
		OrgID:       orgID,
		Name:        req.Name,
		Description: req.Description,
	}
	// Atomic, race-safe cap enforcement (per-org advisory lock + count + insert
	// in one transaction).
	if err := s.repo.CreateGroupTx(ctx, g, lim.MaxGroups); err != nil {
		if errors.Is(err, repository.ErrLimitReached) {
			return nil, fmt.Errorf("%w: batas grup pada paket Anda (%d) telah tercapai. Upgrade paket untuk menambah grup", ErrPlanLimit, lim.MaxGroups)
		}
		return nil, err
	}
	return g, nil
}

func (s *JamaahService) ListGroups(ctx context.Context, orgID uuid.UUID) ([]model.Group, error) {
	groups, err := s.repo.ListGroups(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if groups == nil {
		return []model.Group{}, nil
	}
	return groups, nil
}

func (s *JamaahService) GetGroup(ctx context.Context, groupID, orgID uuid.UUID) (*model.Group, error) {
	return s.repo.GetGroup(ctx, groupID, orgID)
}

func (s *JamaahService) UpdateGroup(ctx context.Context, groupID, orgID uuid.UUID, req model.UpdateGroupRequest) (*model.Group, error) {
	g, err := s.repo.GetGroup(ctx, groupID, orgID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, fmt.Errorf("group not found")
	}
	if req.Name != nil {
		g.Name = *req.Name
	}
	if req.Description != nil {
		g.Description = *req.Description
	}
	if err := s.repo.UpdateGroup(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *JamaahService) DeleteGroup(ctx context.Context, groupID, orgID uuid.UUID) error {
	return s.repo.DeleteGroup(ctx, groupID, orgID)
}

func (s *JamaahService) AddGroupMembers(ctx context.Context, groupID, orgID uuid.UUID, members []model.AddGroupMemberRequest) (int, error) {
	g, err := s.repo.GetGroup(ctx, groupID, orgID)
	if err != nil || g == nil {
		return 0, fmt.Errorf("group not found")
	}
	added := 0
	for _, m := range members {
		gm := &model.GroupMember{
			ID:       uuid.New(),
			OrgID:    orgID,
			GroupID:  groupID,
			MemberID: m.MemberID,
			Name:     m.Name,
			Phone:    m.Phone,
			Notes:    m.Notes,
		}
		if err := s.repo.AddGroupMember(ctx, gm); err == nil {
			added++
		}
	}
	return added, nil
}

func (s *JamaahService) ListGroupMembers(ctx context.Context, groupID, orgID uuid.UUID) ([]model.GroupMember, error) {
	g, err := s.repo.GetGroup(ctx, groupID, orgID)
	if err != nil || g == nil {
		return nil, fmt.Errorf("group not found")
	}
	members, err := s.repo.ListGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if members == nil {
		return []model.GroupMember{}, nil
	}
	return members, nil
}

func (s *JamaahService) UpdateGroupMember(ctx context.Context, groupID, memberID, orgID uuid.UUID, req model.UpdateGroupMemberRequest) error {
	g, err := s.repo.GetGroup(ctx, groupID, orgID)
	if err != nil || g == nil {
		return fmt.Errorf("group not found")
	}
	name, phone, notes := "", "", ""
	if req.Name != nil {
		name = *req.Name
	}
	if req.Phone != nil {
		phone = *req.Phone
	}
	if req.Notes != nil {
		notes = *req.Notes
	}
	return s.repo.UpdateGroupMember(ctx, groupID, memberID, name, phone, notes)
}

func (s *JamaahService) DeleteGroupMember(ctx context.Context, groupID, memberID, orgID uuid.UUID) error {
	g, err := s.repo.GetGroup(ctx, groupID, orgID)
	if err != nil || g == nil {
		return fmt.Errorf("group not found")
	}
	return s.repo.DeleteGroupMember(ctx, groupID, memberID)
}
