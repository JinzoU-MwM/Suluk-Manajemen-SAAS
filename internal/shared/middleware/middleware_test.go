package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestAuthMiddleware_RejectsMissingAndBadAuth(t *testing.T) {
	app := fiber.New()
	app.Use(AuthMiddleware(nil))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	cases := []struct {
		name   string
		header string
		want   int
	}{
		{"missing header", "", 401},
		{"wrong format", "Token abc", 401},
		{"nil manager with bearer", "Bearer abc", 500},
	}
	for _, tc := range cases {
		req := httptest.NewRequest("GET", "/", nil)
		if tc.header != "" {
			req.Header.Set("Authorization", tc.header)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if resp.StatusCode != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, resp.StatusCode, tc.want)
		}
	}
}

func TestGetClaims_MissingReturnsFalse(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		if _, ok := GetClaims(c); ok {
			t.Error("expected GetClaims to report missing claims")
		}
		return c.SendString("ok")
	})
	if _, err := app.Test(httptest.NewRequest("GET", "/", nil)); err != nil {
		t.Fatal(err)
	}
}
