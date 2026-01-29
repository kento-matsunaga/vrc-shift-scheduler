#!/bin/bash
# VRC Shift Scheduler - Database Backup Script
# Backs up PostgreSQL database to Cloudflare R2
#
# Setup Requirements:
#   1. AWS CLI installed: apt install awscli
#   2. AWS credentials configured in ~/.aws/credentials:
#      [default]
#      aws_access_key_id = <R2_ACCESS_KEY_ID>
#      aws_secret_access_key = <R2_SECRET_ACCESS_KEY>
#   3. AWS config in ~/.aws/config:
#      [default]
#      region = auto
#   4. Environment variable R2_ENDPOINT set (or use default)
#
# Cron setup:
#   0 4 * * * /opt/vrcshift/scripts/backup-db.sh >> /var/log/db-backup.log 2>&1
#
# Logrotate setup (to manage log file size):
#   sudo cp scripts/logrotate.d/db-backup /etc/logrotate.d/db-backup

set -euo pipefail

# Configuration (can be overridden by environment variables)
BACKUP_DIR="${BACKUP_DIR:-/tmp/db-backups}"
R2_BUCKET="${R2_BUCKET:-eventshift-db-backup}"
R2_ENDPOINT="${R2_ENDPOINT:-https://80dac8172d46968cd0d2da338cbe7598.r2.cloudflarestorage.com}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
DB_CONTAINER="${DB_CONTAINER:-vrc-shift-db}"
DB_USER="${DB_USER:-vrcshift}"
DB_NAME="${DB_NAME:-vrcshift}"
MIN_BACKUP_SIZE="${MIN_BACKUP_SIZE:-1024}"  # Minimum expected backup size in bytes

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="vrcshift_backup_${TIMESTAMP}.sql.gz"

# Pre-flight checks
preflight_check() {
    local errors=0

    # Check AWS CLI
    if ! command -v aws &> /dev/null; then
        echo "[$(date)] ERROR: AWS CLI is not installed"
        errors=$((errors + 1))
    fi

    # Check Docker
    if ! command -v docker &> /dev/null; then
        echo "[$(date)] ERROR: Docker is not installed"
        errors=$((errors + 1))
    fi

    # Check if DB container is running
    if ! docker ps --format '{{.Names}}' | grep -q "^${DB_CONTAINER}$"; then
        echo "[$(date)] ERROR: Database container '${DB_CONTAINER}' is not running"
        errors=$((errors + 1))
    fi

    # Check AWS credentials
    if ! aws sts get-caller-identity --endpoint-url "$R2_ENDPOINT" &> /dev/null 2>&1; then
        # R2 doesn't support STS, so just check if credentials file exists
        if [ ! -f ~/.aws/credentials ]; then
            echo "[$(date)] ERROR: AWS credentials not configured (~/.aws/credentials not found)"
            errors=$((errors + 1))
        fi
    fi

    if [ $errors -gt 0 ]; then
        echo "[$(date)] Pre-flight check failed with $errors error(s)"
        exit 1
    fi

    echo "[$(date)] Pre-flight checks passed"
}

# Create backup
create_backup() {
    mkdir -p "$BACKUP_DIR"

    echo "[$(date)] Starting database backup..."

    # Create database dump and compress
    if ! docker exec "$DB_CONTAINER" pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_DIR/$BACKUP_FILE"; then
        echo "[$(date)] ERROR: Failed to create database dump"
        exit 1
    fi

    # Verify backup was created and has minimum size
    if [ ! -f "$BACKUP_DIR/$BACKUP_FILE" ]; then
        echo "[$(date)] ERROR: Backup file not created"
        exit 1
    fi

    local backup_size
    backup_size=$(stat -c%s "$BACKUP_DIR/$BACKUP_FILE" 2>/dev/null || stat -f%z "$BACKUP_DIR/$BACKUP_FILE" 2>/dev/null)

    if [ "$backup_size" -lt "$MIN_BACKUP_SIZE" ]; then
        echo "[$(date)] ERROR: Backup file too small (${backup_size} bytes), possibly corrupted"
        rm -f "$BACKUP_DIR/$BACKUP_FILE"
        exit 1
    fi

    local backup_size_human
    backup_size_human=$(du -h "$BACKUP_DIR/$BACKUP_FILE" | cut -f1)
    echo "[$(date)] Backup created: $BACKUP_FILE ($backup_size_human)"
}

# Upload to R2
upload_backup() {
    echo "[$(date)] Uploading to R2..."

    if ! aws s3 cp "$BACKUP_DIR/$BACKUP_FILE" "s3://$R2_BUCKET/$BACKUP_FILE" \
        --endpoint-url "$R2_ENDPOINT"; then
        echo "[$(date)] ERROR: Upload failed"
        exit 1
    fi

    echo "[$(date)] Upload successful"
}

# Clean up old backups
cleanup_old_backups() {
    echo "[$(date)] Cleaning up old backups..."

    local current_timestamp
    current_timestamp=$(date +%s)
    local cleanup_errors=0

    # Get list of files and process
    aws s3 ls "s3://$R2_BUCKET/" --endpoint-url "$R2_ENDPOINT" 2>/dev/null | while read -r line; do
        local file_date file_name
        file_date=$(echo "$line" | awk '{print $1}')
        file_name=$(echo "$line" | awk '{print $4}')

        if [ -z "$file_name" ]; then
            continue
        fi

        # Parse date and calculate age
        local file_timestamp
        file_timestamp=$(date -d "$file_date" +%s 2>/dev/null) || {
            echo "[$(date)] WARNING: Could not parse date for $file_name, skipping"
            continue
        }

        local age_days
        age_days=$(( (current_timestamp - file_timestamp) / 86400 ))

        if [ "$age_days" -gt "$RETENTION_DAYS" ]; then
            echo "[$(date)] Deleting old backup: $file_name (${age_days} days old)"
            if ! aws s3 rm "s3://$R2_BUCKET/$file_name" --endpoint-url "$R2_ENDPOINT" 2>/dev/null; then
                echo "[$(date)] WARNING: Failed to delete $file_name"
                cleanup_errors=$((cleanup_errors + 1))
            fi
        fi
    done

    if [ $cleanup_errors -gt 0 ]; then
        echo "[$(date)] WARNING: $cleanup_errors file(s) failed to delete during cleanup"
    fi
}

# Clean up local files
cleanup_local() {
    rm -f "$BACKUP_DIR/$BACKUP_FILE"
}

# Main execution
main() {
    preflight_check
    create_backup
    upload_backup
    cleanup_old_backups
    cleanup_local

    echo "[$(date)] Backup completed successfully"
}

main "$@"
