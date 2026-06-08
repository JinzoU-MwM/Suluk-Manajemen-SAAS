package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
)

// ListCRM returns the paginated CRM list and attaches each jamaah's outstanding
// invoice balance fetched from the invoice service (best-effort: balances are
// omitted, not fatal, if the invoice service is unavailable).
func (s *JamaahService) ListCRM(ctx context.Context, orgID uuid.UUID, authToken, search string, page, pageSize int) ([]model.CRMJamaahRow, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	rows, total, err := s.repo.ListCRM(ctx, orgID, search, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	if balances := s.fetchBalances(ctx, authToken); balances != nil {
		for i := range rows {
			if b, ok := balances[rows[i].ID]; ok {
				rows[i].TotalAmount = b.TotalAmount
				rows[i].TotalPaid = b.TotalPaid
				rows[i].TotalRemaining = b.TotalRemaining
			}
		}
	}
	return rows, total, nil
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
