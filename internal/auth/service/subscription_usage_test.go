package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/shared/httpclient"
	"github.com/jamaah-in/v2/internal/shared/plan"
)

func addrOf(url string) string { return strings.TrimPrefix(url, "http://") }

// statusResponse must expose the tier's monthly AI-scan quota as UsageLimit so
// the frontend quota bar (which reads subscription.usage_limit) can render.
func TestStatusResponseScanLimit(t *testing.T) {
	cases := map[string]int{
		plan.Gratis:     5,
		plan.Starter:    100,
		plan.Pro:        plan.Unlimited,
		plan.Bisnis:     plan.Unlimited,
		plan.Enterprise: plan.Unlimited,
	}
	for key, want := range cases {
		if got := statusResponse(key, "active", nil).UsageLimit; got != want {
			t.Errorf("%s UsageLimit = %d, want %d", key, got, want)
		}
	}
}

// scanUsageThisMonth fetches the org's monthly count from ai-ocr, forwarding the
// shared internal key, and caches the result so repeat status polls don't hit
// ai-ocr every time.
func TestScanUsageThisMonthFetchesAndCaches(t *testing.T) {
	var hits int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if r.Header.Get("X-Internal-Key") != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"documents_scanned":7,"purchased_scans":100}}`))
	}))
	defer ts.Close()

	s := &AuthService{httpc: httpclient.New(), aiocrAddr: addrOf(ts.URL), internalKey: "secret"}
	org := uuid.New()

	used, purchased := s.scanUsageThisMonth(context.Background(), org)
	if used != 7 || purchased != 100 {
		t.Fatalf("got (used=%d, purchased=%d), want (7, 100)", used, purchased)
	}
	used2, _ := s.scanUsageThisMonth(context.Background(), org)
	if used2 != 7 {
		t.Fatalf("cached used = %d, want 7", used2)
	}
	if n := atomic.LoadInt32(&hits); n != 1 {
		t.Errorf("ai-ocr hits = %d, want 1 (second call should be cached)", n)
	}
}

// A failing/unreachable ai-ocr must never break subscription status: fail open
// with a zero count.
func TestScanUsageThisMonthFailsOpen(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := &AuthService{httpc: httpclient.New(), aiocrAddr: addrOf(ts.URL), internalKey: "secret"}
	if used, purchased := s.scanUsageThisMonth(context.Background(), uuid.New()); used != 0 || purchased != 0 {
		t.Errorf("fail-open got (%d,%d), want (0,0)", used, purchased)
	}
}

// With no ai-ocr address/key configured (e.g. local dev), usage is reported as 0
// without attempting a call.
func TestScanUsageThisMonthUnconfigured(t *testing.T) {
	s := &AuthService{httpc: httpclient.New()}
	if used, purchased := s.scanUsageThisMonth(context.Background(), uuid.New()); used != 0 || purchased != 0 {
		t.Errorf("unconfigured got (%d,%d), want (0,0)", used, purchased)
	}
}

func TestUsageLimitAddsPurchased(t *testing.T) {
	// Starter base 100 + 100 purchased = 200; Pro stays Unlimited.
	if got := effectiveLimit(plan.Get(plan.Starter).MaxScansPerMonth, 100); got != 200 {
		t.Errorf("starter effective limit = %d, want 200", got)
	}
	if got := effectiveLimit(plan.Get(plan.Pro).MaxScansPerMonth, 100); got != plan.Unlimited {
		t.Errorf("pro effective limit = %d, want Unlimited", got)
	}
}
