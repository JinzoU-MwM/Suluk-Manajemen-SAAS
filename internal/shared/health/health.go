// Package health provides a /health handler that verifies real dependencies
// (database, redis, ...) instead of returning a static "ok".
package health

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Check is a named dependency probe. Ping should return nil when healthy.
type Check struct {
	Name string
	Ping func(ctx context.Context) error
}

// Handler returns a Fiber handler that runs all checks (2s budget) and returns
// 503 if any dependency is down, 200 otherwise.
func Handler(service string, checks ...Check) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		results := fiber.Map{}
		healthy := true
		for _, ck := range checks {
			if err := ck.Ping(ctx); err != nil {
				results[ck.Name] = "down"
				healthy = false
			} else {
				results[ck.Name] = "ok"
			}
		}

		status := fiber.StatusOK
		state := "ok"
		if !healthy {
			status = fiber.StatusServiceUnavailable
			state = "degraded"
		}
		return c.Status(status).JSON(fiber.Map{
			"status":  state,
			"service": service,
			"checks":  results,
		})
	}
}
