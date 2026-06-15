package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/inventory/model"
)

// GetMember returns one member's equipment record (for QR rendering).
func (s *InventoryService) GetMember(ctx context.Context, orgID, memberID string) (*model.MemberEquipment, error) {
	return s.repo.GetByMemberID(ctx, orgID, memberID)
}

// Scan records a QR handover scan: resolve the member by token, mark the
// checkpoint, and return the updated record.
func (s *InventoryService) Scan(ctx context.Context, orgID string, scannedBy *uuid.UUID, req model.ScanRequest) (*model.MemberEquipment, error) {
	if req.Token == "" {
		return nil, fmt.Errorf("token is required")
	}
	if req.Checkpoint != "equipment" && req.Checkpoint != "luggage" {
		return nil, fmt.Errorf("checkpoint must be equipment or luggage")
	}
	m, err := s.repo.GetByToken(ctx, orgID, req.Token)
	if err != nil {
		return nil, err
	}
	if err := s.repo.RecordScanTx(ctx, m, req.Checkpoint, req.Items, scannedBy); err != nil {
		return nil, err
	}
	return s.repo.GetByMemberID(ctx, orgID, m.MemberID)
}

// ListCheckpoints returns per-member handover progress for a package.
func (s *InventoryService) ListCheckpoints(ctx context.Context, orgID, packageID string) ([]model.CheckpointMember, error) {
	members, err := s.repo.ListByPackage(ctx, orgID, packageID)
	if err != nil {
		return nil, err
	}
	out := make([]model.CheckpointMember, 0, len(members))
	for _, m := range members {
		out = append(out, model.CheckpointMember{
			MemberID:            m.MemberID,
			Nama:                m.Nama,
			HandoverToken:       m.HandoverToken,
			IsEquipmentReceived: m.IsEquipmentReceived,
			IsLuggageChecked:    m.IsLuggageChecked,
		})
	}
	return out, nil
}
