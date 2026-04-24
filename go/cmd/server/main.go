package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/utils"

	"gonote/internal/models/config"
	"gonote/internal/handlers"
	applogger "gonote/internal/models/logger"
	"gonote/internal/middleware"
	"gonote/internal/services"
)

func main() {
	// Load configuration - check for --config flag or CONFIG_PATH env
	configPath := "config.yaml"
	for i, arg := range os.Args {
		if arg == "--config" && i+1 < len(os.Args) {
			configPath = os.Args[i+1]
		}
	}
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		applogger.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger with config
	applogger.SetEnabled(cfg.Log.Enabled)

	// Security warning for default secret key
	if cfg.Authentication.Enabled && cfg.Authentication.SecretKey == "change_this_to_a_random_secret_key_in_production" {
		applogger.Println("")
		applogger.Println("⚠️  SECURITY WARNING: You are using the default secret_key with authentication enabled!")
		applogger.Println("   Please change 'secret_key' in config.yaml or set AUTHENTICATION_SECRET_KEY environment variable.")
		applogger.Println("   Generate a secure key with: openssl rand -hex 32")
		applogger.Println("")
	}

	// Security warning for default password
	if cfg.Authentication.Enabled && cfg.Authentication.Password == "admin" {
		applogger.Println("")
		applogger.Println("⚠️  SECURITY WARNING: You are using the default password 'admin'!")
		applogger.Println("   Please change 'password' in config.yaml or set AUTHENTICATION_PASSWORD environment variable.")
		applogger.Println("   Example: export AUTHENTICATION_PASSWORD=your_secure_password")
		applogger.Println("")
	}

	// Log secure_cookie status
	if cfg.Authentication.SecureCookie {
		applogger.Println("🔒 Secure Cookie: ENABLED (cookies will only be sent over HTTPS)")
	} else {
		applogger.Println("⚠️  Secure Cookie: DISABLED (not recommended for production HTTPS)")
		applogger.Println("   Set HTTPS=true or X_FORWARDED_PROTO=https env var, or set secure_cookie: true in config.yaml")
	}

	// Ensure required directories exist
	if err := services.EnsureDirectories(cfg.Storage.NotesDir); err != nil {
		applogger.Fatalf("Failed to create directories: %v", err)
	}

	// Initialize session store with secure cookie setting
	middleware.InitSessionStore(cfg.Authentication.SecretKey, cfg.Authentication.SessionMaxAge, cfg.Authentication.SecureCookie)

	// Create Fiber app with body limit from config
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ServerHeader: cfg.App.Name,
		ErrorHandler: middleware.ErrorHandler(cfg.Server.Debug),
		BodyLimit:    cfg.Upload.MaxBodySizeMB * 1024 * 1024, // Convert MB to bytes
	})

	// Middleware
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	
	// HTTP request logger (respects log.enabled config)
	if cfg.Log.Enabled {
		app.Use(logger.New(logger.Config{
			Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
		}))
	}
	
	app.Use(middleware.CORS(cfg.Server.AllowedOrigins))

	// Global rate limiter (configurable via config.yaml)
	app.Use(middleware.RateLimiter())

	// CSRF protection using Double Submit Cookie pattern
	// Generates a CSRF token stored in a cookie, validated against X-CSRF-Token header
	app.Use(csrf.New(csrf.Config{
		KeyLookup:      "cookie:csrf_",
		CookieName:     "csrf_",
		CookieSameSite: "Lax",
		CookieSecure:   cfg.Authentication.SecureCookie,
		CookieHTTPOnly: false, // Must be false so JavaScript can read it
		CookieSessionOnly: false,
		Expiration:     0, // Session cookie (expires when browser closes)
		KeyGenerator:   utils.UUID,
	}))

	// Setup routes
	noteService := setupRoutes(app, cfg)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				applogger.Printf("Shutdown goroutine panic recovered: %v", r)
			}
		}()
		<-quit
		fmt.Println("\nShutting down gracefully...")
		noteService.StopBackgroundScanner() // Stop background scanner
		noteService.StopCacheCleanup()      // Stop cache cleanup goroutine
		handlers.GetWSManager().Stop()      // Stop WebSocket manager and close all connections
		if err := app.Shutdown(); err != nil {
			applogger.Printf("Shutdown error: %v", err)
		}
	}()

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	applogger.Printf("Starting %s v%s on %s", cfg.App.Name, cfg.App.Version, addr)

	if err := app.Listen(addr); err != nil {
		applogger.Fatalf("Failed to start server: %v", err)
	}
}

// resolveStaticPath resolves a static file path, checking multiple locations
// to support both development environment and Docker containers
func resolveStaticPath(name string, paths ...string) string {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			if resolvedPath, err := filepath.EvalSymlinks(p); err == nil {
				return resolvedPath
			}
			if absPath, err := filepath.Abs(p); err == nil {
				return absPath
			}
		}
	}
	// Fallback to first path
	if absPath, err := filepath.Abs(paths[0]); err == nil {
		return absPath
	}
	return paths[0]
}

