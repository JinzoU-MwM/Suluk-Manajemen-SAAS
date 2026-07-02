package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/invoice/repository"
	"github.com/jamaah-in/v2/internal/invoice/service"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
)

func newRefundTestApp(repo *repository.InvoiceRepo) *fiber.App {
	h := NewRefundHandler(service.NewRefundService(repo))
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("claims", &sharedAuth.Claims{OrgID: uuid.New(), UserID: uuid.New()})
		return c.Next()
	})
	app.Get("/refunds/:id", h.GetRefund)
	app.Post("/invoices/:id/refund", h.InitiateRefund)
	return app
}

func TestGetRefundInvalidIDIsBadRequest(t *testing.T) {
	app := newRefundTestApp(nil)
	req := httptest.NewRequest(http.MethodGet, "/refunds/not-a-uuid", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestInitiateRefundZeroAmountIsBadRequest(t *testing.T) {
	app := newRefundTestApp(nil)
	body, _ := json.Marshal(map[string]any{"amount": 0, "refund_pct": 100, "reason": "x"})
	req := httptest.NewRequest(http.MethodPost, "/invoices/"+uuid.NewString()+"/refund", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}
