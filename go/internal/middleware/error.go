package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/logger"
)

// ErrorHandler is a custom error handler for Fiber
func ErrorHandler(debug bool) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Default error code
		code := fiber.StatusInternalServerError
		message := "Internal server error"

		// Check if it's a Fiber error
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			if debug {
				message = e.Message
			} else {
				// Sanitize fiber.Error messages in production too
				message = sanitizeErrorMessage(e.Message)
			}
		} else if debug {
			// In debug mode, show full error
			message = err.Error()
		} else {
			// In production, sanitize error messages to hide sensitive info
			message = sanitizeErrorMessage(err.Error())
		}

		// Log the error with full details (server-side only)
		logger.PrintError("[ERROR] Path=%s IP=%s Error=%v\n", c.Path(), c.IP(), err)

		// Handle 401 errors specially
		if code == 401 {
			// For API requests, return JSON error
			if strings.HasPrefix(c.Path(), "/api/") {
				return c.Status(401).JSON(fiber.Map{
					"detail": message,
				})
			}
			// For page requests, redirect to login
			return c.Redirect("/login", 303)
		}

		// Return JSON error for API routes
		if strings.HasPrefix(c.Path(), "/api/") || strings.HasPrefix(c.Path(), "/share/") {
			return c.Status(code).JSON(fiber.Map{
				"detail": message,
			})
		}

		// Return generic error for other routes
		return c.Status(code).SendString(message)
	}
}

// SanitizeErrorMessage converts detailed errors into user-friendly messages
// to prevent leaking file paths and internal details.
// This is exported for use in handlers that need to return error messages directly.
func SanitizeErrorMessage(errMsg string) string {
	return sanitizeErrorMessage(errMsg)
}

// sanitizeErrorMessage converts detailed errors into user-friendly messages
// to prevent leaking file paths and internal details
func sanitizeErrorMessage(errMsg string) string {
	errLower := strings.ToLower(errMsg)

	// File system errors
	switch {
	case strings.Contains(errLower, "no such file") ||
		strings.Contains(errLower, "file not found") ||
		strings.Contains(errLower, "does not exist") ||
		strings.Contains(errLower, "cannot find"):
		return "Resource not found"

	case strings.Contains(errLower, "permission denied") ||
		strings.Contains(errLower, "access is denied") ||
		strings.Contains(errLower, "cannot access"):
		return "Access denied"

	case strings.Contains(errLower, "invalid path") ||
		strings.Contains(errLower, "invalid argument") ||
		strings.Contains(errLower, "illegal characters"):
		return "Invalid request"

	case strings.Contains(errLower, "already exists"):
		return "Resource already exists"

	case strings.Contains(errLower, "directory not empty"):
		return "Cannot delete non-empty resource"

	case strings.Contains(errLower, "i/o timeout") ||
		strings.Contains(errLower, "context deadline"):
		return "Operation timed out"

	case strings.Contains(errLower, "disk") ||
		strings.Contains(errLower, "no space"):
		return "Storage error"

	default:
		// Return generic message for unknown errors
		return "Internal server error"
	}
}

// HTTPError creates a custom HTTP error
func HTTPError(status int, message string) *fiber.Error {
	return fiber.NewError(status, message)
}
