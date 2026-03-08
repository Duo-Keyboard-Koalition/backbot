#!/bin/bash
# Tailbridge Test Platform - Unix Test Runner
# Run all tests for the Tailbridge test platform

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_PLATFORM_ROOT="$(dirname "$SCRIPT_DIR")"

echo "========================================"
echo "Tailbridge Test Platform"
echo "========================================"
echo ""

# Track test results
MOCK_TESTS_PASSED=true
INTEGRATION_TESTS_PASSED=true

# Function to run mock tests
run_mock_tests() {
    echo "Running Mock Tests..."
    echo "----------------------"
    
    cd "$TEST_PLATFORM_ROOT"
    
    if go test ./mock/... -v -coverprofile=coverage.out; then
        echo "Mock tests PASSED"
        
        # Show coverage
        go tool cover -func=coverage.out | tail -1
    else
        MOCK_TESTS_PASSED=false
        echo "Mock tests FAILED"
    fi
    
    echo ""
}

# Function to run Docker integration tests
run_integration_tests() {
    echo "Running Integration Tests..."
    echo "-----------------------------"
    
    DOCKER_COMPOSE_FILE="$TEST_PLATFORM_ROOT/docker/docker-compose.test.yml"
    
    # Check if .env file exists
    ENV_FILE="$TEST_PLATFORM_ROOT/docker/.env"
    if [ -f "$ENV_FILE" ]; then
        echo "Loading environment from .env file"
    else
        echo "WARNING: No .env file found. Integration tests may fail without TS_AUTHKEY values."
        echo "Create docker/.env with TS_AUTH_KEY_1, TS_AUTH_KEY_2, TS_AUTH_KEY_3"
        echo ""
    fi
    
    # Start Docker containers
    echo "Starting Docker containers..."
    if ! docker-compose -f "$DOCKER_COMPOSE_FILE" up -d; then
        echo "Failed to start Docker containers"
        INTEGRATION_TESTS_PASSED=false
        return
    fi
    
    echo "Waiting for agents to be ready..."
    sleep 30
    
    # Run integration tests
    cd "$TEST_PLATFORM_ROOT"
    if go test ./integration/... -v -tags=integration -timeout=10m; then
        echo "Integration tests PASSED"
    else
        INTEGRATION_TESTS_PASSED=false
        echo "Integration tests FAILED"
    fi
    
    # Cleanup
    echo ""
    echo "Cleaning up Docker containers..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" down
    
    echo ""
}

# Function to show test summary
show_summary() {
    echo "========================================"
    echo "Test Summary"
    echo "========================================"
    echo ""
    
    if [ "$MOCK_TESTS_PASSED" = true ]; then
        echo "[PASS] Mock Tests"
    else
        echo "[FAIL] Mock Tests"
    fi
    
    if [ "$INTEGRATION_TESTS_PASSED" = true ]; then
        echo "[PASS] Integration Tests"
    else
        echo "[FAIL] Integration Tests"
    fi
    
    echo ""
    
    if [ "$MOCK_TESTS_PASSED" = true ] && [ "$INTEGRATION_TESTS_PASSED" = true ]; then
        echo "All tests PASSED!"
        exit 0
    else
        echo "Some tests FAILED!"
        exit 1
    fi
}

# Parse command line arguments
RUN_MOCK_TESTS=true
RUN_INTEGRATION_TESTS=false
RUN_ALL_TESTS=false

for arg in "$@"; do
    case $arg in
        -mock)
            RUN_MOCK_TESTS=true
            RUN_INTEGRATION_TESTS=false
            shift
            ;;
        -integration)
            RUN_MOCK_TESTS=false
            RUN_INTEGRATION_TESTS=true
            shift
            ;;
        -all)
            RUN_ALL_TESTS=true
            RUN_MOCK_TESTS=true
            RUN_INTEGRATION_TESTS=true
            shift
            ;;
        -help|-h)
            echo "Usage: ./run-all-tests.sh [options]"
            echo ""
            echo "Options:"
            echo "  -mock         Run only mock tests (default)"
            echo "  -integration  Run only integration tests"
            echo "  -all          Run all tests"
            echo "  -help, -h     Show this help"
            echo ""
            exit 0
            ;;
    esac
done

# Run tests
if [ "$RUN_MOCK_TESTS" = true ]; then
    run_mock_tests
fi

if [ "$RUN_INTEGRATION_TESTS" = true ] || [ "$RUN_ALL_TESTS" = true ]; then
    run_integration_tests
fi

# Show summary
show_summary
