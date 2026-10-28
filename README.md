# ⌨️ Auto Typer (Go Edition)

**Turn your phone or tablet into a high-speed, secure wireless/USB keyboard for your computer.**

Built in **pure Go** with zero external dependencies or runtime environments (no Python, Node.js, or pip required).

---

## 🚀 Quick Start (Two-File Package)

The compiled package contains just two files:

```
📁 Auto Typer/
├── server.exe    ← Double-click to start server & open Pairing Hub
└── stop.exe      ← Double-click to cleanly shut down the server
```

### 1. Start the Server
Double-click **`server.exe`**.
- Starts the embedded web server and WebSocket engine.
- Automatically launches your default browser to the **Pairing Hub** (`http://localhost:8080/dashboard`).
- If an Android phone is connected via USB-C with USB Debugging enabled, it automatically sets up ADB reverse port forwarding.

### 2. Pair Your Phone
- **Scan the QR Code** displayed on the laptop dashboard using your phone's camera.
- *Alternatively*: Open the URL directly on your phone and enter the **6-digit PIN** shown on your laptop screen.

> 🔒 **Security Guarantee**: Only devices that scan the QR code or enter the secret PIN gain write/typing permissions. Unauthorized visitors on your local network cannot hijack or type into your computer.

### 3. Type or Punch
1. Paste or type your code/text on your phone.
2. Select your desired speed (10–500 WPM) and mode:
   - **🧠 Smart IDE Auto-Indent**: Automatically adjusts for IDE auto-indentation and dedents when needed.
   - **👻 Human-like Jitter**: Adds natural keystroke timing variance to bypass bot detection.
   - **🙈 Oops Mode**: Simulates human typos with natural backspacing and corrections.
   - **📋 Fast Paste**: Instantly injects full text payloads via clipboard.
3. Tap **Start Typing** — focus the target application on your laptop during the countdown.

### 4. Stop the Server
Double-click **`stop.exe`** anytime to gracefully shut down the server and clean up any ADB reverse tunnels.

---

## 🛡️ Safety & Failsafe Features

- **Mouse Corner Failsafe**: Move your mouse cursor into any corner of the laptop screen to immediately abort all typing.
- **Stop Button**: Tap the Stop button on the web app to cancel the current job.
- **Token Verification**: Unpaired network clients are blocked from executing keystrokes or clipboard operations.

---

## 🛠️ Building from Source

Requires [Go 1.22+](https://golang.org):

```bash
# Clone the repository
git clone <repo-url>
cd "Auto typer"

# Build self-contained executables
go build -ldflags="-s -w" -o server.exe ./cmd/server
go build -ldflags="-s -w" -o stop.exe ./cmd/stop
```

All web assets (HTML, CSS, JS, QR code generator) are directly embedded into `server.exe` using `//go:embed`.
