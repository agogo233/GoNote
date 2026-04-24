# Docker Configuration

This directory contains all Docker-related configuration files for GoNote.

## Directory Structure

```
docker/
├── compose/
│   ├── development.yml    # Go backend development configuration
│   └── production.yml     # Production configuration (uses pre-built image)
└── go/
    └── Dockerfile         # Go backend Dockerfile
```

## Quick Start

### Production Deployment (Recommended)

Use the pre-built image from GitHub Container Registry:

```bash
# Using the production compose file
docker-compose -f docker/compose/production.yml up -d
```

Or use the root-level shortcut:
```bash
docker-compose -f docker-compose.ghcr.yml up -d
```

### Development

Build and run from local source:

```bash
# Using the development compose file
docker-compose -f docker/compose/development.yml up -d
```

## Configuration Files

### production.yml
- Uses pre-built image: `ghcr.io/gamosoft/gonote:go`
- Fastest deployment option
- Recommended for production use
- Volume mapping for data persistence

### development.yml
- Builds from local Go source code
- Includes volume mounts for hot reload
- Use for development and testing

## Docker Commands

```bash
# Start containers
docker-compose -f docker/compose/production.yml up -d

# Stop containers
docker-compose -f docker/compose/production.yml down

# View logs
docker-compose -f docker/compose/production.yml logs -f

# Restart containers
docker-compose -f docker/compose/production.yml restart

# Rebuild (development only)
docker-compose -f docker/compose/development.yml up -d --build
```

## Volume Mappings

| Container Path | Purpose | Required |
|----------------|---------|----------|
| `/app/data` | Your notes | Yes |
| `/app/config.yaml` | Configuration | No (bundled) |
| `/app/themes` | Custom themes | No (bundled) |

| `/app/locales` | Translations | No (bundled) |

## Environment Variables

See [project-docs/developer-guide/ENVIRONMENT_VARIABLES.md](../project-docs/developer-guide/ENVIRONMENT_VARIABLES.md) for complete list.

## Building Custom Images

```bash
docker build -f docker/go/Dockerfile -t gonote:custom .
```

## Health Check

The production image includes a health check:

```bash
docker inspect --format='{{.State.Health.Status}}' gonote
```

## Troubleshooting

### Check container logs
```bash
docker logs gonote
```

### Access container shell
```bash
docker exec -it gonote sh
```

### View container details
```bash
docker inspect gonote
```

## Migration from Old Structure

If you were using the old docker-compose files:

| Old Path | New Path |
|----------|----------|
| `go/docker-compose.yml` | `docker/compose/development.yml` |
| `docker-compose.ghcr.yml` | `docker/compose/production.yml` |
| `go/Dockerfile` | `docker/go/Dockerfile` |

The old files remain in place for backward compatibility, but please update to the new structure.
