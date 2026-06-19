package service

import (
	"context"
	"encoding/json"

	"github.com/jamaah-in/v2/internal/inventory/model"
	"github.com/jamaah-in/v2/internal/shared/events"
)

func (s *InventoryService) ListStockItems(ctx context.Context, orgID string) ([]model.StockItem, error) {
	return s.repo.ListStockItems(ctx, orgID)
}
func (s *InventoryService) CreateStockItem(ctx context.Context, orgID string, req model.CreateItemRequest) (model.StockItem, error) {
	return s.repo.CreateStockItem(ctx, orgID, req)
}
func (s *InventoryService) UpdateStockItem(ctx context.Context, orgID, itemID string, req model.UpdateItemRequest) error {
	return s.repo.UpdateStockItem(ctx, orgID, itemID, req)
}
func (s *InventoryService) RestockItem(ctx context.Context, orgID, itemID string, qty int, note, userID string) error {
	return s.repo.RestockItem(ctx, orgID, itemID, qty, note, userID)
}
func (s *InventoryService) AdjustItem(ctx context.Context, orgID, itemID string, delta int, note, userID string) error {
	return s.repo.AdjustItem(ctx, orgID, itemID, delta, note, userID)
}
func (s *InventoryService) ListMovements(ctx context.Context, orgID, itemID string, limit int) ([]model.StockMovement, error) {
	return s.repo.ListMovements(ctx, orgID, itemID, limit)
}
func (s *InventoryService) DeleteStockItem(ctx context.Context, orgID, itemID string) error {
	return s.repo.DeleteStockItem(ctx, orgID, itemID)
}
func (s *InventoryService) GetPackageKit(ctx context.Context, orgID, packageID string) ([]model.PackageKitItem, error) {
	return s.repo.GetPackageKit(ctx, orgID, packageID)
}
func (s *InventoryService) SetPackageKit(ctx context.Context, orgID, packageID string, items []model.KitLine) error {
	return s.repo.SetPackageKit(ctx, orgID, packageID, items)
}

// DeductForDeparture deducts the package kit × member_count when a group departs.
// Idempotent (repo-level). No kit / no members ⇒ no-op.
func (s *InventoryService) DeductForDeparture(ctx context.Context, env *events.Envelope) error {
	var p model.DepartedPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return err
	}
	if p.PackageID == "" || p.MemberCount < 1 || env.OrgID == "" {
		return nil
	}
	kit, err := s.repo.GetPackageKit(ctx, env.OrgID, p.PackageID)
	if err != nil {
		return err
	}
	deductions := ComputeDeductions(kit, p.MemberCount)
	if len(deductions) == 0 {
		return nil
	}
	return s.repo.ApplyDepartureDeduction(ctx, env.OrgID, p.GroupID, p.PackageID, deductions)
}

// StartConsumer subscribes inventory-service to the bus and deducts on departure.
// Non-group.departed events are ACKed and ignored.
func (s *InventoryService) StartConsumer(ctx context.Context, bus *events.Bus) error {
	_, err := bus.Subscribe(ctx, "inventory-deduct", func(ctx context.Context, env *events.Envelope) error {
		if env.EventType != events.EventGroupDeparted {
			return nil
		}
		return s.DeductForDeparture(ctx, env)
	})
	return err
}
