#!/bin/bash

# Test script to start poon-server and poon-git server

set -e

echo "🚀 Starting Poon test environment..."

# Kill any existing processes
pkill -f "poon-server" || true
pkill -f "poon-git" || true
sleep 2

# Generate proto files first (already done)
echo "📦 Proto files already generated..."

# Create log directory
mkdir -p logs

# Set environment variables
export REPO_ROOT="$(pwd)/test/monorepo"
export PORT=50051
export GRPC_SERVER="localhost:50051"

echo "📁 Repository root: $REPO_ROOT"

# Start poon-server in background
echo "🔧 Starting poon-server on port 50051..."
cd ../poon-server
go run . > ../test/logs/poon-server.log 2>&1 &
POON_SERVER_PID=$!
echo "poon-server PID: $POON_SERVER_PID"
cd ../test

# Wait for poon-server to start
sleep 3

# Start poon-git server in background
echo "🌐 Starting poon-git server on port 3000..."
cd ../poon-git
PORT=3000 GRPC_SERVER="localhost:50051" go run . > ../test/logs/poon-git.log 2>&1 &
POON_GIT_PID=$!
echo "poon-git PID: $POON_GIT_PID"
cd ../test

# Wait for servers to start
sleep 3

echo "✅ Servers started successfully!"
echo "📊 Server status:"
echo "  - poon-server: http://localhost:50051 (PID: $POON_SERVER_PID)"
echo "  - poon-git: http://localhost:3000 (PID: $POON_GIT_PID)"

# Save PIDs for cleanup
echo "$POON_SERVER_PID" > poon-server.pid
echo "$POON_GIT_PID" > poon-git.pid

echo ""
echo "🔍 Check server logs:"
echo "  tail -f test/logs/poon-server.log"
echo "  tail -f test/logs/poon-git.log"
echo ""
echo "🛑 To stop servers: ./test/stop-servers.sh"

# Test server connectivity
echo "🧪 Testing server connectivity..."
sleep 2

# Test poon-git health endpoint
if curl -s http://localhost:3000/health > /dev/null; then
    echo "✅ poon-git server is responding"
else
    echo "❌ poon-git server is not responding"
fi

# Keep script running to monitor
echo ""
echo "📝 Servers are running. Press Ctrl+C to stop."
trap "echo '🛑 Stopping servers...'; ./stop-servers.sh" EXIT

# Monitor servers
while true; do
    if ! kill -0 $POON_SERVER_PID 2>/dev/null; then
        echo "❌ poon-server died!"
        break
    fi
    if ! kill -0 $POON_GIT_PID 2>/dev/null; then
        echo "❌ poon-git server died!"
        break
    fi
    sleep 5
done