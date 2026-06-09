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
// → { "transaction": { "status": "completed", ... } }
func (s *InvoiceService) VerifyTransaction(ctx context.Context, orderID string, amount int64) (string, error) {
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
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("pakasir transactiondetail status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Transaction struct {
			Status string `json:"status"`
			Amount int64  `json:"amount"`
		} `json:"transaction"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("decode pakasir response: %w", err)
	}
	return out.Transaction.Status, nil
}
