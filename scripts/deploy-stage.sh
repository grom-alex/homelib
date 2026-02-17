#!/bin/bash
# Stage Deployment Script for HomeLib
# Deploys HomeLib to staging environment via SSH
#
# Usage: ./scripts/deploy-stage.sh [OPTIONS]
#   -t, --tag TAG        Image tag to deploy (default: latest)
#   -h, --host HOST      Stage VM hostname/IP (default: from STAGE_HOST env)
#   -u, --user USER      SSH user (default: deploy)
#   -k, --key PATH       SSH key path (default: ~/.ssh/deploy_key)
#   --skip-health        Skip health check after deployment
#   --dry-run            Show what would be done without executing
#
# Environment Variables (can be set in .env file):
#   STAGE_HOST           Stage VM hostname/IP
#   STAGE_USER           SSH user (default: deploy)
#   DEPLOY_SSH_KEY       Path to SSH private key
#   DOCKER_REGISTRY      Docker registry (default: registry.gromas.ru)
#   IMAGE_PREFIX         Path prefix in registry (default: apps/homelib)
#   IMAGE_TAG            Image tag to deploy (default: latest)
#   NGINX_PORT           Nginx port on staging (default: 80)
#
# Note: Script automatically loads .env from project root if present

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Load shared logging library
# shellcheck source=scripts/lib/logging.sh
. "$SCRIPT_DIR/lib/logging.sh"

# Load .env file if exists (for STAGE_HOST and other defaults)
if [ -f "${PROJECT_ROOT}/.env" ]; then
    # shellcheck disable=SC1091
    set -a
    source "${PROJECT_ROOT}/.env"
    set +a
fi

# Default values
IMAGE_TAG="${IMAGE_TAG:-latest}"
STAGE_HOST="${STAGE_HOST:-}"
STAGE_USER="${STAGE_USER:-deploy}"
SSH_KEY="${DEPLOY_SSH_KEY:-$HOME/.ssh/deploy_key}"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-registry.gromas.ru}"
IMAGE_PREFIX="${IMAGE_PREFIX:-apps/homelib}"
NGINX_PORT="${NGINX_PORT:-80}"
SKIP_HEALTH=false
DRY_RUN=false
HEALTH_TIMEOUT=120
REMOTE_APP_DIR="/opt/homelib"

ssh_exec() {
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] Would execute on ${STAGE_HOST}: $1"
    else
        ssh -i "$SSH_KEY" -o StrictHostKeyChecking=accept-new "${STAGE_USER}@${STAGE_HOST}" "$1"
    fi
}

scp_file() {
    local src="$1"
    local dst="$2"
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] Would copy $src to ${STAGE_HOST}:$dst"
    else
        scp -i "$SSH_KEY" -o StrictHostKeyChecking=accept-new "$src" "${STAGE_USER}@${STAGE_HOST}:$dst"
    fi
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        -h|--host)
            STAGE_HOST="$2"
            shift 2
            ;;
        -u|--user)
            STAGE_USER="$2"
            shift 2
            ;;
        -k|--key)
            SSH_KEY="$2"
            shift 2
            ;;
        --skip-health)
            SKIP_HEALTH=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -t, --tag TAG        Image tag to deploy (default: latest)"
            echo "  -h, --host HOST      Stage VM hostname/IP"
            echo "  -u, --user USER      SSH user (default: deploy)"
            echo "  -k, --key PATH       SSH key path (default: ~/.ssh/deploy_key)"
            echo "  --skip-health        Skip health check after deployment"
            echo "  --dry-run            Show what would be done without executing"
            echo ""
            echo "Environment Variables:"
            echo "  STAGE_HOST           Stage VM hostname/IP"
            echo "  DOCKER_REGISTRY      Docker registry (default: registry.gromas.ru)"
            echo "  IMAGE_PREFIX         Path prefix in registry (default: apps/homelib)"
            echo "  IMAGE_TAG            Image tag to deploy"
            echo "  NGINX_PORT           Nginx port on staging (default: 80)"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Validation
if [ -z "$STAGE_HOST" ]; then
    log_error "STAGE_HOST is required"
    log_error "Set via --host flag or STAGE_HOST environment variable"
    exit 1
fi

if [ ! -f "$SSH_KEY" ] && [ "$DRY_RUN" = false ]; then
    log_error "SSH key not found: $SSH_KEY"
    exit 1
fi

# Display deployment plan
log_section "Deployment Plan"
log_info "Target:          ${STAGE_HOST}"
log_info "User:            ${STAGE_USER}"
log_info "SSH Key:         ${SSH_KEY}"
log_info "Image Tag:       ${IMAGE_TAG}"
log_info "Registry:        ${DOCKER_REGISTRY}/${IMAGE_PREFIX}"
log_info "Nginx Port:      ${NGINX_PORT}"
log_info "Skip Health:     ${SKIP_HEALTH}"
log_info "Dry Run:         ${DRY_RUN}"

if [ "$DRY_RUN" = true ]; then
    log_warn "DRY RUN MODE - No changes will be made"
fi

# Step 1: Test SSH connection
log_section "Step 1: Testing SSH Connection"
if [ "$DRY_RUN" = false ]; then
    if ssh -i "$SSH_KEY" -o StrictHostKeyChecking=accept-new -o ConnectTimeout=10 "${STAGE_USER}@${STAGE_HOST}" "echo 'SSH connection successful'" > /dev/null 2>&1; then
        log_info "SSH connection successful"
    else
        log_error "Cannot connect to ${STAGE_HOST}"
        exit 1
    fi
else
    echo "[DRY RUN] Would test SSH connection to ${STAGE_HOST}"
fi

