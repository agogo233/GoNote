# Security Policy

## Reporting a Vulnerability

We take the security of GoNote seriously. If you discover a security vulnerability, please follow these steps:

### How to Report

1. **Open a GitHub Issue** and mark it as confidential (if you have repository access)
2. **Or contact directly** via GitHub Discussions with a private message

### What to Include

Please provide the following information in your report:

- Description of the vulnerability
- Steps to reproduce
- Potential impact assessment
- Suggested fix (if you have one)

### Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial assessment**: Within 1 week
- **Fix timeline**: Depends on severity
  - Critical: Within 1 week
  - High: Within 2 weeks
  - Medium: Within 1 month
  - Low: Next release cycle

### Security Best Practices for Users

- Always change the default password before exposing to any network
- Use a strong, unique `AUTHENTICATION_SECRET_KEY`
- Run behind HTTPS (reverse proxy recommended)
- Keep Go and dependencies up to date
- Regularly backup your notes data

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.25.x  | ✅        |
| 0.24.x  | ✅        |
| < 0.24  | ❌        |

## Security Features

GoNote includes the following security features:

- **CSRF Protection**: Double Submit Cookie pattern
- **Session Security**: SameSite=Lax cookies, configurable Secure flag
- **Rate Limiting**: Configurable per-endpoint limits
- **Path Validation**: Directory traversal prevention
- **Error Sanitization**: Production mode hides sensitive paths
- **Password Hashing**: bcrypt with configurable cost

---

Thank you for helping keep GoNote secure! 🙏
