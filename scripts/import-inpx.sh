#!/usr/bin/env bash
# Запуск импорта INPX через воркер
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DOCKER_DIR="$(dirname "$SCRIPT_DIR")/docker"
COMPOSE_FILE="${1:-docker-compose.dev.yml}"

echo "=== Запуск импорта INPX ==="
docker compose -f "$DOCKER_DIR/$COMPOSE_FILE" run --rm worker \
  worker -config /etc/homelib/config.yaml -import
