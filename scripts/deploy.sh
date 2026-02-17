#!/bin/sh
# Deployment wrapper script for HomeLib
# Usage: ./scripts/deploy.sh <environment> [image_tag]
#
# Examples:
#   ./scripts/deploy.sh staging sha-abc1234
#   ./scripts/deploy.sh production v1.0.0
#   IMAGE_TAG=v1.0.0 ./scripts/deploy.sh production

set -e  # Exit on error

# ============================================================================
# Script Directory and Library Loading
# ============================================================================
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=scripts/lib/logging.sh
. "$SCRIPT_DIR/lib/logging.sh"

# ============================================================================
# Exit Codes
# ============================================================================
EXIT_SUCCESS=0
EXIT_PREREQUISITE_FAILURE=1
EXIT_EXECUTION_ERROR=2

# ============================================================================
# Usage Information
# ============================================================================
usage() {
    cat <<EOF
Usage: $0 <environment> [image_tag]

Deployment wrapper script for HomeLib — routes to environment-specific deploy scripts

ENVIRONMENTS:
    staging      Deploy to staging via deploy-stage.sh
    production   Deploy to production via deploy-prod.sh

ARGUMENTS:
    image_tag    Docker image tag to deploy (or set IMAGE_TAG env var)

OPTIONS:
    --help, -h   Show this help message

EXAMPLES:
    $0 staging sha-abc1234
    $0 production v1.0.0
    IMAGE_TAG=v1.0.0 $0 production

EXIT CODES:
    0 - Success
    1 - Prerequisite check failed
    2 - Deployment execution error
EOF
}

# ============================================================================
# Argument Parsing
# ============================================================================
if [ $# -lt 1 ]; then
    log_error "Usage: $0 <environment> [image_tag]"
    log_error "Environments: staging, production"
    log_error "Example: $0 staging sha-abc1234"
    exit $EXIT_PREREQUISITE_FAILURE
fi

# Check for --help before anything else
case "$1" in
    --help|-h)
        usage
        exit $EXIT_SUCCESS
        ;;
esac

ENVIRONMENT="$1"
IMAGE_TAG="${2:-${IMAGE_TAG:-}}"

# Validate environment
case "$ENVIRONMENT" in
    staging|production) ;;
    *)
        log_error "Invalid environment: $ENVIRONMENT"
        log_error "Must be one of: staging, production"
        exit $EXIT_PREREQUISITE_FAILURE
        ;;
esac

# Check if IMAGE_TAG is set
if [ -z "$IMAGE_TAG" ]; then
    log_error "IMAGE_TAG not provided"
    log_error "Either pass as second argument or set IMAGE_TAG environment variable"
    log_error "Example: IMAGE_TAG=v1.0.0 $0 $ENVIRONMENT"
    exit $EXIT_PREREQUISITE_FAILURE
fi

export IMAGE_TAG

log_info "Deploying to ${ENVIRONMENT} with image tag: ${IMAGE_TAG}"

# ============================================================================
# Route to Environment-Specific Script
# ============================================================================
case "$ENVIRONMENT" in
    staging)
        if [ -x "$SCRIPT_DIR/deploy-stage.sh" ]; then
            exec "$SCRIPT_DIR/deploy-stage.sh" --tag "$IMAGE_TAG"
        else
            log_error "deploy-stage.sh not found or not executable"
            exit $EXIT_EXECUTION_ERROR
        fi
        ;;
    production)
        if [ -x "$SCRIPT_DIR/deploy-prod.sh" ]; then
            exec "$SCRIPT_DIR/deploy-prod.sh" --tag "$IMAGE_TAG"
        else
            log_error "deploy-prod.sh not found or not executable"
            exit $EXIT_EXECUTION_ERROR
        fi
        ;;
esac
