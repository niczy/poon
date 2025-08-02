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
    # Ensure protoc-gen-go tools are installed
    echo "📦 Ensuring protoc-gen-go tools are available..."
    
    # Always install the tools to ensure they're available in CI
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    
    # Add Go bin directories to PATH for protoc plugins
    export PATH="$PATH:$(go env GOPATH)/bin:$HOME/go/bin"
    
    # Verify tools are available
    echo "🔍 Verifying protoc plugins..."
    which protoc-gen-go || { echo "❌ protoc-gen-go not found"; exit 1; }
    which protoc-gen-go-grpc || { echo "❌ protoc-gen-go-grpc not found"; exit 1; }
    
    # Clean and regenerate
    npm run clean || true
    
    # Create gen directory structure
    mkdir -p gen/go gen/js gen/python gen/ts
    
    # Test Go generation
    if npm run proto:generate:go; then
        echo "✅ Go protobuf generation successful"
        
        # Create go.mod if it doesn't exist
        cd gen/go
        if [ ! -f go.mod ]; then
            echo "📝 Creating go.mod for generated protobuf files..."
            cat > go.mod << EOF
module github.com/nic/poon/poon-proto/gen/go

go 1.23

require (
	google.golang.org/grpc v1.74.2
	google.golang.org/protobuf v1.36.0
)
EOF
            go mod tidy
        fi
        cd ../..
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
    "gen/go/monorepo.pb.go"
    "gen/go/monorepo_grpc.pb.go"
)

for file in "${EXPECTED_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file exists"
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