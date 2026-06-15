package service

// Lead scoring engine (Phase 3A). Deterministic, dependency-free, and
// table-testable: given a few signals it returns a 0-100 score and a
// temperature (hot/warm/cold). Kept pure so the recompute paths (event,
// in-process, lazy refresh) all produce identical results.

// ScoreSignals are the inputs to ComputeLeadScore, gathered by the recompute
// layer from invoice balances, documents, notes/follow-ups, and stage timing.
type ScoreSignals struct {
	Stage              string
	PaidAmount         int64
	TotalAmount        int64
	DocsTotal          int
	DocsReceived       int
	DaysSinceLastTouch int
	DaysInStage        int
}

// Stage contributes a base score reflecting how far the deal has progressed.
var stageBase = map[string]int{
	"prospek":   5,
	"survey":    12,
	"booking":   18,
	"dp":        24,
	"cicilan":   28,
	"lunas":     32,
	"berangkat": 35,
	"selesai":   35,
	"batal":     0,
}

// ComputeLeadScore returns a 0-100 lead score and temperature.
//
// Weights: stage base (max 35) + payment progress (max 35) + document
// completeness (max 20) + recency freshness (max 10) − staleness decay (max 10).
// Terminal overrides: batal → 0/cold; selesai/berangkat → hot.
func ComputeLeadScore(s ScoreSignals) (int, string) {
	if s.Stage == "batal" {
		return 0, "cold"
	}

	score := stageBase[s.Stage] // 0 for unknown stage

	// Payment progress (max 35).
	if s.TotalAmount > 0 && s.PaidAmount > 0 {
		p := int(s.PaidAmount * 35 / s.TotalAmount)
		if p > 35 {
			p = 35
		}
		score += p
	}

	// Document completeness (max 20).
	if s.DocsTotal > 0 && s.DocsReceived > 0 {
		d := s.DocsReceived * 20 / s.DocsTotal
		if d > 20 {
			d = 20
		}
		score += d
	}

	// Recency freshness (max +10) and staleness decay (max −10).
	score += freshnessBonus(s.DaysSinceLastTouch)
	if decay := s.DaysInStage / 10; decay > 0 {
		if decay > 10 {
			decay = 10
		}
		score -= decay
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	temp := temperature(score)
	if s.Stage == "selesai" || s.Stage == "berangkat" {
		temp = "hot"
	}
	return score, temp
}

func freshnessBonus(days int) int {
	switch {
	case days <= 3:
		return 10
	case days <= 7:
		return 7
	case days <= 14:
		return 4
	case days <= 30:
		return 1
	default:
		return 0
	}
}

func temperature(score int) string {
	switch {
	case score >= 66:
		return "hot"
	case score >= 33:
		return "warm"
	default:
		return "cold"
	}
}
