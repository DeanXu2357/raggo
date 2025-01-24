#!/bin/bash

# Exit on any error
set -e

echo "🔄 Restarting RaGGo v2 service..."

# Find and kill existing raggo v2 process
echo "🛑 Stopping existing service..."
if pgrep -f "raggo serve-v2" > /dev/null; then
    pkill -f "raggo serve-v2"
    echo "✅ Stopped existing service"
    # Give it a moment to fully stop
    sleep 2
else
    echo "ℹ️ No existing service found"
fi

# Rebuild the application
echo "🔨 Rebuilding application..."
go build -o raggo main.go

# Start the service
echo "🚀 Starting service..."
./raggo serve-v2 &

echo "✅ Service restarted successfully!"
