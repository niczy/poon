#!/bin/bash

# run_test.sh - Test runner for poon-tests
# This script runs all integration tests for the poon system

set -e  # Exit on any error

echo "🧪 Running integration tests for poon-tests..."
echo "==============================================="

# Change to the poon-tests directory
cd "$(dirname "$0")/.."

# Ensure we have the latest dependencies
echo "📦 Installing/updating Go dependencies..."
go mod download
go mod tidy

# Check if servers are running, if not start them
echo "🚀 Checking server status..."
if ! pgrep -f "poon-server" > /dev/null; then
    echo "⚠️  poon-server not running, attempting to start..."
    if [ -f "./start-servers.sh" ]; then
        ./start-servers.sh
        sleep 3  # Give servers time to start
    else
        echo "❌ start-servers.sh not found"
    fi
fi

# Run all tests with verbose output
echo "🔬 Running integration tests..."
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
echo "✅ All integration tests passed for poon-tests!"
echo "==============================================="