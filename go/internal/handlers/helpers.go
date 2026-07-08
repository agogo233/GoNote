package handlers

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gonote/internal/services"
)

func resolvePathParam(c *fiber.Ctx, notesDir string) (string, bool) {
	decoded, err := url.PathUnescape(c.Params("*"))
	if err != nil {
		c.Status(400).JSON(fiber.Map{"detail": "Invalid path encoding"})
		return "", false
	}
	if !services.ValidatePathSecurity(notesDir, decoded) {
		c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
		return "", false
	}
	return decoded, true
}

func resolvePathParamTrimmed(c *fiber.Ctx, notesDir string) (string, bool) {
	path := strings.TrimPrefix(c.Params("*"), "/")
	decoded, err := url.PathUnescape(path)
	if err != nil {
		c.Status(400).JSON(fiber.Map{"detail": "Invalid path encoding"})
		return "", false
	}
	if !services.ValidatePathSecurity(notesDir, decoded) {
		c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
		return "", false
	}
	return decoded, true
}

func validatePath(c *fiber.Ctx, notesDir, path string) bool {
	if !services.ValidatePathSecurity(notesDir, path) {
		c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
		return false
	}
	return true
}