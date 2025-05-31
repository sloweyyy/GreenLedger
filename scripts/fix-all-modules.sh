#!/bin/bash

# Fix all module paths and dependencies

set -e

echo "🔧 Comprehensive Module Path Fix"
echo "================================"

# Fix all service go.mod files
for service_dir in services/*/; do
    if [ -d "$service_dir" ] && [ -f "$service_dir/go.mod" ]; then
        service_name=$(basename "$service_dir")
        echo "📦 Fixing $service_name module..."
        
        cd "$service_dir"
        
        # Fix the require statement for shared module
        sed -i '' 's|github.com/greenledger/shared|github.com/sloweyyy/GreenLedger/shared|g' go.mod
        
        # Fix the replace statement
        sed -i '' 's|replace github.com/greenledger/shared|replace github.com/sloweyyy/GreenLedger/shared|g' go.mod
        
        # Clean and tidy
        go clean -modcache || true
        go mod tidy
        
        cd ../..
    fi
done

echo "✅ All modules fixed!"
