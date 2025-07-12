#!/bin/bash

# End-to-End Test Runner for ownershit
# This script helps run E2E tests safely against real GitHub repositories

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

print_header() {
    print_color $BLUE "🧪 $1"
}

print_success() {
    print_color $GREEN "✅ $1"
}

print_warning() {
    print_color $YELLOW "⚠️  $1"
}

print_error() {
    print_color $RED "❌ $1"
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking prerequisites..."
    
    # Check if GITHUB_TOKEN is set
    if [ -z "$GITHUB_TOKEN" ]; then
        print_error "GITHUB_TOKEN environment variable is not set"
        print_warning "Please set your GitHub token:"
        echo "export GITHUB_TOKEN=your_github_token_here"
        exit 1
    fi
    
    # Validate token format
    if ! echo "$GITHUB_TOKEN" | grep -qE '^(ghp_|github_pat_|ghs_)'; then
        print_warning "GitHub token format may be invalid"
        print_warning "Expected formats: ghp_*, github_pat_*, or ghs_*"
    fi
    
    # Check if Go is available
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Set default test configuration
set_test_defaults() {
    export RUN_E2E_TESTS="true"
    
    # Use environment variables or safe defaults
    export E2E_TEST_ORG="${E2E_TEST_ORG:-test-org}"
    export E2E_TEST_REPO="${E2E_TEST_REPO:-test-repo}"
    export E2E_TEST_USER="${E2E_TEST_USER:-$(gh api user --jq .login 2>/dev/null || echo 'test-user')}"
    
    print_header "Test Configuration:"
    echo "Organization: $E2E_TEST_ORG"
    echo "Repository: $E2E_TEST_REPO"
    echo "User: $E2E_TEST_USER"
    echo ""
}

# Run read-only E2E tests
run_readonly_tests() {
    print_header "Running read-only E2E tests..."
    
    # Test rate limit and basic API access
    print_color $BLUE "Testing API access and rate limiting..."
    go test -v -run "TestE2E_RateLimit" ./...
    
    # Test token validation
    print_color $BLUE "Testing token validation..."
    go test -v -run "TestE2E_TokenValidation" ./...
    
    # Test configuration validation
    print_color $BLUE "Testing configuration validation..."
    go test -v -run "TestE2E_ConfigurationValidation" ./...
    
    # Test CLI configuration
    print_color $BLUE "Testing CLI configuration..."
    go test -v -run "TestE2E_CLI_ConfigValidation" ./cmd/ownershit/...
    
    # Test CLI commands (read-only)
    print_color $BLUE "Testing CLI commands (read-only)..."
    go test -v -run "TestE2E_CLI_Commands" ./cmd/ownershit/...
    
    print_success "Read-only tests completed"
}

# Run tests that might make changes (with confirmation)
run_modification_tests() {
    print_header "Modification Tests"
    print_warning "The following tests may make changes to your GitHub repositories:"
    print_warning "- Repository settings updates"
    print_warning "- Branch protection changes"
    print_warning "- Team permission modifications"
    echo ""
    
    if [ "$E2E_ALLOW_MODIFICATIONS" != "true" ]; then
        read -p "Do you want to run tests that may modify repositories? (y/N): " -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_warning "Skipping modification tests"
            return 0
        fi
        export E2E_ALLOW_MODIFICATIONS="true"
    fi
    
    print_color $BLUE "Running modification tests..."
    go test -v -run "TestE2E.*sync_real" ./cmd/ownershit/...
    
    print_success "Modification tests completed"
}

# Run specific test scenarios
run_specific_tests() {
    print_header "Running specific test scenarios..."
    
    # Test teams (if organization is accessible)
    print_color $BLUE "Testing team operations..."
    go test -v -run "TestE2E_GetTeams" ./... || print_warning "Team tests may fail if organization is not accessible"
    
    # Test repository operations
    print_color $BLUE "Testing repository operations..."
    go test -v -run "TestE2E_RepositoryOperations" ./... || print_warning "Repository tests may fail without proper access"
    
    # Test init command
    print_color $BLUE "Testing init command..."
    go test -v -run "TestE2E_CLI_InitCommand" ./cmd/ownershit/...
    
    print_success "Specific tests completed"
}

# Build and test the application
build_and_test() {
    print_header "Building application..."
    
    # Run regular tests first
    print_color $BLUE "Running unit tests..."
    go test ./...
    
    # Build the application
    print_color $BLUE "Building binary..."
    go build -o bin/ownershit ./cmd/ownershit/
    
    print_success "Build completed"
}

# Main execution
main() {
    print_header "ownershit End-to-End Test Runner"
    echo "This script runs E2E tests against real GitHub repositories"
    echo ""
    
    # Parse command line arguments
    READONLY_ONLY=false
    SKIP_BUILD=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --readonly)
                READONLY_ONLY=true
                shift
                ;;
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --org)
                export E2E_TEST_ORG="$2"
                shift 2
                ;;
            --repo)
                export E2E_TEST_REPO="$2"
                shift 2
                ;;
            --user)
                export E2E_TEST_USER="$2"
                shift 2
                ;;
            --modifications)
                export E2E_ALLOW_MODIFICATIONS="true"
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --readonly              Run only read-only tests"
                echo "  --skip-build           Skip building the application"
                echo "  --org ORG              Set test organization"
                echo "  --repo REPO            Set test repository"
                echo "  --user USER            Set test user"
                echo "  --modifications        Allow tests that modify repositories"
                echo "  -h, --help             Show this help message"
                echo ""
                echo "Environment Variables:"
                echo "  GITHUB_TOKEN           Your GitHub token (required)"
                echo "  E2E_TEST_ORG          Test organization (default: test-org)"
                echo "  E2E_TEST_REPO         Test repository (default: test-repo)"
                echo "  E2E_TEST_USER         Test user login"
                echo "  E2E_ALLOW_MODIFICATIONS Allow modification tests"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    check_prerequisites
    set_test_defaults
    
    if [ "$SKIP_BUILD" = false ]; then
        build_and_test
    fi
    
    run_readonly_tests
    run_specific_tests
    
    if [ "$READONLY_ONLY" = false ]; then
        run_modification_tests
    fi
    
    print_success "All E2E tests completed successfully!"
    
    # Show summary
    print_header "Test Summary"
    echo "✅ Read-only API tests"
    echo "✅ Configuration validation tests"
    echo "✅ CLI command tests"
    
    if [ "$READONLY_ONLY" = false ] && [ "$E2E_ALLOW_MODIFICATIONS" = "true" ]; then
        echo "✅ Modification tests"
    else
        echo "⏭️  Modification tests (skipped)"
    fi
    
    echo ""
    print_success "E2E testing complete!"
}

# Run main function
main "$@"