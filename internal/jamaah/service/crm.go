package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
)

// scoreStaleAfter is how long a persisted lead_score is trusted before the CRM
// list lazily recomputes it (with the freshly-fetched invoice balance, the most
// accurate payment signal available).
const scoreStaleAfter = 12 * time.Hour

// ListCRM returns the paginated, filterable CRM list and attaches each jamaah's
// outstanding invoice balance from the invoice service (best-effort: balances
// are omitted, not fatal, if it's unavailable). Rows whose score is missing or
// stale are recomputed inline using that balance — bounded to the current page.
func (s *JamaahService) ListCRM(ctx context.Context, orgID uuid.UUID, authToken string, f model.CRMFilter, page, pageSize int) ([]model.CRMJamaahRow, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	rows, total, err := s.repo.ListCRM(ctx, orgID, f, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	balances := s.fetchBalances(ctx, authToken)
	now := time.Now()
	for i := range rows {
		row := &rows[i]
		if b, ok := balances[row.ID]; ok {
			row.TotalAmount = b.TotalAmount
			row.TotalPaid = b.TotalPaid
			row.TotalRemaining = b.TotalRemaining
		}
		if row.PackageID == nil {
			continue // no server-side registration to score
		}
		if row.ScoreUpdatedAt != nil && now.Sub(*row.ScoreUpdatedAt) < scoreStaleAfter {
			continue
		}
		score, temp, rerr := s.RecomputeScore(ctx, orgID, row.ID, *row.PackageID, row.TotalPaid, row.TotalAmount)
		if rerr != nil {
			continue // best-effort; serve the stale/empty score
		}
		row.LeadScore = &score
		row.LeadTemp = temp
		stamped := now
		row.ScoreUpdatedAt = &stamped
	}
	return rows, total, nil
}

// stageOrder is the canonical pipeline ordering for the funnel view.
var stageOrder = []string{"prospek", "survey", "booking", "dp", "cicilan", "lunas", "berangkat", "selesai", "batal"}

// GetPipelineFunnel builds the CRM funnel analytics: per-stage counts/value/avg
// time-in-stage (canonically ordered) plus a lead-source breakdown.
func (s *JamaahService) GetPipelineFunnel(ctx context.Context, orgID uuid.UUID) (*model.PipelineFunnel, error) {
	stageMap, sources, err := s.repo.GetPipelineFunnel(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := &model.PipelineFunnel{Sources: sources}
	for _, st := range stageOrder {
		fs, ok := stageMap[st]
		if !ok {
			fs = model.FunnelStage{Stage: st}
		}
		out.Stages = append(out.Stages, fs)
		out.Total += fs.Count
	}
	return out, nil
}

type invoiceBalance struct {
	JamaahID       uuid.UUID `json:"jamaah_id"`
	TotalAmount    int64     `json:"total_amount"`
	TotalPaid      int64     `json:"total_paid"`
	TotalRemaining int64     `json:"total_remaining"`
}

// fetchBalances is best-effort: the CRM list still renders (without balances) if
// the invoice service is unavailable. The shared client adds timeout + retry.
func (s *JamaahService) fetchBalances(ctx context.Context, authToken string) map[uuid.UUID]invoiceBalance {
	if s.invoiceAddr == "" || authToken == "" {
		return nil
	}
	var balances []invoiceBalance
	if err := s.httpc.GetJSON(ctx, s.invoiceAddr, "/api/v1/invoices/balances", authToken, &balances); err != nil {
		return nil
	}
	m := make(map[uuid.UUID]invoiceBalance, len(balances))
	for _, b := range balances {
		m[b.JamaahID] = b
	}
	return m
}
