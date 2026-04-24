# Shared Assets

This directory contains project-owned assets for GoNote.

## Directory Structure

```
shared/assets/
├── css/            # Compiled CSS files
├── icons/          # Project icons (SVG, PNG)
└── images/         # Project images and graphics
```

## Purpose

This directory is for **project-owned assets** that are created and maintained by the GoNote project.

### What Goes Here

- ✅ Compiled CSS output (Tailwind)
- ✅ Official GoNote logos and icons
- ✅ Marketing graphics and screenshots
- ✅ Email templates and social media images
- ✅ Favicon and PWA assets

### What Does NOT Go Here

- ❌ Third-party libraries (use `shared/frontend/libs/`)
- ❌ User-generated content (use `data/`)
- ❌ Theme files (use `shared/themes/`)

- ❌ Translation files (use `shared/locales/`)

## CSS Output

The compiled Tailwind CSS is output to `css/tailwind.css`:

```bash
# Build CSS
npx tailwindcss -i ../../build/tailwind/input.css -o ./css/tailwind.css --minify
```

## Icon Usage

Icons should be provided in multiple formats:

| Format | Use Case |
|--------|----------|
| SVG | Web, scalable graphics |
| PNG | Fallback, social media |
| ICO | Browser favicon |

### Recommended Sizes

- 16x16 - Browser tab icon
- 32x32 - Small UI elements
- 128x128 - App launcher
- 512x512 - PWA, store listings

## Related Directories

| Directory | Purpose |
|-----------|---------|
| `shared/frontend/libs/` | Third-party library cache |
| `shared/themes/` | CSS themes for the app |

| `shared/locales/` | Translation files |
| `docs/` | Website documentation assets |
