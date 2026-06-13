package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jamaah-in/v2/internal/shared/events"
)

// buildPosting must always return balanced lines (sum debit == sum credit > 0)
// for every supported event type — the core accounting invariant.
func TestBuildPostingBalanced(t *testing.T) {
	mk := func(eventType string, payload map[string]any) *events.Envelope {
		b, _ := json.Marshal(payload)
		return &events.Envelope{
			EventID:    "evt-1",
			EventType:  eventType,
			OrgID:      "00000000-0000-0000-0000-000000000001",
			OccurredAt: time.Now(),
			Payload:    b,
		}
	}

	cases := []struct {
		name    string
		env     *events.Envelope
		wantAcc map[string][2]int64 // code -> {debit, credit}
	}{
		{
			name: "invoice issued",
			env:  mk(events.EventInvoiceIssued, map[string]any{"total_amount": 1000000, "invoice_number": "INV-1"}),
			wantAcc: map[string][2]int64{
				AccPiutangJemaah:   {1000000, 0},
				AccPendapatanPaket: {0, 1000000},
			},
		},
		{
			name: "payment transfer -> bank",
			env:  mk(events.EventPaymentReceived, map[string]any{"amount": 300000, "payment_method": "transfer_bank", "invoice_number": "INV-1"}),
			wantAcc: map[string][2]int64{
				AccBank:          {300000, 0},
				AccPiutangJemaah: {0, 300000},
			},
		},
		{
			name: "payment tunai -> kas",
			env:  mk(events.EventPaymentReceived, map[string]any{"amount": 50000, "payment_method": "tunai", "invoice_number": "INV-2"}),
			wantAcc: map[string][2]int64{
				AccKas:           {50000, 0},
				AccPiutangJemaah: {0, 50000},
			},
		},
		{
			name: "payroll with tax",
			env:  mk(events.EventPayrollPosted, map[string]any{"gross": 5000000, "tax": 500000, "net": 4500000}),
			wantAcc: map[string][2]int64{
				AccBebanGaji:   {5000000, 0},
				AccKas:         {0, 4500000},
				AccHutangPajak: {0, 500000},
			},
		},
		{
			name: "vendor bill",
			env:  mk(events.EventVendorBillCreated, map[string]any{"amount": 2000000, "vendor_name": "Hotel X"}),
			wantAcc: map[string][2]int64{
				AccBebanPerlengkapan: {2000000, 0},
				AccHutangVendor:      {0, 2000000},
			},
		},
		{
			name: "commission accrued",
			env:  mk(events.EventCommissionAccrued, map[string]any{"amount": 150000, "agent_name": "Agen A"}),
			wantAcc: map[string][2]int64{
				AccBebanKomisi:  {150000, 0},
				AccHutangKomisi: {0, 150000},
			},
		},
		{
			name: "savings deposit",
			env:  mk(events.EventSavingsDeposited, map[string]any{"amount": 750000}),
			wantAcc: map[string][2]int64{
				AccKas:            {750000, 0},
				AccHutangTabungan: {0, 750000},
			},
		},
		{
			name: "savings converted",
			env:  mk(events.EventSavingsConverted, map[string]any{"amount": 500000}),
			wantAcc: map[string][2]int64{
				AccHutangTabungan: {500000, 0},
				AccPiutangJemaah:  {0, 500000},
			},
		},
		{
			name: "cash session surplus",
			env:  mk(events.EventPosCashSessionClosed, map[string]any{"difference": 25000}),
			wantAcc: map[string][2]int64{
				AccKas:            {25000, 0},
				AccPendapatanLain: {0, 25000},
			},
		},
		{
			name: "cash session short",
			env:  mk(events.EventPosCashSessionClosed, map[string]any{"difference": -25000}),
			wantAcc: map[string][2]int64{
				AccBebanSelisihKas: {25000, 0},
				AccKas:             {0, 25000},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := buildPosting(tc.env)
			if err != nil {
				t.Fatalf("buildPosting: %v", err)
			}
			var sumD, sumC int64
			got := map[string][2]int64{}
			for _, l := range p.lines {
				sumD += l.Debit
				sumC += l.Credit
				got[l.AccountCode] = [2]int64{l.Debit, l.Credit}
			}
			if sumD != sumC {
				t.Errorf("unbalanced: debit=%d credit=%d", sumD, sumC)
			}
			if sumD == 0 {
				t.Errorf("zero-value journal")
			}
			for code, want := range tc.wantAcc {
				if got[code] != want {
					t.Errorf("account %s: got %v want %v", code, got[code], want)
				}
			}
		})
	}
}

// Unknown event types must be skipped (ErrNoTemplate), not error out the consumer.
func TestBuildPostingUnknownType(t *testing.T) {
	b, _ := json.Marshal(map[string]any{"x": 1})
	env := &events.Envelope{EventType: "something.unmapped", Payload: b}
	if _, err := buildPosting(env); err != ErrNoTemplate {
		t.Fatalf("want ErrNoTemplate, got %v", err)
	}
}

// Malformed payroll (gross != net+tax) must be rejected so the GL never goes
// unbalanced.
func TestBuildPostingPayrollImbalanceRejected(t *testing.T) {
	b, _ := json.Marshal(map[string]any{"gross": 100, "tax": 50, "net": 40}) // 100 != 90
	env := &events.Envelope{EventType: events.EventPayrollPosted, Payload: b}
	if _, err := buildPosting(env); err == nil {
		t.Fatal("expected error for imbalanced payroll, got nil")
	}
}
