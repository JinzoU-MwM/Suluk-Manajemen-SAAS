package service

import (
	"encoding/json"
	"fmt"

	"github.com/jamaah-in/v2/internal/accounting/model"
	"github.com/jamaah-in/v2/internal/shared/events"
)

// posting is the result of mapping an event to a balanced journal.
type posting struct {
	module      string
	description string
	lines       []model.PostingLine
}

// ErrNoTemplate means the event type has no posting rule (consumer skips + acks).
var ErrNoTemplate = fmt.Errorf("no posting template for event type")

// --- payload shapes the producers emit (kept minimal & stable) ---

type paymentPayload struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	InvoiceNumber string `json:"invoice_number"`
}

type invoiceIssuedPayload struct {
	TotalAmount   int64  `json:"total_amount"`
	InvoiceNumber string `json:"invoice_number"`
}

type vendorBillPayload struct {
	Amount     int64  `json:"amount"`
	VendorName string `json:"vendor_name"`
}

type payrollPayload struct {
	Gross int64 `json:"gross"`
	Tax   int64 `json:"tax"`
	Net   int64 `json:"net"`
}

type commissionPayload struct {
	Amount          int64  `json:"amount"`
	AgentName       string `json:"agent_name"`
	Tier            int    `json:"tier"`              // 1 = seller, ≥2 = upline override
	SourceAgentName string `json:"source_agent_name"` // selling agent, for tier ≥2
}

type savingsPayload struct {
	Amount int64 `json:"amount"`
}

type cashSessionPayload struct {
	Difference int64 `json:"difference"` // counted - expected (signed)
}

