#!/bin/sh
# Shared logging library for CI/CD scripts
# POSIX-compliant shell script (T005, T008)
# Reference: FR-023, FR-024

# ANSI color codes for terminal output
# shellcheck disable=SC2034  # Variables used by sourcing scripts
COLOR_RESET='\033[0m'
COLOR_RED='\033[0;31m'
COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[0;33m'
COLOR_BLUE='\033[0;34m'
COLOR_GRAY='\033[0;90m'

# Log level constants
LOG_LEVEL_DEBUG=0
LOG_LEVEL_INFO=1
LOG_LEVEL_WARN=2
LOG_LEVEL_ERROR=3

# Default log level (can be overridden by LOG_LEVEL environment variable)
CURRENT_LOG_LEVEL="${LOG_LEVEL:-${LOG_LEVEL_INFO}}"

# Enable/disable colored output (auto-detect if terminal supports colors)
if [ -t 1 ] && [ -t 2 ]; then
    USE_COLORS=true
else
    USE_COLORS=false
fi

# Helper function to format timestamp
_log_timestamp() {
    date '+%Y-%m-%d %H:%M:%S'
}

# Helper function to apply color if enabled
_log_color() {
    color="$1"
    message="$2"

    if [ "$USE_COLORS" = "true" ]; then
        printf "%b%s%b\n" "$color" "$message" "$COLOR_RESET"
    else
        printf "%s\n" "$message"
    fi
}

# log_debug() - Output debug messages (lowest priority)
# Usage: log_debug "Detailed diagnostic information"
log_debug() {
    if [ "$CURRENT_LOG_LEVEL" -le "$LOG_LEVEL_DEBUG" ]; then
        timestamp=$(_log_timestamp)
        message="$*"
        _log_color "$COLOR_GRAY" "[DEBUG] [$timestamp] $message" >&2
    fi
}

# log_info() - Output informational messages
# Usage: log_info "Starting deployment process"
log_info() {
    if [ "$CURRENT_LOG_LEVEL" -le "$LOG_LEVEL_INFO" ]; then
        timestamp=$(_log_timestamp)
        message="$*"
        _log_color "$COLOR_BLUE" "[INFO]  [$timestamp] $message" >&2
    fi
}

# log_success() - Output success messages
# Usage: log_success "Deployment completed successfully"
log_success() {
    if [ "$CURRENT_LOG_LEVEL" -le "$LOG_LEVEL_INFO" ]; then
        timestamp=$(_log_timestamp)
        message="$*"
        _log_color "$COLOR_GREEN" "[OK]    [$timestamp] $message" >&2
    fi
}

# log_warn() - Output warning messages
# Usage: log_warn "Configuration file not found, using defaults"
log_warn() {
    if [ "$CURRENT_LOG_LEVEL" -le "$LOG_LEVEL_WARN" ]; then
        timestamp=$(_log_timestamp)
        message="$*"
        _log_color "$COLOR_YELLOW" "[WARN]  [$timestamp] $message" >&2
    fi
}

# log_error() - Output error messages
# Usage: log_error "Failed to connect to database"
log_error() {
    if [ "$CURRENT_LOG_LEVEL" -le "$LOG_LEVEL_ERROR" ]; then
        timestamp=$(_log_timestamp)
        message="$*"
        _log_color "$COLOR_RED" "[ERROR] [$timestamp] $message" >&2
    fi
}

# log_fatal() - Output fatal error and exit
# Usage: log_fatal "Critical failure: cannot continue"
log_fatal() {
    timestamp=$(_log_timestamp)
    message="$*"
    _log_color "$COLOR_RED" "[FATAL] [$timestamp] $message" >&2
    exit 1
}

# log_section() - Output a section header for better readability
# Usage: log_section "Running Unit Tests"
log_section() {
    message="$*"
    separator="=================================================="
    printf "\n"
    _log_color "$COLOR_BLUE" "$separator"
    _log_color "$COLOR_BLUE" "  $message"
    _log_color "$COLOR_BLUE" "$separator"
}

# Disable colored output (useful for CI logs)
log_disable_colors() {
    USE_COLORS=false
}

# Enable colored output
log_enable_colors() {
    USE_COLORS=true
}

# Set log level
# Usage: log_set_level "$LOG_LEVEL_DEBUG"
log_set_level() {
    CURRENT_LOG_LEVEL="$1"
}
