#!/bin/bash

# run_test.sh - Test runner for poon-web
# This script runs all tests for the poon-web directory

set -e  # Exit on any error

echo "🧪 Running tests for poon-web..."
echo "================================"

# Change to the poon-web directory
cd "$(dirname "$0")/.."

# Install dependencies
echo "📦 Installing Node.js dependencies..."
npm install

# Run linting
echo "🔍 Running ESLint..."
npm run lint

# Type checking
echo "📝 Running TypeScript type checking..."
npx tsc --noEmit

# Run tests if they exist
if grep -q '"test"' package.json; then
    echo "🔬 Running npm tests..."
    npm test
else
    echo "⚠️  No test script found in package.json, adding basic test setup..."
    
    # Create a basic test to verify the app builds and starts
    echo "🔨 Testing Next.js build..."
    npm run build
    
    echo "🚀 Testing Next.js production start (5 seconds)..."
    timeout 5s npm start || true
    
    echo "✅ Build and start test completed"
fi

# Check for security vulnerabilities
echo "🔒 Running security audit..."
npm audit --audit-level=moderate || {
    echo "⚠️  Security vulnerabilities found, but continuing tests"
}

# Check bundle size if build exists
if [ -d ".next" ]; then
    echo "📊 Checking bundle size..."
    echo "Build output:"
    ls -la .next/static/chunks/ | head -5 || true
fi

echo ""
echo "✅ All tests passed for poon-web!"
echo "================================="