#!/bin/sh
# Local Deployment Execution Script
# Tasks: T045, T048, T052, T061-T063
# Reference: FR-016, FR-023-024, SC-002, SC-009

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
ENVIRONMENT="local"  # local, staging, production
TAG="latest"
SKIP_TESTS=false
VALIDATE=true
COMPOSE_FILE="docker/docker-compose.dev.yml"

# ============================================================================
# Usage Information (FR-024, SC-009)
# ============================================================================
usage() {
    cat <<EOF
Usage: $0 [ENVIRONMENT] [OPTIONS]

Local deployment execution script - deploys without CI infrastructure (FR-016)

ENVIRONMENTS:
    local       Deploy to local environment (default)
    staging     Deploy to staging environment
    production  Deploy to production environment

OPTIONS:
    --tag TAG           Docker image tag to deploy (default: latest) (T062)
    --skip-tests        Skip running tests before deployment (T062)
    --no-validate       Skip health check validation after deployment (T062)
    --compose-file FILE Use specific docker-compose file
    --help, -h          Show this help message

EXAMPLES:
    $0                               # Deploy to local with default settings
    $0 staging --tag sha-abc123      # Deploy to staging with specific tag
    $0 production --tag v1.0.0       # Deploy to production with version tag
    $0 local --skip-tests           # Deploy locally without running tests

PREREQUISITES:
    - Docker and Docker Compose (required)
    - Access to Docker registry (for pulling images)
    - Proper environment configuration (.env files)

DOCUMENTATION:
    See: docs/deployment-guide.md
    Spec: specs/003-cicd-pipeline-rework/spec.md

EXIT CODES:
    0 - Success
    1 - Prerequisite check failed
    2 - Deployment execution error
EOF
}

# ============================================================================
# Argument Parsing
# ============================================================================
parse_arguments() {
    # First positional argument is environment
    if [ $# -gt 0 ] && [ "${1#--}" = "$1" ]; then
        ENVIRONMENT="$1"
        shift
    fi

    while [ $# -gt 0 ]; do
        case "$1" in
            --tag)
                TAG="$2"
                shift 2
                ;;
            --skip-tests)
                SKIP_TESTS=true
                shift
                ;;
            --no-validate)
                VALIDATE=false
                shift
                ;;
            --compose-file)
                COMPOSE_FILE="$2"
                shift 2
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
# Prerequisite Validation (T048, FR-023)
# ============================================================================
check_prerequisites() {
    log_section "Checking Prerequisites"

    # Docker is required
    check_docker
    check_docker_compose

    # Check environment-specific prerequisites
    case "$ENVIRONMENT" in
        local)
            check_file ".env" "Create .env file from .env.example"
            ;;
        staging)
            check_file ".env.staging" "Create .env.staging file"
            check_env_var "STAGE_HOST" "Set STAGE_HOST environment variable"
            ;;
        production)
            check_file ".env.prod" "Create .env.prod file"
            check_env_var "PROD_HOST" "Set PROD_HOST environment variable"
            log_warn "⚠️  Production deployment - proceed with caution!"
            ;;
        *)
            log_error "Unknown environment: $ENVIRONMENT"
            log_error "  → Fix: Use local, staging, or production"
            exit $EXIT_PREREQUISITE_FAILURE
            ;;
    esac

    # Check compose file exists
    check_file "$COMPOSE_FILE" "Ensure docker-compose file exists: $COMPOSE_FILE"

    prereq_summary
}

# ============================================================================
# Run Tests Before Deployment (T061)
# ============================================================================
run_predeploy_tests() {
    if [ "$SKIP_TESTS" = "true" ]; then
        log_warn "Skipping pre-deployment tests (--skip-tests)"
        return 0
    fi

    log_section "Running Pre-Deployment Tests"

    # Run unit tests to ensure code quality
    if [ -x "$SCRIPT_DIR/test-local.sh" ]; then
        log_info "Running unit tests..."
        if ! "$SCRIPT_DIR/test-local.sh" unit; then
            log_error "Pre-deployment tests failed"
            log_error "  → Fix: Fix failing tests or use --skip-tests to bypass"
            exit $EXIT_EXECUTION_ERROR
        fi
        log_success "Pre-deployment tests passed"
    else
        log_warn "test-local.sh not found - skipping tests"
    fi
}

