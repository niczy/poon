#!/bin/bash

# Stop the test servers

echo "🛑 Stopping Poon test servers..."

# Read PIDs if they exist and kill processes
if [ -f "poon-server.pid" ]; then
    PID=$(cat poon-server.pid)
    if kill -0 $PID 2>/dev/null; then
        echo "Stopping poon-server (PID: $PID)"
        kill $PID
        sleep 2
        kill -9 $PID 2>/dev/null || true
    fi
    rm -f poon-server.pid
fi

if [ -f "poon-git.pid" ]; then
    PID=$(cat poon-git.pid)
    if kill -0 $PID 2>/dev/null; then
        echo "Stopping poon-git (PID: $PID)"
        kill $PID
        sleep 2
        kill -9 $PID 2>/dev/null || true
    fi
    rm -f poon-git.pid
fi

# Also kill by process name as backup
pkill -f "poon-server" || true
pkill -f "poon-git" || true

echo "✅ Servers stopped"