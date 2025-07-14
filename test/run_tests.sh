#!/bin/bash

# Scout MCP Integration Test Runner

set -e

echo "🚀 Starting Scout MCP Integration Tests"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "📁 Project root: $PROJECT_ROOT"
echo "🧪 Test directory: $SCRIPT_DIR"

# Build the scout-mcp binary if it doesn't exist
BINARY_PATH="$PROJECT_ROOT/bin/scout-mcp"
if [ ! -f "$BINARY_PATH" ]; then
    echo "🔨 Building scout-mcp binary..."
    cd "$PROJECT_ROOT"
    go build -o bin/scout-mcp cmd/main.go
    echo "✅ Binary built successfully"
else
    echo "✅ Binary already exists: $BINARY_PATH"
fi

# Change to test directory
cd "$SCRIPT_DIR"

# Download test dependencies
echo "📦 Downloading test dependencies..."
go mod download

# Run the tests
echo "🧪 Running integration tests..."
go test -v -timeout 30s ./...

echo "🎉 Tests completed!"
