package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var store *session.Store

// InitSessionStore initializes the session store
// secureCookie should be true when running behind HTTPS
func InitSessionStore(secretKey string, maxAge int, secureCookie bool) {
	store = session.New(session.Config{
		KeyLookup:      "cookie:session_id",
		Expiration:     time.Duration(maxAge) * time.Second,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
		CookieSecure:   secureCookie,
	})
}

// GetSession returns the session for the current request
func GetSession(c *fiber.Ctx) *session.Session {
	sess, _ := store.Get(c)
	return sess
}

// DestroySession destroys the current session
func DestroySession(c *fiber.Ctx) error {
	sess := GetSession(c)
	return sess.Destroy()
}
