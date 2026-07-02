package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/shared/httpclient"
)

func TestDaysUntilDepartureNoAddrConfigured(t *testing.T) {
	s := &RefundService{httpc: httpclient.New()}
	_, ok := s.daysUntilDeparture(context.Background(), uuid.New(), "Bearer x")
	if ok {
		t.Fatal("expected ok=false when packageAddr is unset")
	}
}

func TestDaysUntilDepartureNoAuthToken(t *testing.T) {
	s := &RefundService{httpc: httpclient.New(), packageAddr: "localhost:1"}
	_, ok := s.daysUntilDeparture(context.Background(), uuid.New(), "")
	if ok {
		t.Fatal("expected ok=false when authToken is empty")
	}
}

func TestDaysUntilDepartureReadsPackageServiceResponse(t *testing.T) {
	departure := time.Now().Add(35 * 24 * time.Hour)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"departure_date":"` + departure.Format(time.RFC3339) + `"}}`))
	}))
	defer ts.Close()

	s := &RefundService{httpc: httpclient.New(), packageAddr: strings.TrimPrefix(ts.URL, "http://")}
	days, ok := s.daysUntilDeparture(context.Background(), uuid.New(), "Bearer x")
	if !ok {
		t.Fatal("expected ok=true when package-service returns a departure_date")
	}
	if days < 34 || days > 35 {
		t.Errorf("days = %d, want ~35", days)
	}
}

func TestDaysUntilDepartureNoDepartureDateSet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{}}`))
	}))
	defer ts.Close()

	s := &RefundService{httpc: httpclient.New(), packageAddr: strings.TrimPrefix(ts.URL, "http://")}
	_, ok := s.daysUntilDeparture(context.Background(), uuid.New(), "Bearer x")
	if ok {
		t.Fatal("expected ok=false when the package has no departure_date")
	}
}
