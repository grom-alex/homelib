#!/bin/sh
# Local Build Execution Script for HomeLib
# Builds Docker images for api, worker, and frontend components
# Reference: FR-009, FR-013, FR-015

set -e  # Exit on error

# ============================================================================
# Script Directory and Library Loading
# ============================================================================
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=scripts/lib/logging.sh
. "$SCRIPT_DIR/lib/logging.sh"
# shellcheck source=scripts/lib/prerequisites.sh
. "$SCRIPT_DIR/lib/prerequisites.sh"
# shellcheck source=scripts/lib/docker-utils.sh
. "$SCRIPT_DIR/lib/docker-utils.sh"

# ============================================================================
# Exit Codes (T065)
# ============================================================================
EXIT_SUCCESS=0
EXIT_PREREQUISITE_FAILURE=1
EXIT_EXECUTION_ERROR=2

# ============================================================================
# Default Configuration
# ============================================================================
COMPONENTS="all"  # all, backend (api+worker), frontend, api, worker
TAG="latest"
NO_CACHE=false
PUSH=false
BUILD_TYPE="debug"  # debug, release

# ============================================================================
# Usage Information (FR-024, SC-009)
# ============================================================================
usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Local build execution script - builds Docker images for HomeLib components

OPTIONS:
    --component COMP    Build specific component (backend, frontend, api, worker, all)
                        'backend' builds both api and worker images
    --tag TAG           Docker image tag (default: latest)
    --no-cache          Build without using cache
    --push              Push images to registry after build
    --release           Build release version (optimized, no debug symbols)
    --debug             Build debug version (default)
    --help, -h          Show this help message

EXAMPLES:
    $0                                       # Build all 3 components
    $0 --component backend --tag sha-abc123  # Build api + worker
    $0 --component frontend                  # Build frontend only
    $0 --release --push --tag v1.0.0         # Build release and push

PREREQUISITES:
    - Docker (required)
    - Go 1.25+ (for backend builds)
    - Node.js 22+ (for frontend builds)

EXIT CODES:
    0 - Success
    1 - Prerequisite check failed
    2 - Build execution error
EOF
}

# ============================================================================
# Argument Parsing
# ============================================================================
parse_arguments() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --component)
                COMPONENTS="$2"
                shift 2
                ;;
            --tag)
                TAG="$2"
                shift 2
                ;;
            --no-cache)
                NO_CACHE=true
                shift
                ;;
            --push)
                PUSH=true
                shift
                ;;
            --release)
                BUILD_TYPE="release"
                shift
                ;;
            --debug)
                BUILD_TYPE="debug"
                shift
                ;;
            --help|-h)
                usage
                exit $EXIT_SUCCESS
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit $EXIT_PREREQUISITE_FAILURE
                ;;
        esac
    done
}

# ============================================================================
# Prerequisite Validation (T047, FR-023)
# ============================================================================
check_prerequisites() {
    log_section "Checking Prerequisites"

    # Docker is required for all builds
    check_docker

    # Check component-specific prerequisites
    case "$COMPONENTS" in
        all|backend|api|worker)
            check_go "1.25"
            ;;
    esac

    case "$COMPONENTS" in
        all|frontend)
            check_node "22"
            ;;
    esac

    # Check Docker registry credentials if pushing
    if [ "$PUSH" = "true" ]; then
        check_env_var "DOCKER_REGISTRY_PASSWORD" "Set DOCKER_REGISTRY_PASSWORD to push images"
    fi

    prereq_summary
}

# ============================================================================
# API Build
# ============================================================================
build_api() {
    log_section "Building API Docker Image"

    cd "$SCRIPT_DIR/.." || exit $EXIT_EXECUTION_ERROR

    BUILD_ARGS=""
    if [ "$NO_CACHE" = "true" ]; then
        BUILD_ARGS="$BUILD_ARGS --no-cache"
    fi
    if [ "$BUILD_TYPE" = "release" ]; then
        BUILD_ARGS="$BUILD_ARGS --build-arg BUILD_TYPE=release"
    fi

    # shellcheck disable=SC2086
    if build_image "api" "$TAG" $BUILD_ARGS; then
        log_success "API image built: $DOCKER_REGISTRY/$IMAGE_PREFIX/api:$TAG"

        if [ "$PUSH" = "true" ]; then
            if push_image "api" "$TAG"; then
                log_success "API image pushed to registry"
            else
                log_error "Failed to push API image"
                exit $EXIT_EXECUTION_ERROR
            fi
        fi
    else
        log_error "API build failed"
        exit $EXIT_EXECUTION_ERROR
    fi

    cd "$SCRIPT_DIR" || exit $EXIT_EXECUTION_ERROR
}

# ============================================================================
# Worker Build
# ============================================================================
build_worker() {
    log_section "Building Worker Docker Image"

    cd "$SCRIPT_DIR/.." || exit $EXIT_EXECUTION_ERROR

    BUILD_ARGS=""
    if [ "$NO_CACHE" = "true" ]; then
        BUILD_ARGS="$BUILD_ARGS --no-cache"
    fi
    if [ "$BUILD_TYPE" = "release" ]; then
        BUILD_ARGS="$BUILD_ARGS --build-arg BUILD_TYPE=release"
    fi

    # shellcheck disable=SC2086
    if build_image "worker" "$TAG" $BUILD_ARGS; then
        log_success "Worker image built: $DOCKER_REGISTRY/$IMAGE_PREFIX/worker:$TAG"

        if [ "$PUSH" = "true" ]; then
            if push_image "worker" "$TAG"; then
                log_success "Worker image pushed to registry"
            else
                log_error "Failed to push worker image"
                exit $EXIT_EXECUTION_ERROR
            fi
        fi
    else
        log_error "Worker build failed"
        exit $EXIT_EXECUTION_ERROR
    fi

    cd "$SCRIPT_DIR" || exit $EXIT_EXECUTION_ERROR
}

