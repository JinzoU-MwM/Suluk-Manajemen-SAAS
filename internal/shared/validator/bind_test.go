package validator

import "testing"

type sample struct {
	Name   string `json:"name" validate:"required,min=2"`
	Scheme string `json:"scheme" validate:"required,oneof=a b c"`
	Qty    int    `json:"qty" validate:"min=1"`
}

func TestValidateStruct_ReportsFieldErrors(t *testing.T) {
	errs := ValidateStruct(&sample{Name: "x", Scheme: "z", Qty: 0})
	if len(errs) == 0 {
		t.Fatal("expected validation errors, got none")
	}
	got := map[string]bool{}
	for _, e := range errs {
		got[e.Field] = true
	}
	for _, want := range []string{"name", "scheme", "qty"} {
		if !got[want] {
			t.Errorf("expected an error for field %q; got %+v", want, errs)
		}
	}
}

func TestValidateStruct_Valid(t *testing.T) {
	if errs := ValidateStruct(&sample{Name: "Budi", Scheme: "a", Qty: 3}); errs != nil {
		t.Fatalf("expected no errors, got %+v", errs)
	}
}
