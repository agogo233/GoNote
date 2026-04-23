package handlers

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"gonote/internal/models/config"
	"gonote/internal/middleware"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	config *config.Config
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{config: cfg}
}

// LoginPage renders the login page
func (h *AuthHandler) LoginPage(c *fiber.Ctx) error {
	// If auth is disabled, redirect to home
	if !h.config.Authentication.Enabled {
		return c.Redirect("/", 303)
	}

	// Check if already authenticated
	sess := middleware.GetSession(c)
	authenticated := sess.Get("authenticated")
	if authenticated == true {
		return c.Redirect("/", 303)
	}

	// Read login.html
	data, err := os.ReadFile("frontend/login.html")
	if err != nil {
		// Return a simple login page
		c.Set("Content-Type", "text/html")
		return c.SendString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - GoNote</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; background: #f5f5f5; }
        .login-container { background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); width: 100%; max-width: 400px; }
        h1 { margin-top: 0; color: #333; }
        .form-group { margin-bottom: 1rem; }
        label { display: block; margin-bottom: 0.5rem; color: #666; }
        input[type="password"] { width: 100%; padding: 0.75rem; border: 1px solid #ddd; border-radius: 4px; font-size: 1rem; box-sizing: border-box; }
        button { width: 100%; padding: 0.75rem; background: #0366d6; color: white; border: none; border-radius: 4px; font-size: 1rem; cursor: pointer; }
        button:hover { background: #0257b3; }
        .error { color: #d32f2f; margin-bottom: 1rem; }
    </style>
</head>
<body>
    <div class="login-container">
        <h1>🔐 GoNote</h1>
        <div id="error" class="error" style="display:none;"></div>
        <form id="loginForm">
            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required autofocus>
            </div>
            <button type="submit">Login</button>
        </form>
    </div>
    <script>
        function getCsrfToken() {
            const cookies = document.cookie.split(';');
            for (let i = 0; i < cookies.length; i++) {
                const cookie = cookies[i].trim();
                if (cookie.indexOf('csrf_=') === 0) {
                    return cookie.substring(6);
                }
            }
            return null;
        }
        document.getElementById('loginForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const password = document.getElementById('password').value;
            const errorDiv = document.getElementById('error');
            try {
                const csrfToken = getCsrfToken();
                const headers = { 'Content-Type': 'application/json' };
                if (csrfToken) {
                    headers['X-CSRF-Token'] = csrfToken;
                }
                const res = await fetch('/login', {
                    method: 'POST',
                    headers: headers,
                    body: JSON.stringify({ password })
                });
                const data = await res.json();
                if (data.success) {
                    window.location.href = '/';
                } else {
                    errorDiv.textContent = data.detail || 'Login failed';
                    errorDiv.style.display = 'block';
                }
            } catch (err) {
                errorDiv.textContent = 'Login failed';
                errorDiv.style.display = 'block';
            }
        });
    </script>
</body>
</html>`)
	}

	c.Set("Content-Type", "text/html")
	return c.SendString(string(data))
}

// Login handles login requests
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// If auth is disabled, allow access
	if !h.config.Authentication.Enabled {
		return c.JSON(fiber.Map{"success": true})
	}

	// Parse request
	var req struct {
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "detail": "Invalid request"})
	}

	// Verify password
	if h.config.Authentication.PasswordHash == "" {
		return c.Status(500).JSON(fiber.Map{"success": false, "detail": "Password not configured"})
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(h.config.Authentication.PasswordHash),
		[]byte(req.Password),
	)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "detail": "Invalid password"})
	}

	// 🔒 防止会话固定攻击：先重新生成会话 ID，再设置认证标志
	sess := middleware.GetSession(c)

	// 1. 重新生成会话 ID，销毁旧会话（防止会话固定攻击）
	if err := sess.Regenerate(); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "detail": "Failed to regenerate session"})
	}

	// 2. 在新会话中设置认证标志
	sess.Set("authenticated", true)

	// 3. 保存会话
	if err := sess.Save(); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "detail": "Failed to create session"})
	}

	return c.JSON(fiber.Map{"success": true})
}

// Logout handles logout requests
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	if err := middleware.DestroySession(c); err != nil {
		// Ignore errors
	}

	// For API requests, return JSON
	if strings.HasPrefix(c.Path(), "/api/") {
		return c.JSON(fiber.Map{"success": true})
	}

	// For page requests, redirect to login
	return c.Redirect("/login", 303)
}


