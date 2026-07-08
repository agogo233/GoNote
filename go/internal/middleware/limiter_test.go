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
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled: false,
		},
	}

	app := fiber.New()
	app.Use(RateLimiter(cfg))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

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

	app := fiber.New()
	app.Use(RateLimiter(cfg))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
}

func TestRateLimiterNilConfig(t *testing.T) {
	app := fiber.New()
	app.Use(RateLimiter(nil))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

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

	app := fiber.New()
	app.Use(RateLimiter(cfg))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)
	}

	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 429, resp.StatusCode)
}

func TestEndpointLimiterWithGlobalDisabled(t *testing.T) {
	app := fiber.New()
	app.Use(EndpointLimiter(3, 1))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
}

func TestEndpointLimiterEnabled(t *testing.T) {
	app := fiber.New()
	app.Use(EndpointLimiter(2, 1))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
}

func TestEndpointLimiterPerPath(t *testing.T) {
	app := fiber.New()
	app.Use(EndpointLimiter(1, 10))
	app.Get("/path1", func(c *fiber.Ctx) error {
		return c.SendString("path1")
	})
	app.Get("/path2", func(c *fiber.Ctx) error {
		return c.SendString("path2")
	})

	req := httptest.NewRequest("GET", "/path1", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)

	req1 := httptest.NewRequest("GET", "/path1", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	resp1, _ := app.Test(req1)
	assert.Equal(t, 429, resp1.StatusCode)

	req2 := httptest.NewRequest("GET", "/path2", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	resp2, _ := app.Test(req2)
	assert.Equal(t, 200, resp2.StatusCode)
}

func TestEndpointLimiterSimple(t *testing.T) {
	app := fiber.New()
	app.Use(EndpointLimiterSimple(2))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}

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

	app := fiber.New()
	app.Use(RateLimiter(cfg))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	app.Test(req)

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	resp, _ := app.Test(req2)
	assert.Equal(t, 429, resp.StatusCode)

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

	app := fiber.New()
	app.Use(RateLimiter(cfg))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	app.Test(req)

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	resp, _ := app.Test(req2)
	assert.Equal(t, 429, resp.StatusCode)

	time.Sleep(1100 * time.Millisecond)

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

	app := fiber.New(fiber.Config{
		ProxyHeader:        "X-Forwarded-For",
		EnableIPValidation: false,
	})
	app.Use(RateLimiter(cfg))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("X-Forwarded-For", "10.0.0.2")
	resp2, err := app.Test(req2)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)
}