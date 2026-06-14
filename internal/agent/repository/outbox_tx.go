package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/agent/model"
	"github.com/jamaah-in/v2/internal/shared/outbox"
)

// CreateCommissionTx inserts an agent commission and a commission.accrued outbox
// event in one transaction. Commission id is DB-generated; the outbox
// aggregate_id is set from the returned id inside the same tx.
func (r *AgentRepo) CreateCommissionTx(ctx context.Context, c *model.AgentCommission, eventType string, payload []byte) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(ctx, `
		INSERT INTO agent_commissions (org_id, agent_id, jamaah_id, invoice_id, package_id, jamaah_name, package_name, commission_amount, commission_rate, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at, updated_at
	`, c.OrgID, c.AgentID, c.JamaahID, c.InvoiceID, c.PackageID, c.JamaahName, c.PackageName, c.CommissionAmount, c.CommissionRate, c.Notes).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return err
	}

	orgID, _ := uuid.Parse(c.OrgID)
	aggID, _ := uuid.Parse(c.ID)
	if err := outbox.Insert(ctx, tx, outbox.Event{
		OrgID:         orgID,
		AggregateType: "agent_commission",
		AggregateID:   aggID,
		EventType:     eventType,
		Payload:       payload,
		OccurredAt:    c.CreatedAt,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
