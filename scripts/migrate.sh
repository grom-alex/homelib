#!/usr/bin/env bash
# Проверка статуса миграций HomeLib
#
# Миграции встроены в бинарник (go:embed) и выполняются
# автоматически при запуске API-сервера.
# Этот скрипт — обёртка для проверки статуса.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DOCKER_DIR="$(dirname "$SCRIPT_DIR")/docker"
COMPOSE_FILE="${1:-docker-compose.dev.yml}"

echo "=== Статус миграций ==="
echo "Миграции применяются автоматически при запуске API-сервера."
echo "Проверка логов:"
docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" logs api 2>/dev/null | grep -i migrat || \
  echo "Нет данных о миграциях. API-сервер запущен?"
