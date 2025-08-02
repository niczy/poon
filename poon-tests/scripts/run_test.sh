#!/bin/bash

# run_test.sh - Simple test runner for poon-tests
# This script only runs go test for the poon-tests package

set -e  # Exit on any error

echo "🧪 Running Go tests for poon-tests..."
echo "==================================="

# Change to the poon-tests directory
cd "$(dirname "$0")/.."

# Ensure we have the latest dependencies
echo "📦 Installing/updating Go dependencies..."
go mod download
go mod tidy

# Run all tests with verbose output
echo "🔬 Running Go tests..."
go test -v ./...

echo ""
echo "✅ All Go tests passed for poon-tests!"
echo "==================================="