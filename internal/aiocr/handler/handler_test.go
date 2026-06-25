package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/jamaah-in/v2/internal/aiocr/service"
)

// newTestApp builds a fiber app exposing only the internal scan-usage route,
// backed by a service with a nil repo (GetScanUsageThisMonth returns 0 without a
// DB), so the handler's guard + parsing are testable in isolation.
func newTestApp() *fiber.App {
	h := NewAIOCRHandler(service.NewAIOCRService(nil, nil, zap.NewNop().Sugar()))
	app := fiber.New()
	app.Post("/api/v1/internal/scan-usage", h.ScanUsageInternal)
	return app
}

func TestScanUsageInternalRejectsMissingKey(t *testing.T) {
	t.Setenv("INTERNAL_API_KEY", "testkey")
	app := newTestApp()

	body, _ := json.Marshal(map[string]string{"org_id": uuid.NewString()})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/scan-usage", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// no X-Internal-Key header

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestScanUsageInternalBadBody(t *testing.T) {
	t.Setenv("INTERNAL_API_KEY", "testkey")
	app := newTestApp()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/scan-usage", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", "testkey")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestScanUsageInternalReturnsCount(t *testing.T) {
	t.Setenv("INTERNAL_API_KEY", "testkey")
	app := newTestApp()

	body, _ := json.Marshal(map[string]string{"org_id": uuid.NewString()})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/scan-usage", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", "testkey")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var out struct {
		Data struct {
			DocumentsScanned int `json:"documents_scanned"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Data.DocumentsScanned != 0 {
		t.Errorf("documents_scanned = %d, want 0 (nil repo)", out.Data.DocumentsScanned)
	}
}

func TestScanTopupInternalRejectsMissingKey(t *testing.T) {
	t.Setenv("INTERNAL_API_KEY", "testkey")
	h := NewAIOCRHandler(service.NewAIOCRService(nil, nil, zap.NewNop().Sugar()))
	app := fiber.New()
	app.Post("/api/v1/internal/scan-topup", h.ScanTopupInternal)

	body, _ := json.Marshal(map[string]any{"order_id": uuid.NewString(), "org_id": uuid.NewString(), "scans": 100})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/scan-topup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestScanTopupInternalCreditsWithValidKey(t *testing.T) {
	t.Setenv("INTERNAL_API_KEY", "testkey")
	h := NewAIOCRHandler(service.NewAIOCRService(nil, nil, zap.NewNop().Sugar()))
	app := fiber.New()
	app.Post("/api/v1/internal/scan-topup", h.ScanTopupInternal)

	body, _ := json.Marshal(map[string]any{"order_id": uuid.NewString(), "org_id": uuid.NewString(), "scans": 100})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/scan-topup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", "testkey")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK { // nil repo → credit is a no-op success
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}