// buildPosting maps an event envelope to a balanced journal. Returns
// ErrNoTemplate for event types this engine does not (yet) handle.
func buildPosting(env *events.Envelope) (*posting, error) {
	switch env.EventType {
	case events.EventInvoiceIssued:
		var p invoiceIssuedPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil, fmt.Errorf("decode invoice payload: %w", err)
		}
		if p.TotalAmount <= 0 {
			return nil, fmt.Errorf("invoice total must be > 0")
		}
		memo := "Invoice terbit " + p.InvoiceNumber
		return &posting{
			module:      "invoice",
			description: memo,
			lines: []model.PostingLine{
				{AccountCode: AccPiutangJemaah, Debit: p.TotalAmount, Memo: memo},
				{AccountCode: AccPendapatanPaket, Credit: p.TotalAmount, Memo: memo},
			},
		}, nil

	case events.EventPaymentReceived:
		var p paymentPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil, fmt.Errorf("decode payment payload: %w", err)
		}
		if p.Amount <= 0 {
			return nil, fmt.Errorf("payment amount must be > 0")
		}
		cashAcc := AccBank
		if p.PaymentMethod == "tunai" || p.PaymentMethod == "cash" {
			cashAcc = AccKas
		}
		memo := "Pembayaran invoice " + p.InvoiceNumber
		return &posting{
			module:      "invoice",
			description: memo,
			lines: []model.PostingLine{
				{AccountCode: cashAcc, Debit: p.Amount, Memo: memo},
				{AccountCode: AccPiutangJemaah, Credit: p.Amount, Memo: memo},
			},
		}, nil

	case events.EventOverpaymentReceived:
		var p paymentPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil, fmt.Errorf("decode overpayment payload: %w", err)
		}
		if p.Amount <= 0 {
			return nil, fmt.Errorf("overpayment amount must be > 0")
		}
		cashAcc := AccBank
		if p.PaymentMethod == "tunai" || p.PaymentMethod == "cash" {
			cashAcc = AccKas
		}
		memo := "Kelebihan bayar (titipan) invoice " + p.InvoiceNumber
		return &posting{
			module:      "invoice",
			description: memo,
			lines: []model.PostingLine{
				{AccountCode: cashAcc, Debit: p.Amount, Memo: memo},
				{AccountCode: AccTitipanJemaah, Credit: p.Amount, Memo: memo},
			},
		}, nil

	case events.EventRefundCompleted:
		var p paymentPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil, fmt.Errorf("decode refund payload: %w", err)
		}
		if p.Amount <= 0 {
			return nil, fmt.Errorf("refund amount must be > 0")
		}
		cashAcc := AccBank
		if p.PaymentMethod == "tunai" || p.PaymentMethod == "cash" {
			cashAcc = AccKas
		}
		memo := "Refund invoice " + p.InvoiceNumber
		return &posting{
			module:      "invoice",
			description: memo,
			lines: []model.PostingLine{
				{AccountCode: AccPiutangJemaah, Debit: p.Amount, Memo: memo},
				{AccountCode: cashAcc, Credit: p.Amount, Memo: memo},
			},
		}, nil

	case events.EventInvoiceCancelled:
		var p paymentPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil, fmt.Errorf("decode invoice-cancelled payload: %w", err)
		}
		if p.Amount <= 0 {
			return nil, fmt.Errorf("cancelled amount must be > 0")
		}
		memo := "Pembatalan invoice " + p.InvoiceNumber
		return &posting{
			module:      "invoice",
			description: memo,
			lines: []model.PostingLine{
				{AccountCode: AccPendapatanPaket, Debit: p.Amount, Memo: memo},
				{AccountCode: AccPiutangJemaah, Credit: p.Amount, Memo: memo},
			},
		}, nil

	case events.EventVendorBillCreated:
		var p vendorBillPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil, fmt.Errorf("decode vendor bill payload: %w", err)
		}
		if p.Amount <= 0 {
			return nil, fmt.Errorf("vendor bill amount must be > 0")
		}
		memo := "Tagihan vendor " + p.VendorName
		return &posting{
			module:      "vendor",
			description: memo,
			lines: []model.PostingLine{
				{AccountCode: AccBebanPerlengkapan, Debit: p.Amount, Memo: memo},
				{AccountCode: AccHutangVendor, Credit: p.Amount, Memo: memo},
			},
		}, nil

	case events.EventPayrollPosted:
		var p payrollPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil, fmt.Errorf("decode payroll payload: %w", err)
		}
		if p.Gross <= 0 || p.Net < 0 || p.Tax < 0 || p.Gross != p.Net+p.Tax {
			return nil, fmt.Errorf("payroll amounts invalid (gross must equal net+tax)")
		}
		memo := "Gaji & potongan"
		lines := []model.PostingLine{
			{AccountCode: AccBebanGaji, Debit: p.Gross, Memo: memo},
			{AccountCode: AccKas, Credit: p.Net, Memo: memo},
		}
		if p.Tax > 0 {
			lines = append(lines, model.PostingLine{AccountCode: AccHutangPajak, Credit: p.Tax, Memo: memo})
		}
		return &posting{module: "payroll", description: memo, lines: lines}, nil

	case events.EventCommissionAccrued:
		var p commissionPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil, fmt.Errorf("decode commission payload: %w", err)
		}
		if p.Amount <= 0 {
			return nil, fmt.Errorf("commission amount must be > 0")
		}
		memo := "Komisi agen " + p.AgentName
		if p.Tier >= 2 {
			memo = fmt.Sprintf("Komisi berjenjang tier %d agen %s", p.Tier, p.AgentName)
			if p.SourceAgentName != "" {
				memo += " (dari " + p.SourceAgentName + ")"
			}
		}
		return &posting{
			module:      "agent",
			description: memo,
			lines: []model.PostingLine{
				{AccountCode: AccBebanKomisi, Debit: p.Amount, Memo: memo},
				{AccountCode: AccHutangKomisi, Credit: p.Amount, Memo: memo},
			},
		}, nil

	case events.EventSavingsDeposited:
		var p savingsPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil, fmt.Errorf("decode savings payload: %w", err)
		}
		if p.Amount <= 0 {
			return nil, fmt.Errorf("savings amount must be > 0")
		}
		memo := "Setoran tabungan jemaah"
		return &posting{
			module:      "tabungan",
			description: memo,
			lines: []model.PostingLine{
				{AccountCode: AccKas, Debit: p.Amount, Memo: memo},
				{AccountCode: AccHutangTabungan, Credit: p.Amount, Memo: memo},
			},
		}, nil

	case events.EventSavingsConverted:
		var p savingsPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil, fmt.Errorf("decode savings-converted payload: %w", err)
		}
		if p.Amount <= 0 {
			return nil, fmt.Errorf("conversion amount must be > 0")
		}
		// Saldo tabungan dipakai melunasi piutang jemaah (reklas, tanpa kas):
		// Dr Hutang Tabungan, Cr Piutang Jemaah.
		memo := "Konversi tabungan ke pelunasan"
		return &posting{
			module:      "tabungan",
			description: memo,
			lines: []model.PostingLine{
				{AccountCode: AccHutangTabungan, Debit: p.Amount, Memo: memo},
				{AccountCode: AccPiutangJemaah, Credit: p.Amount, Memo: memo},
			},
		}, nil

	case events.EventPosCashSessionClosed:
		var p cashSessionPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil, fmt.Errorf("decode cash-session payload: %w", err)
		}
		if p.Difference == 0 {
			return nil, ErrNoTemplate // tidak ada selisih → tidak perlu jurnal
		}
		if p.Difference > 0 {
			// Kas fisik lebih dari ekspektasi → Dr Kas, Cr Pendapatan Lain.
			memo := "Selisih kas lebih (tutup kas)"
			return &posting{
				module:      "pos",
				description: memo,
				lines: []model.PostingLine{
					{AccountCode: AccKas, Debit: p.Difference, Memo: memo},
					{AccountCode: AccPendapatanLain, Credit: p.Difference, Memo: memo},
				},
			}, nil
		}
		// Kas fisik kurang → Dr Beban Selisih Kas, Cr Kas.
		short := -p.Difference
		memo := "Selisih kas kurang (tutup kas)"
		return &posting{
			module:      "pos",
			description: memo,
			lines: []model.PostingLine{
				{AccountCode: AccBebanSelisihKas, Debit: short, Memo: memo},
				{AccountCode: AccKas, Credit: short, Memo: memo},
			},
		}, nil
	}

	return nil, ErrNoTemplate
}
