package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/logger"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
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

	// Security warnings for disabled protections
	if !cfg.Authentication.Enabled {
		applogger.Println("")
		applogger.Println("⚠️  SECURITY WARNING: Authentication is DISABLED!")
		applogger.Println("   Anyone can access all notes, media, and settings.")
		applogger.Println("   Enable authentication for any network-exposed deployment.")
		applogger.Println("")
	}

	if len(cfg.Server.AllowedOrigins) == 1 && cfg.Server.AllowedOrigins[0] == "*" {
		applogger.Println("")
		applogger.Println("⚠️  SECURITY WARNING: CORS allows ALL origins (*)!")
		applogger.Println("   Set allowed_origins to specific domains for production.")
		applogger.Println("")
	}

	if !cfg.RateLimit.Enabled {
		applogger.Println("")
		applogger.Println("⚠️  SECURITY WARNING: Rate limiting is DISABLED!")
		applogger.Println("   Enable rate_limit.enabled in config.yaml for production.")
		applogger.Println("")
	}

	// Ensure required directories exist
	if err := services.EnsureDirectories(cfg.Storage.NotesDir); err != nil {
		applogger.Fatalf("Failed to create directories: %v", err)
	}

	// Initialize session store with secure cookie setting
	middleware.InitSessionStore(cfg.Authentication.SecretKey, cfg.Authentication.SessionMaxAge, cfg.Authentication.SecureCookie)

	// Create Fiber app with body limit from config
	app := fiber.New(fiber.Config{
		AppName:                 cfg.App.Name,
		ServerHeader:            cfg.App.Name,
		ErrorHandler:            middleware.ErrorHandler(cfg.Server.Debug),
		BodyLimit:               cfg.Upload.MaxBodySizeMB * 1024 * 1024, // Convert MB to bytes
		ProxyHeader:             cfg.Server.ProxyHeader,
		EnableTrustedProxyCheck: cfg.Server.TrustedProxyCheck,
		TrustedProxies:          cfg.Server.TrustedProxies,
		EnableIPValidation:      cfg.Server.TrustedProxyCheck,
	})

	// Middleware
	// recover 必须放在最前：捕获后续所有中间件/handler 的 panic，避免连接异常断开
	app.Use(fiberrecover.New())

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
	app.Use(middleware.RateLimiter(cfg))

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
	noteService, wsManager := setupRoutes(app, cfg)

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
		applogger.Infof("\nShutting down gracefully...")
		noteService.StopBackgroundScanner() // Stop background scanner
		noteService.StopCacheCleanup()      // Stop cache cleanup goroutine
		wsManager.Stop()      // Stop WebSocket manager and close all connections
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

