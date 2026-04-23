package middleware

import (
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandler_FiberError_DebugMode(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler(true),
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return fiber.NewError(404, "Custom not found error")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Custom not found error")
}

func TestErrorHandler_FiberError_ProductionMode(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler(false),
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return fiber.NewError(404, "Custom not found error")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Internal server error")
}

func TestErrorHandler_GenericError_DebugMode(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler(true),
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return errors.New("Something went wrong")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Something went wrong")
}

func TestErrorHandler_GenericError_ProductionMode(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler(false),
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return errors.New("Something went wrong")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Internal server error")
}

func TestErrorHandler_401_APIRequest(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler(false),
	})

	app.Get("/api/test", func(c *fiber.Ctx) error {
		return fiber.NewError(401, "Unauthorized")
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "detail")
}

func TestErrorHandler_401_PageRequest(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler(false),
	})

	app.Get("/notes/test", func(c *fiber.Ctx) error {
		return fiber.NewError(401, "Unauthorized")
	})

	req := httptest.NewRequest("GET", "/notes/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 303, resp.StatusCode)
	assert.Equal(t, "/login", resp.Header.Get("Location"))
}

func TestErrorHandler_500_InternalServerError(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler(false),
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return fiber.NewError(500, "Internal error")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Internal server error")
}

func TestErrorHandler_ShareRoute(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler(false),
	})

	app.Get("/share/test", func(c *fiber.Ctx) error {
		return fiber.NewError(404, "Share not found")
	})

	req := httptest.NewRequest("GET", "/share/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "detail")
}

func TestSanitizeErrorMessage_FileNotFound(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no such file", "open /path/to/file: no such file or directory", "Resource not found"},
		{"file not found", "file not found: /secret/path", "Resource not found"},
		{"does not exist", "The system cannot find the file specified", "Resource not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeErrorMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeErrorMessage_PermissionDenied(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"permission denied", "open /secret/file: permission denied", "Access denied"},
		{"access is denied", "The system cannot access the file", "Access denied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeErrorMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeErrorMessage_InvalidPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"invalid path", "invalid path: ../../../etc/passwd", "Invalid request"},
		{"invalid argument", "invalid argument", "Invalid request"},
		{"illegal characters", "filename contains illegal characters", "Invalid request"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeErrorMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeErrorMessage_AlreadyExists(t *testing.T) {
	result := sanitizeErrorMessage("file already exists: /path/to/file")
	assert.Equal(t, "Resource already exists", result)
}

func TestSanitizeErrorMessage_DirectoryNotEmpty(t *testing.T) {
	result := sanitizeErrorMessage("directory not empty: /path/to/dir")
	assert.Equal(t, "Cannot delete non-empty resource", result)
}

func TestSanitizeErrorMessage_Timeout(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"i/o timeout", "i/o timeout", "Operation timed out"},
		{"context deadline", "context deadline exceeded", "Operation timed out"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeErrorMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeErrorMessage_DiskError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"disk full", "write /path: no space left on device", "Storage error"},
		{"disk error", "disk I/O error", "Storage error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeErrorMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeErrorMessage_UnknownError(t *testing.T) {
	result := sanitizeErrorMessage("Some unknown error occurred")
	assert.Equal(t, "Internal server error", result)
}

func TestSanitizeErrorMessage_EmptyString(t *testing.T) {
	result := sanitizeErrorMessage("")
	assert.Equal(t, "Internal server error", result)
}

func TestSanitizeErrorMessage_PathLeakPrevention(t *testing.T) {
	leakyErrors := []string{
		"open /home/user/secret/notes/password.md: no such file or directory",
		"access denied: /etc/passwd",
		"cannot write to /var/data/notes: permission denied",
		"failed to read /Users/admin/Documents/secret.md: file not found",
	}

	for _, err := range leakyErrors {
		result := sanitizeErrorMessage(err)
		assert.NotContains(t, result, "/home")
		assert.NotContains(t, result, "/etc")
		assert.NotContains(t, result, "/var")
		assert.NotContains(t, result, "/Users")
		assert.NotContains(t, result, "user")
		assert.NotContains(t, result, "secret")
	}
}

func TestHTTPError(t *testing.T) {
	err := HTTPError(400, "Bad request")

	assert.Equal(t, 400, err.Code)
	assert.Equal(t, "Bad request", err.Message)
}

func TestHTTPError_503(t *testing.T) {
	err := HTTPError(503, "Service unavailable")

	assert.Equal(t, 503, err.Code)
	assert.Equal(t, "Service unavailable", err.Message)
}

func TestErrorHandler_CaseInsensitivity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"NO SUCH FILE", "Resource not found"},
		{"Permission DENIED", "Access denied"},
		{"FILE NOT FOUND", "Resource not found"},
	}

	for _, tt := range tests {
		result := sanitizeErrorMessage(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestErrorHandler_MixedCaseInPath(t *testing.T) {
	leakyError := "Cannot access /Home/User/Secret.md: Permission Denied"
	result := sanitizeErrorMessage(leakyError)

	assert.Equal(t, "Access denied", result)
	assert.NotContains(t, result, "/Home")
	assert.NotContains(t, result, "User")
	assert.NotContains(t, result, "Secret")
}

func TestErrorHandler_ConcurrentAccess(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler(false),
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return fiber.NewError(500, "Error")
	})

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode)
	}
}
