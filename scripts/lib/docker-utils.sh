#!/bin/sh
# Shared Docker utilities library for CI/CD scripts
# POSIX-compliant shell script (T007)
# Reference: FR-015, FR-016

# Source logging library if not already loaded
if ! command -v log_info >/dev/null 2>&1; then
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
    # shellcheck source=scripts/lib/logging.sh
    . "$SCRIPT_DIR/logging.sh"
fi

# Docker registry defaults (can be overridden by environment variables)
DOCKER_REGISTRY="${DOCKER_REGISTRY:-registry.gromas.ru}"
IMAGE_PREFIX="${IMAGE_PREFIX:-apps/homelib}"

# build_image() - Build a Docker image with multi-stage build
# Usage: build_image "backend" "sha-abc1234" [--no-cache]
build_image() {
    component="$1"
    tag="$2"
    shift 2
    extra_args="$*"

    log_info "Building Docker image for $component..."

    # Determine Dockerfile location
    case "$component" in
        api)
            context_dir="backend"
            dockerfile="docker/backend/Dockerfile.api"
            ;;
        worker)
            context_dir="backend"
            dockerfile="docker/backend/Dockerfile.worker"
            ;;
        frontend)
            context_dir="frontend"
            dockerfile="docker/frontend/Dockerfile"
            ;;
        *)
            log_error "Unknown component: $component"
            return 1
            ;;
    esac

    # Validate Dockerfile exists
    if [ ! -f "$dockerfile" ]; then
        log_error "Dockerfile not found: $dockerfile"
        return 1
    fi

    # Build full image name
    image_name="$DOCKER_REGISTRY/$IMAGE_PREFIX/$component:$tag"

    log_info "Image: $image_name"
    log_info "Context: $context_dir"

    # shellcheck disable=SC2086  # Extra args intentionally unquoted
    if docker build \
        -t "$image_name" \
        -f "$dockerfile" \
        $extra_args \
        "$context_dir"; then
        log_success "Image built successfully: $image_name"

        # Also tag as 'latest' for local testing
        docker tag "$image_name" "$DOCKER_REGISTRY/$IMAGE_PREFIX/$component:latest"
        return 0
    else
        log_error "Failed to build image: $component"
        return 1
    fi
}

# push_image() - Push Docker image to registry
# Usage: push_image "backend" "sha-abc1234"
push_image() {
    component="$1"
    tag="$2"

    image_name="$DOCKER_REGISTRY/$IMAGE_PREFIX/$component:$tag"

    log_info "Pushing Docker image: $image_name"

    if docker push "$image_name"; then
        log_success "Image pushed successfully: $image_name"
        return 0
    else
        log_error "Failed to push image: $image_name"
        return 1
    fi
}

# pull_image() - Pull Docker image from registry
# Usage: pull_image "backend" "sha-abc1234"
pull_image() {
    component="$1"
    tag="$2"

    image_name="$DOCKER_REGISTRY/$IMAGE_PREFIX/$component:$tag"

    log_info "Pulling Docker image: $image_name"

    if docker pull "$image_name"; then
        log_success "Image pulled successfully: $image_name"
        return 0
    else
        log_error "Failed to pull image: $image_name"
        return 1
    fi
}

# run_container() - Run a Docker container
# Usage: run_container "backend" "sha-abc1234" [extra docker run args...]
run_container() {
    component="$1"
    tag="$2"
    shift 2
    extra_args="$*"

    image_name="$DOCKER_REGISTRY/$IMAGE_PREFIX/$component:$tag"
    container_name="homelib_${component}_${tag}"

    log_info "Running container: $container_name"

    # shellcheck disable=SC2086  # Extra args intentionally unquoted
    if docker run -d --name "$container_name" $extra_args "$image_name"; then
        log_success "Container started: $container_name"
        return 0
    else
        log_error "Failed to start container: $container_name"
        return 1
    fi
}

# stop_container() - Stop a running container
# Usage: stop_container "container_name"
stop_container() {
    container_name="$1"

    log_info "Stopping container: $container_name"

    if docker stop "$container_name" >/dev/null 2>&1; then
        log_success "Container stopped: $container_name"
        return 0
    else
        log_warn "Failed to stop container (may not be running): $container_name"
        return 1
    fi
}

