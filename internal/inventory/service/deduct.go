package service

import "github.com/jamaah-in/v2/internal/inventory/model"

// ComputeDeductions returns the (item, qty) lines to subtract when a group of
// memberCount jamaah departs. Pure: no DB. Empty when there is nothing to do.
func ComputeDeductions(kit []model.PackageKitItem, memberCount int) []model.Deduction {
	if memberCount < 1 || len(kit) == 0 {
		return nil
	}
	out := make([]model.Deduction, 0, len(kit))
	for _, k := range kit {
		qty := k.QtyPerJamaah * memberCount
		if qty <= 0 {
			continue
		}
		out = append(out, model.Deduction{ItemID: k.ItemID, Qty: qty})
	}
	return out
}
