package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// VerifyTransaction calls Pakasir's Transaction Detail API to authoritatively
// check a transaction's status. Pakasir webhooks are unsigned, so this is the
// trust anchor before we mark an order paid.
//
// GET {base}/api/transactiondetail?project={slug}&amount={amount}&order_id={id}&api_key={key}
// → { "transaction": { "status": "completed", "amount": ..., ... } }
//
// Returns the transaction status AND the authoritative paid amount so the caller
// can confirm the amount actually paid matches the order (the amount query param
// is only a lookup key, not proof of what was charged).
func (s *InvoiceService) VerifyTransaction(ctx context.Context, orderID string, amount int64) (status string, paidAmount int64, err error) {
	base := s.pakasir.BaseURL
	if base == "" {
		base = "https://app.pakasir.com"
	}
	q := url.Values{}
	q.Set("project", s.pakasir.ProjectSlug)
	q.Set("amount", fmt.Sprintf("%d", amount))
	q.Set("order_id", orderID)
	q.Set("api_key", s.pakasir.APIKey)
	endpoint := base + "/api/transactiondetail?" + q.Encode()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", 0, fmt.Errorf("pakasir transactiondetail status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Transaction struct {
			Status string `json:"status"`
			Amount int64  `json:"amount"`
		} `json:"transaction"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", 0, fmt.Errorf("decode pakasir response: %w", err)
	}
	return out.Transaction.Status, out.Transaction.Amount, nil
}
