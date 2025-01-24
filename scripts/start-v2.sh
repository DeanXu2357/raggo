#!/bin/bash

# Exit on any error
set -e

echo "🚀 Starting RaGGo v2..."

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check required commands
echo "🔍 Checking required commands..."
REQUIRED_COMMANDS=("docker" "go" "npm")
for cmd in "${REQUIRED_COMMANDS[@]}"; do
    if ! command_exists "$cmd"; then
        echo "❌ Error: '$cmd' is not installed"
        exit 1
    fi
done

# Setup environment file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Setting up environment file..."
    if [ -f .env.v2.example ]; then
        cp .env.v2.example .env
        echo "✅ Created .env from .env.v2.example"
    else
        echo "❌ Error: .env.v2.example not found"
        exit 1
    fi
fi

# Start services with docker-compose
echo "🐳 Starting services..."
docker compose -f docker-compose-v2.yaml up --build -d
