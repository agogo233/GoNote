# Shared Resources

This directory contains shared resources used by both the Go backend and frontend application.

## Directory Structure

```
shared/
├── frontend/          # Frontend application (HTML, JS, CSS)
│   ├── libs/         # Third-party library cache (CDN-free)
│   ├── app.js        # Main application logic
│   ├── index.html    # SPA entry point
│   ├── login.html    # Login page
│   ├── manifest.json # PWA manifest
│   ├── sw.js         # Service worker
│   └── favicon.svg   # App favicon
├── themes/           # CSS themes (16 built-in themes)
│   ├── cobalt2.css
│   ├── dark.css
│   ├── dracula.css
│   ├── gruvbox-dark.css
│   ├── light.css
│   ├── matcha-light.css
│   ├── monokai.css
│   ├── nord.css
│   ├── vs-blue.css
│   └── vue-high-contrast.css
└── locales/          # Internationalization files
    ├── en-US.json    # English translations
    └── zh-CN.json    # Chinese translations
```

## Concepts

### Frontend Libraries (`frontend/libs/`)
All third-party JavaScript libraries are cached locally to enable CDN-free operation:
- Alpine.js 3.14.1
- Tailwind CSS 3.4.17
- Marked.js 12.0.2
- MathJax 3.2.2
- Highlight.js 11.11.1 (12 language packs)
- Mermaid 11.12.2
- vis-network 10.0.2
- QRCode generator 1.4.4

This ensures the app works offline and without external dependencies.

### Themes (`themes/`)
CSS themes provide alternative color schemes. The app dynamically loads theme files via `/api/themes/<theme-name>` endpoint.

### Localization (`locales/`)
JSON files containing translation strings. The app supports English (en-US) and Chinese (zh-CN) out of the box.

## Static File Serving

The Go backend serves static files from this directory under the `/static/` path:
```
/static/frontend/    → shared/frontend/
/static/themes/      → shared/themes/
/static/locales/     → shared/locales/
```

## Adding New Resources

- **New theme**: Add CSS file to `themes/` and ensure it's registered in the theme loader
- **New translation**: Add JSON file to `locales/` with proper locale code

- **Updated library**: Replace versioned directory under `frontend/libs/` (e.g., `highlight.js/11.11.1/`)

## Related

- See `docker/README.md` for Docker volume mappings
- See `project-docs/developer-guide/ASSETS.md` for asset usage guidelines
