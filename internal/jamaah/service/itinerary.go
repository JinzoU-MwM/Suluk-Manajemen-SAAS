package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/jamaah/model"
)

func (s *JamaahService) ListItineraries(ctx context.Context, orgID, groupID uuid.UUID) ([]model.Itinerary, error) {
	items, err := s.repo.ListItineraries(ctx, groupID, orgID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []model.Itinerary{}
	}
	return items, nil
}

func (s *JamaahService) CreateItinerary(ctx context.Context, orgID, groupID uuid.UUID, req model.CreateItineraryRequest) (*model.Itinerary, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	it := &model.Itinerary{
		ID:          uuid.New(),
		OrgID:       orgID,
		GroupID:     groupID,
		DayNumber:   req.DayNumber,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		SortOrder:   req.SortOrder,
	}
	if it.DayNumber < 1 {
		it.DayNumber = 1
	}
	if err := s.repo.CreateItinerary(ctx, it); err != nil {
		return nil, err
	}
	return it, nil
}

func (s *JamaahService) UpdateItinerary(ctx context.Context, orgID, itemID uuid.UUID, req model.UpdateItineraryRequest) (*model.Itinerary, error) {
	it, err := s.repo.GetItinerary(ctx, itemID, orgID)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, fmt.Errorf("itinerary not found")
	}
	if req.DayNumber != nil {
		it.DayNumber = *req.DayNumber
	}
	if req.Title != nil {
		it.Title = *req.Title
	}
	if req.Description != nil {
		it.Description = *req.Description
	}
	if req.Location != nil {
		it.Location = *req.Location
	}
	if req.StartTime != nil {
		it.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		it.EndTime = req.EndTime
	}
	if req.SortOrder != nil {
		it.SortOrder = *req.SortOrder
	}
	if err := s.repo.UpdateItinerary(ctx, it); err != nil {
		return nil, err
	}
	return it, nil
}

func (s *JamaahService) DeleteItinerary(ctx context.Context, orgID, itemID uuid.UUID) error {
	return s.repo.DeleteItinerary(ctx, itemID, orgID)
}
