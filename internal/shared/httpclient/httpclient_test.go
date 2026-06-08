package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func addrOf(url string) string { return strings.TrimPrefix(url, "http://") }

func TestGetJSON_UnwrapsData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true,"data":{"x":7}}`))
	}))
	defer ts.Close()

	var out struct {
		X int `json:"x"`
	}
	if err := New().GetJSON(context.Background(), addrOf(ts.URL), "/", "", &out); err != nil {
		t.Fatal(err)
	}
	if out.X != 7 {
		t.Errorf("got %d, want 7", out.X)
	}
}

func TestGetRaw_RetriesOn5xxThenSucceeds(t *testing.T) {
	var calls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(500)
			return
		}
		w.Write([]byte(`{"success":true,"data":[]}`))
	}))
	defer ts.Close()

	if _, err := New().GetRaw(context.Background(), addrOf(ts.URL), "/", ""); err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if calls < 2 {
		t.Errorf("expected a retry (>=2 calls), got %d", calls)
	}
}

func TestGetRaw_NoRetryOn4xx(t *testing.T) {
	var calls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(404)
	}))
	defer ts.Close()

	if _, err := New().GetRaw(context.Background(), addrOf(ts.URL), "/", ""); err == nil {
		t.Fatal("expected error on 404")
	}
	if calls != 1 {
		t.Errorf("expected no retry on 4xx, got %d calls", calls)
	}
}
