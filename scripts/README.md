# Maintenance and Utility Scripts

This directory contains shell scripts for common maintenance and operational tasks.

## Scripts

### `backup.sh`
Creates backups of the notes directory.

**Usage:**
```bash
./scripts/backup.sh
```

**Purpose:**
- Backs up the entire `data/notes/` directory
- Creates timestamped tar.gz archives
- Automatically retains the most recent backups (no rotation policy)

### `cleanup.sh`
Cleans temporary files and old cache entries.

**Usage:**
```bash
./scripts/cleanup.sh
```

**Purpose:**
- Deletes files in `data/temp/`
- Removes cache files older than 7 days from `data/cache/`
- Safe to run frequently (cron job recommended)

**Recommended cron schedule:**
```bash
0 2 * * * /path/to/gonote/scripts/cleanup.sh
```

### `migrate.sh`
Data migration and project upgrade utility.

**Usage:**
```bash
./scripts/migrate.sh
```

**Purpose:**
- Migrates data between storage formats
- Upgrades project structure
- Handles version transitions

**Note:** Always backup data before running migrations.

## Directory Structure

```
scripts/
├── backup.sh              # Backup utility
├── cleanup.sh             # Temp/cache cleaner
├── migrate.sh             # Migration tool
├── backups/               # Backup archive storage (created automatically)
└── data/                  # Example data structure (for reference only)
    ├── cache/
    ├── notes/
    └── temp/
```

## Environment

All scripts assume they are run from the **project root directory**.

The `data/` subdirectory in this script directory is **not used** at runtime; it shows the expected directory structure for reference only.

## Important Notes

- Ensure the project root is the current working directory when running scripts
- Check that `data/` exists and is writable
- Review the script headers for any required environment variables
- Test backups periodically to ensure they are valid

## Adding New Scripts

When adding new utility scripts:

1. Use clear, descriptive names ending in `.sh`
2. Include a shebang line (`#!/usr/bin/env bash`)
3. Add comments explaining purpose and usage
4. Handle errors gracefully
5. Follow existing error handling patterns (set `-euo pipefail`)

## Related

- See `docker/README.md` for containerized operations
- See `project-docs/developer-guide/DEPLOY.md` for deployment procedures
- Makefile provides additional commands: `make clean-data`, `make clean-build`, `make clean`
