#!/usr/bin/env bash
# Деплой HomeLib
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"
ENV="${1:-prod}"
COMPOSE_FILE="docker-compose.${ENV}.yml"

if [ ! -f "$DOCKER_DIR/$COMPOSE_FILE" ]; then
  echo "Ошибка: файл $DOCKER_DIR/$COMPOSE_FILE не найден."
  echo "Доступные окружения: dev, stage, prod"
  exit 1
fi

echo "=== Деплой HomeLib (окружение: $ENV) ==="

# Проверка .env
if [ ! -f "$PROJECT_ROOT/.env" ]; then
  echo "Ошибка: файл .env не найден. Скопируйте .env.example и заполните значения."
  exit 1
fi

# Проверка обязательных переменных для не-dev окружений
if [ "$ENV" != "dev" ]; then
  source "$PROJECT_ROOT/.env"
  if [ -z "${DB_PASSWORD:-}" ] || [ -z "${JWT_SECRET:-}" ]; then
    echo "Ошибка: DB_PASSWORD и JWT_SECRET обязательны для $ENV."
    exit 1
  fi
  if [ -z "${LIBRARY_PATH:-}" ]; then
    echo "Ошибка: LIBRARY_PATH обязателен."
    exit 1
  fi
fi

echo "1. Сборка образов..."
docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" --env-file "$PROJECT_ROOT/.env" build

echo "2. Остановка текущих сервисов..."
docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" --env-file "$PROJECT_ROOT/.env" down

echo "3. Запуск сервисов..."
docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" --env-file "$PROJECT_ROOT/.env" up -d

echo "4. Ожидание готовности..."
sleep 5
docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" --env-file "$PROJECT_ROOT/.env" ps

echo "=== Деплой завершён ==="
