# ⌨️ Auto Typer for macOS

> Turn your phone or tablet into an external typing companion on macOS. Full native support for Apple Silicon (M1/M2/M3/M4) and Intel Macs using high-performance CoreGraphics dispatch.

---

## ⚡ Key Highlights on macOS

- **Sub-millisecond Keystroke Overhead**: Powered by macOS `CoreGraphics` (`CGEventPost` / `CGEventKeyboardSetUnicodeString`). No lag, zero dropped keystrokes.
- **Fast Paste Mode**: Instant clipboard injection via macOS native `pbcopy` and simulated `Cmd + V`.
- **Screen Corner Emergency Stop**: Moving mouse to any screen corner immediately halts typing.
- **Multi-Level Jitter & Realistic Typo Mode**: Simulates human typing speeds, pauses, and backspace typo corrections.
- **Dual Connection Modes**: Connect over local Wi-Fi or high-speed USB-C via Android ADB reverse forwarding.

---

## 🔐 Crucial: macOS Accessibility Permissions

macOS security policy requires that any application that simulates keystrokes or monitors global cursor position be granted **Accessibility** permissions.

### Option 1: Automatic Prompt (Recommended)
1. Launch `./start.sh` or `./server`.
2. macOS will automatically display a permission dialog:
   > *"Auto Typer" (or Terminal) would like to control this computer using accessibility features.*
3. Click **Open System Settings**.
4. Toggle **ON** the switch for **Terminal** (or **iTerm2**), and/or the `server` binary.

### Option 2: Run Setup Helper Script
```bash
./setup-permissions.sh
```
This script triggers the permission check and directly opens **System Settings > Privacy & Security > Accessibility**.

---

## 🚀 Quick Start

### 1. Launch the Server
```bash
cd mac
chmod +x start.sh stop.sh server stop
./start.sh
```
- Runs in the background.
- Automatically opens the **Pairing Hub** (`http://localhost:8080/dashboard`) in your default browser.

### 2. Pair Your Phone
- Scan the **QR Code** on the laptop dashboard using your phone's camera.
- Or open the pairing URL on your phone and enter the **6-digit PIN**.

### 3. Start Typing
1. Paste or type code into the phone web app.
2. Choose your typing speed (WPM), Jitter, and Oops typo rate.
3. Tap **Start Typing** and switch focus to your target application (VS Code, Xcode, Terminal, etc.) during the countdown!

### 4. Stop Anytime
- Run `./stop.sh` in the terminal, or
- Click the red **🛑 Stop Server** button on the dashboard, or
- Slam your mouse cursor into any screen corner for an instant emergency abort.

---

## 🛠️ Building from Source

From the project root:

```bash
# Apple Silicon (M1/M2/M3/M4)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o mac/server ./cmd/server
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o mac/stop ./cmd/stop

# Intel (x86_64)
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/server ./cmd/server
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/stop ./cmd/stop
```
Or simply run `./build.sh` inside the `mac/` folder.