func setupRoutes(app *fiber.App, cfg *config.Config) *services.NoteService {
	// Resolve the frontend path with proper symlink handling
	// Fiber's app.Static() doesn't properly follow symlinks, so we resolve
	// the actual target path before passing it to app.Static()
	// Support both development environment (../shared/frontend) and Docker (./frontend)
	frontendPath := resolveStaticPath("frontend", "../shared/frontend", "./frontend")

	// Initialize WebSocket manager
	handlers.InitWSManager()

	// Initialize services with background scanner
	cacheTTL := time.Duration(cfg.Cache.TTL) * time.Second
	scanInterval := time.Duration(cfg.Cache.ScanInterval) * time.Second
	noteService := services.NewNoteServiceWithScanner(cfg.Storage.NotesDir, cacheTTL, cfg.Cache.Capacity, scanInterval)
	noteService.StartCacheCleanup()      // Start cache cleanup goroutine
	noteService.StartBackgroundScanner() // Start background scanner
	
	// Set callback to broadcast WebSocket message when scan completes
	noteService.SetOnScanComplete(func() {
		handlers.BroadcastNotesUpdated()
	})
	
	searchService := services.NewSearchService(cfg.Storage.NotesDir)
	searchIndex := services.NewSearchIndex(cfg.Storage.NotesDir, noteService)
	tagService := services.NewTagService(noteService, cfg.Storage.NotesDir)
	templateService := services.NewTemplateService(cfg.Storage.NotesDir)
	shareService := services.NewShareService(cfg.Storage.NotesDir)
	mediaService := services.NewMediaService(cfg.Storage.NotesDir)
	themePath := resolveStaticPath("themes", "../shared/themes", "./themes")
	localePath := resolveStaticPath("locales", "../shared/locales", "./locales")

	themeService := services.NewThemeService(themePath)
	localeService := services.NewLocaleService(localePath)
	graphService := services.NewGraphService(cfg.Storage.NotesDir)
	exportService := services.NewExportService(cfg.Storage.NotesDir, themePath)

	// Build search index asynchronously on startup
	go func() {
		defer func() {
			if r := recover(); r != nil {
				applogger.Printf("Search index build goroutine panic recovered: %v", r)
			}
		}()
		applogger.Println("Building search index...")
		if err := searchIndex.BuildIndex(); err != nil {
			applogger.Printf("Warning: Failed to build search index: %v", err)
		} else {
			applogger.Printf("Search index built with %d terms", searchIndex.GetIndexSize())
		}
	}()

	// Initialize handlers
	noteHandler := handlers.NewNoteHandlerWithTagService(noteService, tagService, cfg, searchIndex)
	folderHandler := handlers.NewFolderHandlerWithCache(cfg, noteService)
	searchHandler := handlers.NewSearchHandlerWithIndex(searchService, searchIndex, cfg)
	tagHandler := handlers.NewTagHandler(tagService, cfg)
	templateHandler := handlers.NewTemplateHandler(templateService, cfg)
	shareHandler := handlers.NewShareHandler(shareService, exportService, cfg)
	mediaHandler := handlers.NewMediaHandler(mediaService, noteService, cfg)
	themeHandler := handlers.NewThemeHandler(themeService, cfg)
	localeHandler := handlers.NewLocaleHandler(localeService, cfg)
	graphHandler := handlers.NewGraphHandler(graphService, cfg)
	systemHandler := handlers.NewSystemHandler(cfg)
	authHandler := handlers.NewAuthHandler(cfg)
	backlinkService := services.NewBacklinkService(cfg.Storage.NotesDir)
	statisticsService := services.NewStatisticsService(cfg.Storage.NotesDir)
	statisticsHandler := handlers.NewStatisticsHandler(statisticsService, cfg)

	// Public endpoints (no auth required)
	app.Get("/health", systemHandler.HealthCheck)
	app.Get("/sw.js", middleware.EndpointLimiterSimple(30), systemHandler.ServiceWorker)
	app.Get("/api/themes", themeHandler.List)
	app.Get("/api/themes/:id", themeHandler.Get)
	app.Get("/api/locales", localeHandler.List)
	app.Get("/api/locales/:code", localeHandler.Get)
	app.Get("/share/:token", middleware.EndpointLimiterSimple(60), shareHandler.ViewSharedNote)

	// WebSocket endpoint for real-time updates (with authentication check)
	app.Get("/ws", middleware.WSAuthRequired(cfg.Authentication.Enabled), websocket.New(func(c *websocket.Conn) {
		// Register new connection (returns false if manager stopped)
		if !handlers.GetWSManager().Register(c) {
			c.Close()
			return
		}

		// Cleanup on disconnect
		defer func() {
			handlers.GetWSManager().Unregister(c)
		}()

		// Keep connection alive, read incoming messages (ping/pong handled by Fiber)
		var msg map[string]interface{}
		for {
			if err := c.ReadJSON(&msg); err != nil {
				break
			}
			// We don't expect any meaningful messages from clients
			// Just keep reading to detect disconnection
		}
	}))

	// Login routes (public)
	app.Get("/login", authHandler.LoginPage)
	app.Post("/login", middleware.EndpointLimiterSimple(10), authHandler.Login)
	app.Post("/logout", authHandler.Logout)

	// Static files - use resolved absolute path
	app.Static("/static", frontendPath)

	// Protected API routes
	api := app.Group("/api", middleware.AuthRequired(cfg.Authentication.Enabled))

	// System
	api.Get("/config", systemHandler.GetConfig)

	// Notes - register static routes before wildcards
	api.Get("/notes", noteHandler.List)
	api.Post("/notes/move", middleware.EndpointLimiterSimple(30), noteHandler.Move)
	api.Get("/notes/attachments/*", noteHandler.GetAttachments)
	api.Get("/notes/*", noteHandler.Get)
	api.Post("/notes/*", middleware.EndpointLimiterSimple(60), noteHandler.CreateOrUpdate)
	api.Delete("/notes/*", middleware.EndpointLimiterSimple(30), noteHandler.Delete)

	// Folders - register static routes before wildcards
	api.Post("/folders", middleware.EndpointLimiterSimple(30), folderHandler.Create)
	api.Post("/folders/move", middleware.EndpointLimiterSimple(20), folderHandler.Move)
	api.Post("/folders/rename", middleware.EndpointLimiterSimple(30), folderHandler.Rename)
	api.Delete("/folders/*", middleware.EndpointLimiterSimple(20), folderHandler.Delete)

	// Search
	api.Get("/search", searchHandler.Search)

	// Tags
	api.Get("/tags", tagHandler.List)
	api.Get("/tags/*", tagHandler.GetNotesByTag)

	// Templates
	api.Get("/templates", middleware.EndpointLimiterSimple(120), templateHandler.List)
	api.Get("/templates/*", middleware.EndpointLimiterSimple(120), templateHandler.Get)
	api.Post("/templates/create-note", middleware.EndpointLimiterSimple(60), templateHandler.CreateFromTemplate)

	// Share
	api.Post("/share/*", middleware.EndpointLimiterSimple(30), shareHandler.Create)
	api.Get("/share/*", middleware.EndpointLimiterSimple(120), shareHandler.GetStatus)
	api.Delete("/share/*", middleware.EndpointLimiterSimple(30), shareHandler.Revoke)
	api.Get("/shared-notes", middleware.EndpointLimiterSimple(60), shareHandler.ListSharedNotes)

	// Media
	api.Get("/media/orphaned", mediaHandler.ListOrphanedMedia)
	api.Delete("/media/orphaned", mediaHandler.CleanupOrphanedMedia)
	api.Post("/media/move", middleware.EndpointLimiterSimple(30), mediaHandler.Move)
	api.Get("/media/*", mediaHandler.Get)
	api.Post("/upload-media", middleware.EndpointLimiterSimple(20), mediaHandler.Upload)

	// Graph
	api.Get("/graph", graphHandler.Get)

	// Backlinks
	api.Get("/backlinks/*", middleware.EndpointLimiterSimple(60), func(c *fiber.Ctx) error {
		notePath := c.Params("*")
		notePath = strings.TrimPrefix(notePath, "/")
		
		// URL decode the path
		decodedPath, err := url.PathUnescape(notePath)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"detail": "Invalid path encoding"})
		}
		notePath = decodedPath
		
		// Security check
		if !services.ValidatePathSecurity(cfg.Storage.NotesDir, notePath) {
			return c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
		}
		
		backlinks, err := backlinkService.FindBacklinks(notePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
		}
		
		return c.JSON(fiber.Map{
			"success":   true,
			"backlinks": backlinks,
			"count":     len(backlinks),
		})
	})

	// Statistics
	api.Get("/stats/*", middleware.EndpointLimiterSimple(60), statisticsHandler.GetStatistics)

	// SPA fallback - serve index.html for all unmatched routes
	app.Get("/*", middleware.EndpointLimiterSimple(120), middleware.AuthRequired(cfg.Authentication.Enabled), func(c *fiber.Ctx) error {
		path := c.Path()
		// Check if it's an API route that wasn't matched
		if len(path) >= 4 && path[:4] == "/api" {
			return c.Status(404).JSON(fiber.Map{"detail": "Not found"})
		}
		// Also check for static files that might have slipped through (e.g., missing files)
		if len(path) >= 8 && path[:8] == "/static/" {
			return c.Status(404).SendString("Static file not found")
		}
		// Serve index.html for SPA using resolved path
		return c.SendFile(filepath.Join(frontendPath, "index.html"))
	})

	return noteService
}