# ============================================================================
# Local Deployment (T061)
# ============================================================================
deploy_local() {
    log_section "Deploying to Local Environment"

    cd "$SCRIPT_DIR/.." || exit $EXIT_EXECUTION_ERROR

    # Set image tag
    export IMAGE_TAG="$TAG"

    # Stop existing containers
    log_info "Stopping existing containers..."
    docker compose -f "$COMPOSE_FILE" down || true

    # Pull/build images
    log_info "Preparing images..."
    if [ "$TAG" = "latest" ]; then
        log_info "Building images locally..."
        docker compose -f "$COMPOSE_FILE" build
    else
        log_info "Pulling images with tag: $TAG"
        docker compose -f "$COMPOSE_FILE" pull
    fi

    # Start containers
    log_info "Starting containers..."
    docker compose -f "$COMPOSE_FILE" up -d

    cd "$SCRIPT_DIR" || exit $EXIT_EXECUTION_ERROR
}

# ============================================================================
# Staging Deployment (T061)
# ============================================================================
deploy_staging() {
    log_section "Deploying to Staging Environment"

    log_warn "Staging deployment typically uses CD workflow"
    log_warn "This is a manual deployment for testing purposes"

    # Use deploy-stage.sh if available
    if [ -x "$SCRIPT_DIR/deploy-stage.sh" ]; then
        log_info "Using deploy-stage.sh script..."
        "$SCRIPT_DIR/deploy-stage.sh" --tag "$TAG" --host "$STAGE_HOST"
    else
        log_error "deploy-stage.sh not found"
        log_error "  → Fix: Use CD workflow or ensure deploy-stage.sh exists"
        exit $EXIT_EXECUTION_ERROR
    fi
}

# ============================================================================
# Production Deployment (T061)
# ============================================================================
deploy_production() {
    log_section "Deploying to Production Environment"

    log_warn "⚠️  PRODUCTION DEPLOYMENT"
    log_warn "This should typically be done via CD workflow"

    # Confirmation prompt
    printf "Are you sure you want to deploy to PRODUCTION? (yes/no): "
    read -r confirmation
    if [ "$confirmation" != "yes" ]; then
        log_info "Deployment cancelled"
        exit $EXIT_SUCCESS
    fi

    # Use deploy-prod.sh if available
    if [ -x "$SCRIPT_DIR/deploy-prod.sh" ]; then
        log_info "Using deploy-prod.sh script..."
        "$SCRIPT_DIR/deploy-prod.sh" --tag "$TAG" --host "$PROD_HOST"
    else
        log_error "deploy-prod.sh not found"
        log_error "  → Fix: Use CD workflow or ensure deploy-prod.sh exists"
        exit $EXIT_EXECUTION_ERROR
    fi
}

# ============================================================================
# Health Check Validation (T062, T063)
# ============================================================================
validate_deployment() {
    if [ "$VALIDATE" = "false" ]; then
        log_warn "Skipping health check validation (--no-validate)"
        return 0
    fi

    log_section "Validating Deployment"

    # Determine health check URL based on environment (via nginx)
    case "$ENVIRONMENT" in
        local)
            HEALTH_URL="http://localhost:${NGINX_PORT:-80}/health"
            ;;
        staging)
            HEALTH_URL="http://$STAGE_HOST:${NGINX_PORT:-80}/health"
            ;;
        production)
            HEALTH_URL="http://$PROD_HOST:${NGINX_PORT:-80}/health"
            ;;
    esac

    # Wait for health check (T063: use docker-utils.sh)
    if wait_for_url "$HEALTH_URL" 120; then
        log_success "Deployment validation passed"
        return 0
    else
        log_error "Deployment validation failed"
        log_error "  → Problem: Health check endpoint not responding"
        log_error "  → Fix: Check container logs with 'docker compose logs'"
        log_error "  → Docs: See docs/troubleshooting.md"
        exit $EXIT_EXECUTION_ERROR
    fi
}

# ============================================================================
# Main Execution
# ============================================================================
main() {
    # Parse arguments first (for --help)
    parse_arguments "$@"

    log_info "Starting local deployment execution"
    log_info "Environment: $ENVIRONMENT"
    log_info "Tag: $TAG"

    # Check prerequisites
    check_prerequisites

    # Run pre-deployment tests
    run_predeploy_tests

    # Deploy based on environment
    case "$ENVIRONMENT" in
        local)
            deploy_local
            ;;
        staging)
            deploy_staging
            ;;
        production)
            deploy_production
            ;;
    esac

    # Validate deployment
    validate_deployment

    log_success "Deployment completed successfully!"
    log_info "Environment: $ENVIRONMENT"
    log_info "Tag: $TAG"
    log_info "Health Check: $HEALTH_URL"

    exit $EXIT_SUCCESS
}

# Run main function
main "$@"
