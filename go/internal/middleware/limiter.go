package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"gonote/internal/models/config"
)

// RateLimiter creates a global rate limiter based on config settings
func RateLimiter(cfg *config.Config) fiber.Handler {
	if cfg == nil || !cfg.RateLimit.Enabled {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return limiter.New(limiter.Config{
		Max:        cfg.RateLimit.MaxRequests,
		Expiration: time.Duration(cfg.RateLimit.WindowSeconds) * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"detail": "Rate limit exceeded",
			})
		},
	})
}

// EndpointLimiter creates a rate limiter for specific endpoints
func EndpointLimiter(max int, durationSeconds int) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: time.Duration(durationSeconds) * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() + ":" + c.Path()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"detail": "Rate limit exceeded",
			})
		},
	})
}

// EndpointLimiterSimple creates a rate limiter with default duration (60 seconds)
func EndpointLimiterSimple(max int) fiber.Handler {
	return EndpointLimiter(max, 60)
}
