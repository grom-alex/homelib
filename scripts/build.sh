#!/usr/bin/env bash
# Сборка Go-бинарников HomeLib
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKEND_DIR="$PROJECT_ROOT/backend"
OUTPUT_DIR="${1:-$PROJECT_ROOT/build}"

mkdir -p "$OUTPUT_DIR"

echo "=== Сборка API-сервера ==="
cd "$BACKEND_DIR"
CGO_ENABLED=0 GOOS=linux go build -o "$OUTPUT_DIR/api" ./cmd/api

echo "=== Сборка воркера ==="
CGO_ENABLED=0 GOOS=linux go build -o "$OUTPUT_DIR/worker" ./cmd/worker

echo "=== Готово ==="
ls -lh "$OUTPUT_DIR"/api "$OUTPUT_DIR"/worker
