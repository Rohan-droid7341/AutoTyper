#!/usr/bin/env bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR"

echo "=========================================================="
echo "  ⌨️  Auto Typer — macOS Launcher"
echo "=========================================================="

ARCH="$(uname -m)"
BINARY="./server"

if [[ ! -f "$BINARY" ]]; then
  if [[ "$ARCH" == "arm64" && -f "./server-arm64" ]]; then
    BINARY="./server-arm64"
  elif [[ "$ARCH" == "x86_64" && -f "./server-amd64" ]]; then
    BINARY="./server-amd64"
  fi
fi

if [[ ! -f "$BINARY" ]]; then
  echo "❌ Error: Could not find server binary in $DIR."
  echo "Please build it with './build.sh' or download the release package."
  exit 1
fi

chmod +x "$BINARY"

echo "🚀 Starting Auto Typer background service..."
"$BINARY" &
SERVER_PID=$!

echo "✓ Auto Typer running in background (PID: $SERVER_PID)"
echo "📱 Dashboard will automatically open in your default browser."
echo "🛑 To stop Auto Typer anytime, run: ./stop.sh"
echo "=========================================================="
