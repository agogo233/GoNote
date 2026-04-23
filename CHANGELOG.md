# Changelog

All notable changes to GoNote will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Project file reorganization with new directory structure
- Centralized Docker configuration in `docker/` directory
- Categorized documentation in `project-docs/` with subdirectories
- Test directory reorganization with `tests/e2e/` structure
- Build configuration directory (`build/`) for Tailwind
- Shared assets directory (`shared/assets/`) for project-owned resources
- Deployment configuration directory (`deploy/`) for platform-specific configs
- CODEOWNERS file for automated code review assignments
- **Chinese documentation for all user guides** (FEATURES_CN, THEMES_CN, TAGS_CN, TEMPLATES_CN, MATHJAX_CN, MERMAID_CN, PLUGINS_CN, SHARING_CN)
- Updated main README.md with bilingual content (English + Chinese)
- Updated project-docs/README.md with comprehensive documentation index

### Changed
- Moved Tailwind configuration to `build/tailwind/`
- Consolidated documentation from `documentation/` into `project-docs/`
- Reorganized tests into `tests/e2e/` subdirectory
- Migrated `config/` files to appropriate locations (`tests/`, `deploy/`, `build/`)
- Updated `.gitignore` with new build output paths

### Removed
- `scripts/` directory (release scripts no longer needed)
- `config/` directory (files migrated)
- Root-level `docker-compose.ghcr.yml` (use `docker/compose/production.yml`)

### Deprecated
- Root-level `input.css` (use `build/tailwind/input.css`)
- Root-level `tailwind.config.js` (use `build/tailwind/tailwind.config.js`)
- Old docker-compose files (use `docker/compose/`)

### Security
- Added SECURITY_CONTACT.md for security reporting

---

## [0.23.0] - 2026-03-23

### Added
- Go 1.24+ backend with Fiber framework
- Enhanced search with inverted index
- Graph view for note relationships
- Multiple themes (16 built-in)
- Plugin system
- Share tokens for notes
- MathJax support
- Mermaid diagrams
- Syntax highlighting

### Changed
- Primary backend switched from Python to Go
- Improved performance and reduced memory footprint

### Security
- CSRF protection with Double Submit Cookie pattern
- Session security with SameSite cookies
- Rate limiting
- Path validation

---

## [0.3.0] - 2025-XX-XX

### Added
- Initial release with Python FastAPI backend
- Basic note CRUD operations
- Markdown support
- Tag extraction
- Search functionality

---

## Notes

- Version numbers follow semantic versioning (MAJOR.MINOR.PATCH)
- Dates are in YYYY-MM-DD format
- Links to releases: https://github.com/gamosoft/gonote/releases
