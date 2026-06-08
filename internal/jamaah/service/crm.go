package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

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

func (s *JamaahService) fetchBalances(ctx context.Context, authToken string) map[uuid.UUID]invoiceBalance {
	if s.invoiceAddr == "" || authToken == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+s.invoiceAddr+"/api/v1/invoices/balances", nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", authToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var env struct {
		Success bool             `json:"success"`
		Data    []invoiceBalance `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil || !env.Success {
		return nil
	}
	m := make(map[uuid.UUID]invoiceBalance, len(env.Data))
	for _, b := range env.Data {
		m[b.JamaahID] = b
	}
	return m
}
