package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jamaah-in/v2/internal/agent/model"
)

// ErrCommissionNotPayable means PayCommission matched no pending commission for
// the org (already paid, not found, or wrong org).
var ErrCommissionNotPayable = errors.New("commission not found or already paid")

type AgentRepo struct {
	pool *pgxpool.Pool
}

func NewAgentRepo(pool *pgxpool.Pool) *AgentRepo {
	return &AgentRepo{pool: pool}
}

func (r *AgentRepo) CreateAgent(ctx context.Context, a *model.Agent) error {
	if a.Type == "" {
		a.Type = "agent"
	}
	if a.Level <= 0 {
		a.Level = 1
	}
	return r.pool.QueryRow(ctx, `
		INSERT INTO agents (org_id, name, phone, email, address, commission_rate, bank_name, bank_account_number, bank_account_name, notes, parent_id, level, type)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, created_at, updated_at
	`, a.OrgID, a.Name, a.Phone, a.Email, a.Address, a.CommissionRate, a.BankName, a.BankAccountNumber, a.BankAccountName, a.Notes, a.ParentID, a.Level, a.Type).
		Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (r *AgentRepo) ListAgents(ctx context.Context, orgID, search string, page, limit int) ([]model.Agent, int64, error) {
	var total int64
	// Qualify with the "a" alias: the list query self-joins agents (parent), so
	// bare org_id/name/etc would be ambiguous.
	baseWhere := "WHERE a.org_id = $1"
	args := []interface{}{orgID}
	if search != "" {
		baseWhere += " AND (a.name ILIKE $2 OR a.phone ILIKE $2 OR a.email ILIKE $2)"
		args = append(args, "%"+search+"%")
	}
	r.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM agents a %s", baseWhere), args...).Scan(&total)

	offset := (page - 1) * limit
	baseArgCount := len(args)
	query := fmt.Sprintf(`
		SELECT a.id, a.org_id, a.name, a.phone, a.email, a.address, a.commission_rate,
		       a.bank_name, a.bank_account_number, a.bank_account_name, a.notes, a.is_active,
		       a.parent_id, a.level, a.type, p.name, a.created_at, a.updated_at,
		       COALESCE(c.total_comm, 0), COALESCE(c.total_paid, 0), COALESCE(c.total_comm - c.total_paid, 0), COALESCE(c.jamaah_count, 0)
		FROM agents a
		LEFT JOIN agents p ON p.id = a.parent_id
		LEFT JOIN (
			SELECT agent_id,
			       SUM(commission_amount) as total_comm,
			       SUM(CASE WHEN status = 'paid' THEN commission_amount ELSE 0 END) as total_paid,
			       COUNT(DISTINCT jamaah_id) as jamaah_count
			FROM agent_commissions WHERE org_id = $1 GROUP BY agent_id
		) c ON a.id = c.agent_id
		%s ORDER BY a.name LIMIT $%d OFFSET $%d
	`, baseWhere, baseArgCount+1, baseArgCount+2)
	selectArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var agents []model.Agent
	for rows.Next() {
		var a model.Agent
		var parentName *string
		if err := rows.Scan(
			&a.ID, &a.OrgID, &a.Name, &a.Phone, &a.Email, &a.Address, &a.CommissionRate,
			&a.BankName, &a.BankAccountNumber, &a.BankAccountName, &a.Notes, &a.IsActive,
			&a.ParentID, &a.Level, &a.Type, &parentName, &a.CreatedAt, &a.UpdatedAt,
			&a.TotalCommissions, &a.TotalPaid, &a.TotalOutstanding, &a.TotalJamaah,
		); err != nil {
			return nil, 0, err
		}
		if parentName != nil {
			a.ParentName = *parentName
		}
		agents = append(agents, a)
	}
	return agents, total, rows.Err()
}

func (r *AgentRepo) GetAgent(ctx context.Context, id, orgID string) (*model.Agent, error) {
	var a model.Agent
	var parentName *string
	err := r.pool.QueryRow(ctx, `
		SELECT a.id, a.org_id, a.name, a.phone, a.email, a.address, a.commission_rate,
		       a.bank_name, a.bank_account_number, a.bank_account_name, a.notes, a.is_active,
		       a.parent_id, a.level, a.type, p.name, a.created_at, a.updated_at,
		       COALESCE(c.total_comm, 0), COALESCE(c.total_paid, 0), COALESCE(c.total_comm - c.total_paid, 0), COALESCE(c.jamaah_count, 0)
		FROM agents a
		LEFT JOIN agents p ON p.id = a.parent_id
		LEFT JOIN (
			SELECT agent_id,
			       SUM(commission_amount) as total_comm,
			       SUM(CASE WHEN status = 'paid' THEN commission_amount ELSE 0 END) as total_paid,
			       COUNT(DISTINCT jamaah_id) as jamaah_count
			FROM agent_commissions WHERE org_id = $2 GROUP BY agent_id
		) c ON a.id = c.agent_id
		WHERE a.id = $1 AND a.org_id = $2
	`, id, orgID).Scan(
		&a.ID, &a.OrgID, &a.Name, &a.Phone, &a.Email, &a.Address, &a.CommissionRate,
		&a.BankName, &a.BankAccountNumber, &a.BankAccountName, &a.Notes, &a.IsActive,
		&a.ParentID, &a.Level, &a.Type, &parentName, &a.CreatedAt, &a.UpdatedAt,
		&a.TotalCommissions, &a.TotalPaid, &a.TotalOutstanding, &a.TotalJamaah,
	)
	if err != nil {
		return nil, fmt.Errorf("agent not found")
	}
	if parentName != nil {
		a.ParentName = *parentName
	}
	return &a, nil
}

func (r *AgentRepo) UpdateAgent(ctx context.Context, id, orgID string, updates map[string]interface{}) error {
	query := "UPDATE agents SET updated_at = NOW()"
	args := []interface{}{}
	idx := 1
	for k, v := range updates {
		query += fmt.Sprintf(", %s = $%d", k, idx)
		args = append(args, v)
		idx++
	}
	query += fmt.Sprintf(" WHERE id = $%d AND org_id = $%d", idx, idx+1)
	args = append(args, id, orgID)
	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *AgentRepo) CreateCommission(ctx context.Context, c *model.AgentCommission) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO agent_commissions (org_id, agent_id, jamaah_id, invoice_id, package_id, jamaah_name, package_name, commission_amount, commission_rate, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at, updated_at
	`, c.OrgID, c.AgentID, c.JamaahID, c.InvoiceID, c.PackageID, c.JamaahName, c.PackageName, c.CommissionAmount, c.CommissionRate, c.Notes).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *AgentRepo) ListCommissions(ctx context.Context, orgID, agentID, status, tierLevel string, page, limit int) ([]model.AgentCommission, int64, error) {
	var total int64
	baseWhere := "WHERE c.org_id = $1"
	args := []interface{}{orgID}
	argIdx := 2
	if agentID != "" {
		baseWhere += fmt.Sprintf(" AND c.agent_id = $%d", argIdx)
		args = append(args, agentID)
		argIdx++
	}
	if status != "" && status != "all" {
		baseWhere += fmt.Sprintf(" AND c.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if tierLevel != "" {
		baseWhere += fmt.Sprintf(" AND c.tier_level = $%d", argIdx)
		args = append(args, tierLevel)
		argIdx++
	}
	r.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM agent_commissions c %s", baseWhere), args...).Scan(&total)

	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT c.id, c.org_id, c.agent_id, c.jamaah_id, c.invoice_id, c.package_id, c.jamaah_name, c.package_name,
		       c.commission_amount, c.commission_rate, c.status, c.paid_at, c.notes, c.tier_level, c.source_commission_id, c.created_at, c.updated_at, a.name
		FROM agent_commissions c JOIN agents a ON c.agent_id = a.id AND a.org_id = c.org_id
		%s ORDER BY c.created_at DESC LIMIT $%d OFFSET $%d
	`, baseWhere, argIdx, argIdx+1)
	selectArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []model.AgentCommission
	for rows.Next() {
		var c model.AgentCommission
		if err := rows.Scan(
			&c.ID, &c.OrgID, &c.AgentID, &c.JamaahID, &c.InvoiceID, &c.PackageID, &c.JamaahName, &c.PackageName,
			&c.CommissionAmount, &c.CommissionRate, &c.Status, &c.PaidAt, &c.Notes, &c.TierLevel, &c.SourceCommissionID, &c.CreatedAt, &c.UpdatedAt, &c.AgentName,
		); err != nil {
			return nil, 0, err
		}
		result = append(result, c)
	}
	return result, total, rows.Err()
}

