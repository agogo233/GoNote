package config

import (
	"os"
	"testing"
)

func TestDetectHTTPSAndSetSecureCookie_HTTPSEnvVar(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"HTTPS=true", "true", true},
		{"HTTPS=1", "1", true},
		{"HTTPS=on", "on", true},
		{"HTTPS=ON", "ON", true},
		{"HTTPS=False", "false", false},
		{"HTTPS=0", "0", false},
		{"HTTPS=off", "off", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset
			os.Unsetenv("HTTPS")
			os.Unsetenv("X_FORWARDED_PROTO")

			cfg := &Config{
				Server: ServerConfig{
					AllowedOrigins: []string{"*"},
				},
				Authentication: AuthConfig{
					SecureCookie: false,
				},
			}

			os.Setenv("HTTPS", tt.envValue)
			defer os.Unsetenv("HTTPS")

			detected, _ := DetectHTTPSAndSetSecureCookie(cfg)
			if detected != tt.expected {
				t.Errorf("expected detected=%v, got %v", tt.expected, detected)
			}
			if cfg.Authentication.SecureCookie != tt.expected {
				t.Errorf("expected SecureCookie=%v, got %v", tt.expected, cfg.Authentication.SecureCookie)
			}
		})
	}
}

func TestDetectHTTPSAndSetSecureCookie_XForwardedProto(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"X_FORWARDED_PROTO=https", "https", true},
		{"X_FORWARDED_PROTO=HTTPS", "HTTPS", true},
		{"X_FORWARDED_PROTO=http", "http", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset
			os.Unsetenv("HTTPS")
			os.Unsetenv("X_FORWARDED_PROTO")

			cfg := &Config{
				Server: ServerConfig{
					AllowedOrigins: []string{"*"},
				},
				Authentication: AuthConfig{
					SecureCookie: false,
				},
			}

			os.Setenv("X_FORWARDED_PROTO", tt.envValue)
			defer os.Unsetenv("X_FORWARDED_PROTO")

			detected, _ := DetectHTTPSAndSetSecureCookie(cfg)
			if detected != tt.expected {
				t.Errorf("expected detected=%v, got %v", tt.expected, detected)
			}
			if cfg.Authentication.SecureCookie != tt.expected {
				t.Errorf("expected SecureCookie=%v, got %v", tt.expected, cfg.Authentication.SecureCookie)
			}
		})
	}
}

func TestDetectHTTPSAndSetSecureCookie_AllowedOrigins(t *testing.T) {
	tests := []struct {
		name      string
		origins   []string
		expected  bool
	}{
		{"https origin", []string{"https://example.com"}, true},
		{"multiple origins with https", []string{"http://localhost:8000", "https://example.com"}, true},
		{"https with wildcard", []string{"https://*.example.com"}, true},
		{"http only", []string{"http://localhost:8000"}, false},
		{"wildcard only", []string{"*"}, false},
		{"mixed case HTTPS", []string{"HTTPS://example.com"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset
			os.Unsetenv("HTTPS")
			os.Unsetenv("X_FORWARDED_PROTO")

			cfg := &Config{
				Server: ServerConfig{
					AllowedOrigins: tt.origins,
				},
				Authentication: AuthConfig{
					SecureCookie: false,
				},
			}

			detected, _ := DetectHTTPSAndSetSecureCookie(cfg)
			if detected != tt.expected {
				t.Errorf("expected detected=%v, got %v", tt.expected, detected)
			}
			if cfg.Authentication.SecureCookie != tt.expected {
				t.Errorf("expected SecureCookie=%v, got %v", tt.expected, cfg.Authentication.SecureCookie)
			}
		})
	}
}

func TestDetectHTTPSAndSetSecureCookie_AlreadySet(t *testing.T) {
	// Reset
	os.Unsetenv("HTTPS")
	os.Unsetenv("X_FORWARDED_PROTO")

	cfg := &Config{
		Server: ServerConfig{
			AllowedOrigins: []string{"http://localhost"},
		},
		Authentication: AuthConfig{
			SecureCookie: true, // Already set
		},
	}

	detected, _ := DetectHTTPSAndSetSecureCookie(cfg)
	if detected {
		t.Error("expected no detection when SecureCookie already true")
	}
	if !cfg.Authentication.SecureCookie {
		t.Error("SecureCookie should remain true")
	}
}

func TestDetectHTTPSAndSetSecureCookie_Priority(t *testing.T) {
	// HTTPS env var should take priority over X_FORWARDED_PROTO
	os.Unsetenv("HTTPS")
	os.Unsetenv("X_FORWARDED_PROTO")

	cfg := &Config{
		Server: ServerConfig{
			AllowedOrigins: []string{"*"},
		},
		Authentication: AuthConfig{
			SecureCookie: false,
		},
	}

	// Set both env vars, HTTPS=true should win
	os.Setenv("HTTPS", "true")
	os.Setenv("X_FORWARDED_PROTO", "http") // This would normally not enable secure cookie
	defer os.Unsetenv("HTTPS")
	defer os.Unsetenv("X_FORWARDED_PROTO")

	detected, source := DetectHTTPSAndSetSecureCookie(cfg)
	if !detected {
		t.Error("expected HTTPS=true to enable secure cookie")
	}
	if source != "HTTPS env var" {
		t.Errorf("expected source='HTTPS env var', got '%s'", source)
	}
}
