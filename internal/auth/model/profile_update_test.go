package model

import "testing"

func sp(s string) *string { return &s }
func bp(b bool) *bool     { return &b }

func TestApplyProfileUpdatePartial(t *testing.T) {
	u := &User{Name: "Old", AvatarColor: "blue", NotifyUsageLimit: true, NotifyExpiry: true}
	if err := ApplyProfileUpdate(u, ProfileUpdate{Name: sp("New"), City: sp("Bandung")}); err != nil {
		t.Fatalf("err = %v", err)
	}
	if u.Name != "New" {
		t.Fatalf("name = %q, want New", u.Name)
	}
	if u.City == nil || *u.City != "Bandung" {
		t.Fatalf("city = %v, want Bandung", u.City)
	}
	if u.AvatarColor != "blue" { // untouched
		t.Fatalf("avatar = %q, want blue (unchanged)", u.AvatarColor)
	}
	if u.Phone != nil {
		t.Fatalf("phone = %v, want nil (untouched)", u.Phone)
	}
	if u.Bio != nil {
		t.Fatalf("bio = %v, want nil (untouched)", u.Bio)
	}
}

func TestApplyProfileUpdateRejectsEmptyName(t *testing.T) {
	u := &User{Name: "Old"}
	if err := ApplyProfileUpdate(u, ProfileUpdate{Name: sp("  ")}); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestApplyProfileUpdateClampsAvatarColor(t *testing.T) {
	u := &User{AvatarColor: "rose"}
	if err := ApplyProfileUpdate(u, ProfileUpdate{AvatarColor: sp("neon")}); err != nil {
		t.Fatalf("err = %v", err)
	}
	if u.AvatarColor != "blue" {
		t.Fatalf("avatar = %q, want blue (clamped)", u.AvatarColor)
	}

	u2 := &User{AvatarColor: "blue"}
	if err := ApplyProfileUpdate(u2, ProfileUpdate{AvatarColor: sp("emerald")}); err != nil {
		t.Fatalf("err = %v", err)
	}
	if u2.AvatarColor != "emerald" {
		t.Fatalf("avatar = %q, want emerald (valid color preserved)", u2.AvatarColor)
	}
}

func TestApplyProfileUpdateBoolPointers(t *testing.T) {
	u := &User{NotifyUsageLimit: true, NotifyExpiry: true}
	if err := ApplyProfileUpdate(u, ProfileUpdate{NotifyExpiry: bp(false)}); err != nil {
		t.Fatalf("err = %v", err)
	}
	if u.NotifyExpiry != false || u.NotifyUsageLimit != true {
		t.Fatalf("notify = (%v,%v), want (true,false)", u.NotifyUsageLimit, u.NotifyExpiry)
	}
}
