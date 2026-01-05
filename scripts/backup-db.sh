#!/bin/bash
# VRC Shift Scheduler - Database Backup Script
# Backs up PostgreSQL database to Cloudflare R2
# Run via cron: 0 4 * * * /opt/vrcshift/scripts/backup-db.sh >> /var/log/db-backup.log 2>&1

set -e

# Configuration
BACKUP_DIR="/tmp/db-backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="vrcshift_backup_${TIMESTAMP}.sql.gz"
R2_BUCKET="eventshift-db-backup"
R2_ENDPOINT="https://80dac8172d46968cd0d2da338cbe7598.r2.cloudflarestorage.com"
RETENTION_DAYS=30

# Docker container name
DB_CONTAINER="vrc-shift-db"
DB_USER="vrcshift"
DB_NAME="vrcshift"

# Create backup directory
mkdir -p "$BACKUP_DIR"

echo "[$(date)] Starting database backup..."

# Create database dump and compress
docker exec "$DB_CONTAINER" pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_DIR/$BACKUP_FILE"

# Check if backup was created successfully
if [ -f "$BACKUP_DIR/$BACKUP_FILE" ]; then
    BACKUP_SIZE=$(du -h "$BACKUP_DIR/$BACKUP_FILE" | cut -f1)
    echo "[$(date)] Backup created: $BACKUP_FILE ($BACKUP_SIZE)"
else
    echo "[$(date)] ERROR: Backup file not created"
    exit 1
fi

# Upload to R2 using AWS CLI (S3-compatible)
echo "[$(date)] Uploading to R2..."
aws s3 cp "$BACKUP_DIR/$BACKUP_FILE" "s3://$R2_BUCKET/$BACKUP_FILE" \
    --endpoint-url "$R2_ENDPOINT"

if [ $? -eq 0 ]; then
    echo "[$(date)] Upload successful"
else
    echo "[$(date)] ERROR: Upload failed"
    exit 1
fi

# Clean up old backups (keep last RETENTION_DAYS days)
echo "[$(date)] Cleaning up old backups..."

# Delete old files from R2
aws s3 ls "s3://$R2_BUCKET/" --endpoint-url "$R2_ENDPOINT" | while read -r line; do
    file_date=$(echo "$line" | awk '{print $1}')
    file_name=$(echo "$line" | awk '{print $4}')

    if [ -n "$file_name" ]; then
        # Calculate age in days
        file_timestamp=$(date -d "$file_date" +%s 2>/dev/null || echo "0")
        current_timestamp=$(date +%s)
        age_days=$(( (current_timestamp - file_timestamp) / 86400 ))

        if [ "$age_days" -gt "$RETENTION_DAYS" ]; then
            echo "[$(date)] Deleting old backup: $file_name (${age_days} days old)"
            aws s3 rm "s3://$R2_BUCKET/$file_name" --endpoint-url "$R2_ENDPOINT"
        fi
    fi
done

# Clean up local backup file
rm -f "$BACKUP_DIR/$BACKUP_FILE"

echo "[$(date)] Backup completed successfully"
