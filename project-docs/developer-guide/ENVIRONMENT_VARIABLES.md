# 🔧 Environment Variables

GoNote supports environment variables to override configuration settings, allowing different behavior in different deployment environments (local, staging, production).

## 📋 Available Environment Variables

### Core Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `PORT` | integer | `9000` | HTTP port for the application (Docker, Go backend) |

> **Note**: Advanced server settings (CORS origins, debug mode) are configured via `config.yaml` only, not via environment variables. See [config.yaml](#advanced-server-configuration) for details.

### Authentication

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `AUTHENTICATION_ENABLED` | boolean | `config.yaml` | Enable/disable authentication |
| `AUTHENTICATION_PASSWORD` | string | `admin` | Plain text password (hashed automatically at startup) |
| `AUTHENTICATION_PASSWORD_HASH` | string | - | Pre-hashed bcrypt password (for advanced users) |
| `AUTHENTICATION_SECRET_KEY` | string | `config.yaml` | Session secret key (for session security) |
| `AUTHENTICATION_SECURE_COOKIE` | boolean | `false` | Force enable secure cookies (auto-detected if not set) |

> **Password Priority:** `AUTHENTICATION_PASSWORD` takes precedence over `AUTHENTICATION_PASSWORD_HASH`. If both are set, the plain text password is used.

#### 🔒 HTTPS Auto-Detection for Secure Cookies

When `secure_cookie` is not explicitly set to `true`, the system automatically detects HTTPS and enables secure cookies:

| Detection Method | Environment Variable | Value to Trigger |
|-----------------|---------------------|------------------|
| HTTPS flag | `HTTPS` | `true`, `1`, or `on` |
| Reverse proxy | `X_FORWARDED_PROTO` | `https` |
| Allowed origins | `allowed_origins` in config | Contains `https://` URL |

**Example: PaaS Deployment (Render, Railway, etc.)**
```bash
# Most PaaS platforms set HTTPS=true automatically
# No additional configuration needed!
```

**Example: Reverse Proxy (Nginx, Caddy, etc.)**
```bash
# Set the forwarded protocol
export X_FORWARDED_PROTO=https
```

**Example: Docker Compose**
```yaml
environment:
  - X_FORWARDED_PROTO=https
```

> **Security Note:** Secure cookies are only sent over HTTPS connections. If you're behind a reverse proxy terminating SSL, ensure the proxy forwards the correct protocol headers.

#### Example: Setting password via environment variable

```bash
# Docker
docker run -e AUTHENTICATION_ENABLED=true -e AUTHENTICATION_PASSWORD=mysecretpassword ...

# Docker Compose (in .env file or docker-compose.yml)
AUTHENTICATION_PASSWORD=mysecretpassword
```

### Demo Mode

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `DEMO_MODE` | boolean | `false` | Enable demo mode (enables rate limiting and other demo restrictions) |

### Support

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `ALREADY_DONATED` | boolean | `false` | Hides the support buttons in the Settings pane |

> ⚠️ **Disclaimer:** No verification exists. But legend says that setting this to `true` without donating causes your next `git push` to fail silently. Just once. When it matters most.
>
> Haven't donated yet? [☕ Buy me a coffee](https://ko-fi.com/gamosoft) - it takes 30 seconds and makes my day!

## 🎯 Configuration Priority

Configuration is loaded in this order (later overrides earlier):

1. **`config.yaml`** - Default configuration file
2. **Environment Variables** - Runtime overrides
3. **Command Line** - Highest priority (if applicable)

## 🔧 Advanced Server Configuration

The following settings are available in `config.yaml` only (not via environment variables):

### CORS (Cross-Origin Resource Sharing)

```yaml
server:
  # List of allowed origins for CORS
  # Default: ["*"] allows all origins (fine for self-hosted)
  # Production: specify your domains
  allowed_origins: ["*"]
  
  # Examples for production:
  # allowed_origins: ["http://localhost:8000", "https://yourdomain.com"]
  # allowed_origins: ["https://*.yourdomain.com"]  # Wildcard subdomain
```

**Security Note:**
- `["*"]` is **safe for self-hosted** deployments on private networks
- For **public deployments**, specify exact origins to prevent unauthorized API access
- This prevents CSRF attacks when authentication is enabled

### Debug Mode

```yaml
server:
  # Enable detailed error messages in API responses
  # Default: false (production-safe)
  # Set to true for development/troubleshooting
  debug: false
```

**⚠️ CRITICAL**: Never enable `debug: true` in production!

When `debug: true`:
- Full error stack traces are returned to users
- Internal paths and system details are exposed
- Security vulnerabilities may be revealed

When `debug: false` (recommended):
- Generic error messages are returned
- Full error details are logged server-side only
- Production-safe error handling

---

## 📚 Related Documentation

- **Authentication**: [AUTHENTICATION.md](AUTHENTICATION.md)
- **API Rate Limiting**: [API.md](API.md#rate-limiting)

---

**Pro Tip:** Use environment variables for **deployment-specific** settings, and `config.yaml` for **application defaults**. This keeps your configuration flexible and maintainable! 🎯

