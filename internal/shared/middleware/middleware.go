// Package middleware provides shared Fiber middleware used by every service:
// JWT auth, safe claims access, request IDs, and structured request logging.
package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

const claimsKey = "claims"

// AuthMiddleware validates the Bearer JWT and stores the claims in c.Locals.
// It always either sets claims and calls Next, or returns a 401/500 — so handlers
// downstream can rely on GetClaims succeeding.
func AuthMiddleware(jwtMgr *sharedAuth.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, "missing authorization header")
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			return response.Unauthorized(c, "invalid authorization format")
		}
		if jwtMgr == nil {
			return response.InternalError(c, "authentication is not configured")
		}
		claims, err := jwtMgr.ValidateToken(tokenStr)
		if err != nil {
			return response.Unauthorized(c, "invalid or expired token")
		}
		c.Locals(claimsKey, claims)
		return c.Next()
	}
}

// GetClaims safely retrieves the authenticated claims set by AuthMiddleware.
func GetClaims(c *fiber.Ctx) (*sharedAuth.Claims, bool) {
	v := c.Locals(claimsKey)
	if v == nil {
		return nil, false
	}
	claims, ok := v.(*sharedAuth.Claims)
	return claims, ok
}

// RequireRole returns middleware that allows the request only if the
// authenticated user's role is one of the given roles (403 otherwise). Must run
// after AuthMiddleware. Use to gate write/management routes by role.
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *fiber.Ctx) error {
		claims, ok := GetClaims(c)
		if !ok {
			return response.Unauthorized(c, "unauthorized")
		}
		if !allowed[claims.Role] {
			return response.Forbidden(c, "akses ditolak untuk peran Anda")
		}
		return c.Next()
	}
}

// RequireStaff blocks external agents (role "agent") from staff/back-office
// routes. External agents are confined to the /b2b portal; every other route
// group should run this after AuthMiddleware so an agent token can't reach
// org-wide staff data by calling the legacy endpoints directly.
func RequireStaff(c *fiber.Ctx) error {
	claims, ok := GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if claims.Role == "agent" {
		return response.Forbidden(c, "akses ditolak: gunakan portal agen")
	}
	return c.Next()
}

// RequireAgentScope gates the B2B portal: the caller must be an external agent
// (role "agent") with a linked AgentID claim. Must run after AuthMiddleware.
// Handlers downstream can rely on claims.AgentID being non-nil.
func RequireAgentScope(c *fiber.Ctx) error {
	claims, ok := GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	if claims.Role != "agent" || claims.AgentID == nil {
		return response.Forbidden(c, "portal agen hanya untuk akun agen")
	}
	return c.Next()
}

// RequireClaims returns the claims or writes a 401 and returns ok=false.
func RequireClaims(c *fiber.Ctx) (*sharedAuth.Claims, error) {
	claims, ok := GetClaims(c)
	if !ok {
		return nil, response.Unauthorized(c, "unauthorized")
	}
	return claims, nil
}

// RequestID assigns/propagates an X-Request-ID for cross-service correlation.
func RequestID() fiber.Handler {
	return requestid.New()
}

// RequestLogger logs one structured line per request (method, path, status, latency, id).
func RequestLogger(logger *zap.SugaredLogger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		rid, _ := c.Locals(requestid.ConfigDefault.ContextKey).(string)
		logger.Infow("request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency_ms", time.Since(start).Milliseconds(),
			"request_id", rid,
			"ip", c.IP(),
		)
		return err
	}
}
