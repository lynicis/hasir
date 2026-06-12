#!/bin/bash
# ==============================================================================
# Hasir Backup Script
# Performs PostgreSQL database dump, gzips it, and applies retention policy
# (7 daily backups, 4 weekly backups).
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

cd "$REPO_DIR"

# Load environment variables
if [ -f .env ]; then
  # shellcheck disable=SC1091
  source .env
else
  echo "No .env file found. Run ./scripts/setup.sh first."
  exit 1
fi

BACKUP_DIR="${REPO_DIR}/backups"
mkdir -p "$BACKUP_DIR"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
DB_BACKUP_FILE="${BACKUP_DIR}/hasir_db_${TIMESTAMP}.sql.gz"

echo "=== 1. Starting PostgreSQL Backup ==="
# Dump database from postgres service using env credentials
docker compose exec -T postgres pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" | gzip > "$DB_BACKUP_FILE"
echo "Database backup saved to: $DB_BACKUP_FILE"

echo "=== 2. Applying Retention Policy ==="
# Retention: Keep daily backups for 7 days.
# Keep weekly backups (backups taken on Sunday, Day 7) for 28 days (4 weeks).
find "$BACKUP_DIR" -name "hasir_db_*.sql.gz" -type f -mtime +7 | while read -r file; do
  # Determine day of the week (1=Monday, 7=Sunday)
  if [[ "$OSTYPE" == "darwin"* ]]; then
    DAY_OF_WEEK=$(date -r "$(stat -f %m "$file")" +%u)
  else
    DAY_OF_WEEK=$(date -r "$file" +%u)
  fi
  
  if [ "$DAY_OF_WEEK" -ne 7 ]; then
    echo "Pruning old daily backup: $file"
    rm "$file"
  fi
done

# Prune weekly backups older than 28 days
find "$BACKUP_DIR" -name "hasir_db_*.sql.gz" -type f -mtime +28 -delete
echo "Retention policy applied."

# Optional S3 Upload if AWS CLI is configured/env vars exist
if command -v aws &> /dev/null && [ "${AWS_BACKUP_BUCKET:-}" != "" ]; then
  echo "=== 3. Uploading to Amazon S3 ==="
  aws s3 cp "$DB_BACKUP_FILE" "s3://${AWS_BACKUP_BUCKET}/db/$(basename "$DB_BACKUP_FILE")"
  echo "Backup uploaded to S3."
fi

echo "=== Backup Completed Successfully ==="
