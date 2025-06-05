#!/bin/bash

# Wind Test Suite
# This script runs comprehensive tests for the Wind CLI tool

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "██╗    ██╗██╗███╗   ██╗██████╗     ████████╗███████╗███████╗████████╗"
echo "██║    ██║██║████╗  ██║██╔══██╗    ╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝"
echo "██║ █╗ ██║██║██╔██╗ ██║██║  ██║       ██║   █████╗  ███████╗   ██║   "
echo "██║███╗██║██║██║╚██╗██║██║  ██║       ██║   ██╔══╝  ╚════██║   ██║   "
echo "╚███╔███╔╝██║██║ ╚████║██████╔╝       ██║   ███████╗███████║   ██║   "
echo " ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═════╝        ╚═╝   ╚══════╝╚══════╝   ╚═╝   "
echo -e "${NC}"
echo

# Function to print section headers
print_section() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

print_success "Go version: $(go version)"
echo

# Parse command line arguments
RUN_UNIT=true
RUN_INTEGRATION=true
RUN_BENCHMARKS=false
RUN_COVERAGE=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --unit-only)
            RUN_INTEGRATION=false
            RUN_BENCHMARKS=false
            shift
            ;;
        --integration-only)
            RUN_UNIT=false
            RUN_BENCHMARKS=false
            shift
            ;;
        --benchmarks)
            RUN_BENCHMARKS=true
            shift
            ;;
        --coverage)
            RUN_COVERAGE=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --all)
            RUN_UNIT=true
            RUN_INTEGRATION=true
            RUN_BENCHMARKS=true
            RUN_COVERAGE=true
            shift
            ;;
        --help)
            echo "Wind Test Suite"
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --unit-only         Run only unit tests"
            echo "  --integration-only  Run only integration tests"
            echo "  --benchmarks        Run benchmark tests"
            echo "  --coverage          Generate test coverage report"
            echo "  --verbose           Enable verbose output"
            echo "  --all              Run all tests and benchmarks"
            echo "  --help             Show this help message"
            echo ""
            echo "Default: Run unit and integration tests"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Set verbose flag for go test
VERBOSE_FLAG=""
if [ "$VERBOSE" = true ]; then
    VERBOSE_FLAG="-v"
fi

# Ensure we're in the correct directory
if [ ! -f "main.go" ]; then
    print_error "main.go not found. Please run this script from the Wind project root."
    exit 1
fi

# Create test output directory
mkdir -p test-results

# Run unit tests
if [ "$RUN_UNIT" = true ]; then
    print_section "Running Unit Tests"
    
    if go test $VERBOSE_FLAG -run "^Test" -short ./... 2>&1 | tee test-results/unit-tests.log; then
        print_success "Unit tests passed"
    else
        print_error "Unit tests failed"
        exit 1
    fi
    echo
fi

# Run integration tests
if [ "$RUN_INTEGRATION" = true ]; then
    print_section "Running Integration Tests"
    print_warning "Integration tests may take longer as they test real file operations"
    
    if go test $VERBOSE_FLAG -run "^TestIntegration" ./... > test-results/integration-tests.log 2>&1; then
        print_success "Integration tests passed"
        cat test-results/integration-tests.log
    else
        print_error "Integration tests failed"
        cat test-results/integration-tests.log
        exit 1
    fi
    echo
fi

# Run benchmark tests
if [ "$RUN_BENCHMARKS" = true ]; then
    print_section "Running Benchmark Tests"
    print_warning "Benchmarks may take several minutes to complete"
    
    echo "Running file scanning benchmarks..."
    go test -bench="^BenchmarkScanFiles" -benchmem ./... 2>&1 | tee test-results/bench-scan.log
    
    echo "Running change detection benchmarks..."
    go test -bench="^BenchmarkCheckForChanges" -benchmem ./... 2>&1 | tee test-results/bench-changes.log
    
    echo "Running project detection benchmarks..."
    go test -bench="^BenchmarkDetectProjectStructure" -benchmem ./... 2>&1 | tee test-results/bench-detection.log
    
    echo "Running file extension benchmarks..."
    go test -bench="^BenchmarkShouldWatch" -benchmem ./... 2>&1 | tee test-results/bench-extensions.log
    
    echo "Running complete workflow benchmarks..."
    go test -bench="^BenchmarkCompleteWorkflow" -benchmem ./... 2>&1 | tee test-results/bench-workflow.log
    
    print_success "Benchmark tests completed"
    echo
fi

# Generate test coverage
if [ "$RUN_COVERAGE" = true ]; then
    print_section "Generating Test Coverage Report"
    
    # Run tests with coverage
    go test -coverprofile=test-results/coverage.out ./... > /dev/null 2>&1
    
    # Generate HTML coverage report
    go tool cover -html=test-results/coverage.out -o test-results/coverage.html
    
    # Generate text coverage report
    go tool cover -func=test-results/coverage.out | tee test-results/coverage.txt
    
    # Get coverage percentage
    COVERAGE=$(go tool cover -func=test-results/coverage.out | grep total: | awk '{print $3}')
    
    print_success "Coverage report generated: $COVERAGE"
    print_success "HTML report: test-results/coverage.html"
    print_success "Text report: test-results/coverage.txt"
    echo
fi

# Test build process
print_section "Testing Build Process"

if go build -o test-results/wind-test ./...; then
    print_success "Build successful"
    
    # Test basic commands
    if ./test-results/wind-test version > /dev/null 2>&1; then
        print_success "Version command works"
    else
        print_warning "Version command failed"
    fi
    
    if ./test-results/wind-test help > /dev/null 2>&1; then
        print_success "Help command works"
    else
        print_warning "Help command failed"
    fi
    
    # Clean up test binary
    rm -f test-results/wind-test
else
    print_error "Build failed"
    exit 1
fi

echo

# Test summary
print_section "Test Summary"

if [ "$RUN_UNIT" = true ]; then
    print_success "Unit tests: PASSED"
fi

if [ "$RUN_INTEGRATION" = true ]; then
    print_success "Integration tests: PASSED"
fi

if [ "$RUN_BENCHMARKS" = true ]; then
    print_success "Benchmark tests: COMPLETED"
fi

if [ "$RUN_COVERAGE" = true ]; then
    print_success "Coverage report: GENERATED ($COVERAGE)"
fi

print_success "Build test: PASSED"

echo
print_success "All tests completed successfully!"
print_success "Test results saved in: test-results/"

# Show test result files
echo
echo "Generated files:"
ls -la test-results/ | grep -v "^total" | tail -n +2 | while read line; do
    filename=$(echo $line | awk '{print $9}')
    if [ "$filename" != "." ] && [ "$filename" != ".." ]; then
        echo "  - test-results/$filename"
    fi
done 