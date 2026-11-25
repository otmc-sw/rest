/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/

package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// LocalsKey is the fiber.Ctx.Locals key holding the verified *Claims.
const LocalsKey = "auth.claims"

// ClaimsFromCtx returns the verified claims set by RequireAuth, or nil.
func ClaimsFromCtx(c *fiber.Ctx) *Claims {
	if v, ok := c.Locals(LocalsKey).(*Claims); ok {
		return v
	}
	return nil
}

// RequireAuth is the one-line integration for resource servers:
//
//	app.Use(auth.RequireAuth(auth.Config{Authority: "http://localhost:5007"}))
//
// It builds a Provider once (discovery/JWKS are lazy) and verifies the
// Authorization: Bearer token on every request. On success the verified
// *Claims are available via ClaimsFromCtx / c.Locals(auth.LocalsKey).
func RequireAuth(cfg Config) fiber.Handler {
	p := NewProvider(cfg)
	return p.Middleware()
}

// Middleware returns the Fiber handler for this provider.
func (p *Provider) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             "invalid_token",
				"error_description": "Missing bearer token",
			})
		}
		claims, err := p.Verify(c.Context(), strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             "invalid_token",
				"error_description": "Invalid or expired token",
			})
		}
		c.Locals(LocalsKey, claims)
		return c.Next()
	}
}
