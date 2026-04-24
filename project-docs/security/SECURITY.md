# Security Guide - GoNote

This document provides security recommendations for deploying and running GoNote in production environments.

---

## 🔐 Quick Security Checklist

Before exposing GoNote to the internet, complete the following:

- [ ] **Change default password** from `admin` to a strong password
- [ ] **Generate a random secret key** for session encryption
- [ ] **Enable authentication** (`authentication.enabled: true`)
- [ ] **Enable rate limiting** (`rate_limit.enabled: true`)
- [ ] **Configure CORS** with specific allowed origins (not `*`)
- [ ] **Enable secure cookies** if using HTTPS (`secure_cookie: true`)
- [ ] **Update Go version** to latest stable (minimum 1.24.13+)

---

## 🚨 Critical Security Settings

### 1. Authentication

**Default (INSECURE):**
```yaml
authentication:
  enabled: false
  password: "admin"
  secret_key: "change_this_to_a_random_secret_key_in_production"
```

**Production (SECURE):**
```yaml
authentication:
  enabled: true
  password: "YourStrongPassword123!"  # Change this!
  secret_key: "a3f8b2c1d4e5f6789012345678901234567890abcdef"  # Generate new!
  session_max_age: 604800  # 7 days
  secure_cookie: true
```

**Generate a secure secret key:**
```bash
# Using OpenSSL
openssl rand -hex 32



# Using Go
go run -e 'package main; import "crypto/rand"; import "fmt"; import "encoding/hex"; func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(hex.EncodeToString(b)) }'
```

### 2. Rate Limiting

**Default (disabled for local development):**
```yaml
rate_limit:
  enabled: false
  max_requests: 30
  window_seconds: 1
```

**Production (enabled):**
```yaml
rate_limit:
  enabled: true
  max_requests: 30
  window_seconds: 1
```

Rate limiting protects against:
- Brute force attacks on login
- API abuse
- Denial of Service (DoS)

### 3. CORS Configuration

**Default (permissive):**
```yaml
server:
  allowed_origins: ["*"]
```

**Production (restrictive):**
```yaml
server:
  allowed_origins:
    - "https://yourdomain.com"
    - "https://www.yourdomain.com"
```

### 4. Secure Cookies

When using HTTPS (recommended for production):

```yaml
authentication:
  secure_cookie: true
```

**Auto-detection:** The application automatically enables secure cookies if:
- `HTTPS=true` environment variable is set
- `X_FORWARDED_PROTO=https` (reverse proxy scenario)
- `allowed_origins` contains HTTPS URLs

---

## 🛡️ Security Features

### Built-in Protections

| Feature | Description | Status |
|---------|-------------|--------|
| **CSRF Protection** | Double Submit Cookie pattern with `X-CSRF-Token` header | ✅ Enabled |
| **Path Traversal Prevention** | `ValidatePathSecurity()` validates all file paths | ✅ Enabled |
| **Session Security** | HTTPOnly cookies, SameSite=Lax, configurable Secure | ✅ Enabled |
| **File Upload Validation** | Size limits, MIME type checking, atomic writes | ✅ Enabled |
| **Password Hashing** | bcrypt with cost factor 12 | ✅ Enabled |
| **Graceful Shutdown** | Proper cleanup of goroutines and connections | ✅ Enabled |

### File Upload Security

```yaml
upload:
  max_file_size_mb: 50       # Adjust based on needs
  max_body_size_mb: 100
  allowed_types: []          # Empty = allow all, or specify MIME types
```

**Recommended restrictions for production:**
```yaml
upload:
  allowed_types:
    - image/jpeg
    - image/png
    - image/gif
    - image/webp
    - application/pdf
```

---

## ⚠️ Known Vulnerabilities

### Go Standard Library (as of Go 1.24)

GoNote inherits 15 vulnerabilities from Go 1.24 standard library. These are **not code bugs** but known issues in Go itself.

**Critical vulnerabilities fixed in Go 1.25.8:**
- GO-2026-4602: FileInfo escape in `os` package
- GO-2026-4601: IPv6 parsing in `net/url`
- GO-2026-4340: TLS handshake encryption
- GO-2026-4337: TLS session resumption

