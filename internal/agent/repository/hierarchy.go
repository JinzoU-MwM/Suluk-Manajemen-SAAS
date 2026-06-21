package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/jamaah-in/v2/internal/agent/model"
)

// AncestorIDs returns the agent plus all of its ancestors (walking parent_id up).
// Used for cycle detection: a proposed parent is invalid if the agent being
// re-parented appears among the proposed parent's ancestor chain.
func (r *AgentRepo) AncestorIDs(ctx context.Context, agentID, orgID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		WITH RECURSIVE ul AS (
			SELECT id, parent_id, 0 AS depth FROM agents WHERE id = $1 AND org_id = $2
			UNION ALL
			SELECT a.id, a.parent_id, ul.depth + 1 FROM agents a JOIN ul ON a.id = ul.parent_id
			WHERE a.org_id = $2 AND ul.depth < 32
		)
		SELECT id FROM ul`, agentID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// UplineAgents returns the seller's ancestors nearest-first, up to maxDepth
// levels. Depth 1 = direct parent (commission tier 2), depth 2 = grandparent
// (tier 3), etc. Drives the commission cascade.
func (r *AgentRepo) UplineAgents(ctx context.Context, agentID, orgID string, maxDepth int) ([]model.DownlineNode, error) {
	rows, err := r.pool.Query(ctx, `
		WITH RECURSIVE ul AS (
			SELECT id, parent_id, name, level, is_active, 0 AS depth
			FROM agents WHERE id = $1 AND org_id = $3
			UNION ALL
			SELECT a.id, a.parent_id, a.name, a.level, a.is_active, ul.depth + 1
			FROM agents a JOIN ul ON a.id = ul.parent_id WHERE a.org_id = $3 AND ul.depth < $2
		)
		SELECT id, name, parent_id, level, depth, is_active
		FROM ul WHERE depth > 0 AND depth <= $2 ORDER BY depth`, agentID, maxDepth, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNodes(rows)
}

// Upline returns the agent's ancestor chain (nearest-first) with commission
// aggregates, for the /agents/:id/upline endpoint.
func (r *AgentRepo) Upline(ctx context.Context, agentID, orgID string) ([]model.DownlineNode, error) {
	rows, err := r.pool.Query(ctx, `
		WITH RECURSIVE ul AS (
			SELECT id, parent_id, name, level, is_active, 0 AS depth
			FROM agents WHERE id = $1 AND org_id = $2
			UNION ALL
			SELECT a.id, a.parent_id, a.name, a.level, a.is_active, ul.depth + 1
			FROM agents a JOIN ul ON a.id = ul.parent_id WHERE a.org_id = $2 AND ul.depth < 32
		)
		SELECT u.id, u.name, u.parent_id, u.level, u.depth, u.is_active,
		       COALESCE(c.total_comm, 0), COALESCE(c.jamaah_count, 0)
		FROM ul u
		LEFT JOIN (
			SELECT agent_id, SUM(commission_amount) AS total_comm, COUNT(DISTINCT jamaah_id) AS jamaah_count
			FROM agent_commissions WHERE org_id = $2 GROUP BY agent_id
		) c ON c.agent_id = u.id
		WHERE u.depth > 0 ORDER BY u.depth`, agentID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAggNodes(rows)
}

// Downline returns all descendants of an agent (depth-ordered) with commission
// aggregates, for the /agents/:id/downline tree endpoint.
func (r *AgentRepo) Downline(ctx context.Context, agentID, orgID string) ([]model.DownlineNode, error) {
	rows, err := r.pool.Query(ctx, `
		WITH RECURSIVE dl AS (
			SELECT id, parent_id, name, level, is_active, 0 AS depth
			FROM agents WHERE id = $1 AND org_id = $2
			UNION ALL
			SELECT a.id, a.parent_id, a.name, a.level, a.is_active, dl.depth + 1
			FROM agents a JOIN dl ON a.parent_id = dl.id WHERE a.org_id = $2 AND dl.depth < 32
		)
		SELECT d.id, d.name, d.parent_id, d.level, d.depth, d.is_active,
		       COALESCE(c.total_comm, 0), COALESCE(c.jamaah_count, 0)
		FROM dl d
		LEFT JOIN (
			SELECT agent_id, SUM(commission_amount) AS total_comm, COUNT(DISTINCT jamaah_id) AS jamaah_count
			FROM agent_commissions WHERE org_id = $2 GROUP BY agent_id
		) c ON c.agent_id = d.id
		WHERE d.depth > 0 ORDER BY d.depth, d.name`, agentID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAggNodes(rows)
}

func scanNodes(rows pgx.Rows) ([]model.DownlineNode, error) {
	nodes := []model.DownlineNode{}
	for rows.Next() {
		var n model.DownlineNode
		if err := rows.Scan(&n.ID, &n.Name, &n.ParentID, &n.Level, &n.Depth, &n.IsActive); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func scanAggNodes(rows pgx.Rows) ([]model.DownlineNode, error) {
	nodes := []model.DownlineNode{}
	for rows.Next() {
		var n model.DownlineNode
		if err := rows.Scan(&n.ID, &n.Name, &n.ParentID, &n.Level, &n.Depth, &n.IsActive, &n.TotalCommissions, &n.TotalJamaah); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

// ListTiers returns the org's configured upline override rates (may be empty,
// in which case the service falls back to code defaults).
func (r *AgentRepo) ListTiers(ctx context.Context, orgID string) ([]model.CommissionTier, error) {
	rows, err := r.pool.Query(ctx, `SELECT level, rate_pct FROM commission_tiers WHERE org_id = $1 ORDER BY level`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tiers := []model.CommissionTier{}
	for rows.Next() {
		var t model.CommissionTier
		if err := rows.Scan(&t.Level, &t.RatePct); err != nil {
			return nil, err
		}
		tiers = append(tiers, t)
	}
	return tiers, rows.Err()
}

// UpsertTiers replaces the org's tier configuration in one transaction.
func (r *AgentRepo) UpsertTiers(ctx context.Context, orgID string, tiers []model.CommissionTier) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM commission_tiers WHERE org_id = $1`, orgID); err != nil {
		return err
	}
	for _, t := range tiers {
		if _, err := tx.Exec(ctx, `INSERT INTO commission_tiers (org_id, level, rate_pct) VALUES ($1, $2, $3)`,
			orgID, t.Level, t.RatePct); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
