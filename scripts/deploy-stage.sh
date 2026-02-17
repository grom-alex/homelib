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
#   --build              Build images locally and transfer to staging server
#                        (use when Docker registry is not available)
#
# Environment Variables (can be set in .env file):
#   STAGE_HOST           Stage VM hostname/IP
#   STAGE_USER           SSH user (default: deploy)
#   DEPLOY_SSH_KEY       Path to SSH private key
#   DOCKER_REGISTRY      Docker registry (default: registry.gromas.ru)
#   IMAGE_PREFIX         Path prefix in registry (default: apps/homelib)
#   IMAGE_TAG            Image tag to deploy (default: latest)
#
# Note: Script automatically loads .env from project root if present

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

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
SKIP_HEALTH=false
DRY_RUN=false
HEALTH_TIMEOUT=60
REMOTE_APP_DIR="/opt/homelib"

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
    echo -e "${BLUE}===${NC} $1 ${BLUE}===${NC}"
    echo ""
}

ssh_exec() {
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] Would execute on ${STAGE_HOST}: $1"
    else
        ssh -i "$SSH_KEY" -o StrictHostKeyChecking=no "${STAGE_USER}@${STAGE_HOST}" "$1"
    fi
}

scp_file() {
    local src="$1"
    local dst="$2"
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] Would copy $src to ${STAGE_HOST}:$dst"
    else
        scp -i "$SSH_KEY" -o StrictHostKeyChecking=no "$src" "${STAGE_USER}@${STAGE_HOST}:$dst"
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
log_step "Deployment Plan"
log_info "Target:          ${STAGE_HOST}"
log_info "User:            ${STAGE_USER}"
log_info "SSH Key:         ${SSH_KEY}"
log_info "Image Tag:       ${IMAGE_TAG}"
log_info "Registry:        ${DOCKER_REGISTRY}/${IMAGE_PREFIX}"
log_info "Skip Health:     ${SKIP_HEALTH}"
log_info "Dry Run:         ${DRY_RUN}"

if [ "$DRY_RUN" = true ]; then
    log_warn "DRY RUN MODE - No changes will be made"
fi

# Step 1: Test SSH connection
log_step "Step 1: Testing SSH Connection"
if [ "$DRY_RUN" = false ]; then
    if ssh -i "$SSH_KEY" -o StrictHostKeyChecking=no -o ConnectTimeout=10 "${STAGE_USER}@${STAGE_HOST}" "echo 'SSH connection successful'" > /dev/null 2>&1; then
        log_info "SSH connection successful"
    else
        log_error "Cannot connect to ${STAGE_HOST}"
        exit 1
    fi
else
    echo "[DRY RUN] Would test SSH connection to ${STAGE_HOST}"
fi

# Step 2: Backup current state for rollback (FR-017)
log_step "Step 2: Backing Up Current State"
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
log_step "Step 3: Copying Deployment Files"
ssh_exec "mkdir -p ${REMOTE_APP_DIR}/nginx"
scp_file "${PROJECT_ROOT}/docker/docker-compose.stage.yml" "${REMOTE_APP_DIR}/docker-compose.yml"
scp_file "${PROJECT_ROOT}/docker/nginx/nginx.prod.conf" "${REMOTE_APP_DIR}/nginx/nginx.conf"
scp_file "${PROJECT_ROOT}/config/config.stage.yaml" "${REMOTE_APP_DIR}/config.yaml"
log_info "Deployment files copied"

# Step 4: Pull and deploy (FR-015)
log_step "Step 4: Pulling and Deploying"
ssh_exec "
    cd ${REMOTE_APP_DIR}

    # Set environment
    export IMAGE_TAG='${IMAGE_TAG}'
    export DOCKER_REGISTRY='${DOCKER_REGISTRY}'
    export IMAGE_PREFIX='${IMAGE_PREFIX}'

    echo 'Pulling images...'
    docker compose pull

    echo 'Starting services...'
    docker compose up -d --remove-orphans

    echo 'Deployment commands executed'
"
log_info "Deployment complete"

# Step 5: Health check (FR-016)
if [ "$SKIP_HEALTH" = false ]; then
    log_step "Step 5: Health Check (${HEALTH_TIMEOUT}s timeout)"

    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] Would perform health check on ${STAGE_HOST}"
    else
        ATTEMPTS=$((HEALTH_TIMEOUT / 5))
        HEALTHY=false

        for i in $(seq 1 $ATTEMPTS); do
            echo "Health check attempt $i/${ATTEMPTS}..."

            # Check health via nginx reverse proxy (port 80)
            if curl -sf "http://${STAGE_HOST}/health" > /dev/null 2>&1; then
                HEALTHY=true
                echo "  Health check: OK"
                break
            else
                echo "  Health check: FAIL"
            fi

            sleep 5
        done

        if [ "$HEALTHY" = true ]; then
            log_info "All services healthy!"
        else
            log_error "Health check failed after ${HEALTH_TIMEOUT}s"
            log_warn "Initiating automatic rollback..."

            # Attempt rollback
            ssh_exec "
                cd ${REMOTE_APP_DIR}

                if [ -f .previous-images ]; then
                    echo 'Rolling back...'
                    docker compose down

                    # Extract previous tag
                    PREV_TAG=\$(cat .previous-images | grep backend | head -1 | sed 's/.*://')
                    if [ -n \"\$PREV_TAG\" ]; then
                        export IMAGE_TAG=\"\$PREV_TAG\"
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
    log_step "Step 5: Health Check Skipped"
    log_warn "Health check was skipped with --skip-health flag"
fi

# Summary
log_step "Deployment Summary"
echo ""
log_info "Stage deployment completed successfully!"
echo ""
echo "  Environment:  Stage"
echo "  Host:         ${STAGE_HOST}"
echo "  Image Tag:    ${IMAGE_TAG}"
echo "  URL:          http://${STAGE_HOST}"
echo ""

if [ "$DRY_RUN" = false ]; then
    log_info "View logs: ssh ${STAGE_USER}@${STAGE_HOST} 'cd ${REMOTE_APP_DIR} && docker compose logs -f'"
fi

exit 0