# ============================================================================
# Frontend Build
# ============================================================================
build_frontend() {
    log_section "Building Frontend Docker Image"

    cd "$SCRIPT_DIR/.." || exit $EXIT_EXECUTION_ERROR

    # Build arguments (docker-utils.sh handles --file)
    BUILD_ARGS=""

    if [ "$NO_CACHE" = "true" ]; then
        BUILD_ARGS="$BUILD_ARGS --no-cache"
    fi

    if [ "$BUILD_TYPE" = "release" ]; then
        BUILD_ARGS="$BUILD_ARGS --build-arg NODE_ENV=production"
    fi

    # Use docker-utils.sh function (T063)
    if build_image "frontend" "$TAG" $BUILD_ARGS; then
        log_success "Frontend image built: $DOCKER_REGISTRY/$IMAGE_PREFIX/frontend:$TAG"

        if [ "$PUSH" = "true" ]; then
            if push_image "frontend" "$TAG"; then
                log_success "Frontend image pushed to registry"
            else
                log_error "Failed to push frontend image"
                exit $EXIT_EXECUTION_ERROR
            fi
        fi
    else
        log_error "Frontend build failed"
        exit $EXIT_EXECUTION_ERROR
    fi

    cd "$SCRIPT_DIR" || exit $EXIT_EXECUTION_ERROR
}

# ============================================================================
# Binary Build (without Docker)
# ============================================================================
build_backend_binary() {
    log_section "Building Backend Binaries (without Docker)"

    cd "$SCRIPT_DIR/../backend" || {
        log_error "Backend directory not found"
        exit $EXIT_EXECUTION_ERROR
    }

    mkdir -p bin

    for target in api worker; do
        if [ "$BUILD_TYPE" = "release" ]; then
            log_info "Building release binary: $target..."
            CGO_ENABLED=0 go build -ldflags="-w -s" -o "bin/$target" "./cmd/$target"
        else
            log_info "Building debug binary: $target..."
            go build -o "bin/$target" "./cmd/$target"
        fi

        if [ -f "bin/$target" ]; then
            log_success "Binary built: backend/bin/$target"
            ls -lh "bin/$target"
        else
            log_error "Binary not found after build: $target"
            exit $EXIT_EXECUTION_ERROR
        fi
    done

    cd "$SCRIPT_DIR" || exit $EXIT_EXECUTION_ERROR
}

build_frontend_dist() {
    log_section "Building Frontend Distribution (without Docker)"

    cd "$SCRIPT_DIR/../frontend" || {
        log_error "Frontend directory not found"
        exit $EXIT_EXECUTION_ERROR
    }

    log_info "Installing dependencies..."
    npm ci

    if [ "$BUILD_TYPE" = "release" ]; then
        log_info "Building production bundle..."
        NODE_ENV=production npm run build
    else
        log_info "Building development bundle..."
        npm run build -- --mode development
    fi

    if [ -d dist ]; then
        log_success "Frontend built: frontend/dist/"
        du -sh dist
    else
        log_error "Frontend dist not found after build"
        exit $EXIT_EXECUTION_ERROR
    fi

    cd "$SCRIPT_DIR" || exit $EXIT_EXECUTION_ERROR
}

# ============================================================================
# Main Execution
# ============================================================================
main() {
    # Parse arguments first (for --help)
    parse_arguments "$@"

    log_info "Starting local build execution"
    log_info "Components: $COMPONENTS"
    log_info "Tag: $TAG"
    log_info "Build Type: $BUILD_TYPE"

    # Check prerequisites
    check_prerequisites

    # Determine build method
    if docker info >/dev/null 2>&1; then
        BUILD_METHOD="docker"
    else
        BUILD_METHOD="binary"
        log_warn "Docker not available - building binaries only"
    fi

    # Build based on component selection
    case "$COMPONENTS" in
        api)
            if [ "$BUILD_METHOD" = "docker" ]; then
                build_api
            else
                build_backend_binary
            fi
            ;;
        worker)
            if [ "$BUILD_METHOD" = "docker" ]; then
                build_worker
            else
                build_backend_binary
            fi
            ;;
        backend)
            # backend builds both api and worker
            if [ "$BUILD_METHOD" = "docker" ]; then
                build_api
                build_worker
            else
                build_backend_binary
            fi
            ;;
        frontend)
            if [ "$BUILD_METHOD" = "docker" ]; then
                build_frontend
            else
                build_frontend_dist
            fi
            ;;
        all)
            if [ "$BUILD_METHOD" = "docker" ]; then
                build_api
                build_worker
                build_frontend
            else
                build_backend_binary
                build_frontend_dist
            fi
            ;;
        *)
            log_error "Unknown component: $COMPONENTS"
            log_error "  → Fix: Use --component with backend, frontend, api, worker, or all"
            exit $EXIT_PREREQUISITE_FAILURE
            ;;
    esac

    log_success "All builds completed successfully!"
    exit $EXIT_SUCCESS
}

# Run main function
main "$@"
