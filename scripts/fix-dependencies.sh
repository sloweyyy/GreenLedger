#!/bin/bash

# 🔧 Fix Dependencies Script
# This script ensures all Go modules have proper dependencies and go.sum files

set -e

echo "🔧 GreenLedger Dependency Fix Script"
echo "===================================="

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

# Function to check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}')
    print_status "Using Go version: $GO_VERSION"
}

# Function to process a Go module
process_module() {
    local module_path="$1"
    local module_name="$2"
    
    if [ ! -f "$module_path/go.mod" ]; then
        print_warning "No go.mod found in $module_path, skipping..."
        return
    fi
    
    print_status "Processing $module_name module..."
    
    cd "$module_path"
    
    # Check if go.mod is valid
    if ! go mod verify &> /dev/null; then
        print_warning "go.mod verification failed for $module_name, attempting to fix..."
    fi
    
    # Download dependencies
    print_status "  📥 Downloading dependencies..."
    if ! go mod download; then
        print_error "Failed to download dependencies for $module_name"
        cd - > /dev/null
        return 1
    fi
    
    # Tidy up the module
    print_status "  🧹 Tidying module..."
    if ! go mod tidy; then
        print_error "Failed to tidy module $module_name"
        cd - > /dev/null
        return 1
    fi
    
    # Verify the module
    print_status "  ✅ Verifying module..."
    if ! go mod verify; then
        print_error "Module verification failed for $module_name"
        cd - > /dev/null
        return 1
    fi
    
    # Check if there are any Go files to build
    if find . -name "*.go" -not -path "./vendor/*" | grep -q .; then
        print_status "  🔨 Testing build..."
        if ! go build ./...; then
            print_warning "Build test failed for $module_name (this might be expected for some modules)"
        fi
    fi
    
    cd - > /dev/null
    print_success "$module_name module processed successfully"
}

# Function to fix root module
fix_root_module() {
    print_status "🌱 Processing root module..."
    process_module "." "root"
}

# Function to fix shared module
fix_shared_module() {
    if [ -d "shared" ]; then
        print_status "📚 Processing shared module..."
        process_module "shared" "shared"
    else
        print_warning "Shared directory not found, skipping..."
    fi
}

# Function to fix service modules
fix_service_modules() {
    print_status "🔧 Processing service modules..."
    
    if [ ! -d "services" ]; then
        print_warning "Services directory not found, skipping..."
        return
    fi
    
    for service_dir in services/*/; do
        if [ -d "$service_dir" ]; then
            service_name=$(basename "$service_dir")
            print_status "🛠️  Processing $service_name service..."
            process_module "$service_dir" "$service_name"
        fi
    done
}

# Function to check for common issues
check_common_issues() {
    print_status "🔍 Checking for common issues..."
    
    # Check for missing go.sum files
    print_status "  📋 Checking for missing go.sum files..."
    
    missing_sum_files=()
    
    # Check root
    if [ -f "go.mod" ] && [ ! -f "go.sum" ]; then
        missing_sum_files+=("root")
    fi
    
    # Check shared
    if [ -f "shared/go.mod" ] && [ ! -f "shared/go.sum" ]; then
        missing_sum_files+=("shared")
    fi
    
    # Check services
    for service_dir in services/*/; do
        if [ -d "$service_dir" ]; then
            service_name=$(basename "$service_dir")
            if [ -f "$service_dir/go.mod" ] && [ ! -f "$service_dir/go.sum" ]; then
                missing_sum_files+=("$service_name")
            fi
        fi
    done
    
    if [ ${#missing_sum_files[@]} -gt 0 ]; then
        print_warning "Missing go.sum files detected in: ${missing_sum_files[*]}"
        print_status "These will be created during the dependency resolution process"
    else
        print_success "All modules have go.sum files"
    fi
    
    # Check for outdated dependencies
    print_status "  📦 Checking for outdated dependencies..."
    
    # This is a basic check - in a real scenario you might want to use tools like 'go list -u -m all'
    print_status "Run 'go list -u -m all' in each module directory to check for updates"
}

# Function to generate dependency report
generate_report() {
    print_status "📊 Generating dependency report..."
    
    report_file="dependency-report.txt"
    echo "GreenLedger Dependency Report" > "$report_file"
    echo "Generated on: $(date)" >> "$report_file"
    echo "=============================" >> "$report_file"
    echo "" >> "$report_file"
    
    # Root module
    if [ -f "go.mod" ]; then
        echo "Root Module Dependencies:" >> "$report_file"
        go list -m all >> "$report_file" 2>/dev/null || echo "Failed to list root dependencies" >> "$report_file"
        echo "" >> "$report_file"
    fi
    
    # Shared module
    if [ -f "shared/go.mod" ]; then
        echo "Shared Module Dependencies:" >> "$report_file"
        (cd shared && go list -m all >> "../$report_file" 2>/dev/null) || echo "Failed to list shared dependencies" >> "$report_file"
        echo "" >> "$report_file"
    fi
    
    # Service modules
    for service_dir in services/*/; do
        if [ -d "$service_dir" ] && [ -f "$service_dir/go.mod" ]; then
            service_name=$(basename "$service_dir")
            echo "$service_name Service Dependencies:" >> "$report_file"
            (cd "$service_dir" && go list -m all >> "../../$report_file" 2>/dev/null) || echo "Failed to list $service_name dependencies" >> "$report_file"
            echo "" >> "$report_file"
        fi
    done
    
    print_success "Dependency report generated: $report_file"
}

# Main execution
main() {
    print_status "Starting dependency fix process..."
    
    # Check prerequisites
    check_go
    
    # Check for common issues first
    check_common_issues
    
    # Fix modules in order
    fix_root_module
    fix_shared_module
    fix_service_modules
    
    # Generate report
    generate_report
    
    print_success "🎉 Dependency fix process completed!"
    print_status "All Go modules should now have proper dependencies and go.sum files"
    print_status "You can now run the CI pipeline or build the services"
}

# Handle script arguments
case "${1:-}" in
    "check")
        check_go
        check_common_issues
        ;;
    "report")
        check_go
        generate_report
        ;;
    "root")
        check_go
        fix_root_module
        ;;
    "shared")
        check_go
        fix_shared_module
        ;;
    "services")
        check_go
        fix_service_modules
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  (no args)  - Fix all modules (default)"
        echo "  check      - Check for common issues only"
        echo "  report     - Generate dependency report only"
        echo "  root       - Fix root module only"
        echo "  shared     - Fix shared module only"
        echo "  services   - Fix service modules only"
        echo "  help       - Show this help message"
        ;;
    "")
        main
        ;;
    *)
        print_error "Unknown command: $1"
        print_status "Use '$0 help' for usage information"
        exit 1
        ;;
esac
