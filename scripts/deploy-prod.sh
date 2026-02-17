#!/bin/bash
# Production Deployment Script for HomeLib
# Deploys HomeLib to production environment via SSH with backup and rollback
#
# CAUTION: This script deploys to PRODUCTION. Use with care.
#
# Usage: ./scripts/deploy-prod.sh [OPTIONS]
#   -t, --tag TAG        Image tag to deploy (REQUIRED)
#   -h, --host HOST      Production VM hostname/IP (REQUIRED)
#   -u, --user USER      SSH user (default: deploy)
#   -k, --key PATH       SSH key path (default: ~/.ssh/deploy_key)
#   --skip-confirm       Skip confirmation prompt (dangerous!)
#   --skip-health        Skip health check after deployment
#   --dry-run            Show what would be done without executing
#
# Environment Variables:
#   PROD_HOST            Production VM hostname/IP
#   PROD_USER            SSH user (default: deploy)
#   DEPLOY_SSH_KEY       Path to SSH private key
#   DOCKER_REGISTRY      Docker registry (default: registry.gromas.ru)
#   IMAGE_PREFIX         Path prefix in registry (default: apps/homelib)
#   NGINX_PORT           Nginx port (default: 80)

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Load shared logging library
# shellcheck source=scripts/lib/logging.sh
. "$SCRIPT_DIR/lib/logging.sh"

# Default values
IMAGE_TAG=""
PROD_HOST="${PROD_HOST:-}"
PROD_USER="${PROD_USER:-deploy}"
SSH_KEY="${DEPLOY_SSH_KEY:-$HOME/.ssh/deploy_key}"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-registry.gromas.ru}"
IMAGE_PREFIX="${IMAGE_PREFIX:-apps/homelib}"
NGINX_PORT="${NGINX_PORT:-80}"
SKIP_CONFIRM=false
SKIP_HEALTH=false
DRY_RUN=false
HEALTH_TIMEOUT=60
REMOTE_APP_DIR="/opt/homelib"

ssh_exec() {
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] Would execute on ${PROD_HOST}: $1"
    else
        ssh -i "$SSH_KEY" -o StrictHostKeyChecking=accept-new "${PROD_USER}@${PROD_HOST}" "$1"
    fi
}

scp_file() {
    local src="$1"
    local dst="$2"
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] Would copy $src to ${PROD_HOST}:$dst"
    else
        scp -i "$SSH_KEY" -o StrictHostKeyChecking=accept-new "$src" "${PROD_USER}@${PROD_HOST}:$dst"
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
            PROD_HOST="$2"
            shift 2
            ;;
        -u|--user)
            PROD_USER="$2"
            shift 2
            ;;
        -k|--key)
            SSH_KEY="$2"
            shift 2
            ;;
        --skip-confirm)
            SKIP_CONFIRM=true
            shift
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
            echo "CAUTION: This script deploys to PRODUCTION!"
            echo ""
            echo "Options:"
            echo "  -t, --tag TAG        Image tag to deploy (REQUIRED)"
            echo "  -h, --host HOST      Production VM hostname/IP (REQUIRED)"
            echo "  -u, --user USER      SSH user (default: deploy)"
            echo "  -k, --key PATH       SSH key path (default: ~/.ssh/deploy_key)"
            echo "  --skip-confirm       Skip confirmation prompt (dangerous!)"
            echo "  --skip-health        Skip health check after deployment"
            echo "  --dry-run            Show what would be done without executing"
            echo ""
            echo "Environment Variables:"
            echo "  PROD_HOST            Production VM hostname/IP"
            echo "  DOCKER_REGISTRY      Docker registry (default: registry.gromas.ru)"
            echo "  IMAGE_PREFIX         Path prefix in registry (default: apps/homelib)"
            echo "  NGINX_PORT           Nginx port (default: 80)"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Validation
if [ -z "$IMAGE_TAG" ]; then
    log_error "IMAGE_TAG is required"
    log_error "Use --tag flag or set IMAGE_TAG environment variable"
    exit 1
fi

if [ -z "$PROD_HOST" ]; then
    log_error "PROD_HOST is required"
    log_error "Use --host flag or set PROD_HOST environment variable"
    exit 1
fi

if [ ! -f "$SSH_KEY" ] && [ "$DRY_RUN" = false ]; then
    log_error "SSH key not found: $SSH_KEY"
    exit 1
fi

# Display deployment plan
log_section "Production Deployment Plan"
log_warn "PRODUCTION DEPLOYMENT"
log_info "Target:     ${PROD_HOST}"
log_info "User:       ${PROD_USER}"
log_info "Image Tag:  ${IMAGE_TAG}"
log_info "Registry:   ${DOCKER_REGISTRY}/${IMAGE_PREFIX}"
log_info "Nginx Port: ${NGINX_PORT}"

if [ "$DRY_RUN" = true ]; then
    log_warn "DRY RUN MODE - No changes will be made"
fi

# Confirmation
if [ "$SKIP_CONFIRM" = false ] && [ "$DRY_RUN" = false ]; then
    echo ""
    log_warn "You are about to deploy to PRODUCTION!"
    log_warn "This will affect live users."
    echo ""
    read -p "Type 'yes' to confirm: " -r
    echo
    if [[ ! $REPLY =~ ^yes$ ]]; then
        log_info "Deployment cancelled"
        exit 0
    fi
fi

# Step 1: Test SSH connection
log_section "Step 1: Testing SSH Connection"
if [ "$DRY_RUN" = false ]; then
    if ssh -i "$SSH_KEY" -o StrictHostKeyChecking=accept-new -o ConnectTimeout=10 "${PROD_USER}@${PROD_HOST}" "echo 'SSH connection successful'" > /dev/null 2>&1; then
        log_info "SSH connection successful"
    else
        log_error "Cannot connect to ${PROD_HOST}"
        exit 1
    fi
