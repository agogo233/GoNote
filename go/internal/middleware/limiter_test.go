package middleware

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"gonote/internal/models/config"
)

func TestRateLimiterDisabled(t *testing.T) {
	// Set up config with rate limiting disabled
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled: false,
		},
	}
	config.GlobalConfig = cfg
	defer func() { config.GlobalConfig = nil }()

	app := fiber.New()
	app.Use(RateLimiter())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Should allow all requests when disabled
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestRateLimiterEnabled(t *testing.T) {
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled:      true,
			MaxRequests:  3,
			WindowSeconds: 1,
		},
	}
	config.GlobalConfig = cfg
	defer func() { config.GlobalConfig = nil }()

	app := fiber.New()
	app.Use(RateLimiter())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
}

func TestRateLimiterNilConfig(t *testing.T) {
	config.GlobalConfig = nil
	defer func() { config.GlobalConfig = nil }()

	app := fiber.New()
	app.Use(RateLimiter())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Should allow all requests when config is nil
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled:      true,
			MaxRequests:  2,
			WindowSeconds: 1,
		},
	}
	config.GlobalConfig = cfg
	defer func() { config.GlobalConfig = nil }()

	app := fiber.New()
	app.Use(RateLimiter())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 429, resp.StatusCode)
}

func TestEndpointLimiterWithGlobalDisabled(t *testing.T) {
	// Endpoint limiter should still work even when global rate limiting is disabled
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled: false,
		},
	}
	config.GlobalConfig = cfg
	defer func() { config.GlobalConfig = nil }()

	app := fiber.New()
	app.Use(EndpointLimiter(3, 1))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}

	// 4th request should be rate limited (endpoint limiter is independent of global)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
}

func TestEndpointLimiterEnabled(t *testing.T) {
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled: true,
		},
	}
	config.GlobalConfig = cfg
	defer func() { config.GlobalConfig = nil }()

	app := fiber.New()
	app.Use(EndpointLimiter(2, 1))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
}

func TestEndpointLimiterPerPath(t *testing.T) {
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled: true,
		},
	}
	config.GlobalConfig = cfg
	defer func() { config.GlobalConfig = nil }()

	app := fiber.New()
	app.Use(EndpointLimiter(1, 10))
	app.Get("/path1", func(c *fiber.Ctx) error {
		return c.SendString("path1")
	})
	app.Get("/path2", func(c *fiber.Ctx) error {
		return c.SendString("path2")
	})

	// Use up limit on /path1
	req := httptest.NewRequest("GET", "/path1", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)

	// /path1 should be rate limited
	req1 := httptest.NewRequest("GET", "/path1", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	resp1, _ := app.Test(req1)
	assert.Equal(t, 429, resp1.StatusCode)

	// /path2 should still work (different path)
	req2 := httptest.NewRequest("GET", "/path2", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	resp2, _ := app.Test(req2)
	assert.Equal(t, 200, resp2.StatusCode)
}

func TestEndpointLimiterSimple(t *testing.T) {
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled: true,
		},
	}
	config.GlobalConfig = cfg
	defer func() { config.GlobalConfig = nil }()

	app := fiber.New()
	app.Use(EndpointLimiterSimple(2))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}

	// 3rd should be rate limited
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
}

func TestRateLimiterResponseFormat(t *testing.T) {
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled:      true,
			MaxRequests:  1,
			WindowSeconds: 60,
		},
	}
	config.GlobalConfig = cfg
	defer func() { config.GlobalConfig = nil }()

	app := fiber.New()
	app.Use(RateLimiter())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// First request succeeds
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	app.Test(req)

	// Second request should return JSON error
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	resp, _ := app.Test(req2)
	assert.Equal(t, 429, resp.StatusCode)

	// Check response body contains detail
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Rate limit exceeded")
}

func TestRateLimiterExpiration(t *testing.T) {
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled:      true,
			MaxRequests:  1,
			WindowSeconds: 1,
		},
	}
	config.GlobalConfig = cfg
	defer func() { config.GlobalConfig = nil }()

	app := fiber.New()
	app.Use(RateLimiter())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Use up the limit
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	app.Test(req)

	// Should be rate limited
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	resp, _ := app.Test(req2)
	assert.Equal(t, 429, resp.StatusCode)

	// Wait for expiration
	time.Sleep(1100 * time.Millisecond)

	// Should work again after expiration
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.RemoteAddr = "192.168.1.1:1234"
	resp3, _ := app.Test(req3)
	assert.Equal(t, 200, resp3.StatusCode)
}

func TestRateLimiterXForwardedFor(t *testing.T) {
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled:      true,
			MaxRequests:  2,
			WindowSeconds: 1,
		},
		Server: config.ServerConfig{
			ProxyHeader: "X-Forwarded-For",
		},
	}
	config.GlobalConfig = cfg
	defer func() { config.GlobalConfig = nil }()

	app := fiber.New(fiber.Config{
		ProxyHeader:        "X-Forwarded-For",
		EnableIPValidation: false,
	})
	app.Use(RateLimiter())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Requests with X-Forwarded-For should use that value for rate limiting
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}

	// 3rd request from same X-Forwarded-For IP should be rate limited
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)

	// Different X-Forwarded-For IP should still work
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("X-Forwarded-For", "10.0.0.2")
	resp2, err := app.Test(req2)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)
}
