package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func (s *AuthService) CreateBranch(ctx context.Context, parentOrgID, name string) (map[string]interface{}, error) {
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	id, err := s.repo.CreateBranch(ctx, uuid.MustParse(parentOrgID), name, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}
	return map[string]interface{}{"id": id.String(), "name": name, "slug": slug}, nil
}

func (s *AuthService) ListBranches(ctx context.Context, parentOrgID uuid.UUID) ([]interface{}, error) {
	branches, err := s.repo.ListBranches(ctx, parentOrgID)
	if err != nil {
		return nil, err
	}
	var result []interface{}
	for _, b := range branches {
		result = append(result, b)
	}
	return result, nil
}

// GetConsolidatedStats aggregates totals across the parent org and all its
// branches. Because each service owns a separate database, the figures cannot be
// queried locally; instead we mint a short-lived token scoped to each org and
// call the per-org endpoints of the jamaah and invoice services, then sum the
// results. If JWT or the downstream addresses are unavailable, only the branch
// count (which lives in this database) is returned.
func (s *AuthService) GetConsolidatedStats(ctx context.Context, parentOrgID string) (map[string]interface{}, error) {
	parentID, err := uuid.Parse(parentOrgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}

	branches, err := s.repo.ListBranches(ctx, parentID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"total_branches": len(branches),
		"total_jamaah":   0,
		"total_revenue":  int64(0),
		"total_invoices": 0,
	}

	if s.jwt == nil || s.jamaahAddr == "" || s.invoiceAddr == "" {
		return result, nil
	}

	orgIDs := make([]uuid.UUID, 0, len(branches)+1)
	orgIDs = append(orgIDs, parentID)
	for _, b := range branches {
		orgIDs = append(orgIDs, b.ID)
	}

	var totalJamaah, totalInvoices int
	var totalRevenue int64
	partial := false

	for _, orgID := range orgIDs {
		tp, err := s.jwt.GenerateTokenPair(uuid.Nil, orgID, "owner", "consolidated@internal")
		if err != nil {
			partial = true
			continue
		}
		token := "Bearer " + tp.AccessToken

		var analytics struct {
			TotalJamaah int `json:"total_jamaah"`
		}
		if err := s.fetchJSON(ctx, s.jamaahAddr, "/api/v1/analytics/dashboard", token, &analytics); err != nil {
			partial = true
		} else {
			totalJamaah += analytics.TotalJamaah
		}

		var summary struct {
			TotalInvoices int64 `json:"total_invoices"`
			TotalPaid     int64 `json:"total_paid"`
		}
		if err := s.fetchJSON(ctx, s.invoiceAddr, "/api/v1/invoices/summary", token, &summary); err != nil {
			partial = true
		} else {
			totalRevenue += summary.TotalPaid
			totalInvoices += int(summary.TotalInvoices)
		}
	}

	result["total_jamaah"] = totalJamaah
	result["total_revenue"] = totalRevenue
	result["total_invoices"] = totalInvoices
	if partial {
		result["partial"] = true
	}
	return result, nil
}

// fetchJSON does an authenticated GET to a sibling service and unwraps the
// standard {success, data} envelope into out.
func (s *AuthService) fetchJSON(ctx context.Context, addr, path, authToken string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", authToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("service %s returned status %d", path, resp.StatusCode)
	}

	var envelope struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return err
	}
	if !envelope.Success {
		return fmt.Errorf("service %s returned error", path)
	}
	return json.Unmarshal(envelope.Data, out)
}
