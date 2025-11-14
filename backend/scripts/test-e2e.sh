#!/bin/bash

# @kthulu:core
# End-to-end testing script for Kthulu application

set -e

echo "🧪 Starting E2E tests for Kthulu..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
E2E_DIR="e2e"
BACKEND_DIR="backend"
FRONTEND_DIR="frontend"
TEST_RESULTS_DIR="test-results"
REPORTS_DIR="reports"

# Environment variables
export NODE_ENV=test
export CI=${CI:-false}
export BASE_URL=${BASE_URL:-http://localhost:8080}

# Function to print colored output
print_status() {
    echo -e "${BLUE}[E2E]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[E2E]${NC} ✅ $1"
}

print_error() {
    echo -e "${RED}[E2E]${NC} ❌ $1"
}

print_warning() {
    echo -e "${YELLOW}[E2E]${NC} ⚠️  $1"
}

# Function to cleanup processes
cleanup() {
    print_status "Cleaning up processes..."
    
    # Kill backend server if running
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
        wait $BACKEND_PID 2>/dev/null || true
    fi
    
    # Kill any remaining processes on port 8080
    lsof -ti:8080 | xargs kill -9 2>/dev/null || true
    
    print_status "Cleanup completed"
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Function to wait for server
wait_for_server() {
    local url=$1
    local max_attempts=30
    local attempt=1
    
    print_status "Waiting for server at $url..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url/health" > /dev/null 2>&1; then
            print_success "Server is ready!"
            return 0
        fi
        
        print_status "Attempt $attempt/$max_attempts - Server not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_error "Server failed to start within timeout"
    return 1
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Node.js is installed
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed"
        exit 1
    fi
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi
    
    # Check if npm is installed
    if ! command -v npm &> /dev/null; then
        print_error "npm is not installed"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to setup test environment
setup_test_environment() {
    print_status "Setting up test environment..."
    
    # Create test results directory
    mkdir -p "$TEST_RESULTS_DIR"
    mkdir -p "$REPORTS_DIR"
    
    # Install E2E test dependencies
    if [ ! -d "$E2E_DIR/node_modules" ]; then
        print_status "Installing E2E test dependencies..."
        cd "$E2E_DIR"
        npm install
        cd ..
    fi
    
    # Install Playwright browsers
    print_status "Installing Playwright browsers..."
    cd "$E2E_DIR"
    npx playwright install
    cd ..
    
    print_success "Test environment setup completed"
}

# Function to build application
build_application() {
    print_status "Building application..."
    
    # Build backend
    print_status "Building backend..."
    cd "$BACKEND_DIR"
    go build -o kthulu-app cmd/service/main.go
    cd ..
    
    # Build frontend
    print_status "Building frontend..."
    cd "$FRONTEND_DIR"
    npm run build
    cd ..
    
    print_success "Application build completed"
}

# Function to start backend server
start_backend_server() {
    print_status "Starting backend server..."
    
    cd "$BACKEND_DIR"
    
    # Set test environment variables
    export DB_DRIVER=sqlite
    export DB_URL=":memory:"
    export JWT_SECRET="test-jwt-secret-for-e2e-tests"
    export JWT_REFRESH_SECRET="test-jwt-refresh-secret-for-e2e-tests"
    export SERVER_PORT=8080
    export LOG_LEVEL=warn
    
    # Start server in background
    ./kthulu-app &
    BACKEND_PID=$!
    
    cd ..
    
    # Wait for server to be ready
    if ! wait_for_server "$BASE_URL"; then
        print_error "Failed to start backend server"
        exit 1
    fi
    
    print_success "Backend server started (PID: $BACKEND_PID)"
}

# Function to run backend integration tests
run_backend_integration_tests() {
    print_status "Running backend integration tests..."
    
    cd "$BACKEND_DIR"
    
    # Run integration tests
    go test -v -tags=integration ./internal/integration/... \
        -coverprofile="../$TEST_RESULTS_DIR/backend-integration-coverage.out" \
        -json > "../$TEST_RESULTS_DIR/backend-integration-results.json" 2>&1
    
    local exit_code=$?
    
    # Generate coverage report
    if [ -f "../$TEST_RESULTS_DIR/backend-integration-coverage.out" ]; then
        go tool cover -html="../$TEST_RESULTS_DIR/backend-integration-coverage.out" \
            -o "../$REPORTS_DIR/backend-integration-coverage.html"
    fi
    
    cd ..
    
    if [ $exit_code -eq 0 ]; then
        print_success "Backend integration tests passed"
    else
        print_error "Backend integration tests failed"
        return $exit_code
    fi
}

# Function to run Playwright E2E tests
run_playwright_tests() {
    print_status "Running Playwright E2E tests..."
    
    cd "$E2E_DIR"
    
    # Run Playwright tests
    local test_command="npx playwright test"
    
    # Add CI-specific options
    if [ "$CI" = "true" ]; then
        test_command="$test_command --reporter=html,json,junit"
    else
        test_command="$test_command --reporter=html"
    fi
    
    # Run tests
    $test_command
    local exit_code=$?
    
    # Move reports to main reports directory
    if [ -d "playwright-report" ]; then
        cp -r playwright-report "../$REPORTS_DIR/playwright-report"
    fi
    
    if [ -d "test-results" ]; then
        cp -r test-results/* "../$TEST_RESULTS_DIR/"
    fi
    
    cd ..
    
    if [ $exit_code -eq 0 ]; then
        print_success "Playwright E2E tests passed"
    else
        print_error "Playwright E2E tests failed"
        return $exit_code
    fi
}

# Function to run performance tests
run_performance_tests() {
    print_status "Running performance tests..."
    
    # Simple performance test using curl
    local response_time=$(curl -o /dev/null -s -w '%{time_total}' "$BASE_URL/health")
    local response_code=$(curl -o /dev/null -s -w '%{http_code}' "$BASE_URL/health")
    
    echo "Health endpoint response time: ${response_time}s"
    echo "Health endpoint response code: $response_code"
    
    # Check if response time is reasonable (< 1 second)
    if (( $(echo "$response_time < 1.0" | bc -l) )); then
        print_success "Performance test passed (${response_time}s)"
    else
        print_warning "Performance test warning: slow response (${response_time}s)"
    fi
    
    # Save performance metrics
    echo "{\"health_endpoint_response_time\": $response_time, \"health_endpoint_response_code\": $response_code}" \
        > "$TEST_RESULTS_DIR/performance-metrics.json"
}

# Function to generate test report
generate_test_report() {
    print_status "Generating test report..."
    
    local report_file="$REPORTS_DIR/e2e-test-report.html"
    
    cat > "$report_file" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Kthulu E2E Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .success { color: green; }
        .error { color: red; }
        .warning { color: orange; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Kthulu E2E Test Report</h1>
        <p>Generated on: $(date)</p>
        <p>Base URL: $BASE_URL</p>
    </div>
    
    <div class="section">
        <h2>Test Results</h2>
        <ul>
            <li><a href="backend-integration-coverage.html">Backend Integration Coverage</a></li>
            <li><a href="playwright-report/index.html">Playwright Test Report</a></li>
        </ul>
    </div>
    
    <div class="section">
        <h2>Performance Metrics</h2>
        <p>Health endpoint response time: ${response_time:-"N/A"}s</p>
    </div>
</body>
</html>
EOF
    
    print_success "Test report generated: $report_file"
}

# Main execution
main() {
    print_status "Starting Kthulu E2E test suite..."
    
    # Parse command line arguments
    local run_backend_tests=true
    local run_frontend_tests=true
    local run_perf_tests=true
    local build_app=true
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-backend)
                run_backend_tests=false
                shift
                ;;
            --no-frontend)
                run_frontend_tests=false
                shift
                ;;
            --no-performance)
                run_perf_tests=false
                shift
                ;;
            --no-build)
                build_app=false
                shift
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo "Options:"
                echo "  --no-backend      Skip backend integration tests"
                echo "  --no-frontend     Skip frontend E2E tests"
                echo "  --no-performance  Skip performance tests"
                echo "  --no-build        Skip application build"
                echo "  --help            Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Execute test pipeline
    check_prerequisites
    setup_test_environment
    
    if [ "$build_app" = true ]; then
        build_application
    fi
    
    start_backend_server
    
    local overall_exit_code=0
    
    if [ "$run_backend_tests" = true ]; then
        if ! run_backend_integration_tests; then
            overall_exit_code=1
        fi
    fi
    
    if [ "$run_frontend_tests" = true ]; then
        if ! run_playwright_tests; then
            overall_exit_code=1
        fi
    fi
    
    if [ "$run_perf_tests" = true ]; then
        run_performance_tests
    fi
    
    generate_test_report
    
    if [ $overall_exit_code -eq 0 ]; then
        print_success "All E2E tests completed successfully! 🎉"
        print_status "Reports available in: $REPORTS_DIR/"
    else
        print_error "Some E2E tests failed! 😞"
        print_status "Check reports in: $REPORTS_DIR/"
    fi
    
    exit $overall_exit_code
}

# Run main function with all arguments
main "$@"