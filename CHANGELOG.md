# Changelog

All notable changes to GoNote will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- New documentation structure in `project-docs/` with bilingual (English/Chinese) support
- Comprehensive developer guides (API, environment variables, testing)
- Security documentation (SECURITY.md, AUTHENTICATION.md)
- Contribute guidelines (CONTRIBUTING.md)

### Changed
- Project structure reorganization - data paths now unified under project root
- Migration from `go/data/` to `./data/` for consistent Docker and local operation
- Configuration system improvements with environment variable overrides

### Fixed
- Docker build path issues
- Configuration schema validation
- CORS default values

### Removed
- Python-related content (no longer used)
- Plugin system references

---

## [0.25.0] - 2026-04-24

### Added
- Full bilingual support (English/Chinese) for all user-facing documentation
- Dedicated developer guide section with API documentation and environment variable reference
- Security hardening guide and authentication setup documentation
- E2E test suite with 44 spec files covering 12 feature areas
- Playwright test fixtures for consistent test data
- Test coverage reporting

### Changed
- Migrated to Go 1.24+ with Fiber framework
- Refactored project structure for better modularity
- Improved Docker setup with separate development and production compose files
- Enhanced configuration validation and error handling
- Optimized search indexing for better performance
- Improved background service handling for graceful shutdown

### Fixed
- Multiple configuration path resolution issues
- Build inconsistencies between different environments
- Security defaults and cookie settings
- Data directory handling for Docker and local runs

### Removed
- Legacy deployment documentation (replaced by comprehensive guides)
- Unused dependencies and deprecated code

---

## [0.24.x] - Pre-2026

### Added
- Initial public release with basic note-taking functionality
- Markdown file storage
- Basic authentication
- Theme system
- Tag organization
- Search functionality
- Graph view
- Sharing capabilities

---

## Version Legend

- **0.25.0** - Major documentation and infrastructure update
- **0.24.x** - Early releases before restructuring

---

## Converting Git History to Changelog

Future changelog entries will be automatically generated from git commit messages following Conventional Commits format. Example:

```bash
# Generate release notes from git tags and commits
git log --pretty=format:"- %s (%h)" v0.25.0..HEAD
```

---

## How to Cite This Changelog

This file follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) standards. Each version section describes:

- **Added** for new features
- **Changed** for modifications to existing functionality
- **Deprecated** for soon-to-be-removed features
- **Removed** for now-removed features
- **Fixed** for bug fixes
- **Security** for security-related changes

---

**Note:** Detailed commit history can be viewed in the git repository:
```bash
git log --oneline --all
```
