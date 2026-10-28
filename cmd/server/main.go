package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"autotyper/internal/adb"
	"autotyper/internal/auth"
	"autotyper/internal/web"
)

type LockInfo struct {
	PID        int    `json:"pid"`
	Port       int    `json:"port"`
	AdminToken string `json:"admin_token"`
}

func getLockFilePath() string {
	return filepath.Join(os.TempDir(), "autotyper.lock")
}

func writeLockFile(port int, adminToken string) error {
	info := LockInfo{
		PID:        os.Getpid(),
		Port:       port,
		AdminToken: adminToken,
	}
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getLockFilePath(), data, 0600)
}

func removeLockFile() {
	_ = os.Remove(getLockFilePath())
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func main() {
	log.Println("──────────────────────────────────────────────────────────")
	log.Println("  ⌨️  Auto Typer Server — Initializing...")
	log.Println("──────────────────────────────────────────────────────────")

	// 1. Create Auth Manager
	authMgr, err := auth.New()
	if err != nil {
		log.Fatalf("❌ Failed to initialize auth manager: %v", err)
	}

	// 2. Select Port (default 8080, fallback if in use)
	port := 8080
	for p := 8080; p <= 8090; p++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", p))
		if err == nil {
			_ = ln.Close()
			port = p
			break
		}
	}

	// 3. Save Lock File for stop.exe
	if err := writeLockFile(port, authMgr.AdminToken); err != nil {
		log.Printf("⚠️  Could not write lock file: %v", err)
	}
	defer removeLockFile()

	// 4. Check ADB & Setup Reverse Port Forwarding
	adbStatus := adb.Check()
	if adbStatus.Connected {
		log.Printf("🔌 ADB USB device detected (%s). Setting up reverse tunnel on port %d...", adbStatus.DeviceID, port)
		if err := adb.SetupReverse(port); err != nil {
			log.Printf("⚠️  ADB reverse setup warning: %v", err)
		} else {
			log.Printf("✓ ADB reverse forwarding active: phone can use http://localhost:%d", port)
		}
	} else if adbStatus.Installed {
		log.Printf("ℹ️  ADB is installed, but no device connected. Running in Wi-Fi / Local Network mode.")
	} else {
		log.Printf("ℹ️  Running in Wi-Fi / Local Network mode.")
	}
	defer adb.RemoveReverse()

	// 5. Initialize Web Server
	shutdownChan := make(chan struct{}, 1)
	serverInstance := web.NewServer(port, authMgr, shutdownChan)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: serverInstance.Handler(),
	}

	go func() {
		log.Printf("🌐 Auto Typer running at http://localhost:%d", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	dashboardURL := fmt.Sprintf("http://localhost:%d/dashboard", port)
	log.Printf("📱 Pairing Dashboard: %s", dashboardURL)
	log.Printf("🔑 Pairing PIN:        %s", authMgr.PairingPIN)
	log.Println("──────────────────────────────────────────────────────────")
	log.Println("  💡 Moving mouse to screen corner activates Emergency Stop")
	log.Println("  🛑 Run stop.exe or press Ctrl+C to shut down")
	log.Println("──────────────────────────────────────────────────────────")

	// Open pairing dashboard automatically
	go func() {
		time.Sleep(600 * time.Millisecond)
		openBrowser(dashboardURL)
	}()

	// 6. Handle Shutdown Signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("\n👋 Received interrupt signal. Shutting down...")
	case <-shutdownChan:
		log.Println("\n🛑 Shutdown requested via stop command or dashboard...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = httpServer.Shutdown(ctx)
	removeLockFile()
	adb.RemoveReverse()
	log.Println("✅ Auto Typer server cleanly stopped. Goodbye!")
}
