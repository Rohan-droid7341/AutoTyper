#!/usr/bin/env bash
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR"

echo "=========================================================="
echo "  🛑 Auto Typer — macOS Stop Utility"
echo "=========================================================="

ARCH="$(uname -m)"
STOP_BIN="./stop"

if [[ ! -f "$STOP_BIN" ]]; then
  if [[ "$ARCH" == "arm64" && -f "./stop-arm64" ]]; then
    STOP_BIN="./stop-arm64"
  elif [[ "$ARCH" == "x86_64" && -f "./stop-amd64" ]]; then
    STOP_BIN="./stop-amd64"
  fi
fi

if [[ -f "$STOP_BIN" ]]; then
  chmod +x "$STOP_BIN"
  "$STOP_BIN"
else
  echo "Stopping running Auto Typer processes..."
  pkill -f "server" 2>/dev/null || true
  adb reverse --remove-all 2>/dev/null || true
  rm -f /tmp/autotyper.lock
  echo "✅ Stopped successfully."
fi
