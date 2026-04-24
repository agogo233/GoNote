#!/bin/bash
# backup.sh - Backup script for GoNote notes
# Creates weekly backup of notes directory

set -e

BASE_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DATA_DIR="$BASE_DIR/data"
BACKUP_DIR="$BASE_DIR/backups"
DATE=$(date +%Y%m%d_%H%M%S)

echo "[$(date)] Starting backup..."

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Create backup archive
if [ -d "$DATA_DIR/notes" ]; then
    tar -czf "$BACKUP_DIR/notes_backup_$DATE.tar.gz" -C "$BASE_DIR" data/notes
    echo "[$(date)] Backup created: $BACKUP_DIR/notes_backup_$DATE.tar.gz"
else
    echo "[$(date)] No notes directory to backup."
fi

# Keep only last 4 backups
cd "$BACKUP_DIR"
ls -t notes_backup_*.tar.gz 2>/dev/null | tail -n +5 | xargs -r rm
echo "[$(date)] Backup completed."