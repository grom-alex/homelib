#!/bin/bash
# Build and push Docker images to registry for HomeLib
# Builds 3 images: api, worker, frontend
#
# Usage: ./scripts/build-and-push.sh [OPTIONS]
#
# Image tag format: v<VERSION>-sha-<COMMIT>
# Version is read from ./version file in project root
#
# Examples:
#   ./scripts/build-and-push.sh                  # v0.1.0-sha-abc1234
#   ./scripts/build-and-push.sh --bump patch     # Bump 0.1.0 → 0.1.1, build
#   ./scripts/build-and-push.sh --bump minor     # Bump 0.1.0 → 0.2.0, build
#   ./scripts/build-and-push.sh --bump major     # Bump 0.1.0 → 1.0.0, build
#   ./scripts/build-and-push.sh --skip-tests     # Skip tests

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
VERSION_FILE="${PROJECT_ROOT}/version"

# Defaults
BUMP=""
SKIP_TESTS=false

# Help
usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Build, test, and push Docker images for HomeLib to registry.
Builds 3 images: api, worker, frontend.

IMAGE TAG FORMAT:
    v<VERSION>-sha-<COMMIT>     (e.g., v0.1.0-sha-abc1234)

    Version is read from ./version file in project root.
    Use --bump to increment before building.

OPTIONS:
    --bump <part>   Bump version before build (patch|minor|major)
    --skip-tests    Skip running tests before build
    --help, -h      Show this help message

VERSION BUMP RULES:
    patch   Bug fixes, minor improvements, dependency updates
            (0.1.0 → 0.1.1)
    minor   New features, new API endpoints, non-breaking changes
            (0.1.0 → 0.2.0)
    major   Breaking API changes, incompatible DB migrations, architecture changes
            (0.1.0 → 1.0.0)

EXAMPLES:
    $0                      # Build with current version
    $0 --bump patch         # Bug fix release
    $0 --bump minor         # Feature release
    $0 --skip-tests         # Quick rebuild without tests

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
while [[ $# -gt 0 ]]; do
    case $1 in
        --bump)
            BUMP="$2"
            if [[ ! "$BUMP" =~ ^(patch|minor|major)$ ]]; then
                log_error "Invalid bump type: $BUMP (use: patch, minor, major)"
                exit 1
            fi
            shift 2
            ;;
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Read version from file
if [ ! -f "$VERSION_FILE" ]; then
    log_error "Version file not found: $VERSION_FILE"
    log_error "Create it with: echo '0.1.0' > version"
    exit 1
fi

CURRENT_VERSION=$(tr -d '[:space:]' < "$VERSION_FILE")

if [[ ! "$CURRENT_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    log_error "Invalid version format in $VERSION_FILE: '$CURRENT_VERSION'"
    log_error "Expected: MAJOR.MINOR.PATCH (e.g., 0.1.0)"
    exit 1
fi

# Bump version if requested
if [ -n "$BUMP" ]; then
    IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

    case "$BUMP" in
        major)
            MAJOR=$((MAJOR + 1))
            MINOR=0
            PATCH=0
            ;;
        minor)
            MINOR=$((MINOR + 1))
            PATCH=0
            ;;
        patch)
            PATCH=$((PATCH + 1))
            ;;
    esac

    NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
    log_info "Version bump: ${CURRENT_VERSION} → ${NEW_VERSION} (${BUMP})"
    echo "$NEW_VERSION" > "$VERSION_FILE"
    CURRENT_VERSION="$NEW_VERSION"
fi

# Build image tag: v<VERSION>-sha-<COMMIT>
SHORT_SHA=$(git rev-parse --short HEAD)
IMAGE_TAG="v${CURRENT_VERSION}-sha-${SHORT_SHA}"

log_info "Version: ${CURRENT_VERSION}"
log_info "Image tag: ${IMAGE_TAG}"

cd "${PROJECT_ROOT}"

# Step 1: Run tests
if [ "$SKIP_TESTS" = true ]; then
    log_warn "Skipping tests (--skip-tests)"
else
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
    if npm run test -- --run || npx vitest run; then
        log_info "Frontend tests passed"
    else
        log_error "Frontend tests failed!"
        exit 1
    fi

    cd "${PROJECT_ROOT}"
fi

# Step 2: Build Docker images (api, worker, frontend)
log_section "Step 2: Building Docker images"

for component in api worker frontend; do
    log_info "Building ${component} image..."

    # Determine Dockerfile and context
    case "$component" in
        api)
            DOCKERFILE="docker/backend/Dockerfile.api"
            CONTEXT="backend/"
            ;;
        worker)
            DOCKERFILE="docker/backend/Dockerfile.worker"
            CONTEXT="backend/"
            ;;
        frontend)
            DOCKERFILE="docker/frontend/Dockerfile"
            CONTEXT="frontend/"
            ;;
    esac

    TAGS=(
        "-t" "${REGISTRY}/${IMAGE_PREFIX}/${component}:${IMAGE_TAG}"
        "-t" "${REGISTRY}/${IMAGE_PREFIX}/${component}:latest"
    )

    if ! docker build -f "$DOCKERFILE" "${TAGS[@]}" "$CONTEXT"; then
        log_error "${component} image build failed!"
        exit 1
    fi
    log_info "${component} image built successfully"
done

# Step 3: Push to registry
log_section "Step 3: Pushing images to registry"

for component in api worker frontend; do
    log_info "Pushing ${component} images..."
    docker push "${REGISTRY}/${IMAGE_PREFIX}/${component}:${IMAGE_TAG}"
    docker push "${REGISTRY}/${IMAGE_PREFIX}/${component}:latest"
    log_info "${component} images pushed"
done

# Summary
log_section "Build Summary"

echo "Version: ${CURRENT_VERSION}"
echo "Tag:     ${IMAGE_TAG}"
echo ""
echo "Images built and pushed:"
for component in api worker frontend; do
    echo "  ${REGISTRY}/${IMAGE_PREFIX}/${component}:${IMAGE_TAG}"
done
echo ""
echo "Next steps:"
echo "  Deploy to staging:    ./scripts/deploy-stage.sh --tag ${IMAGE_TAG}"
echo "  Deploy to production: ./scripts/deploy-prod.sh --tag ${IMAGE_TAG} --host <PROD_HOST>"
echo ""

exit 0
