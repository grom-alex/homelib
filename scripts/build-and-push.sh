#!/bin/bash
# Build and push Docker images to registry for HomeLib
# Builds 3 images: api, worker, frontend
#
# Usage: ./scripts/build-and-push.sh [version]
#
# Examples:
#   ./scripts/build-and-push.sh              # Use git SHA
#   ./scripts/build-and-push.sh v1.0.0       # Production release

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Load shared logging library
# shellcheck source=scripts/lib/logging.sh
. "$SCRIPT_DIR/lib/logging.sh"

# Registry configuration (configurable via environment variables)
REGISTRY="${DOCKER_REGISTRY:-registry.gromas.ru}"
IMAGE_PREFIX="${IMAGE_PREFIX:-apps/homelib}"

# Help
usage() {
    cat <<EOF
Usage: $0 [OPTIONS] [version]

Build, test, and push Docker images for HomeLib to registry.
Builds 3 images: api, worker, frontend.

ARGUMENTS:
    version         Version tag (default: sha-<git-hash>)

OPTIONS:
    --help, -h      Show this help message

TAGGING STRATEGY:
    - Always tagged with: <version>, sha-<hash>
    - Semver releases (vX.Y.Z) also tagged as: latest

EXAMPLES:
    $0                  # Build with git SHA tag
    $0 v1.0.0           # Production release with semver + latest

PREREQUISITES:
    - Docker (required)
    - Go 1.25+ (for backend tests)
    - Node.js 22+ (for frontend tests)
    - Access to Docker registry

EXIT CODES:
    0 - Success
    1 - Tests failed or build error
EOF
}

# Parse arguments
case "${1:-}" in
    --help|-h)
        usage
        exit 0
        ;;
esac

# Extract short SHA (used for all tagging)
SHORT_SHA=$(git rev-parse --short HEAD)

# Get version
VERSION=${1:-}
if [ -z "$VERSION" ]; then
    VERSION="sha-${SHORT_SHA}"
    log_info "No version specified, using: ${VERSION}"
else
    log_info "Building version: ${VERSION}"
fi

cd "${PROJECT_ROOT}"

# Step 1: Run tests
log_section "Step 1: Running tests"

log_info "Running backend tests..."
cd "${PROJECT_ROOT}/backend"
if go test ./...; then
    log_info "Backend tests passed"
else
    log_error "Backend tests failed!"
    exit 1
fi

cd "${PROJECT_ROOT}/frontend"
log_info "Running frontend tests..."
if npm run test -- --run 2>/dev/null || npx vitest run 2>/dev/null; then
    log_info "Frontend tests passed"
else
    log_error "Frontend tests failed!"
    exit 1
fi

cd "${PROJECT_ROOT}"

# Step 2: Build Docker images (api, worker, frontend)
log_section "Step 2: Building Docker images"

# API
log_info "Building API image..."
API_TAGS=(
    "-t" "${REGISTRY}/${IMAGE_PREFIX}/api:${VERSION}"
)
# Add SHA tag only if version is not already sha-based
if [ "$VERSION" != "sha-${SHORT_SHA}" ]; then
    API_TAGS+=("-t" "${REGISTRY}/${IMAGE_PREFIX}/api:sha-${SHORT_SHA}")
fi
if [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    API_TAGS+=("-t" "${REGISTRY}/${IMAGE_PREFIX}/api:latest")
    log_info "Production release detected, adding 'latest' tag"
fi

if ! docker build -f docker/backend/Dockerfile.api "${API_TAGS[@]}" backend/; then
    log_error "API image build failed!"
    exit 1
fi
log_info "API image built successfully"

# Worker
log_info "Building Worker image..."
WORKER_TAGS=(
    "-t" "${REGISTRY}/${IMAGE_PREFIX}/worker:${VERSION}"
)
if [ "$VERSION" != "sha-${SHORT_SHA}" ]; then
    WORKER_TAGS+=("-t" "${REGISTRY}/${IMAGE_PREFIX}/worker:sha-${SHORT_SHA}")
fi
if [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    WORKER_TAGS+=("-t" "${REGISTRY}/${IMAGE_PREFIX}/worker:latest")
fi

if ! docker build -f docker/backend/Dockerfile.worker "${WORKER_TAGS[@]}" backend/; then
    log_error "Worker image build failed!"
    exit 1
fi
log_info "Worker image built successfully"

# Frontend
log_info "Building Frontend image..."
FRONTEND_TAGS=(
    "-t" "${REGISTRY}/${IMAGE_PREFIX}/frontend:${VERSION}"
)
if [ "$VERSION" != "sha-${SHORT_SHA}" ]; then
    FRONTEND_TAGS+=("-t" "${REGISTRY}/${IMAGE_PREFIX}/frontend:sha-${SHORT_SHA}")
fi
if [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    FRONTEND_TAGS+=("-t" "${REGISTRY}/${IMAGE_PREFIX}/frontend:latest")
fi

if ! docker build -f docker/frontend/Dockerfile "${FRONTEND_TAGS[@]}" frontend/; then
    log_error "Frontend image build failed!"
    exit 1
fi
log_info "Frontend image built successfully"

# Step 3: Push to registry
log_section "Step 3: Pushing images to registry"

for component in api worker frontend; do
    log_info "Pushing ${component} images..."
    docker push "${REGISTRY}/${IMAGE_PREFIX}/${component}:${VERSION}"

    if [ "$VERSION" != "sha-${SHORT_SHA}" ]; then
        docker push "${REGISTRY}/${IMAGE_PREFIX}/${component}:sha-${SHORT_SHA}"
    fi

    if [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        docker push "${REGISTRY}/${IMAGE_PREFIX}/${component}:latest"
    fi

    log_info "${component} images pushed"
done

# Summary
log_section "Build Summary"

echo "Images built and pushed:"
echo ""
for component in api worker frontend; do
    echo "${component}:"
    echo "  - ${REGISTRY}/${IMAGE_PREFIX}/${component}:${VERSION}"
    if [ "$VERSION" != "sha-${SHORT_SHA}" ]; then
        echo "  - ${REGISTRY}/${IMAGE_PREFIX}/${component}:sha-${SHORT_SHA}"
    fi
    if [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "  - ${REGISTRY}/${IMAGE_PREFIX}/${component}:latest"
    fi
    echo ""
done

echo "Next steps:"
echo "  Deploy to staging: ./scripts/deploy-stage.sh --tag ${VERSION}"
echo "  Deploy to production: ./scripts/deploy-prod.sh --tag ${VERSION} --host <PROD_HOST>"
echo ""

exit 0
