package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS creates a CORS middleware with the specified allowed origins
func CORS(allowedOrigins []string) fiber.Handler {
	// Check if wildcard is used
	isWildcard := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"

	return cors.New(cors.Config{
		AllowOrigins:     strings.Join(allowedOrigins, ","),
		AllowCredentials: !isWildcard, // Cannot use credentials with wildcard
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-CSRF-Token",
		MaxAge:           86400, // 24 hours
	})
}