# remove_container() - Remove a container
# Usage: remove_container "container_name"
remove_container() {
    container_name="$1"

    log_info "Removing container: $container_name"

    if docker rm -f "$container_name" >/dev/null 2>&1; then
        log_success "Container removed: $container_name"
        return 0
    else
        log_warn "Failed to remove container (may not exist): $container_name"
        return 1
    fi
}

# wait_for_health() - Wait for container health check to pass
# Usage: wait_for_health "container_name" [timeout_seconds]
wait_for_health() {
    container_name="$1"
    timeout="${2:-60}"

    log_info "Waiting for container health check: $container_name (timeout: ${timeout}s)"

    elapsed=0
    interval=2

    while [ $elapsed -lt "$timeout" ]; do
        # Get container health status
        health_status=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null)

        case "$health_status" in
            healthy)
                log_success "Container is healthy: $container_name"
                return 0
                ;;
            unhealthy)
                log_error "Container is unhealthy: $container_name"
                docker logs --tail=20 "$container_name"
                return 1
                ;;
            starting)
                log_debug "Container health check is starting... ($elapsed/${timeout}s)"
                ;;
            *)
                log_debug "Container has no health check or is not running"
                # If no health check, just verify container is running
                if docker ps --filter "name=$container_name" --format '{{.Names}}' | grep -q "$container_name"; then
                    log_success "Container is running: $container_name"
                    return 0
                fi
                ;;
        esac

        sleep $interval
        elapsed=$((elapsed + interval))
    done

    log_error "Timeout waiting for container health: $container_name"
    docker logs --tail=20 "$container_name"
    return 1
}

# wait_for_url() - Wait for HTTP endpoint to return 200 OK
# Usage: wait_for_url "http://localhost:8080/health" [timeout_seconds]
wait_for_url() {
    url="$1"
    timeout="${2:-60}"

    log_info "Waiting for URL to respond: $url (timeout: ${timeout}s)"

    elapsed=0
    interval=2

    while [ $elapsed -lt "$timeout" ]; do
        # Try curl first, fall back to wget
        if command -v curl >/dev/null 2>&1; then
            if curl -sf "$url" >/dev/null 2>&1; then
                log_success "URL is responding: $url"
                return 0
            fi
        elif command -v wget >/dev/null 2>&1; then
            if wget -q -O- "$url" >/dev/null 2>&1; then
                log_success "URL is responding: $url"
                return 0
            fi
        else
            log_error "Neither curl nor wget is available"
            return 1
        fi

        log_debug "Waiting for URL... ($elapsed/${timeout}s)"
        sleep $interval
        elapsed=$((elapsed + interval))
    done

    log_error "Timeout waiting for URL: $url"
    return 1
}

# docker_login() - Login to Docker registry
# Usage: docker_login [registry] [username] [password]
docker_login() {
    registry="${1:-$DOCKER_REGISTRY}"
    username="${2:-$DOCKER_REGISTRY_USERNAME}"
    password="${3:-$DOCKER_REGISTRY_PASSWORD}"

    if [ -z "$password" ]; then
        log_error "Docker registry password not provided"
        log_error "  → Set DOCKER_REGISTRY_PASSWORD environment variable"
        return 1
    fi

    log_info "Logging in to Docker registry: $registry"

    if echo "$password" | docker login "$registry" -u "$username" --password-stdin >/dev/null 2>&1; then
        log_success "Docker login successful"
        return 0
    else
        log_error "Docker login failed"
        return 1
    fi
}

# get_image_digest() - Get SHA256 digest of an image
# Usage: get_image_digest "backend" "sha-abc1234"
get_image_digest() {
    component="$1"
    tag="$2"

    image_name="$DOCKER_REGISTRY/$IMAGE_PREFIX/$component:$tag"

    digest=$(docker inspect --format='{{index .RepoDigests 0}}' "$image_name" 2>/dev/null | cut -d'@' -f2)

    if [ -n "$digest" ]; then
        echo "$digest"
        return 0
    else
        log_error "Failed to get digest for image: $image_name"
        return 1
    fi
}

# cleanup_old_images() - Remove old Docker images to free space
# Usage: cleanup_old_images [days_to_keep]
cleanup_old_images() {
    days="${1:-7}"

    log_info "Cleaning up Docker images older than $days days..."

    # Remove dangling images
    if docker image prune -f --filter "until=${days}d" >/dev/null 2>&1; then
        log_success "Docker cleanup completed"
        return 0
    else
        log_warn "Docker cleanup encountered issues"
        return 1
    fi
}