func setupRoutes(app *fiber.App, cfg *config.Config) (*services.NoteService, *handlers.WSManager) {
	// Resolve the frontend path with proper symlink handling
	frontendPath := resolveStaticPath("frontend", "./shared/frontend", "../shared/frontend", "./frontend")

	// Initialize WebSocket manager
	wsManager := handlers.InitWSManager(cfg.Server.WSMaxConnections)

	// Initialize services with background scanner
	cacheTTL := time.Duration(cfg.Cache.TTL) * time.Second
	scanInterval := time.Duration(cfg.Cache.ScanInterval) * time.Second
	noteService := services.NewNoteServiceWithScanner(cfg.Storage.NotesDir, cacheTTL, cfg.Cache.Capacity, scanInterval)
	noteService.StartCacheCleanup()
	noteService.StartBackgroundScanner()

	linkIndex := services.NewLinkIndex(cfg.Storage.NotesDir)
	noteService.SetLinkIndex(linkIndex)

	noteService.SetOnScanComplete(func() {
		handlers.BroadcastNotesUpdated(wsManager)
	})

	searchService := services.NewSearchService(cfg.Storage.NotesDir, noteService)
	searchIndex := services.NewSearchIndex(cfg.Storage.NotesDir, noteService)
	noteService.SetSearchIndex(searchIndex)
	tagService := services.NewTagService(noteService, cfg.Storage.NotesDir)
	templateService := services.NewTemplateService(cfg.Storage.NotesDir)
	shareService := services.NewShareService(cfg.Storage.NotesDir)
	mediaService := services.NewMediaService(cfg.Storage.NotesDir)
	themePath := resolveStaticPath("themes", "../shared/themes", "./themes")
	localePath := resolveStaticPath("locales", "../shared/locales", "./locales")

	themeService := services.NewThemeService(themePath)
	localeService := services.NewLocaleService(localePath)
	graphService := services.NewGraphService(cfg.Storage.NotesDir, noteService)
	graphService.SetLinkIndex(linkIndex)
	exportService := services.NewExportService(cfg.Storage.NotesDir, themePath)

	// Build search index asynchronously on startup
	if cfg.Search.Enabled {
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
	}

	// Initialize handlers
	noteHandler := handlers.NewNoteHandlerWithTagService(noteService, tagService, cfg, searchIndex, shareService)
	folderHandler := handlers.NewFolderHandlerWithCache(cfg, noteService)
	searchHandler := handlers.NewSearchHandlerWithIndex(searchService, searchIndex, cfg)
	tagHandler := handlers.NewTagHandler(tagService, cfg)
	templateHandler := handlers.NewTemplateHandler(templateService, cfg)
	shareHandler := handlers.NewShareHandler(shareService, exportService, cfg, themePath)
	mediaHandler := handlers.NewMediaHandler(mediaService, noteService, cfg)
	themeHandler := handlers.NewThemeHandler(themeService, cfg)
	localeHandler := handlers.NewLocaleHandler(localeService, cfg)
	graphHandler := handlers.NewGraphHandler(graphService, cfg)
	systemHandler := handlers.NewSystemHandler(cfg, frontendPath)
	systemHandler.SetReadinessDeps(noteService, searchIndex)
	authHandler := handlers.NewAuthHandler(cfg)
	backlinkHandler := handlers.NewBacklinkHandler(cfg, linkIndex)
	statisticsService := services.NewStatisticsService(cfg.Storage.NotesDir)
	statisticsHandler := handlers.NewStatisticsHandler(statisticsService, cfg)

	// Public endpoints (no auth required)
	systemHandler.RegisterRoutes(app)
	themeHandler.RegisterRoutes(app)
	localeHandler.RegisterRoutes(app)
	authHandler.RegisterRoutes(app)
	app.Get("/share/:token", middleware.EndpointLimiterSimple(60), shareHandler.ViewSharedNote)

	// WebSocket endpoint
	app.Get("/ws", middleware.WSAuthRequired(cfg.Authentication.Enabled), websocket.New(func(c *websocket.Conn) {
		c.SetReadLimit(4096)
		c.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.SetPongHandler(func(string) error {
			c.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		if !wsManager.Register(c) {
			c.Close()
			return
		}

		defer func() {
			wsManager.Unregister(c)
		}()

		var msg map[string]interface{}
		for {
			if err := c.ReadJSON(&msg); err != nil {
				break
			}
		}
	}, websocket.Config{
		Origins: cfg.Server.AllowedOrigins,
	}))

	// Static files
	app.Static("/static", frontendPath)

	// Protected API routes
	api := app.Group("/api", middleware.AuthRequired(cfg.Authentication.Enabled))

	systemHandler.RegisterAPIRoutes(api)
	noteHandler.RegisterRoutes(api)
	folderHandler.RegisterRoutes(api)
	tagHandler.RegisterRoutes(api)
	templateHandler.RegisterRoutes(api)
	shareHandler.RegisterRoutes(api, app)
	mediaHandler.RegisterRoutes(api)
	graphHandler.RegisterRoutes(api)
	statisticsHandler.RegisterRoutes(api)
	backlinkHandler.RegisterRoutes(api)

	// Search (only registered if search is enabled)
	if cfg.Search.Enabled {
		searchHandler.RegisterRoutes(api)
	}

	// SPA fallback
	app.Get("/*", middleware.EndpointLimiterSimple(120), middleware.AuthRequired(cfg.Authentication.Enabled), func(c *fiber.Ctx) error {
		path := c.Path()
		if len(path) >= 4 && path[:4] == "/api" {
			return c.Status(404).JSON(fiber.Map{"detail": "Not found"})
		}
		if len(path) >= 8 && path[:8] == "/static/" {
			return c.Status(404).SendString("Static file not found")
		}
		return c.SendFile(filepath.Join(frontendPath, "index.html"))
	})

	return noteService, wsManager
}
