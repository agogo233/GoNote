# Build Configuration

This directory contains build configuration files for the GoNote project.

## Directory Structure

```
build/
└── tailwind/
    ├── input.css           # Tailwind CSS input source
    ├── tailwind.config.js  # Tailwind configuration
    └── postcss.config.js   # PostCSS configuration
```

## Building CSS

### Development

```bash
# From project root
npm install
npx tailwindcss -i ./build/tailwind/input.css -o ./shared/frontend/libs/tailwind/tailwind.css --watch
```

### Production

```bash
# From project root
npm install
npx tailwindcss -i ./build/tailwind/input.css -o ./shared/frontend/libs/tailwind/tailwind.css --minify
```

## Using Make

```bash
# Install dependencies
make deps

# Build CSS
make css-build

# Watch for changes
make css-watch
```

## Output

The compiled CSS is output to `shared/frontend/libs/tailwind/tailwind.css` and is used by the frontend application.
