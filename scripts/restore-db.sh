#!/usr/bin/env bash
# Восстановление базы данных HomeLib из бэкапа
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DOCKER_DIR="$(dirname "$SCRIPT_DIR")/docker"

if [ -z "${1:-}" ]; then
  echo "Использование: $0 <файл_бэкапа.sql.gz> [compose-файл]"
  echo "Пример: $0 backups/homelib_20260217.sql.gz docker-compose.dev.yml"
  exit 1
fi

BACKUP_FILE="$1"
COMPOSE_FILE="${2:-docker-compose.dev.yml}"

if [ ! -f "$BACKUP_FILE" ]; then
  echo "Ошибка: файл $BACKUP_FILE не найден."
  exit 1
fi

echo "ВНИМАНИЕ: Текущая база данных будет перезаписана!"
echo "Сервисы api и worker будут остановлены на время восстановления."
read -rp "Продолжить? (y/N): " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
  echo "Отменено."
  exit 0
fi

echo "=== Остановка api и worker ==="
docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" stop api worker

echo "=== Восстановление из $BACKUP_FILE ==="
gunzip -c "$BACKUP_FILE" | docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" exec -T postgres \
  psql -U homelib homelib

echo "=== Запуск api и worker ==="
docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" start api worker

echo "Восстановление завершено."
