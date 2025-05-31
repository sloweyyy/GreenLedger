#!/bin/bash

# 🌱 GreenLedger Development Setup Script
# This script sets up the local development environment

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check Go version
check_go_version() {
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go 1.19 or later."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.19"
    
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_error "Go version $GO_VERSION is too old. Please install Go $REQUIRED_VERSION or later."
        exit 1
    fi
    
    print_success "Go version $GO_VERSION is compatible"
}

# Function to check disk space
check_disk_space() {
    AVAILABLE_SPACE=$(df . | tail -1 | awk '{print $4}')
    REQUIRED_SPACE=1048576  # 1GB in KB
    
    if [ "$AVAILABLE_SPACE" -lt "$REQUIRED_SPACE" ]; then
        print_warning "Low disk space detected. Available: $(($AVAILABLE_SPACE / 1024))MB"
        print_warning "Please free up some disk space before continuing."
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Function to create directories
create_directories() {
    print_status "Creating necessary directories..."
    
    mkdir -p data
    mkdir -p logs
    mkdir -p tmp
    mkdir -p bin
    
    print_success "Directories created"
}

# Function to setup environment file
setup_environment() {
    print_status "Setting up environment configuration..."
    
    if [ ! -f .env ]; then
        if [ -f .env.example ]; then
            cp .env.example .env
            print_success "Created .env from .env.example"
            print_warning "Please review and update .env with your local settings"
        else
            print_warning ".env.example not found, creating basic .env"
            cat > .env << EOF
# Basic development environment
ENVIRONMENT=development
LOG_LEVEL=debug

# Database (SQLite for local development)
DB_TYPE=sqlite
DB_PATH=./data/

# Service ports
CALCULATOR_PORT=8081
TRACKER_PORT=8082
WALLET_PORT=8083
USER_AUTH_PORT=8084
REPORTING_PORT=8085

# Security
JWT_SECRET=local-development-secret-key-change-me

# Features
REDIS_ENABLED=false
KAFKA_ENABLED=false
EOF
            print_success "Created basic .env file"
        fi
    else
        print_success ".env file already exists"
    fi
}

# Function to install development tools
install_dev_tools() {
    print_status "Installing development tools..."
    
    # Install Air for live reloading
    if ! command_exists air; then
        print_status "Installing Air for live reloading..."
        go install github.com/cosmtrek/air@latest
        print_success "Air installed"
    else
        print_success "Air already installed"
    fi
    
    # Install golangci-lint
    if ! command_exists golangci-lint; then
        print_status "Installing golangci-lint..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
        print_success "golangci-lint installed"
    else
        print_success "golangci-lint already installed"
    fi
    
    # Install gosec
    if ! command_exists gosec; then
        print_status "Installing gosec..."
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        print_success "gosec installed"
    else
        print_success "gosec already installed"
    fi
}

# Function to fix dependencies
fix_dependencies() {
    print_status "Fixing Go module dependencies..."
    
    if [ -f scripts/fix-dependencies.sh ]; then
        chmod +x scripts/fix-dependencies.sh
        ./scripts/fix-dependencies.sh
    else
        print_status "Running go mod tidy on all modules..."
        go mod tidy
        
        if [ -d shared ]; then
            cd shared && go mod tidy && cd ..
        fi
        
        for service in services/*/; do
            if [ -f "$service/go.mod" ]; then
                print_status "Fixing dependencies for $service"
                cd "$service" && go mod tidy && cd ../..
            fi
        done
    fi
    
    print_success "Dependencies fixed"
}

# Function to run tests
run_tests() {
    print_status "Running tests to verify setup..."
    
    # Test shared module
    if [ -d shared ]; then
        print_status "Testing shared module..."
        cd shared && go test ./... && cd ..
    fi
    
    # Test services
    for service in services/*/; do
        if [ -f "$service/go.mod" ]; then
            service_name=$(basename "$service")
            print_status "Testing $service_name service..."
            cd "$service" && go test ./... && cd ../..
        fi
    done
    
    print_success "All tests passed"
}

# Function to create run scripts
create_run_scripts() {
    print_status "Creating service run scripts..."
    
    mkdir -p scripts/run
    
    # Create individual service run scripts
    for service in services/*/; do
        if [ -f "$service/cmd/main.go" ]; then
            service_name=$(basename "$service")
            script_file="scripts/run/$service_name.sh"
            
            cat > "$script_file" << EOF
#!/bin/bash
# Run $service_name service

cd services/$service_name
export DB_PATH="../../data/${service_name}.db"
export LOG_LEVEL=debug
export ENVIRONMENT=development

echo "🚀 Starting $service_name service..."
go run ./cmd/main.go
EOF
            chmod +x "$script_file"
        fi
    done
    
    # Create run-all script
    cat > scripts/run/all.sh << 'EOF'
#!/bin/bash
# Run all services concurrently

echo "🌱 Starting all GreenLedger services..."

# Function to run service in background
run_service() {
    local service=$1
    local port=$2
    echo "Starting $service on port $port..."
    ./scripts/run/$service.sh &
}

# Start all services
run_service calculator 8081
run_service tracker 8082
run_service wallet 8083
run_service user-auth 8084

echo "All services started!"
echo "Press Ctrl+C to stop all services"

# Wait for all background processes
wait
EOF
    chmod +x scripts/run/all.sh
    
    print_success "Run scripts created in scripts/run/"
}

# Function to display next steps
show_next_steps() {
    echo
    print_success "🎉 Development environment setup complete!"
    echo
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Review and update .env file with your settings"
    echo "2. Run individual services:"
    echo "   ./scripts/run/calculator.sh"
    echo "   ./scripts/run/tracker.sh"
    echo "   ./scripts/run/wallet.sh"
    echo "   ./scripts/run/user-auth.sh"
    echo
    echo "3. Or run all services at once:"
    echo "   ./scripts/run/all.sh"
    echo
    echo "4. Test the services:"
    echo "   curl http://localhost:8081/health"
    echo "   curl http://localhost:8082/health"
    echo "   curl http://localhost:8083/health"
    echo "   curl http://localhost:8084/health"
    echo
    echo "5. For live reloading during development:"
    echo "   cd services/calculator && air"
    echo
    echo "6. Run linting:"
    echo "   make lint"
    echo
    echo "7. Run tests:"
    echo "   make test"
    echo
    print_success "Happy coding! 🌱"
}

# Main execution
main() {
    echo -e "${GREEN}"
    echo "🌱 GreenLedger Development Setup"
    echo "================================"
    echo -e "${NC}"
    
    check_go_version
    check_disk_space
    create_directories
    setup_environment
    install_dev_tools
    fix_dependencies
    create_run_scripts
    
    # Only run tests if not in CI or if explicitly requested
    if [ "${SKIP_TESTS:-false}" != "true" ]; then
        run_tests
    fi
    
    show_next_steps
}

# Run main function
main "$@"
EOF
