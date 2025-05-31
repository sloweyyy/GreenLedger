#!/bin/bash

# Quick fix script for immediate CI issues

set -e

echo "🔧 Quick Fix for CI Issues"
echo "=========================="

# Fix root module
echo "📦 Fixing root module..."
go mod download
go mod tidy

# Fix shared module
echo "📦 Fixing shared module..."
if [ -d "shared" ]; then
    cd shared
    go mod download
    go mod tidy
    cd ..
fi

# Fix calculator module
echo "📦 Fixing calculator module..."
if [ -d "services/calculator" ]; then
    cd services/calculator
    go mod download
    go mod tidy
    cd ../..
fi

# Fix tracker module (skip for now due to import cycle)
echo "⚠️ Skipping tracker module (import cycle needs manual fix)"

# Fix wallet module
echo "📦 Fixing wallet module..."
if [ -d "services/wallet" ]; then
    cd services/wallet
    go mod download
    go mod tidy
    cd ../..
fi

# Fix user-auth module
echo "📦 Fixing user-auth module..."
if [ -d "services/user-auth" ]; then
    cd services/user-auth
    go mod download
    go mod tidy
    cd ../..
fi

# Fix reporting module
echo "📦 Fixing reporting module..."
if [ -d "services/reporting" ]; then
    cd services/reporting
    go mod download
    go mod tidy
    cd ../..
fi

# Fix certifier module (already fixed)
echo "✅ Certifier module already fixed"

echo "✅ Quick fix completed!"
echo "Note: Tracker module needs manual import cycle fix"
