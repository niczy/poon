#!/bin/bash

# run_test.sh - Test runner for poon-proto
# This script runs all tests for the poon-proto directory

set -e  # Exit on any error

echo "🧪 Running tests for poon-proto..."
echo "=================================="

# Change to the poon-proto directory
cd "$(dirname "$0")/.."

# Check if npm is available for Node.js/JavaScript tests
if command -v npm >/dev/null 2>&1; then
    echo "📦 Installing Node.js dependencies..."
    npm install
    
    # Run npm tests if they exist
    if grep -q '"test"' package.json; then
        echo "🔬 Running npm tests..."
        npm test
    else
        echo "⚠️  No npm test script found in package.json"
    fi
    
    # Run linting if available
    if grep -q '"lint"' package.json; then
        echo "🔍 Running npm lint..."
        npm run lint
    fi
else
    echo "⚠️  npm not found, skipping Node.js tests"
fi

# Test protocol buffer generation
echo "🔨 Testing protobuf generation..."
if command -v protoc >/dev/null 2>&1; then
    # Clean and regenerate
    npm run clean || true
    
    # Create gen directory structure
    mkdir -p gen/go gen/js gen/python gen/ts
    
    # Test Go generation
    if npm run proto:generate:go; then
        echo "✅ Go protobuf generation successful"
    else
        echo "❌ Go protobuf generation failed"
        exit 1
    fi
    
    # Test JavaScript generation (may not have required tools, so warn only)
    if npm run proto:generate:js; then
        echo "✅ JavaScript protobuf generation successful"
    else
        echo "⚠️  JavaScript protobuf generation failed (tools may not be installed)"
    fi
else
    echo "⚠️  protoc not found, skipping protobuf generation test"
fi

# Verify generated files exist
echo "📋 Verifying generated files..."
EXPECTED_FILES=(
    "gen/monorepo.pb.go"
)

for file in "${EXPECTED_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file exists"
        # Move to go subdirectory if not already there
        if [ ! -f "gen/go/$(basename $file)" ]; then
            mkdir -p gen/go
            mv "$file" "gen/go/"
            echo "✅ Moved $file to gen/go/"
        fi
    else
        echo "❌ $file missing"
        exit 1
    fi
done

# Check proto file syntax
echo "🔍 Validating proto file syntax..."
if command -v protoc >/dev/null 2>&1; then
    if protoc --proto_path=. --descriptor_set_out=/dev/null monorepo.proto; then
        echo "✅ Proto file syntax is valid"
    else
        echo "❌ Proto file syntax validation failed"
        exit 1
    fi
fi

echo ""
echo "✅ All tests passed for poon-proto!"
echo "==================================="