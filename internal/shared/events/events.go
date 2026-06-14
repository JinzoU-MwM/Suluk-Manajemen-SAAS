// Package events is the Suluk Integration Bus: a thin wrapper over NATS
// JetStream plus the canonical event envelope and subject vocabulary shared by
// every service. Producers publish via the transactional outbox relay; the
// accounting-service consumes to post double-entry journals.
package events

import (
	"encoding/json"
	"time"
)

// Stream + subject vocabulary. All Suluk domain events live under `suluk.>` in a
// single JetStream stream so the accounting consumer can subscribe with one
// wildcard and the bus stays simple to operate.
const (
	StreamName = "SULUK_EVENTS"
	SubjectAll = "suluk.>"
	subjectPrefix = "suluk."
)

// Event types (the suffix after `suluk.`). Keep these stable: they are the
// contract between producers and the accounting posting engine.
const (
	EventInvoiceIssued    = "invoice.issued"      // invoice terbit (Dr Piutang Jemaah, Cr Pendapatan Paket)
	EventPaymentReceived  = "payment.received"   // jemaah membayar invoice (Dr Kas/Bank, Cr Piutang)
	EventRefundCompleted  = "refund.completed"    // refund cair (Dr Piutang/again, Cr Kas)
	EventVendorBillCreated = "vendor.bill.created" // tagihan vendor (Dr Beban/Persediaan, Cr Hutang Vendor)
	EventPayrollPosted    = "payroll.posted"      // slip gaji (Dr Beban Gaji, Cr Kas/Hutang)
	EventCommissionAccrued = "commission.accrued"  // komisi agen (Dr Beban Komisi, Cr Hutang Komisi)
	EventSavingsDeposited = "savings.deposited"    // setoran tabungan (Dr Kas, Cr Hutang Tabungan)
	EventSavingsConverted = "savings.converted"    // tabungan jadi pelunasan (Dr Hutang Tabungan, Cr Piutang)
	EventPosCashSessionClosed = "pos.cash.session.closed" // tutup kas: selisih kas lebih/kurang
)

// Subject returns the fully-qualified NATS subject for an event type.
func Subject(eventType string) string { return subjectPrefix + eventType }

// Envelope is the canonical wire format for every event on the bus. EventID is
// the producer's outbox row id — it doubles as the JetStream Msg-Id (publish
// dedup) and the accounting consumer's processed_events key (post-once), giving
// exactly-once-effective journaling end to end.
type Envelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	OrgID         string          `json:"org_id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload"`
}

// Marshal serializes the envelope for publishing.
func (e *Envelope) Marshal() ([]byte, error) { return json.Marshal(e) }

// ParseEnvelope decodes a wire message back into an Envelope.
func ParseEnvelope(data []byte) (*Envelope, error) {
	var e Envelope
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}
