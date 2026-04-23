package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func init() {
	// Initialize session store for tests
	InitSessionStore("test-secret-key-for-testing", 3600, false)
}

func TestAuthRequired_Disabled(t *testing.T) {
	app := fiber.New(fiber.Config{
		AppName: "GoNote Test",
	})
	app.Use(AuthRequired(false))
	app.Get("/api/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestAuthRequired_API_Unauthenticated(t *testing.T) {
	app := fiber.New(fiber.Config{
		AppName: "GoNote Test",
	})
	app.Use(AuthRequired(true))
	app.Get("/api/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != 401 {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthRequired_Page_Redirect(t *testing.T) {
	app := fiber.New(fiber.Config{
		AppName: "GoNote Test",
	})
	app.Use(AuthRequired(true))
	app.Get("/notes", func(c *fiber.Ctx) error {
		return c.SendString("notes")
	})

	req := httptest.NewRequest("GET", "/notes", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// Should redirect to login (303 See Other)
	if resp.StatusCode != 303 {
		t.Errorf("Expected status 303, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location != "/login" {
		t.Errorf("Expected redirect to /login, got %s", location)
	}
}

func TestAuthRequired_API_Prefix_Match(t *testing.T) {
	// Test that only /api/ prefix returns JSON error
	app := fiber.New(fiber.Config{
		AppName: "GoNote Test",
	})
	app.Use(AuthRequired(true))

	// API routes should return 401 JSON
	app.Get("/api/notes", func(c *fiber.Ctx) error {
		return c.SendString("notes")
	})

	// Non-API routes should redirect
	app.Get("/settings", func(c *fiber.Ctx) error {
		return c.SendString("settings")
	})

	// Test API route
	req := httptest.NewRequest("GET", "/api/notes", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 401 {
		t.Errorf("API route expected 401, got %d", resp.StatusCode)
	}

	// Test non-API route
	req = httptest.NewRequest("GET", "/settings", nil)
	resp, _ = app.Test(req, -1)
	if resp.StatusCode != 303 {
		t.Errorf("Non-API route expected 303, got %d", resp.StatusCode)
	}
}

func TestAuthRequired_Enabled_Blocks_Request(t *testing.T) {
	app := fiber.New(fiber.Config{
		AppName: "GoNote Test",
	})
	app.Use(AuthRequired(true))
	app.Get("/api/protected", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"secret": "data"})
	})

	req := httptest.NewRequest("GET", "/api/protected", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != 401 {
		t.Errorf("Expected 401 for protected route, got %d", resp.StatusCode)
	}
}

func TestAuthRequired_Multiple_Routes(t *testing.T) {
	app := fiber.New(fiber.Config{
		AppName: "GoNote Test",
	})
	app.Use(AuthRequired(true))

	app.Get("/api/notes", func(c *fiber.Ctx) error { return c.SendString("notes") })
	app.Post("/api/notes", func(c *fiber.Ctx) error { return c.SendString("created") })
	app.Delete("/api/notes/1", func(c *fiber.Ctx) error { return c.SendString("deleted") })

	// All methods should be blocked
	methods := []string{"GET", "POST", "DELETE"}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/notes", nil)
		if method == "DELETE" {
			req = httptest.NewRequest(method, "/api/notes/1", nil)
		}
		resp, _ := app.Test(req, -1)
		if resp.StatusCode != 401 {
			t.Errorf("Method %s expected 401, got %d", method, resp.StatusCode)
		}
	}
}

func TestAuthRequired_Static_Routes(t *testing.T) {
	// Static routes should not be affected by auth middleware if placed after
	app := fiber.New(fiber.Config{
		AppName: "GoNote Test",
	})

	// Place auth middleware only on API routes
	app.Use("/api", AuthRequired(true))

	// Static route should be accessible
	app.Get("/static/file.js", func(c *fiber.Ctx) error {
		return c.SendString("js content")
	})

	// API route should be blocked
	app.Get("/api/data", func(c *fiber.Ctx) error {
		return c.SendString("data")
	})

	// Test static route (should pass)
	req := httptest.NewRequest("GET", "/static/file.js", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 200 {
		t.Errorf("Static route expected 200, got %d", resp.StatusCode)
	}

	// Test API route (should be blocked)
	req = httptest.NewRequest("GET", "/api/data", nil)
	resp, _ = app.Test(req, -1)
	if resp.StatusCode != 401 {
		t.Errorf("API route expected 401, got %d", resp.StatusCode)
	}
}
