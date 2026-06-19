package service

import (
	"testing"

	"github.com/jamaah-in/v2/internal/inventory/model"
)

func TestComputeDeductionsMultipliesByHeadcount(t *testing.T) {
	kit := []model.PackageKitItem{
		{ItemID: "koper", QtyPerJamaah: 1},
		{ItemID: "ihram", QtyPerJamaah: 2},
	}
	got := ComputeDeductions(kit, 30)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ItemID != "koper" || got[0].Qty != 30 {
		t.Fatalf("koper = %+v, want qty 30", got[0])
	}
	if got[1].ItemID != "ihram" || got[1].Qty != 60 {
		t.Fatalf("ihram = %+v, want qty 60", got[1])
	}
}

func TestComputeDeductionsZeroMembersIsEmpty(t *testing.T) {
	kit := []model.PackageKitItem{{ItemID: "koper", QtyPerJamaah: 1}}
	if got := ComputeDeductions(kit, 0); len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestComputeDeductionsEmptyKitIsEmpty(t *testing.T) {
	if got := ComputeDeductions(nil, 30); len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}
