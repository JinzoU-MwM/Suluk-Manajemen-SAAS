package health

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func run(t *testing.T, checks ...Check) int {
	t.Helper()
	app := fiber.New()
	app.Get("/health", Handler("test", checks...))
	resp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}

func TestHealth_OK(t *testing.T) {
	if got := run(t, Check{Name: "db", Ping: func(context.Context) error { return nil }}); got != 200 {
		t.Errorf("got %d, want 200", got)
	}
}

func TestHealth_Down(t *testing.T) {
	if got := run(t, Check{Name: "db", Ping: func(context.Context) error { return errors.New("down") }}); got != 503 {
		t.Errorf("got %d, want 503", got)
	}
}
