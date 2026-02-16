#!/usr/bin/env bash
# Бэкап базы данных HomeLib
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DOCKER_DIR="$(dirname "$SCRIPT_DIR")/docker"
COMPOSE_FILE="${1:-docker-compose.dev.yml}"
BACKUP_DIR="${2:-$SCRIPT_DIR/../backups}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/homelib_${TIMESTAMP}.sql.gz"

mkdir -p "$BACKUP_DIR"

echo "=== Создание бэкапа базы данных ==="
docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" exec -T postgres \
  pg_dump -U homelib homelib | gzip > "$BACKUP_FILE"

echo "Бэкап сохранён: $BACKUP_FILE"
ls -lh "$BACKUP_FILE"
