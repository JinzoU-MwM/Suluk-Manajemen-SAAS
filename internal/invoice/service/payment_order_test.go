package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/invoice/model"
	"github.com/jamaah-in/v2/internal/shared/httpclient"
	"github.com/jamaah-in/v2/internal/shared/plan"
)

func addrOf(u string) string { return strings.TrimPrefix(u, "http://") }

// A non-Starter caller cannot buy a top-up.
func TestCreateTopupOrderRejectsNonStarter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":{"plan":"gratis"}}`))
	}))
	defer ts.Close()

	s := &InvoiceService{httpc: httpclient.New(), authAddr: addrOf(ts.URL)}
	_, err := s.CreateTopupOrder(context.Background(), uuid.New(), uuid.New(), "Bearer x")
	if err == nil {
		t.Fatal("expected error for non-Starter caller")
	}
}

// callerPlan surfaces the org's tier from auth status, forwarding the token.
func TestCallerPlanReadsStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"plan":"starter"}}`))
	}))
	defer ts.Close()

	s := &InvoiceService{httpc: httpclient.New(), authAddr: addrOf(ts.URL)}
	got, err := s.callerPlan(context.Background(), "Bearer tok")
	if err != nil {
		t.Fatal(err)
	}
	if got != plan.Starter {
		t.Errorf("callerPlan = %q, want starter", got)
	}
}

// creditScanTopup forwards the order to aiocr with the internal key + correct
// scan quantity.
func TestCreditScanTopupCallsAiocr(t *testing.T) {
	var gotKey string
	var gotBody struct {
		OrderID string `json:"order_id"`
		OrgID   string `json:"org_id"`
		Scans   int    `json:"scans"`
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Internal-Key")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"success":true,"data":{"credited":true}}`))
	}))
	defer ts.Close()

	s := &InvoiceService{httpc: httpclient.New(), aiocrAddr: addrOf(ts.URL), internalKey: "k"}
	order := &model.PaymentOrder{ID: uuid.New(), OrgID: uuid.New(), Purpose: "scan_topup"}
	if err := s.creditScanTopup(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	if gotKey != "k" {
		t.Errorf("X-Internal-Key = %q, want k", gotKey)
	}
	if gotBody.Scans != plan.ScanTopupScans {
		t.Errorf("scans = %d, want %d", gotBody.Scans, plan.ScanTopupScans)
	}
	if gotBody.OrderID != order.ID.String() {
		t.Errorf("order_id = %q, want %q", gotBody.OrderID, order.ID.String())
	}
}
