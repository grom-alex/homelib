#!/usr/bin/env bash
# Запуск импорта INPX через воркер
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=scripts/lib/logging.sh
. "$SCRIPT_DIR/lib/logging.sh"
DOCKER_DIR="$(dirname "$SCRIPT_DIR")/docker"
COMPOSE_FILE="${1:-docker-compose.dev.yml}"

log_info "Запуск импорта INPX..."
docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" run --rm worker \
  worker -config /etc/homelib/config.yaml -import
