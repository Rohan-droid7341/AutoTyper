# ⌨️ Auto Typer for Windows

> Turn your mobile phone or tablet into an external typing companion for your Windows PC or laptop. Simulates natural human typing cadences, adjacent QWERTY key typos, automatic backspace corrections, and smart IDE indentation over local Wi-Fi or USB.

---

## 📁 Package Contents

```
📁 windows/
├── server.exe           ← Double-click to launch (runs silently in background)
├── stop.exe             ← Double-click to cleanly shut down the server
├── allow-firewall.bat   ← Run once as Administrator if local Wi-Fi blocks incoming connections
└── build.bat            ← Recompile binaries from source
```

---

## 🚀 Quick Start

### 1. Launch Server
Double-click **`server.exe`**.
- Runs in the background (no console window).
- Automatically opens the **Pairing Hub** (`http://localhost:8080/dashboard`).

### 2. Allow Firewall (If Needed)
If you connect your phone over local Wi-Fi and the phone cannot load the dashboard, right-click **`allow-firewall.bat`** and select **Run as Administrator** to open the server port.

### 3. Pair Your Phone
- Scan the **QR Code** on the laptop dashboard using your phone's camera.
- *Or*: Navigate to the displayed pairing URL and type the **6-digit PIN**.

### 4. Start Typing
1. Paste or type code into the phone web app.
2. Configure speed (10–500 WPM), Jitter level, and Oops typo rate.
3. Tap **Start Typing** and switch focus to your target app (VS Code, Notepad, terminal) during the countdown!

### 5. Stop Anytime
- Double-click **`stop.exe`**, or
- Click the red **🛑 Stop Server** button on the dashboard, or
- Move your mouse cursor into any of the 4 screen corners for an instant emergency abort.

---

## 🛠️ Building from Source

Requires [Go 1.22+](https://golang.org):

```cmd
cd windows
build.bat
```
Or from the project root:
```bash
go build -ldflags="-H=windowsgui -s -w" -o windows/server.exe ./cmd/server
go build -ldflags="-s -w" -o windows/stop.exe ./cmd/stop
```
