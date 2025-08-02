#!/bin/bash

# run_test.sh - Test runner for poon-cli
# This script runs all tests for the poon-cli directory

set -e  # Exit on any error

echo "🧪 Running tests for poon-cli..."
echo "================================"

# Change to the poon-cli directory
cd "$(dirname "$0")/.."

# Ensure we have the latest dependencies
echo "📦 Installing/updating Go dependencies..."
go mod download
go mod tidy

# Run all tests with verbose output (if any exist)
echo "🔬 Running Go tests..."
if ls *_test.go >/dev/null 2>&1; then
    go test -v -race -cover ./...
else
    echo "⚠️  No test files found (*_test.go)"
fi

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

# Test that the CLI builds successfully
echo "🔨 Testing CLI build..."
go build -o poon-cli-test ./...
rm -f poon-cli-test

echo ""
echo "✅ All tests passed for poon-cli!"
echo "================================="