#!/usr/bin/env bash
# Сборка Go-бинарников HomeLib
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=scripts/lib/logging.sh
. "$SCRIPT_DIR/lib/logging.sh"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKEND_DIR="$PROJECT_ROOT/backend"
OUTPUT_DIR="${1:-$PROJECT_ROOT/build}"

mkdir -p "$OUTPUT_DIR"

log_info "Сборка API-сервера..."
cd "$BACKEND_DIR"
CGO_ENABLED=0 GOOS=linux go build -o "$OUTPUT_DIR/api" ./cmd/api

log_info "Сборка воркера..."
CGO_ENABLED=0 GOOS=linux go build -o "$OUTPUT_DIR/worker" ./cmd/worker

log_success "Сборка завершена"
ls -lh "$OUTPUT_DIR"/api "$OUTPUT_DIR"/worker
