package service

import (
	"context"

	"github.com/jamaah-in/v2/internal/inventory/model"
	"github.com/jamaah-in/v2/internal/inventory/repository"
)

type InventoryService struct {
	repo *repository.InventoryRepo
}

func NewInventoryService(repo *repository.InventoryRepo) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) SyncMembers(ctx context.Context, orgID string, req model.SyncMembersRequest) error {
	return s.repo.UpsertMembers(ctx, orgID, req)
}

func (s *InventoryService) GetForecast(ctx context.Context, orgID, packageID string) (*model.ForecastResponse, error) {
	members, err := s.repo.ListByPackage(ctx, orgID, packageID)
	if err != nil {
		return nil, err
	}

	resp := &model.ForecastResponse{
		Requirements:  make(map[string]int),
		SizeBreakdown: make(map[string]int),
		Details:       make([]model.MemberDetail, 0, len(members)),
	}

	for _, m := range members {
		resp.TotalMembers++
		if m.BajuSize != "" {
			resp.SizeBreakdown[m.BajuSize]++
		} else {
			resp.SizeBreakdown["N/A"]++
		}

		switch m.Gender {
		case "male":
			resp.Requirements["ihram"]++
		case "female":
			resp.Requirements["mukena"]++
		}
		resp.Requirements["koper"]++
		resp.Requirements["baju"]++

		resp.Details = append(resp.Details, model.MemberDetail{
			MemberID:            m.MemberID,
			Nama:                m.Nama,
			Gender:              m.Gender,
			BajuSize:            m.BajuSize,
			FamilyID:            m.FamilyID,
			IsEquipmentReceived: m.IsEquipmentReceived,
		})
	}

	return resp, nil
}

func (s *InventoryService) GetFulfillment(ctx context.Context, orgID, packageID string) (*model.FulfillmentResponse, error) {
	received, err := s.repo.ListReceived(ctx, orgID, packageID)
	if err != nil {
		return nil, err
	}

	pending, err := s.repo.ListPending(ctx, orgID, packageID)
	if err != nil {
		return nil, err
	}

	resp := &model.FulfillmentResponse{
		ReceivedCount: len(received),
		PendingCount:  len(pending),
		Received:      make([]model.MemberBrief, len(received)),
		Pending:       make([]model.MemberBrief, len(pending)),
	}

	for i, m := range received {
		resp.Received[i] = model.MemberBrief{
			ID:                  m.ID,
			Nama:                m.Nama,
			IsEquipmentReceived: true,
		}
	}
	for i, m := range pending {
		resp.Pending[i] = model.MemberBrief{
			ID:                  m.ID,
			Nama:                m.Nama,
			IsEquipmentReceived: false,
		}
	}

	return resp, nil
}

func (s *InventoryService) MarkReceived(ctx context.Context, orgID, packageID string, req model.MarkReceivedRequest) (int64, error) {
	return s.repo.MarkReceived(ctx, orgID, packageID, req.MemberIDs, req.ItemsReceived)
}

func (s *InventoryService) UpdateOperational(ctx context.Context, orgID, memberID string, req model.UpdateOperationalRequest) error {
	return s.repo.UpdateOperational(ctx, orgID, memberID, req.BajuSize, req.FamilyID)
}
