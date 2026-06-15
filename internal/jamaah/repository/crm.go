package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
)

// ListCRM returns jamaah profiles joined with their most-recent package
// registration (status/package/room/price + lead score), paginated, searchable
// and filterable by stage/temperature/min-score. Invoice balances are attached
// by the service layer.
func (r *JamaahRepo) ListCRM(ctx context.Context, orgID uuid.UUID, f model.CRMFilter, offset, limit int) ([]model.CRMJamaahRow, int, error) {
	// Shared FROM: the LATERAL pins each profile to its newest registration, so
	// stage/temp/score filters in the WHERE act on that registration.
	from := `FROM jamaah_profiles p
		LEFT JOIN LATERAL (
			SELECT package_id, room_type, pipeline_status, price_snapshot, discount_amount,
				lead_score, lead_temp, stage_entered_at, score_updated_at
			FROM jamaah_package_registrations
			WHERE jamaah_id = p.id AND org_id = p.org_id
			ORDER BY registered_at DESC
			LIMIT 1
		) r ON TRUE`

	where := ` WHERE p.org_id = $1`
	args := []any{orgID}
	add := func(cond string, val any) {
		args = append(args, val)
		where += fmt.Sprintf(cond, len(args))
	}
	if f.Search != "" {
		add(` AND (p.nama ILIKE $%[1]d OR p.no_identitas ILIKE $%[1]d OR p.no_paspor ILIKE $%[1]d OR p.no_hp ILIKE $%[1]d OR p.email ILIKE $%[1]d)`, "%"+f.Search+"%")
	}
	if f.Stage != "" {
		add(` AND r.pipeline_status = $%d`, f.Stage)
	}
	if f.Temp != "" {
		add(` AND r.lead_temp = $%d`, f.Temp)
	}
	if f.MinScore > 0 {
		add(` AND r.lead_score >= $%d`, f.MinScore)
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) `+from+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	order := ` ORDER BY p.created_at DESC`
	if f.Sort == "score" {
		order = ` ORDER BY r.lead_score DESC NULLS LAST, p.created_at DESC`
	}

	listQuery := `
		SELECT p.id, p.nama, COALESCE(p.no_hp, ''), COALESCE(p.no_identitas, ''), COALESCE(p.no_paspor, ''),
			COALESCE(p.email, ''), COALESCE(p.gender, ''),
			r.package_id, COALESCE(r.room_type, ''), COALESCE(r.pipeline_status, ''),
			COALESCE(r.price_snapshot, 0), COALESCE(r.discount_amount, 0),
			r.lead_score, COALESCE(r.lead_temp, ''), r.stage_entered_at, r.score_updated_at,
			COALESCE(EXTRACT(EPOCH FROM (NOW() - r.stage_entered_at)) / 86400, 0)::int
		` + from + where + order +
		fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]model.CRMJamaahRow, 0)
	for rows.Next() {
		var row model.CRMJamaahRow
		if err := rows.Scan(&row.ID, &row.Nama, &row.NoHP, &row.NoIdentitas, &row.NoPaspor,
			&row.Email, &row.Gender, &row.PackageID, &row.RoomType, &row.PipelineStatus,
			&row.PriceSnapshot, &row.DiscountAmount,
			&row.LeadScore, &row.LeadTemp, &row.StageEnteredAt, &row.ScoreUpdatedAt, &row.DaysInStage); err != nil {
			return nil, 0, err
		}
		out = append(out, row)
	}
	return out, total, rows.Err()
}

// ListLeadsByAgents returns jamaah whose referring agent is in the given set
// (an agent's own id plus their downline), hydrated with the latest
// registration's pipeline/score. Powers the B2B portal's "my leads".
func (r *JamaahRepo) ListLeadsByAgents(ctx context.Context, orgID uuid.UUID, agentIDs []uuid.UUID) ([]model.AgentLeadRow, error) {
	if len(agentIDs) == 0 {
		return []model.AgentLeadRow{}, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.nama, COALESCE(p.no_hp, ''), p.referring_agent_id,
			COALESCE(r.pipeline_status, ''), r.lead_score, COALESCE(r.lead_temp, '')
		FROM jamaah_profiles p
		LEFT JOIN LATERAL (
			SELECT pipeline_status, lead_score, lead_temp
			FROM jamaah_package_registrations
			WHERE jamaah_id = p.id AND org_id = p.org_id
			ORDER BY registered_at DESC
			LIMIT 1
		) r ON TRUE
		WHERE p.org_id = $1 AND p.referring_agent_id = ANY($2)
		ORDER BY p.created_at DESC`, orgID, agentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.AgentLeadRow{}
	for rows.Next() {
		var l model.AgentLeadRow
		if err := rows.Scan(&l.ID, &l.Nama, &l.NoHP, &l.ReferringAgentID, &l.PipelineStatus, &l.LeadScore, &l.LeadTemp); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// GetPipelineFunnel returns per-stage counts/value/avg-time and a lead-source
// breakdown for the CRM analytics view. Stage ordering is applied by the service.
func (r *JamaahRepo) GetPipelineFunnel(ctx context.Context, orgID uuid.UUID) (map[string]model.FunnelStage, []model.FunnelSource, error) {
	stageRows, err := r.pool.Query(ctx, `
		SELECT pipeline_status,
			COUNT(*),
			COALESCE(SUM(GREATEST(price_snapshot - discount_amount, 0)), 0),
			COALESCE(AVG(EXTRACT(EPOCH FROM (NOW() - stage_entered_at)) / 86400), 0)
		FROM jamaah_package_registrations
		WHERE org_id = $1
		GROUP BY pipeline_status`, orgID)
	if err != nil {
		return nil, nil, err
	}
	defer stageRows.Close()
	stages := make(map[string]model.FunnelStage)
	for stageRows.Next() {
		var s model.FunnelStage
		if err := stageRows.Scan(&s.Stage, &s.Count, &s.TotalValue, &s.AvgDaysInStage); err != nil {
			return nil, nil, err
		}
		stages[s.Stage] = s
	}
	if err := stageRows.Err(); err != nil {
		return nil, nil, err
	}

	srcRows, err := r.pool.Query(ctx, `
		SELECT COALESCE(NULLIF(p.lead_source, ''), 'unknown'), COUNT(*)
		FROM jamaah_package_registrations r
		JOIN jamaah_profiles p ON p.id = r.jamaah_id
		WHERE r.org_id = $1
		GROUP BY 1
		ORDER BY 2 DESC`, orgID)
	if err != nil {
		return nil, nil, err
	}
	defer srcRows.Close()
	sources := []model.FunnelSource{}
	for srcRows.Next() {
		var s model.FunnelSource
		if err := srcRows.Scan(&s.Source, &s.Count); err != nil {
			return nil, nil, err
		}
		sources = append(sources, s)
	}
	return stages, sources, srcRows.Err()
}
