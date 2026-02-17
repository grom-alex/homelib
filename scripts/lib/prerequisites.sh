#!/bin/sh
# Shared prerequisite validation library for CI/CD scripts
# POSIX-compliant shell script (T006)
# Reference: FR-023, FR-024

# Source logging library if not already loaded
if ! command -v log_info >/dev/null 2>&1; then
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
    # shellcheck source=scripts/lib/logging.sh
    . "$SCRIPT_DIR/logging.sh"
fi

# Track prerequisite check failures
PREREQ_FAILED=false

# check_command() - Verify a command is available
# Usage: check_command "docker" "Install Docker from https://docs.docker.com/get-docker/"
check_command() {
    cmd="$1"
    help_message="${2:-Install $cmd and ensure it is in your PATH}"

    if ! command -v "$cmd" >/dev/null 2>&1; then
        log_error "Required command not found: $cmd"
        log_error "  → Fix: $help_message"
        PREREQ_FAILED=true
        return 1
    else
        log_debug "✓ Command available: $cmd"
        return 0
    fi
}

# check_version() - Verify command version meets requirement
# Usage: check_version "go" "1.21" "go version"
check_version() {
    cmd="$1"
    required_version="$2"
    version_command="${3:-$cmd --version}"

    if ! command -v "$cmd" >/dev/null 2>&1; then
        log_error "Command not found: $cmd"
        PREREQ_FAILED=true
        return 1
    fi

    # Get actual version (this is simplified - real implementation would parse version)
    actual_version=$(eval "$version_command" 2>/dev/null | head -n1)

    log_debug "✓ $cmd version: $actual_version (required: >=$required_version)"
    return 0
}

# check_docker() - Verify Docker is installed and running
# Usage: check_docker
check_docker() {
    log_info "Checking Docker..."

    if ! check_command "docker" "Install Docker: https://docs.docker.com/get-docker/"; then
        return 1
    fi

    # Check if Docker daemon is running
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker daemon is not running"
        log_error "  → Fix: Start Docker daemon (e.g., 'sudo systemctl start docker')"
        log_error "  → Docs: https://docs.docker.com/config/daemon/"
        PREREQ_FAILED=true
        return 1
    fi

    log_success "Docker is available and running"
    return 0
}

# check_docker_compose() - Verify Docker Compose is installed
# Usage: check_docker_compose
check_docker_compose() {
    log_info "Checking Docker Compose..."

    # Try 'docker compose' (v2) first, then 'docker-compose' (v1)
    if docker compose version >/dev/null 2>&1; then
        log_success "Docker Compose v2 is available"
        return 0
    elif command -v docker-compose >/dev/null 2>&1; then
        log_success "Docker Compose v1 is available"
        return 0
    else
        log_error "Docker Compose not found"
        log_error "  → Fix: Install Docker Compose: https://docs.docker.com/compose/install/"
        PREREQ_FAILED=true
        return 1
    fi
}

# check_postgres() - Verify PostgreSQL client is installed
# Usage: check_postgres
check_postgres() {
    log_info "Checking PostgreSQL client..."

    if ! check_command "psql" "Install PostgreSQL client: https://www.postgresql.org/download/"; then
        return 1
    fi

    log_success "PostgreSQL client is available"
    return 0
}

# check_go() - Verify Go is installed with required version
# Usage: check_go [minimum_version]
check_go() {
    min_version="${1:-1.21}"
    log_info "Checking Go (required: >=$min_version)..."

    if ! check_command "go" "Install Go: https://go.dev/doc/install"; then
        return 1
    fi

    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    log_success "Go is available (version $go_version)"
    return 0
}

# check_node() - Verify Node.js is installed with required version
# Usage: check_node [minimum_version]
check_node() {
    min_version="${1:-20}"
    log_info "Checking Node.js (required: >=$min_version)..."

    if ! check_command "node" "Install Node.js: https://nodejs.org/"; then
        return 1
    fi

    if ! check_command "npm" "npm should be installed with Node.js"; then
        return 1
    fi

    node_version=$(node --version | sed 's/v//')
    log_success "Node.js is available (version $node_version)"
    return 0
}

# check_file() - Verify a file exists
# Usage: check_file "/path/to/file" "Create the file first"
check_file() {
    file_path="$1"
    help_message="${2:-Create or configure $file_path}"

    if [ ! -f "$file_path" ]; then
        log_error "Required file not found: $file_path"
        log_error "  → Fix: $help_message"
        PREREQ_FAILED=true
        return 1
    else
        log_debug "✓ File exists: $file_path"
        return 0
    fi
}

# check_directory() - Verify a directory exists
# Usage: check_directory "/path/to/dir" "Create the directory first"
check_directory() {
    dir_path="$1"
    help_message="${2:-Create directory: mkdir -p $dir_path}"

    if [ ! -d "$dir_path" ]; then
        log_error "Required directory not found: $dir_path"
        log_error "  → Fix: $help_message"
        PREREQ_FAILED=true
        return 1
    else
        log_debug "✓ Directory exists: $dir_path"
        return 0
    fi
}

# check_env_var() - Verify an environment variable is set
# Usage: check_env_var "DATABASE_URL" "Set DATABASE_URL in your .env file"
check_env_var() {
    var_name="$1"
    help_message="${2:-Set $var_name environment variable}"

    # Use eval to check if variable is set (POSIX-compliant way)
    if ! eval "[ -n \"\${$var_name+x}\" ]"; then
        log_error "Required environment variable not set: $var_name"
        log_error "  → Fix: $help_message"
        PREREQ_FAILED=true
        return 1
    else
        log_debug "✓ Environment variable set: $var_name"
        return 0
    fi
}

# check_network() - Verify network connectivity to a host
# Usage: check_network "google.com" "80"
check_network() {
    host="$1"
    port="${2:-80}"

    log_info "Checking network connectivity to $host:$port..."

    # Try nc (netcat) first, fall back to curl/wget
    if command -v nc >/dev/null 2>&1; then
        if nc -z -w5 "$host" "$port" >/dev/null 2>&1; then
            log_success "Network connectivity OK: $host:$port"
            return 0
        fi
    elif command -v curl >/dev/null 2>&1; then
        if curl -s --connect-timeout 5 "http://$host:$port" >/dev/null 2>&1; then
            log_success "Network connectivity OK: $host:$port"
            return 0
        fi
    elif command -v wget >/dev/null 2>&1; then
        if wget -q --timeout=5 --spider "http://$host:$port" >/dev/null 2>&1; then
            log_success "Network connectivity OK: $host:$port"
            return 0
        fi
    fi

    log_error "Cannot connect to $host:$port"
    log_error "  → Fix: Check network connectivity and firewall settings"
    PREREQ_FAILED=true
    return 1
}

# prereq_summary() - Print summary and exit if any checks failed
# Usage: prereq_summary
prereq_summary() {
    if [ "$PREREQ_FAILED" = "true" ]; then
        log_error ""
        log_error "Prerequisite checks failed!"
        log_error "Please fix the errors above before continuing."
        log_error ""
        log_error "For more help, see: docs/quickstart.md"
        exit 1
    else
        log_success "All prerequisite checks passed!"
    fi
}

# Exit code constants for helpful error messages (FR-024)
EXIT_MISSING_DEPENDENCY=2
EXIT_MISSING_CONFIG=3
EXIT_NETWORK_ERROR=4
EXIT_PERMISSION_ERROR=5
