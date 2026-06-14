package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/jamaah-in/v2/internal/accounting/service"
	sharedConfig "github.com/jamaah-in/v2/internal/shared/config"
	sharedDB "github.com/jamaah-in/v2/internal/shared/database"
	"github.com/jamaah-in/v2/internal/shared/events"
)

// runBackfill replays historical invoices + payments from the jamaah_invoice DB
// into the accounting GL via the same posting path. Deterministic event ids make
// it idempotent: re-running posts nothing new (processed_events dedups).
//
// invoice.issued → Dr Piutang / Cr Pendapatan; payment.received → Dr Kas|Bank /
// Cr Piutang. Together the GL is balanced: assets (cash+AR) = revenue (equity).
func runBackfill(ctx context.Context, cfg *sharedConfig.Config, svc *service.Service, logger *zap.SugaredLogger) error {
	invCfg := cfg.Database
	invCfg.DBName = "jamaah_invoice"
	pool, err := sharedDB.Connect(ctx, invCfg.DSN())
	if err != nil {
		return fmt.Errorf("connect invoice db: %w", err)
	}
	defer sharedDB.Close(pool)

	// 1) invoices (exclude cancelled) → invoice.issued
	invRows, err := pool.Query(ctx, `SELECT id, org_id, invoice_number, total_amount, issued_at
		FROM invoices WHERE status != 'batal'`)
	if err != nil {
		return fmt.Errorf("query invoices: %w", err)
	}
	var issued, payments int
	type inv struct {
		id, org, number string
		total           int64
		at              time.Time
	}
	invs := []inv{}
	for invRows.Next() {
		var v inv
		if err := invRows.Scan(&v.id, &v.org, &v.number, &v.total, &v.at); err != nil {
			invRows.Close()
			return err
		}
		invs = append(invs, v)
	}
	invRows.Close()
	for _, v := range invs {
		payload, _ := json.Marshal(map[string]any{"total_amount": v.total, "invoice_number": v.number})
		env := &events.Envelope{
			EventID:       "backfill:invoice:" + v.id,
			EventType:     events.EventInvoiceIssued,
			OrgID:         v.org,
			AggregateType: "invoice",
			AggregateID:   v.id,
			OccurredAt:    v.at,
			Payload:       payload,
		}
		posted, perr := svc.PostFromEvent(ctx, env)
		if perr != nil {
			logger.Warnf("backfill invoice %s: %v", v.id, perr)
			continue
		}
		if posted {
			issued++
		}
	}

	// 2) payments → payment.received
	payRows, err := pool.Query(ctx, `SELECT p.id, p.org_id, p.amount, p.payment_method, p.invoice_id, i.invoice_number, p.paid_at
		FROM payments p JOIN invoices i ON i.id = p.invoice_id WHERE i.status != 'batal'`)
	if err != nil {
		return fmt.Errorf("query payments: %w", err)
	}
	type pay struct {
		id, org, method, invID, number string
		amount                         int64
		at                             time.Time
	}
	pays := []pay{}
	for payRows.Next() {
		var p pay
		if err := payRows.Scan(&p.id, &p.org, &p.amount, &p.method, &p.invID, &p.number, &p.at); err != nil {
			payRows.Close()
			return err
		}
		pays = append(pays, p)
	}
	payRows.Close()
	for _, p := range pays {
		payload, _ := json.Marshal(map[string]any{"amount": p.amount, "payment_method": p.method, "invoice_number": p.number})
		env := &events.Envelope{
			EventID:       "backfill:payment:" + p.id,
			EventType:     events.EventPaymentReceived,
			OrgID:         p.org,
			AggregateType: "invoice",
			AggregateID:   p.invID,
			OccurredAt:    p.at,
			Payload:       payload,
		}
		posted, perr := svc.PostFromEvent(ctx, env)
		if perr != nil {
			logger.Warnf("backfill payment %s: %v", p.id, perr)
			continue
		}
		if posted {
			payments++
		}
	}

	logger.Infof("backfill: posted %d invoice-issued + %d payment journals", issued, payments)
	return nil
}
