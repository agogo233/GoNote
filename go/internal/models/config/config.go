package config

import (
	"bytes"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	App            AppConfig       `yaml:"app"`
	Log            LogConfig       `yaml:"log"`
	Server         ServerConfig    `yaml:"server"`
	Storage        StorageConfig   `yaml:"storage"`
	Search         SearchConfig    `yaml:"search"`
	Authentication AuthConfig      `yaml:"authentication"`
	Cache          CacheConfig     `yaml:"cache"`
	RateLimit      RateLimitConfig `yaml:"rate_limit"`
	Upload         UploadConfig    `yaml:"upload"`
}

// UploadConfig holds file upload settings
type UploadConfig struct {
	MaxFileSizeMB int      `yaml:"max_file_size_mb"` // Max file size in MB (default: 50)
	MaxBodySizeMB int      `yaml:"max_body_size_mb"` // Max request body in MB (default: 100)
	AllowedTypes  []string `yaml:"allowed_types"`    // Allowed MIME types (empty = all)
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	Enabled       bool `yaml:"enabled"`        // Enable rate limiting globally
	MaxRequests   int  `yaml:"max_requests"`   // Max requests per window
	WindowSeconds int  `yaml:"window_seconds"` // Window duration in seconds
}

// AppConfig holds application-level settings
type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// LogConfig holds logging settings
type LogConfig struct {
	Enabled bool `yaml:"enabled"` // Enable/disable all logging
}

// ServerConfig holds server settings
type ServerConfig struct {
	Host                  string   `yaml:"host"`
	Port                  int      `yaml:"port"`
	AllowedOrigins        []string `yaml:"allowed_origins"`
	Debug                 bool     `yaml:"debug"`
	ProxyHeader           string   `yaml:"proxy_header"`
	TrustedProxyCheck     bool     `yaml:"trusted_proxy_check"`
	TrustedProxies        []string `yaml:"trusted_proxies"`
	WSMaxConnections      int      `yaml:"ws_max_connections"`
}

// StorageConfig holds storage paths
type StorageConfig struct {
	NotesDir string `yaml:"notes_dir"`
}

