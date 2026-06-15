package service

import "testing"

func TestSaleBase(t *testing.T) {
	cases := []struct {
		name   string
		amount int64
		rate   float64
		want   int64
	}{
		{"5pct of 10jt", 500_000, 5.0, 10_000_000},
		{"2.5pct", 250_000, 2.5, 10_000_000},
		{"zero rate falls back to amount", 500_000, 0, 500_000},
		{"negative rate falls back", 500_000, -1, 500_000},
		{"rounds to nearest", 333_333, 3.0, 11_111_100},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := saleBase(c.amount, c.rate); got != c.want {
				t.Fatalf("saleBase(%d, %v) = %d, want %d", c.amount, c.rate, got, c.want)
			}
		})
	}
}

func TestTieredAmount(t *testing.T) {
	cases := []struct {
		name string
		base int64
		rate float64
		want int64
	}{
		{"tier2 2pct", 10_000_000, 2.0, 200_000},
		{"tier3 1pct", 10_000_000, 1.0, 100_000},
		{"zero base", 0, 2.0, 0},
		{"zero rate", 10_000_000, 0, 0},
		{"rounds", 11_111_100, 1.0, 111_111},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := tieredAmount(c.base, c.rate); got != c.want {
				t.Fatalf("tieredAmount(%d, %v) = %d, want %d", c.base, c.rate, got, c.want)
			}
		})
	}
}

// A full berjenjang payout: 5% seller on a 10jt sale, with default L2=2% L3=1%.
// Seller 500k, tier2 200k, tier3 100k — each a distinct override on the same base.
func TestCascadeMathEndToEnd(t *testing.T) {
	sellerAmount := int64(500_000)
	sellerRate := 5.0
	base := saleBase(sellerAmount, sellerRate)
	if base != 10_000_000 {
		t.Fatalf("base = %d, want 10jt", base)
	}
	tier2 := tieredAmount(base, 2.0)
	tier3 := tieredAmount(base, 1.0)
	if tier2 != 200_000 || tier3 != 100_000 {
		t.Fatalf("tier2=%d tier3=%d, want 200000/100000", tier2, tier3)
	}
}
