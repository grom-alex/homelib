#!/bin/bash
# Build and push Docker images to registry for HomeLib
# Builds 3 images: api, worker, frontend
#
# Usage: ./scripts/build-and-push.sh [version]
#
# Examples:
#   ./scripts/build-and-push.sh              # Use git SHA
#   ./scripts/build-and-push.sh v1.0.0       # Production release

set -e  # Exit on error
set -u  # Exit on undefined variable

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Registry configuration (configurable via environment variables)
REGISTRY="${DOCKER_REGISTRY:-registry.gromas.ru}"
IMAGE_PREFIX="${IMAGE_PREFIX:-apps/homelib}"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo ""
    echo -e "${BLUE}==>${NC} $1"
    echo ""
}

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

# Get version
VERSION=${1:-}
if [ -z "$VERSION" ]; then
    SHORT_SHA=$(git rev-parse --short HEAD)
    VERSION="sha-${SHORT_SHA}"
    log_info "No version specified, using: ${VERSION}"
else
    log_info "Building version: ${VERSION}"
fi

# Extract short SHA for tagging
SHORT_SHA=$(git rev-parse --short HEAD)

cd "${PROJECT_ROOT}"

# Step 1: Run tests
log_step "Step 1: Running tests"

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
log_step "Step 2: Building Docker images"

# API
log_info "Building API image..."
API_TAGS=(
    "-t" "${REGISTRY}/${IMAGE_PREFIX}/api:${VERSION}"
    "-t" "${REGISTRY}/${IMAGE_PREFIX}/api:sha-${SHORT_SHA}"
)
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
    "-t" "${REGISTRY}/${IMAGE_PREFIX}/worker:sha-${SHORT_SHA}"
)
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
    "-t" "${REGISTRY}/${IMAGE_PREFIX}/frontend:sha-${SHORT_SHA}"
)
if [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    FRONTEND_TAGS+=("-t" "${REGISTRY}/${IMAGE_PREFIX}/frontend:latest")
fi

if ! docker build -f docker/frontend/Dockerfile "${FRONTEND_TAGS[@]}" frontend/; then
    log_error "Frontend image build failed!"
    exit 1
fi
log_info "Frontend image built successfully"

# Step 3: Push to registry
log_step "Step 3: Pushing images to registry"

for component in api worker frontend; do
    log_info "Pushing ${component} images..."
    docker push "${REGISTRY}/${IMAGE_PREFIX}/${component}:${VERSION}"
    docker push "${REGISTRY}/${IMAGE_PREFIX}/${component}:sha-${SHORT_SHA}"

    if [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        docker push "${REGISTRY}/${IMAGE_PREFIX}/${component}:latest"
    fi

    log_info "${component} images pushed"
done

# Summary
log_step "Build Summary"

echo "Images built and pushed:"
echo ""
for component in api worker frontend; do
    echo "${component}:"
    echo "  - ${REGISTRY}/${IMAGE_PREFIX}/${component}:${VERSION}"
    echo "  - ${REGISTRY}/${IMAGE_PREFIX}/${component}:sha-${SHORT_SHA}"
    if [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "  - ${REGISTRY}/${IMAGE_PREFIX}/${component}:latest"
    fi
    echo ""
done

echo "Next steps:"
echo "  Deploy to staging: IMAGE_TAG=${VERSION} ./scripts/deploy.sh staging"
echo "  Deploy to production: IMAGE_TAG=${VERSION} ./scripts/deploy.sh production"
echo ""

exit 0
