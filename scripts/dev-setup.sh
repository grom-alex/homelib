#!/usr/bin/env bash
# Настройка окружения разработчика HomeLib
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=== Настройка окружения разработчика HomeLib ==="

# 1. Проверка зависимостей
echo "1. Проверка зависимостей..."
for cmd in go node npm docker; do
  if command -v "$cmd" &>/dev/null; then
    echo "   $cmd: $(command -v "$cmd")"
  else
    echo "   ОШИБКА: $cmd не найден. Установите перед продолжением."
    exit 1
  fi
done

# 2. Создание .env
if [ ! -f "$PROJECT_ROOT/.env" ]; then
  echo "2. Создание .env из .env.example..."
  cp "$PROJECT_ROOT/.env.example" "$PROJECT_ROOT/.env"
  echo "   Отредактируйте .env при необходимости."
else
  echo "2. .env уже существует, пропускаем."
fi

# 3. Установка Go-зависимостей
echo "3. Установка Go-зависимостей..."
cd "$PROJECT_ROOT/backend"
go mod download

# 4. Установка Node-зависимостей
echo "4. Установка Node-зависимостей..."
cd "$PROJECT_ROOT/frontend"
npm ci

# 5. Запуск Docker Compose
echo "5. Запуск сервисов (docker compose)..."
docker compose -f "$PROJECT_ROOT/docker/docker-compose.dev.yml" up -d

echo ""
echo "=== Готово! ==="
echo "HomeLib доступен на http://localhost"
echo "Для остановки: make stop"
