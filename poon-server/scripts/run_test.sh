#!/bin/bash

# run_test.sh - Test runner for poon-server
# This script runs all tests for the poon-server directory

set -e  # Exit on any error

echo "🧪 Running tests for poon-server..."
echo "==================================="

# Change to the poon-server directory
cd "$(dirname "$0")/.."

# Ensure we have the latest dependencies
echo "📦 Installing/updating Go dependencies..."
go mod download
go mod tidy

# Run all tests with verbose output
echo "🔬 Running Go tests..."
go test -v -race -cover ./...

# Check if there are any linting issues
if command -v golint >/dev/null 2>&1; then
    echo "🔍 Running golint..."
    golint ./...
else
    echo "⚠️  golint not installed, skipping linting"
fi

# Run go vet for static analysis
echo "🔧 Running go vet..."
go vet ./...

# Check go formatting
echo "📝 Checking go formatting..."
if [ "$(gofmt -l . | wc -l)" -gt 0 ]; then
    echo "❌ Code is not properly formatted. Run 'gofmt -w .' to fix."
    gofmt -l .
    exit 1
else
    echo "✅ Code is properly formatted"
fi

echo ""
echo "✅ All tests passed for poon-server!"
echo "===================================="