**Recommendation:** 
- For production: Use **Go 1.24.13+** (fixes 11 vulnerabilities)
- For maximum security: Wait for **Go 1.25.8** (fixes all)

**Mitigation:** The application's `ValidatePathSecurity()` function provides additional protection against path traversal attacks, reducing the impact of some vulnerabilities.

### Application-Level Fixes

The following issues have been fixed in the codebase:

| Issue | File | Status |
|-------|------|--------|
| Invalid regex `(?!` | `internal/services/statistics.go` | ✅ Fixed |
| Deprecated `strings.Title()` | `internal/services/theme.go` | ✅ Fixed |
| Unused functions | `internal/services/backlink.go` | ✅ Removed |
| Redundant bool comparisons | `internal/services/backlink.go` | ✅ Fixed |

---

## 🔒 Deployment Scenarios

### Local Development (Default Config)

```yaml
authentication:
  enabled: false
rate_limit:
  enabled: false
server:
  allowed_origins: ["*"]
```

**Risk Level:** ✅ Safe for localhost only

### Self-Hosted (Home Network)

```yaml
authentication:
  enabled: true
  password: "ChangeMe123!"
  secret_key: "generate_random_key"
rate_limit:
  enabled: false  # Optional for trusted network
server:
  allowed_origins: ["http://192.168.1.100:8000"]
```

**Risk Level:** ✅ Low (trusted network)

### Production (Public Internet)

```yaml
authentication:
  enabled: true
  password: "StrongPassword123!"
  secret_key: "random_64_char_hex_string"
  secure_cookie: true
rate_limit:
  enabled: true
  max_requests: 30
  window_seconds: 1
server:
  allowed_origins:
    - "https://yourdomain.com"
  debug: false
upload:
  max_file_size_mb: 10
  allowed_types:
    - image/jpeg
    - image/png
    - image/gif
```

**Risk Level:** ⚠️ Requires all security measures enabled

---

## 🚀 Docker Deployment Security

### Environment Variables (Recommended)

```bash
# Authentication
AUTHENTICATION_ENABLED=true
AUTHENTICATION_PASSWORD=YourStrongPassword123!
AUTHENTICATION_SECRET_KEY=$(openssl rand -hex 32)
AUTHENTICATION_SECURE_COOKIE=true

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX=30

# Server
DEBUG=false
ALLOWED_ORIGINS=https://yourdomain.com

# HTTPS Detection (if behind reverse proxy)
X_FORWARDED_PROTO=https
```

### Docker Compose Example

```yaml
version: '3.8'
services:
  gonote:
    image: gonote/gonote:latest
    environment:
      - AUTHENTICATION_ENABLED=true
      - AUTHENTICATION_PASSWORD=${AUTH_PASSWORD:-ChangeMe123!}
      - AUTHENTICATION_SECRET_KEY=${AUTH_SECRET_KEY:-change_me}
      - RATE_LIMIT_ENABLED=true
      - DEBUG=false
    volumes:
      - ./data:/app/data
      - ./config.yaml:/app/config.yaml
    ports:
      - "8000:8000"
    restart: unless-stopped
```

---

## 📋 Security Audit Commands

Run these commands to verify your installation:

```bash
# Check for vulnerabilities
cd go && go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Run static analysis
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...

# Run tests with race detector
go test -race ./...

# Build and verify
go build ./...
```

---

## 🆘 Security Incident Response

If you suspect a security breach:

1. **Immediately change passwords:**
   ```bash
   # Update config.yaml
   password: "NewStrongPassword!"
   ```

2. **Regenerate secret key:**
   ```bash
   openssl rand -hex 32
   ```

3. **Review logs:**
   ```bash
   # Check application logs
   docker logs gonote
   ```

4. **Update to latest version:**
   ```bash
   git pull origin main
   docker-compose pull
   docker-compose up -d
   ```

---

## 📚 Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://go.dev/doc/security/)
- [Fiber Security](https://docs.gofiber.io/security/)

---

## 🤝 Reporting Security Issues

If you discover a security vulnerability, please report it privately by opening an issue on GitHub and marking it as confidential, or contact the maintainers directly.

**Do not** disclose security vulnerabilities publicly until they have been addressed.
