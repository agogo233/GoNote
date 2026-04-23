package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"gonote/internal/models/config"
	"gonote/internal/middleware"
)

func init() {
	// Initialize session store for auth tests
	middleware.InitSessionStore("test-secret-key-for-auth-tests", 3600, false)
}

func TestAuthHandler_LoginPage(t *testing.T) {
	t.Run("redirects to home when auth disabled", func(t *testing.T) {
		cfg := &config.Config{
			Authentication: config.AuthConfig{
				Enabled: false,
			},
		}
		handler := NewAuthHandler(cfg)

		app := fiber.New()
		app.Get("/login", handler.LoginPage)

		req := httptest.NewRequest("GET", "/login", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 303, resp.StatusCode)
		assert.Equal(t, "/", resp.Header.Get("Location"))
	})

	t.Run("returns login page when auth enabled", func(t *testing.T) {
		cfg := &config.Config{
			Authentication: config.AuthConfig{
				Enabled: true,
			},
		}
		handler := NewAuthHandler(cfg)

		app := fiber.New()
		app.Get("/login", handler.LoginPage)

		req := httptest.NewRequest("GET", "/login", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "GoNote")
		assert.Contains(t, string(body), "Password")
	})
}

func TestAuthHandler_Login(t *testing.T) {
	// Create a hashed password for testing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	t.Run("returns success when auth disabled", func(t *testing.T) {
		cfg := &config.Config{
			Authentication: config.AuthConfig{
				Enabled: false,
			},
		}
		handler := NewAuthHandler(cfg)

		app := fiber.New()
		app.Post("/login", handler.Login)

		body, _ := json.Marshal(map[string]string{"password": "any"})
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["success"].(bool))
	})

	t.Run("returns success with correct password", func(t *testing.T) {
		cfg := &config.Config{
			Authentication: config.AuthConfig{
				Enabled:      true,
				PasswordHash: string(hashedPassword),
			},
		}
		handler := NewAuthHandler(cfg)

		app := fiber.New()
		app.Post("/login", handler.Login)

		body, _ := json.Marshal(map[string]string{"password": "test-password"})
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["success"].(bool))
	})

	t.Run("returns error with wrong password", func(t *testing.T) {
		cfg := &config.Config{
			Authentication: config.AuthConfig{
				Enabled:      true,
				PasswordHash: string(hashedPassword),
			},
		}
		handler := NewAuthHandler(cfg)

		app := fiber.New()
		app.Post("/login", handler.Login)

		body, _ := json.Marshal(map[string]string{"password": "wrong-password"})
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.False(t, result["success"].(bool))
		assert.Contains(t, result["detail"], "Invalid password")
	})

	t.Run("returns error with invalid request body", func(t *testing.T) {
		cfg := &config.Config{
			Authentication: config.AuthConfig{
				Enabled:      true,
				PasswordHash: string(hashedPassword),
			},
		}
		handler := NewAuthHandler(cfg)

		app := fiber.New()
		app.Post("/login", handler.Login)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("returns error when password not configured", func(t *testing.T) {
		cfg := &config.Config{
			Authentication: config.AuthConfig{
				Enabled:      true,
				PasswordHash: "",
			},
		}
		handler := NewAuthHandler(cfg)

		app := fiber.New()
		app.Post("/login", handler.Login)

		body, _ := json.Marshal(map[string]string{"password": "any"})
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	cfg := &config.Config{
		Authentication: config.AuthConfig{
			Enabled: true,
		},
	}
	handler := NewAuthHandler(cfg)

	t.Run("returns JSON for API requests", func(t *testing.T) {
		app := fiber.New()
		app.Post("/api/logout", handler.Logout)

		req := httptest.NewRequest("POST", "/api/logout", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["success"].(bool))
	})

	t.Run("redirects to login for page requests", func(t *testing.T) {
		app := fiber.New()
		app.Post("/logout", handler.Logout)

		req := httptest.NewRequest("POST", "/logout", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 303, resp.StatusCode)
		assert.Equal(t, "/login", resp.Header.Get("Location"))
	})
}

func TestNewAuthHandler(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name: "TestApp",
		},
	}
	handler := NewAuthHandler(cfg)

	assert.NotNil(t, handler)
	assert.Equal(t, cfg, handler.config)
}
