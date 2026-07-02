package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/jamaah-in/v2/internal/shared/httpclient"
	"github.com/jamaah-in/v2/internal/shared/plan"
)

func addrOf(u string) string { return strings.TrimPrefix(u, "http://") }

func TestFetchQuotaFailsOpenWithoutAuthAddr(t *testing.T) {
	s := &AIOCRService{httpc: httpclient.New()}
	used, limit := s.fetchQuota(context.Background(), uuid.New(), "Bearer tok")
	if used != 0 || limit != plan.Unlimited {
		t.Errorf("fetchQuota() = (%d, %d), want (0, Unlimited)", used, limit)
	}
}

func TestFetchQuotaFailsOpenWithoutAuthToken(t *testing.T) {
	s := &AIOCRService{httpc: httpclient.New(), authAddr: "auth-service:50051"}
	used, limit := s.fetchQuota(context.Background(), uuid.New(), "")
	if used != 0 || limit != plan.Unlimited {
		t.Errorf("fetchQuota() = (%d, %d), want (0, Unlimited)", used, limit)
	}
}

func TestFetchQuotaReadsUsageAndLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/subscription/status" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"usage_count":5,"usage_limit":100}}`))
	}))
	defer ts.Close()

	s := &AIOCRService{httpc: httpclient.New(), authAddr: addrOf(ts.URL)}
	used, limit := s.fetchQuota(context.Background(), uuid.New(), "Bearer tok")
	if used != 5 || limit != 100 {
		t.Errorf("fetchQuota() = (%d, %d), want (5, 100)", used, limit)
	}
}

func TestFetchQuotaFailsOpenOnServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := &AIOCRService{httpc: httpclient.New(), authAddr: addrOf(ts.URL)}
	used, limit := s.fetchQuota(context.Background(), uuid.New(), "Bearer tok")
	if used != 0 || limit != plan.Unlimited {
		t.Errorf("fetchQuota() = (%d, %d), want (0, Unlimited) on server error", used, limit)
	}
}

func TestFetchQuotaCachesWithinTTL(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"success":true,"data":{"usage_count":1,"usage_limit":5}}`))
	}))
	defer ts.Close()

	s := &AIOCRService{httpc: httpclient.New(), authAddr: addrOf(ts.URL)}
	orgID := uuid.New()
	s.fetchQuota(context.Background(), orgID, "Bearer tok")
	s.fetchQuota(context.Background(), orgID, "Bearer tok")
	if calls != 1 {
		t.Errorf("expected 1 upstream call (second served from cache), got %d", calls)
	}
}

// TestProcessDocumentsSyncBlocksAtQuota is the AIOCR-1 regression: once an org
// has used its full monthly quota, no OCR call may be made at all.
func TestProcessDocumentsSyncBlocksAtQuota(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":{"usage_count":5,"usage_limit":5}}`))
	}))
	defer ts.Close()

	fa := &fakeAnalyzer{}
	svc := &AIOCRService{analyzer: fa, logger: zap.NewNop().Sugar(), httpc: httpclient.New(), authAddr: addrOf(ts.URL)}
	files := []SyncFile{{FileName: "a.png", ContentType: "image/png", Data: []byte("x")}}

	_, err := svc.ProcessDocumentsSync(context.Background(), uuid.New(), files, "default", "Bearer tok")
	if err == nil {
		t.Fatal("expected ErrScanQuotaExceeded, got nil")
	}
	if !strings.Contains(err.Error(), "kuota") {
		t.Errorf("error = %v, want quota-exceeded message", err)
	}
	if fa.inFlight != 0 || fa.maxSeen != 0 {
		t.Errorf("analyzer should never have been called once quota is exhausted, maxSeen=%d", fa.maxSeen)
	}
}
