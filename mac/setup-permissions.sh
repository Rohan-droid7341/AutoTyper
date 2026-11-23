#!/usr/bin/env bash
set -e

echo "=========================================================="
echo "  ⌨️  Auto Typer — macOS Accessibility Setup"
echo "=========================================================="
echo ""
echo "macOS requires Accessibility permissions for Auto Typer to"
echo "simulate keystrokes and type into other applications (VS Code, terminal, etc.)."
echo ""

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
  echo "ℹ️  This setup script is intended for macOS."
fi

echo "👉 Triggering macOS Accessibility permission dialog..."
osascript -e 'tell application "System Events" to key code 0' 2>/dev/null || true

echo "👉 Opening System Settings -> Privacy & Security -> Accessibility..."
open "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"

echo ""
echo "----------------------------------------------------------"
echo "📋 STEPS TO COMPLETE:"
echo "1. In System Settings, locate 'Accessibility' under 'Privacy & Security'."
echo "2. Toggle ON the switch next to Terminal (or your terminal emulator, e.g. iTerm2)"
echo "   and/or the 'server' binary."
echo "3. Once granted, launch Auto Typer by running './start.sh' or './server'."
echo "----------------------------------------------------------"
echo "✅ Setup prompt launched successfully."
