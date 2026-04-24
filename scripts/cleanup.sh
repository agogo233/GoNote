#!/bin/bash
# cleanup.sh - Automated cleanup script for GoNote
# Cleans temporary files and old cache

set -e

BASE_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DATA_DIR="$BASE_DIR/data"
LOG_DIR="$BASE_DIR/logs"

echo "[$(date)] Starting cleanup..."

# Clean temporary files
if [ -d "$DATA_DIR/temp" ]; then
    echo "Cleaning temporary files..."
    find "$DATA_DIR/temp" -type f -mtime +1 -delete 2>/dev/null || true
fi

# Clean cache files older than 7 days
if [ -d "$DATA_DIR/cache" ]; then
    echo "Cleaning cache files older than 7 days..."
    find "$DATA_DIR/cache" -type f -mtime +7 -delete 2>/dev/null || true
fi

# Rotate logs
mkdir -p "$LOG_DIR"
find "$LOG_DIR" -name "*.log" -mtime +30 -delete 2>/dev/null || true

echo "[$(date)] Cleanup completed."

# Create log entry
echo "$(date): Cleanup completed" >> "$LOG_DIR/cleanup.log"