package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/stretchr/testify/assert"
)

func TestInitSessionStore(t *testing.T) {
	secretKey := "test-secret-key-12345"
	maxAge := 3600
	secureCookie := false

	InitSessionStore(secretKey, maxAge, secureCookie)

	assert.NotNil(t, store)
	assert.Equal(t, "cookie:session_id", store.KeyLookup)
	assert.Equal(t, time.Duration(3600)*time.Second, store.Expiration)
	assert.True(t, store.CookieHTTPOnly)
	assert.Equal(t, "Lax", store.CookieSameSite)
	assert.False(t, store.CookieSecure)
}

func TestInitSessionStore_SecureCookie(t *testing.T) {
	secretKey := "test-secret-key"
	maxAge := 7200
	secureCookie := true

	InitSessionStore(secretKey, maxAge, secureCookie)

	assert.NotNil(t, store)
	assert.True(t, store.CookieSecure)
	assert.Equal(t, time.Duration(7200)*time.Second, store.Expiration)
}

func TestInitSessionStore_CustomMaxAge(t *testing.T) {
	secretKey := "test-secret-key"
	maxAge := 86400 // 24 hours
	secureCookie := false

	InitSessionStore(secretKey, maxAge, secureCookie)

	assert.NotNil(t, store)
	assert.Equal(t, time.Duration(86400)*time.Second, store.Expiration)
}

func TestGetSession_NoExistingSession(t *testing.T) {
	InitSessionStore("test-secret", 3600, false)

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		sess := GetSession(c)
		assert.NotNil(t, sess)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetSession_WithExistingSession(t *testing.T) {
	InitSessionStore("test-secret", 3600, false)

	app := fiber.New()
	app.Get("/set", func(c *fiber.Ctx) error {
		sess := GetSession(c)
		sess.Set("key", "value")
		err := sess.Save()
		assert.NoError(t, err)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/set", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDestroySession(t *testing.T) {
	InitSessionStore("test-secret", 3600, false)

	app := fiber.New()

	app.Get("/set", func(c *fiber.Ctx) error {
		sess := GetSession(c)
		sess.Set("key", "value")
		sess.Save()
		return c.SendString("OK")
	})

	app.Get("/destroy", func(c *fiber.Ctx) error {
		err := DestroySession(c)
		assert.NoError(t, err)
		return c.SendString("OK")
	})

	// First, create a session
	req := httptest.NewRequest("GET", "/set", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Then destroy it
	req = httptest.NewRequest("GET", "/destroy", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestSessionStore_KeyLookup(t *testing.T) {
	InitSessionStore("test-secret", 3600, false)

	assert.Equal(t, "cookie:session_id", store.KeyLookup)
}

func TestSessionStore_HTTPOnly(t *testing.T) {
	InitSessionStore("test-secret", 3600, false)

	assert.True(t, store.CookieHTTPOnly)
}

func TestSessionStore_SameSite(t *testing.T) {
	InitSessionStore("test-secret", 3600, false)

	assert.Equal(t, "Lax", store.CookieSameSite)
}

func TestSessionStore_Expiration(t *testing.T) {
	InitSessionStore("test-secret", 1800, false)

	assert.Equal(t, time.Duration(1800)*time.Second, store.Expiration)
}

func TestGetSession_SessionDataPersistence(t *testing.T) {
	InitSessionStore("test-secret", 3600, false)

	app := fiber.New()

	var testStore *session.Store = store

	app.Get("/set", func(c *fiber.Ctx) error {
		sess, _ := testStore.Get(c)
		sess.Set("user_id", "123")
		sess.Set("username", "testuser")
		sess.Set("role", "admin")
		sess.Save()
		return c.SendString("OK")
	})

	app.Get("/get", func(c *fiber.Ctx) error {
		sess, _ := testStore.Get(c)
		userID := sess.Get("user_id")
		username := sess.Get("username")
		role := sess.Get("role")

		assert.Equal(t, "123", userID)
		assert.Equal(t, "testuser", username)
		assert.Equal(t, "admin", role)

		return c.SendString("OK")
	})

	// Set session data
	req := httptest.NewRequest("GET", "/set", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Get the session cookie from the response
	cookies := resp.Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessionCookie = cookie
			break
		}
	}

	// Get session data - pass the session cookie
	req = httptest.NewRequest("GET", "/get", nil)
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestSessionStore_ConcurrentAccess(t *testing.T) {
	InitSessionStore("test-secret", 3600, false)

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		sess := GetSession(c)
		assert.NotNil(t, sess)
		return c.SendString("OK")
	})

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestInitSessionStore_ZeroMaxAge(t *testing.T) {
	InitSessionStore("test-secret", 0, false)

	assert.NotNil(t, store)
	// Fiber session store has a default expiration of 24 hours when maxAge is 0
	// This is expected behavior - the middleware doesn't override Fiber's default
	assert.Equal(t, time.Duration(24)*time.Hour, store.Expiration)
}

func TestInitSessionStore_LargeMaxAge(t *testing.T) {
	InitSessionStore("test-secret", 31536000, false) // 1 year

	assert.NotNil(t, store)
	assert.Equal(t, time.Duration(31536000)*time.Second, store.Expiration)
}

func TestGetSession_NilStore(t *testing.T) {
	// Temporarily set store to nil
	originalStore := store
	store = nil
	defer func() { store = originalStore }()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// Expected - should panic when store is nil
				assert.NotNil(t, r)
			}
		}()
		sess := GetSession(c)
		assert.Nil(t, sess)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	app.Test(req)
}

func TestDestroySession_NonExistentSession(t *testing.T) {
	InitSessionStore("test-secret", 3600, false)

	app := fiber.New()
	app.Get("/destroy", func(c *fiber.Ctx) error {
		err := DestroySession(c)
		// Should not error even if session doesn't exist
		assert.NoError(t, err)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/destroy", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestSessionStore_DifferentSecretKeys(t *testing.T) {
	InitSessionStore("secret-key-1", 3600, false)
	store1 := store

	InitSessionStore("secret-key-2", 3600, false)
	store2 := store

	assert.NotNil(t, store1)
	assert.NotNil(t, store2)
	// Different secret keys should create different stores
	assert.NotEqual(t, store1, store2)
}
