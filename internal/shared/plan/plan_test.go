package plan

import "testing"

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"free":       Gratis,
		"FREE":       Gratis,
		"business":   Bisnis,
		"":           Gratis,
		"bogus":      Gratis,
		"pro":        Pro,
		" starter ":  Starter,
		"enterprise": Enterprise,
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRankOrdering(t *testing.T) {
	if !(Rank(Gratis) < Rank(Starter) && Rank(Starter) < Rank(Pro) &&
		Rank(Pro) < Rank(Bisnis) && Rank(Bisnis) < Rank(Enterprise)) {
		t.Fatal("ranks are not strictly increasing gratis<starter<pro<bisnis<enterprise")
	}
	// legacy aliases must rank with their modern equivalents
	if Rank("free") != Rank(Gratis) {
		t.Error("legacy free should rank as gratis")
	}
	if Rank("business") != Rank(Bisnis) {
		t.Error("legacy business should rank as bisnis")
	}
}

func TestIsProOrHigher(t *testing.T) {
	for _, p := range []string{Gratis, Starter} {
		if IsProOrHigher(p) {
			t.Errorf("%s should NOT be pro-or-higher", p)
		}
	}
	for _, p := range []string{Pro, Bisnis, Enterprise} {
		if !IsProOrHigher(p) {
			t.Errorf("%s should be pro-or-higher", p)
		}
	}
}

func TestPriceFor(t *testing.T) {
	cases := []struct {
		plan, period string
		want         int64
		wantErr      bool
	}{
		{Pro, "monthly", 299000, false},
		{Pro, "yearly", 2990000, false},
		{Pro, "annual", 2990000, false},
		{Pro, "", 299000, false},
		{Starter, "monthly", 149000, false},
		{Bisnis, "yearly", 5990000, false},
		{Gratis, "monthly", 0, true},     // not purchasable
		{Enterprise, "monthly", 0, true}, // not purchasable
		{Pro, "weekly", 0, true},         // bad period
		{"bogus", "monthly", 0, true},    // unknown plan
	}
	for _, c := range cases {
		got, err := PriceFor(c.plan, c.period)
		if c.wantErr {
			if err == nil {
				t.Errorf("PriceFor(%q,%q) expected error, got %d", c.plan, c.period, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("PriceFor(%q,%q) unexpected error: %v", c.plan, c.period, err)
		}
		if got != c.want {
			t.Errorf("PriceFor(%q,%q) = %d, want %d", c.plan, c.period, got, c.want)
		}
	}
}

func TestNormalizePeriod(t *testing.T) {
	cases := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{"monthly", PeriodMonthly, true},
		{"month", PeriodMonthly, true},
		{"", PeriodMonthly, true},
		{" Monthly ", PeriodMonthly, true},
		{"yearly", PeriodYearly, true},
		{"annual", PeriodYearly, true},
		{"year", PeriodYearly, true},
		{"YEARLY", PeriodYearly, true},
		{"weekly", "", false},
		{"bogus", "", false},
	}
	for _, c := range cases {
		got, ok := NormalizePeriod(c.in)
		if ok != c.wantOK || got != c.want {
			t.Errorf("NormalizePeriod(%q) = (%q,%v), want (%q,%v)", c.in, got, ok, c.want, c.wantOK)
		}
	}
}

func TestAtLeast(t *testing.T) {
	if !AtLeast(Bisnis, Pro) {
		t.Error("bisnis should be at least pro")
	}
	if AtLeast(Starter, Pro) {
		t.Error("starter should not be at least pro")
	}
}