# Step 2: Backup current state for rollback (FR-017)
log_section "Step 2: Backing Up Current State"
ssh_exec "
    cd ${REMOTE_APP_DIR}

    # Save current image tags for rollback
    if docker compose ps -q 2>/dev/null | head -1 > /dev/null; then
        echo 'Saving current state...'
        docker compose config 2>/dev/null | grep 'image:' > .previous-images || true
        docker compose ps --format 'table {{.Name}}\t{{.Status}}' > .previous-status || true
        echo 'Current state saved'
    else
        echo 'No running containers - fresh deployment'
    fi
"
log_info "Backup complete"

# Step 3: Copy deployment files
log_section "Step 3: Copying Deployment Files"
ssh_exec "mkdir -p ${REMOTE_APP_DIR}/nginx"
scp_file "${PROJECT_ROOT}/docker/docker-compose.stage.yml" "${REMOTE_APP_DIR}/docker-compose.yml"
scp_file "${PROJECT_ROOT}/docker/nginx/nginx.prod.conf" "${REMOTE_APP_DIR}/nginx/nginx.conf"
scp_file "${PROJECT_ROOT}/config/config.stage.yaml" "${REMOTE_APP_DIR}/config.yaml"
log_info "Deployment files copied"

# Step 4: Update .env on remote server with IMAGE_TAG
log_section "Step 4: Updating Remote Environment"
ssh_exec "
    cd ${REMOTE_APP_DIR}

    # Update IMAGE_TAG in .env (create if not exists)
    if [ -f .env ]; then
        # Replace IMAGE_TAG if exists, append if not
        if grep -q '^IMAGE_TAG=' .env; then
            sed -i 's|^IMAGE_TAG=.*|IMAGE_TAG=${IMAGE_TAG}|' .env
        else
            echo 'IMAGE_TAG=${IMAGE_TAG}' >> .env
        fi
    else
        echo 'IMAGE_TAG=${IMAGE_TAG}' > .env
    fi

    echo 'Remote .env updated with IMAGE_TAG=${IMAGE_TAG}'
"
log_info "Remote environment updated"

# Step 5: Pull and deploy (FR-015)
log_section "Step 5: Pulling and Deploying"
ssh_exec "
    cd ${REMOTE_APP_DIR}

    echo 'Pulling images...'
    docker compose pull

    echo 'Starting services...'
    docker compose up -d --remove-orphans

    # Перезапуск nginx для обновления DNS (upstream IP мог измениться)
    echo 'Restarting nginx to refresh upstream DNS...'
    docker compose restart nginx

    # Ожидание готовности контейнеров
    echo 'Waiting for containers to start...'
    WAIT_TIMEOUT=90
    ELAPSED=0
    while [ \$ELAPSED -lt \$WAIT_TIMEOUT ]; do
        STATUSES=\$(docker compose ps --format '{{.Status}}' 2>/dev/null)
        if echo \"\$STATUSES\" | grep -qi 'starting'; then
            sleep 5
            ELAPSED=\$((ELAPSED + 5))
            echo \"  Waiting for containers... (\${ELAPSED}s/\${WAIT_TIMEOUT}s)\"
        else
            echo 'All containers started'
            break
        fi
    done

    docker compose ps
"
log_info "Deployment complete"

# Step 6: Health check (FR-016)
if [ "$SKIP_HEALTH" = false ]; then
    log_section "Step 6: Health Check (${HEALTH_TIMEOUT}s timeout)"

    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] Would perform health check on ${STAGE_HOST}:${NGINX_PORT}"
    else
        ATTEMPTS=$((HEALTH_TIMEOUT / 5))
        HEALTHY=false

        for i in $(seq 1 $ATTEMPTS); do
            echo "Health check attempt $i/${ATTEMPTS}..."

            # Check health via nginx reverse proxy with correct port
            if curl -sf "http://${STAGE_HOST}:${NGINX_PORT}/health" > /dev/null 2>&1; then
                HEALTHY=true
                echo "  Health check: OK"
                break
            else
                echo "  Health check: FAIL"
            fi

            sleep 5
        done

        if [ "$HEALTHY" = true ]; then
            log_success "All services healthy!"
        else
            log_error "Health check failed after ${HEALTH_TIMEOUT}s"
            log_warn "Initiating automatic rollback..."

            # Attempt rollback
            ssh_exec "
                cd ${REMOTE_APP_DIR}

                if [ -f .previous-images ]; then
                    echo 'Rolling back...'
                    docker compose down

                    # Extract previous tag (search for api image)
                    PREV_TAG=\$(grep 'api' .previous-images | head -1 | sed 's/.*://')
                    if [ -n \"\$PREV_TAG\" ]; then
                        # Update .env with previous tag
                        sed -i \"s|^IMAGE_TAG=.*|IMAGE_TAG=\$PREV_TAG|\" .env
                        docker compose up -d
                        echo \"Rolled back to \$PREV_TAG\"
                    else
                        echo 'Could not determine previous tag'
                        exit 1
                    fi
                else
                    echo 'No previous state for rollback'
                fi
            " || true

            exit 1
        fi
    fi
else
    log_section "Step 6: Health Check Skipped"
    log_warn "Health check was skipped with --skip-health flag"
fi

# Summary
log_section "Deployment Summary"

log_success "Stage deployment completed successfully!"
log_info "Environment:  Stage"
log_info "Host:         ${STAGE_HOST}"
log_info "Image Tag:    ${IMAGE_TAG}"
log_info "URL:          http://${STAGE_HOST}:${NGINX_PORT}"

if [ "$DRY_RUN" = false ]; then
    log_info "View logs: ssh ${STAGE_USER}@${STAGE_HOST} 'cd ${REMOTE_APP_DIR} && docker compose logs -f'"
fi

exit 0
