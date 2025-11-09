#!/bin/bash

# Integration Tests Runner Script
# This script sets up the test environment, runs integration tests, and cleans up

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.test.yml"
LOGS_DIR="${PROJECT_ROOT}/tests/logs"
TEST_TIMEOUT=${TEST_TIMEOUT:-600}  # 10 minutes default timeout

# Flags
SKIP_BUILD=false
KEEP_RUNNING=false
VERBOSE=false
SHORT_MODE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --keep-running)
            KEEP_RUNNING=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --short)
            SHORT_MODE=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --skip-build     Skip building Docker images"
            echo "  --keep-running   Keep containers running after tests"
            echo "  --verbose, -v    Show verbose output"
            echo "  --short          Run only short tests (skip load/timeout tests)"
            echo "  --help, -h       Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Function to print colored messages
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to cleanup
cleanup() {
    local exit_code=$?

    if [ "$KEEP_RUNNING" = false ]; then
        log_info "Cleaning up test environment..."

        # Save logs if tests failed
        if [ $exit_code -ne 0 ]; then
            log_info "Collecting logs from failed test run..."
            mkdir -p "$LOGS_DIR"

            docker-compose -f "$COMPOSE_FILE" logs --no-color > "$LOGS_DIR/all-services.log" 2>&1 || true
            docker-compose -f "$COMPOSE_FILE" logs --no-color user-service-test > "$LOGS_DIR/user-service.log" 2>&1 || true
            docker-compose -f "$COMPOSE_FILE" logs --no-color order-service-test > "$LOGS_DIR/order-service.log" 2>&1 || true
            docker-compose -f "$COMPOSE_FILE" logs --no-color notification-service-test > "$LOGS_DIR/notification-service.log" 2>&1 || true
            docker-compose -f "$COMPOSE_FILE" logs --no-color api-gateway-test > "$LOGS_DIR/api-gateway.log" 2>&1 || true

            log_warning "Logs saved to: $LOGS_DIR"
        fi

        # Stop and remove containers
        docker-compose -f "$COMPOSE_FILE" down -v --remove-orphans > /dev/null 2>&1 || true
        log_success "Cleanup complete"
    else
        log_warning "Keeping containers running (--keep-running flag set)"
        log_info "To stop manually, run: docker-compose -f $COMPOSE_FILE down -v"
    fi

    exit $exit_code
}

# Register cleanup function
trap cleanup EXIT INT TERM

# Function to wait for service health
wait_for_service() {
    local service_name=$1
    local max_attempts=60
    local attempt=1

    log_info "Waiting for $service_name to be healthy..."

    while [ $attempt -le $max_attempts ]; do
        if docker-compose -f "$COMPOSE_FILE" ps | grep "$service_name" | grep -q "healthy\|Up"; then
            log_success "$service_name is ready"
            return 0
        fi

        if [ $VERBOSE = true ]; then
            echo -n "."
        fi

        sleep 2
        attempt=$((attempt + 1))
    done

    log_error "$service_name did not become healthy in time"
    docker-compose -f "$COMPOSE_FILE" logs "$service_name"
    return 1
}

# Main execution starts here
log_info "Starting integration tests..."
log_info "Project root: $PROJECT_ROOT"

# Navigate to project root
cd "$PROJECT_ROOT"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    log_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    log_error "docker-compose is not installed. Please install it and try again."
    exit 1
fi

# Stop any existing test containers
log_info "Stopping any existing test containers..."
docker-compose -f "$COMPOSE_FILE" down -v --remove-orphans > /dev/null 2>&1 || true

# Build services if needed
if [ "$SKIP_BUILD" = false ]; then
    log_info "Building Docker images..."
    if [ $VERBOSE = true ]; then
        docker-compose -f "$COMPOSE_FILE" build
    else
        docker-compose -f "$COMPOSE_FILE" build > /dev/null 2>&1
    fi
    log_success "Images built successfully"
else
    log_warning "Skipping build (--skip-build flag set)"
fi

# Start services
log_info "Starting test environment..."
if [ $VERBOSE = true ]; then
    docker-compose -f "$COMPOSE_FILE" up -d
else
    docker-compose -f "$COMPOSE_FILE" up -d > /dev/null 2>&1
fi

# Wait for all services to be healthy
log_info "Waiting for all services to be healthy (this may take a minute)..."

wait_for_service "postgres-test"
wait_for_service "mongodb-test"
wait_for_service "rabbitmq-test"
wait_for_service "mailhog"
wait_for_service "user-service-test"
wait_for_service "order-service-test"
wait_for_service "notification-service-test"
wait_for_service "api-gateway-test"

# Give services extra time to fully initialize
log_info "Allowing services additional time to initialize connections..."
sleep 5

log_success "All services are healthy and ready"

# Show service status
if [ $VERBOSE = true ]; then
    log_info "Service status:"
    docker-compose -f "$COMPOSE_FILE" ps
fi

# Run integration tests
log_info "Running integration tests..."

# Set test flags
TEST_FLAGS="-v"
if [ $SHORT_MODE = true ]; then
    TEST_FLAGS="$TEST_FLAGS -short"
fi

# Navigate to tests directory
cd "${PROJECT_ROOT}/tests/integration"

# Run tests with timeout
if [ $VERBOSE = true ]; then
    log_info "Running: go test $TEST_FLAGS -timeout ${TEST_TIMEOUT}s ./..."
    timeout ${TEST_TIMEOUT}s go test $TEST_FLAGS -timeout ${TEST_TIMEOUT}s ./...
    TEST_EXIT_CODE=$?
else
    timeout ${TEST_TIMEOUT}s go test $TEST_FLAGS -timeout ${TEST_TIMEOUT}s ./... 2>&1 | tee "${LOGS_DIR}/test-output.log"
    TEST_EXIT_CODE=${PIPESTATUS[0]}
fi

# Check test results
if [ $TEST_EXIT_CODE -eq 0 ]; then
    log_success "All integration tests passed! ✓"

    # Show test summary
    echo ""
    echo "=================================="
    echo "  Integration Tests Summary"
    echo "=================================="
    echo "Status: PASSED"
    echo "Services tested:"
    echo "  - User Service"
    echo "  - Order Service"
    echo "  - Notification Service"
    echo "  - API Gateway"
    echo "=================================="
    echo ""

    exit 0
else
    log_error "Integration tests failed! ✗"

    # Show failed test information
    echo ""
    echo "=================================="
    echo "  Integration Tests Failed"
    echo "=================================="
    echo "Exit code: $TEST_EXIT_CODE"
    echo "Logs available at: $LOGS_DIR"
    echo ""
    echo "To debug, you can:"
    echo "  1. Check logs in: $LOGS_DIR"
    echo "  2. Run with --verbose flag for more details"
    echo "  3. Run with --keep-running to inspect services"
    echo "=================================="
    echo ""

    exit $TEST_EXIT_CODE
fi