else
    echo "[DRY RUN] Would test SSH connection to ${PROD_HOST}"
fi

# Step 2: Check current state
log_section "Step 2: Checking Current State"
ssh_exec "
    cd ${REMOTE_APP_DIR}
    echo 'Current containers:'
    docker compose ps 2>/dev/null || echo 'No containers running'
    echo ''
    echo 'Current images:'
    docker compose config 2>/dev/null | grep 'image:' || echo 'No compose config'
"

# Step 3: Backup current state
log_section "Step 3: Backing Up Current State"
ssh_exec "
    cd ${REMOTE_APP_DIR}

    # Save current image tags for rollback
    docker compose config 2>/dev/null | grep 'image:' > .previous-images || true

    # Save container status
    docker compose ps --format 'table {{.Name}}\t{{.Status}}' > .previous-status 2>/dev/null || true

    echo 'Backup saved to .previous-images and .previous-status'
"
log_info "State backup complete"

# Step 4: Setup JWT Keys
log_section "Step 4: Setting Up JWT Keys"
ssh_exec "
    KEYS_DIR='${REMOTE_APP_DIR}/keys'
    mkdir -p \"\$KEYS_DIR\"

    # Check if keys exist
    if [ -f \"\$KEYS_DIR/private.pem\" ] && [ -f \"\$KEYS_DIR/public.pem\" ]; then
        echo 'JWT keys already exist, ensuring correct permissions...'
    else
        echo 'Generating new JWT keys...'
        openssl genrsa -out \"\$KEYS_DIR/private.pem\" 2048 2>/dev/null
        openssl rsa -in \"\$KEYS_DIR/private.pem\" -pubout -out \"\$KEYS_DIR/public.pem\" 2>/dev/null
        echo 'JWT keys generated successfully'
    fi

    # Private key readable only by owner, public key world-readable
    chmod 600 \"\$KEYS_DIR/private.pem\"
    chmod 644 \"\$KEYS_DIR/public.pem\"
"
log_info "JWT keys ready"

# Step 5: Copy deployment files
log_section "Step 5: Copying Deployment Files"
ssh_exec "mkdir -p ${REMOTE_APP_DIR}/nginx"
scp_file "${PROJECT_ROOT}/docker/docker-compose.prod.yml" "${REMOTE_APP_DIR}/docker-compose.yml"
scp_file "${PROJECT_ROOT}/docker/nginx/nginx.prod.conf" "${REMOTE_APP_DIR}/nginx/nginx.conf"
scp_file "${PROJECT_ROOT}/config/config.prod.yaml" "${REMOTE_APP_DIR}/config.yaml"
log_info "Deployment files copied"

# Step 6: Update .env on remote server and deploy
log_section "Step 6: Pulling and Deploying"
ssh_exec "
    cd ${REMOTE_APP_DIR}

    # Update IMAGE_TAG in .env (create if not exists)
    if [ -f .env ]; then
        if grep -q '^IMAGE_TAG=' .env; then
            sed -i 's|^IMAGE_TAG=.*|IMAGE_TAG=${IMAGE_TAG}|' .env
        else
            echo 'IMAGE_TAG=${IMAGE_TAG}' >> .env
        fi
    else
        echo 'IMAGE_TAG=${IMAGE_TAG}' > .env
    fi

    echo 'Pulling images...'
    docker compose pull

    echo 'Starting services...'
    docker compose up -d --remove-orphans

    echo 'Deployment commands executed'
"
log_info "Deployment complete"

# Step 7: Health check
if [ "$SKIP_HEALTH" = false ]; then
    log_section "Step 7: Health Check (${HEALTH_TIMEOUT}s timeout)"

    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] Would perform health check on ${PROD_HOST}:${NGINX_PORT}"
    else
        ATTEMPTS=$((HEALTH_TIMEOUT / 5))
        HEALTHY=false

        for i in $(seq 1 $ATTEMPTS); do
            echo "Health check attempt $i/${ATTEMPTS}..."

            # Check health via nginx reverse proxy with correct port
            if curl -sf "http://${PROD_HOST}:${NGINX_PORT}/health" > /dev/null 2>&1; then
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
            log_warn ""
            log_warn "MANUAL INTERVENTION REQUIRED!"
            log_warn "To rollback, run:"
            log_warn "  ssh ${PROD_USER}@${PROD_HOST}"
            log_warn "  cd ${REMOTE_APP_DIR}"
            log_warn "  # Check previous tag: cat .previous-images"
            log_warn "  # Update .env with previous IMAGE_TAG and run: docker compose up -d"
            log_warn ""
            exit 1
        fi
    fi
else
    log_section "Step 7: Health Check Skipped"
    log_warn "Health check was skipped with --skip-health flag"
    log_warn "Manually verify the deployment!"
fi

# Summary
log_section "Deployment Summary"

log_success "PRODUCTION DEPLOYMENT SUCCESSFUL"
log_info "Environment:  Production"
log_info "Host:         ${PROD_HOST}"
log_info "Image Tag:    ${IMAGE_TAG}"
log_info "URL:          http://${PROD_HOST}:${NGINX_PORT}"

if [ "$DRY_RUN" = false ]; then
    log_info "Post-Deployment Actions:"
    log_info "1. Monitor logs: ssh ${PROD_USER}@${PROD_HOST} 'cd ${REMOTE_APP_DIR} && docker compose logs -f'"
    log_info "2. Verify user flows work correctly"
    log_info "3. Monitor error rates in logs"
fi

exit 0
