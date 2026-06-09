// Package plan is the single source of truth for subscription tiers: prices,
// per-tier limits, rank ordering, and validation. Both auth-service (status,
// activation, pricing) and invoice-service (payment order pricing) depend on it
// so a tier is defined in exactly one place on the backend.
package plan

import (
	"fmt"
	"strings"
)

// Tier identifiers (stored verbatim in subscriptions.plan).
const (
	Gratis     = "gratis"
	Starter    = "starter"
	Pro        = "pro"
	Bisnis     = "bisnis"
	Enterprise = "enterprise"
)

// Billing periods accepted by PriceFor.
const (
	PeriodMonthly = "monthly"
	PeriodYearly  = "yearly"
)

// Unlimited marks a limit field as having no cap.
const Unlimited = -1

// Tier describes one subscription plan. Prices are in IDR (no decimals).
// A limit of Unlimited (-1) means no cap. Annual price 0 on a purchasable tier
// would be invalid; only non-purchasable tiers (gratis/enterprise) carry 0.
type Tier struct {
	Key          string   `json:"key"`
	Name         string   `json:"name"`
	Rank         int      `json:"rank"`
	MonthlyPrice int64    `json:"monthly_price"`
	AnnualPrice  int64    `json:"annual_price"`
	MaxJamaah    int      `json:"max_jamaah"`
	MaxGroups    int      `json:"max_groups"`
	MaxUsers     int      `json:"max_users"`
	Purchasable  bool     `json:"purchasable"`
	Features     []string `json:"features"`
}

// Catalog is the canonical tier definition. Annual ≈ 10× monthly (~2 months free).
var Catalog = map[string]Tier{
	Gratis: {
		Key: Gratis, Name: "Gratis", Rank: 0,
		MonthlyPrice: 0, AnnualPrice: 0,
		MaxJamaah: 50, MaxGroups: 2, MaxUsers: 1,
		Purchasable: false,
		Features:    []string{"Hingga 50 jamaah", "Data jamaah & grup", "Manajemen paket dasar", "1 pengguna"},
	},
	Starter: {
		Key: Starter, Name: "Starter", Rank: 1,
		MonthlyPrice: 149000, AnnualPrice: 1490000,
		MaxJamaah: 250, MaxGroups: 5, MaxUsers: 3,
		Purchasable: true,
		Features:    []string{"Hingga 250 jamaah", "CRM & pembayaran", "AI Scanner dokumen", "Hingga 3 pengguna", "Laporan dasar"},
	},
	Pro: {
		Key: Pro, Name: "Pro", Rank: 2,
		MonthlyPrice: 299000, AnnualPrice: 2990000,
		MaxJamaah: Unlimited, MaxGroups: Unlimited, MaxUsers: 10,
		Purchasable: true,
		Features:    []string{"Jamaah tak terbatas", "Semua modul (CRM, Keuangan, Kontrak)", "AI Scanner tanpa batas", "Hingga 10 pengguna", "Laporan & ekspor lanjutan"},
	},
	Bisnis: {
		Key: Bisnis, Name: "Bisnis", Rank: 3,
		MonthlyPrice: 599000, AnnualPrice: 5990000,
		MaxJamaah: Unlimited, MaxGroups: Unlimited, MaxUsers: 25,
		Purchasable: true,
		Features:    []string{"Semua fitur Pro", "Multi-cabang & multi-PT", "Hingga 25 pengguna", "Dukungan prioritas"},
	},
	Enterprise: {
		Key: Enterprise, Name: "Enterprise", Rank: 4,
		MonthlyPrice: 0, AnnualPrice: 0,
		MaxJamaah: Unlimited, MaxGroups: Unlimited, MaxUsers: Unlimited,
		Purchasable: false,
		Features:    []string{"Semua fitur Bisnis", "Akses API & integrasi", "Pengguna tak terbatas", "Dukungan prioritas 24/7"},
	},
}

// Ordered is the catalog in display/rank order (low → high).
var Ordered = []Tier{Catalog[Gratis], Catalog[Starter], Catalog[Pro], Catalog[Bisnis], Catalog[Enterprise]}

// proRank is the threshold at which the advanced modules unlock.
var proRank = Catalog[Pro].Rank

// Valid reports whether key is a known tier.
func Valid(key string) bool {
	_, ok := Catalog[key]
	return ok
}

// Normalize maps legacy/unknown plan strings onto the current tiers.
// Legacy "free" → gratis, "business" → bisnis; anything unknown/empty → gratis.
func Normalize(key string) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case Gratis, Starter, Pro, Bisnis, Enterprise:
		return strings.ToLower(strings.TrimSpace(key))
	case "free":
		return Gratis
	case "business":
		return Bisnis
	default:
		return Gratis
	}
}

// Rank returns the tier rank (after Normalize). Higher = more capable.
func Rank(key string) int {
	return Catalog[Normalize(key)].Rank
}

// AtLeast reports whether key's tier is >= min's tier.
func AtLeast(key, min string) bool {
	return Rank(key) >= Rank(min)
}

// IsProOrHigher reports whether key unlocks the advanced modules.
func IsProOrHigher(key string) bool {
	return Rank(key) >= proRank
}

// Get returns the tier for key (after Normalize); always found.
func Get(key string) Tier {
	return Catalog[Normalize(key)]
}

// NormalizePeriod canonicalizes a billing period to PeriodMonthly or
// PeriodYearly. Accepts "monthly"/"month"/"" → monthly and
// "yearly"/"annual"/"year" → yearly. Anything else returns ok=false so callers
// reject it rather than silently defaulting (which previously let an
// annually-charged order grant only a one-month expiry). This is the single
// source of period interpretation shared by pricing and activation.
func NormalizePeriod(period string) (canonical string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case PeriodMonthly, "month", "":
		return PeriodMonthly, true
	case PeriodYearly, "annual", "year":
		return PeriodYearly, true
	default:
		return "", false
	}
}

// PriceFor returns the IDR amount for a purchasable tier and period.
// period accepts the forms understood by NormalizePeriod.
func PriceFor(key, period string) (int64, error) {
	t, ok := Catalog[key]
	if !ok {
		return 0, fmt.Errorf("unknown plan: %q", key)
	}
	if !t.Purchasable {
		return 0, fmt.Errorf("plan %q is not purchasable", key)
	}
	canonical, ok := NormalizePeriod(period)
	if !ok {
		return 0, fmt.Errorf("invalid period: %q (want monthly or yearly)", period)
	}
	if canonical == PeriodYearly {
		return t.AnnualPrice, nil
	}
	return t.MonthlyPrice, nil
}
