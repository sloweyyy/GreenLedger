#!/bin/bash

# Fix module paths script
# Updates all import paths from github.com/greenledger to github.com/sloweyyy/GreenLedger

set -e

echo "🔧 Fixing module paths..."
echo "========================="

# Function to update imports in a directory
update_imports() {
    local dir="$1"
    echo "📦 Updating imports in $dir..."
    
    # Find all .go files and update import paths
    find "$dir" -name "*.go" -type f -exec sed -i '' 's|github.com/greenledger|github.com/sloweyyy/GreenLedger|g' {} \;
}

# Update root directory
if [ -f "go.mod" ]; then
    echo "📦 Updating root module imports..."
    update_imports "."
fi

# Update shared module
if [ -d "shared" ]; then
    echo "📦 Updating shared module imports..."
    update_imports "shared"
fi

# Update service modules
for service_dir in services/*/; do
    if [ -d "$service_dir" ]; then
        service_name=$(basename "$service_dir")
        echo "📦 Updating $service_name module imports..."
        
        # Update go.mod file
        if [ -f "$service_dir/go.mod" ]; then
            sed -i '' "s|github.com/greenledger/services/$service_name|github.com/sloweyyy/GreenLedger/services/$service_name|g" "$service_dir/go.mod"
        fi
        
        # Update import statements in Go files
        update_imports "$service_dir"
    fi
done

echo "✅ Module paths updated!"
echo "Now running go mod tidy on all modules..."

# Tidy all modules
go mod tidy

if [ -d "shared" ]; then
    cd shared && go mod tidy && cd ..
fi

for service_dir in services/*/; do
    if [ -d "$service_dir" ] && [ -f "$service_dir/go.mod" ]; then
        service_name=$(basename "$service_dir")
        echo "🧹 Tidying $service_name module..."
        cd "$service_dir" && go mod tidy && cd ../..
    fi
done

echo "🎉 All module paths fixed and tidied!"
