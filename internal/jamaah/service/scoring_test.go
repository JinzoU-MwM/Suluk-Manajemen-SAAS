package service

import "testing"

func TestComputeLeadScore(t *testing.T) {
	cases := []struct {
		name      string
		in        ScoreSignals
		wantScore int
		wantTemp  string
	}{
		{
			name:      "batal forces zero cold",
			in:        ScoreSignals{Stage: "batal", PaidAmount: 9_000_000, TotalAmount: 9_000_000, DocsTotal: 5, DocsReceived: 5},
			wantScore: 0,
			wantTemp:  "cold",
		},
		{
			name:      "lunas fully paid all docs fresh is hot",
			in:        ScoreSignals{Stage: "lunas", PaidAmount: 9_000_000, TotalAmount: 9_000_000, DocsTotal: 5, DocsReceived: 5, DaysSinceLastTouch: 1, DaysInStage: 2},
			wantScore: 97, // 32 + 35 + 20 + 10 - 0
			wantTemp:  "hot",
		},
		{
			name:      "berangkat is always hot",
			in:        ScoreSignals{Stage: "berangkat", DaysSinceLastTouch: 99, DaysInStage: 200},
			wantScore: 25, // 35 + 0 + 0 + 0 - 10(decay cap)
			wantTemp:  "hot",
		},
		{
			name:      "fresh prospect no money is cold",
			in:        ScoreSignals{Stage: "prospek", DaysSinceLastTouch: 1, DaysInStage: 0},
			wantScore: 15, // 5 + 0 + 0 + 10 - 0
			wantTemp:  "cold",
		},
		{
			name:      "warm boundary 33",
			in:        ScoreSignals{Stage: "prospek", PaidAmount: 80, TotalAmount: 100, DaysSinceLastTouch: 99, DaysInStage: 0},
			wantScore: 33, // 5 + 28(=35*80/100) + 0 + 0 - 0
			wantTemp:  "warm",
		},
		{
			name:      "cold boundary 32",
			in:        ScoreSignals{Stage: "prospek", PaidAmount: 27, TotalAmount: 35, DaysSinceLastTouch: 99, DaysInStage: 0},
			wantScore: 32, // 5 + 27 + 0 + 0 - 0
			wantTemp:  "cold",
		},
		{
			name:      "hot boundary 66",
			in:        ScoreSignals{Stage: "dp", PaidAmount: 100, TotalAmount: 100, DocsTotal: 10, DocsReceived: 5, DaysSinceLastTouch: 1, DaysInStage: 0},
			wantScore: 79, // 24 + 35 + 10 + 10 - 0
			wantTemp:  "hot",
		},
		{
			name:      "staleness decay caps at 10",
			in:        ScoreSignals{Stage: "cicilan", PaidAmount: 50, TotalAmount: 100, DaysSinceLastTouch: 40, DaysInStage: 365},
			wantScore: 35, // 28 + 17(=35*50/100) + 0 + 0 - 10
			wantTemp:  "warm",
		},
		{
			name:      "zero total payment contributes nothing",
			in:        ScoreSignals{Stage: "survey", PaidAmount: 5, TotalAmount: 0, DaysSinceLastTouch: 2},
			wantScore: 22, // 12 + 0 + 0 + 10
			wantTemp:  "cold",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotScore, gotTemp := ComputeLeadScore(c.in)
			if gotScore != c.wantScore {
				t.Fatalf("score = %d, want %d", gotScore, c.wantScore)
			}
			if gotTemp != c.wantTemp {
				t.Fatalf("temp = %q, want %q", gotTemp, c.wantTemp)
			}
		})
	}
}
