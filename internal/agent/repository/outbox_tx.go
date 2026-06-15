package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/agent/model"
	"github.com/jamaah-in/v2/internal/shared/events"
	"github.com/jamaah-in/v2/internal/shared/outbox"
)

// TierCommission is one upline commission row (the seller's overrides) plus the
// outbox payload to emit for it. Built by the service's cascade engine.
type TierCommission struct {
	Commission *model.AgentCommission
	Payload    []byte
}

const insertCommissionSQL = `
	INSERT INTO agent_commissions
		(org_id, agent_id, jamaah_id, invoice_id, package_id, jamaah_name, package_name,
		 commission_amount, commission_rate, notes, tier_level, source_commission_id)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	RETURNING id, created_at, updated_at`

// CreateCommissionCascadeTx inserts the seller's commission (tier 1) plus every
// upline tier commission, and one commission.accrued outbox event per row, all
// in a single transaction — so a berjenjang payout is all-or-nothing and the
// accounting consumer posts a balanced journal for each tier. Each tier row's
// source_commission_id points back to the seller commission.
func (r *AgentRepo) CreateCommissionCascadeTx(ctx context.Context, seller *model.AgentCommission, sellerPayload []byte, tiers []TierCommission) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(ctx, insertCommissionSQL,
		seller.OrgID, seller.AgentID, seller.JamaahID, seller.InvoiceID, seller.PackageID,
		seller.JamaahName, seller.PackageName, seller.CommissionAmount, seller.CommissionRate,
		seller.Notes, seller.TierLevel, nil,
	).Scan(&seller.ID, &seller.CreatedAt, &seller.UpdatedAt); err != nil {
		return err
	}
	if err := insertOutbox(ctx, tx, seller, sellerPayload); err != nil {
		return err
	}

	for _, t := range tiers {
		c := t.Commission
		if err := tx.QueryRow(ctx, insertCommissionSQL,
			c.OrgID, c.AgentID, c.JamaahID, c.InvoiceID, c.PackageID,
			c.JamaahName, c.PackageName, c.CommissionAmount, c.CommissionRate,
			c.Notes, c.TierLevel, seller.ID,
		).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return err
		}
		c.SourceCommissionID = &seller.ID
		if err := insertOutbox(ctx, tx, c, t.Payload); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func insertOutbox(ctx context.Context, tx outbox.Querier, c *model.AgentCommission, payload []byte) error {
	orgID, _ := uuid.Parse(c.OrgID)
	aggID, _ := uuid.Parse(c.ID)
	return outbox.Insert(ctx, tx, outbox.Event{
		OrgID:         orgID,
		AggregateType: "agent_commission",
		AggregateID:   aggID,
		EventType:     events.EventCommissionAccrued,
		Payload:       payload,
		OccurredAt:    c.CreatedAt,
	})
}
