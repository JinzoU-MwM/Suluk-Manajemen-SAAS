package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/shared/httpclient"
)

func TestCancelDanglingInvoicesCancelsAndRefundsMatchingInvoice(t *testing.T) {
	jamaahID := uuid.New()
	packageID := uuid.New()
	invoiceID := uuid.New()
	var sawCancel, sawRefund bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/invoices/jamaah/"+jamaahID.String():
			list, _ := json.Marshal([]map[string]any{{
				"id": invoiceID, "package_id": packageID, "status": "belum_bayar",
				"amount_paid": 200000, "amount_remaining": 800000,
			}})
			_, _ = w.Write([]byte(`{"success":true,"data":` + string(list) + `}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/invoices/"+invoiceID.String()+"/cancel":
			sawCancel = true
			_, _ = w.Write([]byte(`{"success":true,"data":null}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/invoices/"+invoiceID.String()+"/refund":
			sawRefund = true
			_, _ = w.Write([]byte(`{"success":true,"data":null}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	s := &JamaahService{httpc: httpclient.New(), invoiceAddr: strings.TrimPrefix(ts.URL, "http://")}
	s.cancelDanglingInvoices(context.Background(), jamaahID, packageID, "Bearer x")

	if !sawCancel {
		t.Error("expected the dangling invoice to be cancelled")
	}
	if !sawRefund {
		t.Error("expected a refund to be initiated for the amount already paid")
	}
}

func TestCancelDanglingInvoicesSkipsAlreadyCancelledAndOtherPackages(t *testing.T) {
	jamaahID := uuid.New()
	packageID := uuid.New()
	otherPackageID := uuid.New()
	var writeCalls int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			list, _ := json.Marshal([]map[string]any{
				{"id": uuid.New(), "package_id": packageID, "status": "batal", "amount_paid": 100000, "amount_remaining": 0},
				{"id": uuid.New(), "package_id": otherPackageID, "status": "belum_bayar", "amount_paid": 0, "amount_remaining": 500000},
			})
			_, _ = w.Write([]byte(`{"success":true,"data":` + string(list) + `}`))
			return
		}
		writeCalls++
		_, _ = w.Write([]byte(`{"success":true,"data":null}`))
	}))
	defer ts.Close()

	s := &JamaahService{httpc: httpclient.New(), invoiceAddr: strings.TrimPrefix(ts.URL, "http://")}
	s.cancelDanglingInvoices(context.Background(), jamaahID, packageID, "Bearer x")

	if writeCalls != 0 {
		t.Errorf("expected no cancel/refund calls (one invoice already batal, one for a different package), got %d", writeCalls)
	}
}

func TestCancelDanglingInvoicesNoAuthTokenIsNoop(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Write([]byte(`{"success":true,"data":[]}`))
	}))
	defer ts.Close()

	s := &JamaahService{httpc: httpclient.New(), invoiceAddr: strings.TrimPrefix(ts.URL, "http://")}
	s.cancelDanglingInvoices(context.Background(), uuid.New(), uuid.New(), "")

	if called {
		t.Error("expected no HTTP call when authToken is empty")
	}
}
