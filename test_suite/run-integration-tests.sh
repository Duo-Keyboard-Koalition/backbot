#!/usr/bin/env bash
# SentinelAI Integration Test Runner
# Runs all integration tests with real APIs (NO MOCKS)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
TEST_DIR="$SCRIPT_DIR/integration"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
TEST_MODE="all"
VERBOSITY="-v"
PARALLEL=""
COVERAGE=""
SKIP_MARKERS=""

usage() {
    cat << EOF
SentinelAI Integration Test Runner

Usage: $(basename "$0") [OPTIONS]

Options:
    -m, --mode MODE       Test mode: all, gemini, tailscale, darci, e2e (default: all)
    -q, --quiet           Quiet mode (minimal output)
    -vv, --verbose        Extra verbose output
    -n, --parallel N      Run tests in parallel with N workers
    -c, --coverage        Generate coverage report
    -s, --skip MARKER     Skip tests with marker (e.g., "slow", "api_cost")
    -k, --keyword EXPR    Run tests matching keyword expression
    -h, --help            Show this help message

Examples:
    $(basename "$0")                          # Run all tests
    $(basename "$0") -m gemini                # Run only Gemini tests
    $(basename "$0") -m tailscale -n auto     # Run Tailscale tests in parallel
    $(basename "$0") -s slow -s api_cost      # Skip slow and costly tests
    $(basename "$0") -k "test_agent"          # Run tests with "test_agent" in name

Test Markers:
    gemini      Tests requiring Gemini API
    tailscale   Tests requiring Tailscale connection
    e2e         End-to-end workflow tests
    slow        Slow-running tests (>30s)
    api_cost    Tests consuming API quota

EOF
    exit 1
}

log_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

log_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

log_error() {
    echo -e "${RED}✗ $1${NC}"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check Python
    if ! command -v python &> /dev/null; then
        log_error "Python not found. Please install Python 3.10+"
        exit 1
    fi
    
    # Check pytest
    if ! python -m pytest --version &> /dev/null; then
        log_error "pytest not found. Install with: pip install pytest pytest-asyncio pytest-cov pytest-timeout"
        exit 1
    fi
    
    # Check .env.test
    if [ ! -f "$PROJECT_ROOT/.env.test" ]; then
        log_warning ".env.test not found. Copy .env.test.example and configure API keys."
        log_info "Creating from example..."
        cp "$PROJECT_ROOT/.env.test.example" "$PROJECT_ROOT/.env.test"
    fi
    
    # Check Tailscale (for tailscale tests)
    if [[ "$TEST_MODE" == "tailscale" || "$TEST_MODE" == "all" ]]; then
        if command -v tailscale &> /dev/null; then
            if ! tailscale status &> /dev/null; then
                log_warning "Tailscale not connected. Tailscale tests may be skipped."
            else
                log_success "Tailscale connected"
            fi
        else
            log_warning "tailscale CLI not found. Tailscale tests may be skipped."
        fi
    fi
    
    log_success "Prerequisites check complete"
}

run_tests() {
    local pytest_args=("$VERBOSITY")
    
    # Add test directory
    pytest_args+=("$TEST_DIR")
    
    # Add mode-specific markers
    case "$TEST_MODE" in
        gemini)
            pytest_args+=("-m" "gemini")
            log_info "Running Gemini API tests..."
            ;;
        tailscale)
            pytest_args+=("-m" "tailscale")
            log_info "Running Tailscale tests..."
            ;;
        darci)
            pytest_args+=("-k" "darci")
            log_info "Running DarCI tests..."
            ;;
        e2e)
            pytest_args+=("-m" "e2e")
            log_info "Running E2E tests..."
            ;;
        all)
            log_info "Running all integration tests..."
            ;;
        *)
            log_error "Unknown mode: $TEST_MODE"
            exit 1
            ;;
    esac
    
    # Add parallel execution
    if [ -n "$PARALLEL" ]; then
        pytest_args+=("$PARALLEL")
        log_info "Parallel execution enabled"
    fi
    
    # Add coverage
    if [ -n "$COVERAGE" ]; then
        pytest_args+=("--cov=backend" "--cov=darci" "--cov-report=html" "--cov-report=term")
        log_info "Coverage report enabled"
    fi
    
    # Add skip markers
    if [ -n "$SKIP_MARKERS" ]; then
        for marker in $SKIP_MARKERS; do
            pytest_args+=("-m" "not $marker")
        done
    fi
    
    # Add keyword filter
    if [ -n "$KEYWORD" ]; then
        pytest_args+=("-k" "$KEYWORD")
    fi
    
    # Run tests
    log_info "Executing: pytest ${pytest_args[*]}"
    echo ""
    
    if python -m pytest "${pytest_args[@]}"; then
        log_success "All tests passed!"
    else
        log_error "Some tests failed"
        exit 1
    fi
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--mode)
            TEST_MODE="$2"
            shift 2
            ;;
        -q|--quiet)
            VERBOSITY="-q"
            shift
            ;;
        -vv|--verbose)
            VERBOSITY="-vv"
            shift
            ;;
        -n|--parallel)
            PARALLEL="-n $2"
            shift 2
            ;;
        -c|--coverage)
            COVERAGE="yes"
            shift
            ;;
        -s|--skip)
            SKIP_MARKERS="$SKIP_MARKERS $2"
            shift 2
            ;;
        -k|--keyword)
            KEYWORD="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Main execution
echo ""
echo "========================================"
echo "  SentinelAI Integration Test Runner   "
echo "========================================"
echo ""

check_prerequisites

echo ""
run_tests

echo ""
log_success "Test run complete!"
