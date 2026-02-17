#!/usr/bin/env bash
# Настройка окружения разработчика HomeLib
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=scripts/lib/logging.sh
. "$SCRIPT_DIR/lib/logging.sh"
# shellcheck source=scripts/lib/prerequisites.sh
. "$SCRIPT_DIR/lib/prerequisites.sh"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

log_section "Настройка окружения разработчика HomeLib"

# 1. Проверка зависимостей
log_info "1. Проверка зависимостей..."
check_docker
check_docker_compose
check_go "1.25"
check_node "22"
prereq_summary

# 2. Создание .env
if [ ! -f "$PROJECT_ROOT/.env" ]; then
  log_info "2. Создание .env из .env.example..."
  cp "$PROJECT_ROOT/.env.example" "$PROJECT_ROOT/.env"
  log_warn "Отредактируйте .env при необходимости."
else
  log_info "2. .env уже существует, пропускаем."
fi

# 3. Установка Go-зависимостей
log_info "3. Установка Go-зависимостей..."
cd "$PROJECT_ROOT/backend"
go mod download

# 4. Установка Node-зависимостей
log_info "4. Установка Node-зависимостей..."
cd "$PROJECT_ROOT/frontend"
npm ci

# 5. Запуск Docker Compose
log_info "5. Запуск сервисов (docker compose)..."
docker compose -f "$PROJECT_ROOT/docker/docker-compose.dev.yml" up -d

log_success "Готово! HomeLib доступен на http://localhost"
log_info "Для остановки: make stop"
