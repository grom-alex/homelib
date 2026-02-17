#!/usr/bin/env bash
# Восстановление базы данных HomeLib из бэкапа
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=scripts/lib/logging.sh
. "$SCRIPT_DIR/lib/logging.sh"
DOCKER_DIR="$(dirname "$SCRIPT_DIR")/docker"

if [ -z "${1:-}" ]; then
  log_error "Использование: $0 <файл_бэкапа.sql.gz> [compose-файл]"
  log_error "Пример: $0 backups/homelib_20260217.sql.gz docker-compose.dev.yml"
  exit 1
fi

BACKUP_FILE="$1"
COMPOSE_FILE="${2:-docker-compose.dev.yml}"

if [ ! -f "$BACKUP_FILE" ]; then
  log_error "Файл $BACKUP_FILE не найден."
  exit 1
fi

log_warn "ВНИМАНИЕ: Текущая база данных будет перезаписана!"
log_warn "Сервисы api и worker будут остановлены на время восстановления."
read -rp "Продолжить? (y/N): " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
  log_info "Отменено."
  exit 0
fi

log_info "Остановка api и worker..."
docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" stop api worker

log_info "Восстановление из $BACKUP_FILE..."
gunzip -c "$BACKUP_FILE" | docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" exec -T postgres \
  psql -U homelib homelib

log_info "Запуск api и worker..."
docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" start api worker

log_success "Восстановление завершено."
