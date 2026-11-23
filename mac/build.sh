#!/usr/bin/env bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$DIR"

echo "Building Auto Typer for macOS..."

echo "1. Building for Apple Silicon (ARM64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o mac/server-arm64 ./cmd/server
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o mac/stop-arm64 ./cmd/stop

echo "2. Building for Intel (AMD64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/server-amd64 ./cmd/server
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/stop-amd64 ./cmd/stop

# Provide standard server/stop symlink or default to arm64
cp mac/server-arm64 mac/server
cp mac/stop-arm64 mac/stop

chmod +x mac/server mac/stop mac/server-* mac/stop-* mac/*.sh

echo "✅ macOS binaries built successfully in mac/!"