// SearchConfig holds search settings
type SearchConfig struct {
	Enabled bool `yaml:"enabled"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	Enabled       bool   `yaml:"enabled"`
	SecretKey     string `yaml:"secret_key"`
	PasswordHash  string `yaml:"password_hash"`
	Password      string `yaml:"password"`
	SessionMaxAge int    `yaml:"session_max_age"`
	SecureCookie  bool   `yaml:"secure_cookie"` // Set true for HTTPS in production
}

// CacheConfig holds cache settings
type CacheConfig struct {
	TTL           int `yaml:"ttl"`            // seconds
	Capacity      int `yaml:"capacity"`       // max items in cache
	ScanInterval  int `yaml:"scan_interval"`  // background scan interval in seconds (default: 30)
}

// Global flags set from environment variables
var (
	DemoMode       bool
	AlreadyDonated bool
	GlobalConfig   *Config // Global config reference for middleware access
)

// Load reads configuration from YAML file and applies environment variable overrides
func Load(configPath string) (*Config, error) {
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}

	// Load version from VERSION file
	if err := loadVersion(cfg); err != nil {
		return nil, err
	}

	// Apply defaults
	applyDefaults(cfg)

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	// Hash password if provided in plaintext
	hashPasswordIfNeeded(cfg)

	// Auto-detect HTTPS and set secure_cookie if needed
	DetectHTTPSAndSetSecureCookie(cfg)

	// Set global config reference for middleware access
	GlobalConfig = cfg

	return cfg, nil
}

// applyDefaults sets default values for optional config fields.
// Called after YAML load and before env overrides.
// These defaults are purely defensive — config.yaml 中已显式设置的值不会被覆盖
// (零值检测: 0/""/false/nil 才会触发默认值)。
func applyDefaults(cfg *Config) {
	// 日志默认关闭（注释保留：设计意图——不允许 true 走默认值，零值即关闭）

	// 服务器默认值
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 9000
	}
	if len(cfg.Server.AllowedOrigins) == 0 {
		cfg.Server.AllowedOrigins = []string{"*"} // Allow all origins by default
	}
	if cfg.Server.WSMaxConnections == 0 {
		cfg.Server.WSMaxConnections = 100 // Default max 100 concurrent WS connections
	}

	// 存储默认值
	if cfg.Storage.NotesDir == "" {
		cfg.Storage.NotesDir = "./data/notes"
	}

	// 认证默认值
	// session_max_age=0 会导致会话立即过期，必须给一个值
	if cfg.Authentication.SessionMaxAge == 0 {
		cfg.Authentication.SessionMaxAge = 86400 // 24 hours default
	}

	// 缓存默认值
	if cfg.Cache.TTL == 0 {
		cfg.Cache.TTL = 60 // 60 seconds default
	}
	if cfg.Cache.Capacity == 0 {
		cfg.Cache.Capacity = 1000 // 1000 items default
	}
	if cfg.Cache.ScanInterval == 0 {
		cfg.Cache.ScanInterval = 30 // 30 seconds default
	}

	// 限流默认值
	// 当 RateLimit.Enabled=true 但 max/window 未设置时——零值会导致拒绝所有请求或除零
	if cfg.RateLimit.MaxRequests == 0 {
		cfg.RateLimit.MaxRequests = 100 // 100 requests per window
	}
	if cfg.RateLimit.WindowSeconds == 0 {
		cfg.RateLimit.WindowSeconds = 60 // 60 seconds window
	}

	// 上传默认值
	if cfg.Upload.MaxFileSizeMB == 0 {
		cfg.Upload.MaxFileSizeMB = 50 // 50MB default
	}
	if cfg.Upload.MaxBodySizeMB == 0 {
		cfg.Upload.MaxBodySizeMB = 100 // 100MB default
	}
}

// loadVersion reads the version from VERSION file
func loadVersion(cfg *Config) error {
	versionPath := "VERSION"
	data, err := os.ReadFile(versionPath)
	if err != nil {
		return err
	}
	cfg.App.Version = strings.TrimSpace(string(data))
	return nil
}

// applyEnvOverrides applies environment variable overrides to configuration
func applyEnvOverrides(cfg *Config) {
	// ========== 日志配置 ==========
	// LOG_ENABLED override
	if v := os.Getenv("LOG_ENABLED"); v != "" {
		cfg.Log.Enabled = strings.ToLower(v) == "true"
	}

	// ========== 服务器配置 ==========
	// HOST override
	if v := os.Getenv("HOST"); v != "" {
		cfg.Server.Host = v
	}

	// PORT override
	if v := os.Getenv("PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil && port > 0 && port <= 65535 {
			cfg.Server.Port = port
		}
	}

	// DEBUG override
	if v := os.Getenv("DEBUG"); v != "" {
		cfg.Server.Debug = strings.ToLower(v) == "true"
	}

	// ALLOWED_ORIGINS override (comma-separated)
	if v := os.Getenv("ALLOWED_ORIGINS"); v != "" {
		origins := strings.Split(v, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		cfg.Server.AllowedOrigins = origins
	}

	// ========== 认证配置 ==========
	// AUTHENTICATION_ENABLED override
	if v := os.Getenv("AUTHENTICATION_ENABLED"); v != "" {
		cfg.Authentication.Enabled = strings.ToLower(v) == "true"
	}

	// AUTHENTICATION_PASSWORD override (will be hashed)
	if v := os.Getenv("AUTHENTICATION_PASSWORD"); v != "" {
		cfg.Authentication.Password = v
	}

	// AUTHENTICATION_PASSWORD_HASH override (pre-hashed)
	if v := os.Getenv("AUTHENTICATION_PASSWORD_HASH"); v != "" {
		cfg.Authentication.PasswordHash = v
	}

	// AUTHENTICATION_SECRET_KEY override
	if v := os.Getenv("AUTHENTICATION_SECRET_KEY"); v != "" {
		cfg.Authentication.SecretKey = v
	}

	// AUTHENTICATION_SECURE_COOKIE override
	if v := os.Getenv("AUTHENTICATION_SECURE_COOKIE"); v != "" {
		cfg.Authentication.SecureCookie = strings.ToLower(v) == "true"
	}

	// ========== 限流配置 ==========
	// RATE_LIMIT_ENABLED override
	if v := os.Getenv("RATE_LIMIT_ENABLED"); v != "" {
		cfg.RateLimit.Enabled = strings.ToLower(v) == "true"
	}

	// RATE_LIMIT_MAX override
	if v := os.Getenv("RATE_LIMIT_MAX"); v != "" {
		if max, err := strconv.Atoi(v); err == nil && max > 0 {
			cfg.RateLimit.MaxRequests = max
		}
	}

	// RATE_LIMIT_WINDOW override
	if v := os.Getenv("RATE_LIMIT_WINDOW"); v != "" {
		if window, err := strconv.Atoi(v); err == nil && window > 0 {
			cfg.RateLimit.WindowSeconds = window
		}
	}

	// ========== 上传配置 ==========
	// UPLOAD_MAX_FILE_SIZE_MB override
	if v := os.Getenv("UPLOAD_MAX_FILE_SIZE_MB"); v != "" {
		if size, err := strconv.Atoi(v); err == nil && size > 0 {
			cfg.Upload.MaxFileSizeMB = size
		}
	}

	// UPLOAD_MAX_BODY_SIZE_MB override
	if v := os.Getenv("UPLOAD_MAX_BODY_SIZE_MB"); v != "" {
		if size, err := strconv.Atoi(v); err == nil && size > 0 {
			cfg.Upload.MaxBodySizeMB = size
		}
	}

	// ========== 演示模式 ==========
	// DEMO_MODE
	DemoMode = strings.ToLower(os.Getenv("DEMO_MODE")) == "true"

	// ALREADY_DONATED
	AlreadyDonated = strings.ToLower(os.Getenv("ALREADY_DONATED")) == "true"

	// ========== 代理配置 ==========
	// PROXY_HEADER override (e.g. "X-Forwarded-For")
	if v := os.Getenv("PROXY_HEADER"); v != "" {
		cfg.Server.ProxyHeader = v
	}

	// TRUSTED_PROXY_CHECK override
	if v := os.Getenv("TRUSTED_PROXY_CHECK"); v != "" {
		cfg.Server.TrustedProxyCheck = strings.ToLower(v) == "true"
	}

	// TRUSTED_PROXIES override (comma-separated)
	if v := os.Getenv("TRUSTED_PROXIES"); v != "" {
		proxies := strings.Split(v, ",")
		for i, p := range proxies {
			proxies[i] = strings.TrimSpace(p)
		}
		cfg.Server.TrustedProxies = proxies
	}

	// ========== 存储配置 ==========
	// STORAGE_NOTES_DIR override
	if v := os.Getenv("STORAGE_NOTES_DIR"); v != "" {
		cfg.Storage.NotesDir = v
	}

	// ========== 搜索配置 ==========
	// SEARCH_ENABLED override
	if v := os.Getenv("SEARCH_ENABLED"); v != "" {
		cfg.Search.Enabled = strings.ToLower(v) == "true"
	}

	// ========== 会话配置 ==========
	// AUTHENTICATION_SESSION_MAX_AGE override
	if v := os.Getenv("AUTHENTICATION_SESSION_MAX_AGE"); v != "" {
		if age, err := strconv.Atoi(v); err == nil && age > 0 {
			cfg.Authentication.SessionMaxAge = age
		}
	}

	// ========== 缓存配置 ==========
	// CACHE_TTL override
	if v := os.Getenv("CACHE_TTL"); v != "" {
		if ttl, err := strconv.Atoi(v); err == nil && ttl > 0 {
			cfg.Cache.TTL = ttl
		}
	}

	// CACHE_CAPACITY override
	if v := os.Getenv("CACHE_CAPACITY"); v != "" {
		if cap, err := strconv.Atoi(v); err == nil && cap > 0 {
			cfg.Cache.Capacity = cap
		}
	}

	// CACHE_SCAN_INTERVAL override
	if v := os.Getenv("CACHE_SCAN_INTERVAL"); v != "" {
		if interval, err := strconv.Atoi(v); err == nil && interval > 0 {
			cfg.Cache.ScanInterval = interval
		}
	}

	// ========== 上传类型配置 ==========
	// UPLOAD_ALLOWED_TYPES override (comma-separated MIME types)
	if v := os.Getenv("UPLOAD_ALLOWED_TYPES"); v != "" {
		types := strings.Split(v, ",")
		for i, t := range types {
			types[i] = strings.TrimSpace(t)
		}
		cfg.Upload.AllowedTypes = types
	}
}

// hashPasswordIfNeeded hashes plaintext password at startup
func hashPasswordIfNeeded(cfg *Config) {
	if cfg.Authentication.Password != "" {
		hash, err := bcrypt.GenerateFromPassword(
			[]byte(cfg.Authentication.Password),
			bcrypt.DefaultCost,
		)
		if err == nil {
			cfg.Authentication.PasswordHash = string(hash)
			cfg.Authentication.Password = "" // Clear plaintext for security
		}
	}
}

// DetectHTTPSAndSetSecureCookie detects if the app is running behind HTTPS
// and automatically enables secure_cookie if not explicitly set
// Returns true if HTTPS was detected and secure_cookie was auto-enabled
func DetectHTTPSAndSetSecureCookie(cfg *Config) (detected bool, source string) {
	// If secure_cookie is already explicitly set to true, no need to detect
	if cfg.Authentication.SecureCookie {
		return false, ""
	}

	// Method 1: Check HTTPS environment variable (common in PaaS platforms)
	if v := os.Getenv("HTTPS"); v != "" {
		v = strings.ToLower(v)
		if v == "true" || v == "1" || v == "on" {
			cfg.Authentication.SecureCookie = true
			return true, "HTTPS env var"
		}
	}

	// Method 2: Check X-Forwarded-Proto (reverse proxy scenario)
	if v := os.Getenv("X_FORWARDED_PROTO"); v != "" {
		if strings.ToLower(v) == "https" {
			cfg.Authentication.SecureCookie = true
			return true, "X_FORWARDED_PROTO env var"
		}
	}

	// Method 3: Check if any allowed_origin starts with https://
	for _, origin := range cfg.Server.AllowedOrigins {
		if strings.HasPrefix(strings.ToLower(origin), "https://") {
			cfg.Authentication.SecureCookie = true
			return true, "allowed_origins (https:// detected)"
		}
	}

	return false, ""
}