func (r *AgentRepo) PayCommission(ctx context.Context, id, orgID string) error {
	now := time.Now()
	tag, err := r.pool.Exec(ctx, `
		UPDATE agent_commissions SET status = 'paid', paid_at = $1, updated_at = $1 WHERE id = $2 AND org_id = $3 AND status = 'pending'
	`, now, id, orgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCommissionNotPayable
	}
	return nil
}

func (r *AgentRepo) GetAgentCommissions(ctx context.Context, agentID, orgID string) ([]model.AgentCommission, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.org_id, c.agent_id, c.jamaah_id, c.invoice_id, c.package_id, c.jamaah_name, c.package_name,
		       c.commission_amount, c.commission_rate, c.status, c.paid_at, c.notes, c.tier_level, c.source_commission_id, c.created_at, c.updated_at, a.name
		FROM agent_commissions c JOIN agents a ON c.agent_id = a.id AND a.org_id = c.org_id
		WHERE c.agent_id = $1 AND c.org_id = $2 ORDER BY c.created_at DESC
	`, agentID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.AgentCommission
	for rows.Next() {
		var c model.AgentCommission
		if err := rows.Scan(
			&c.ID, &c.OrgID, &c.AgentID, &c.JamaahID, &c.InvoiceID, &c.PackageID, &c.JamaahName, &c.PackageName,
			&c.CommissionAmount, &c.CommissionRate, &c.Status, &c.PaidAt, &c.Notes, &c.TierLevel, &c.SourceCommissionID, &c.CreatedAt, &c.UpdatedAt, &c.AgentName,
		); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
