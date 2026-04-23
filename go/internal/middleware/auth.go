package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthRequired creates an authentication middleware
func AuthRequired(authEnabled bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If auth is disabled, allow all requests
		if !authEnabled {
			return c.Next()
		}

		// Check if user is authenticated
		sess := GetSession(c)
		authenticated := sess.Get("authenticated")

		if authenticated == nil || authenticated != true {
			// For API requests, return JSON error
			if strings.HasPrefix(c.Path(), "/api/") {
				return c.Status(401).JSON(fiber.Map{
					"detail": "Not authenticated",
				})
			}
			// For page requests, redirect to login
			return c.Redirect("/login", 303)
		}

		return c.Next()
	}
}

// WSAuthRequired creates an authentication middleware for WebSocket connections.
// Returns 401 if authentication is enabled but user is not authenticated.
func WSAuthRequired(authEnabled bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If auth is disabled, allow all connections
		if !authEnabled {
			return c.Next()
		}

		// Check if user is authenticated
		sess := GetSession(c)
		authenticated := sess.Get("authenticated")

		if authenticated == nil || authenticated != true {
			// Return 401 for unauthorized WebSocket upgrade attempts
			return c.Status(401).JSON(fiber.Map{
				"detail": "Not authenticated",
			})
		}

		return c.Next()
	}
}

// IsAuthenticated checks if the current session is authenticated.
// Can be used to check auth status without blocking the request.
func IsAuthenticated(c *fiber.Ctx) bool {
	sess := GetSession(c)
	authenticated := sess.Get("authenticated")
	return authenticated == true
}
