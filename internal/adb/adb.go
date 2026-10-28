package adb

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Status contains ADB diagnostic info.
type Status struct {
	Installed bool   `json:"installed"`
	Connected bool   `json:"connected"`
	DeviceID  string `json:"device_id,omitempty"`
	Message   string `json:"message"`
}

// Check checks for ADB installation and connected Android devices.
func Check() Status {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "adb", "version")
	if err := cmd.Run(); err != nil {
		return Status{
			Installed: false,
			Connected: false,
			Message:   "ADB is not installed or not found on system PATH",
		}
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	cmdDevices := exec.CommandContext(ctx2, "adb", "devices")
	output, err := cmdDevices.Output()
	if err != nil {
		return Status{
			Installed: true,
			Connected: false,
			Message:   "Failed to query connected ADB devices",
		}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 && fields[1] == "device" {
			return Status{
				Installed: true,
				Connected: true,
				DeviceID:  fields[0],
				Message:   fmt.Sprintf("Android device connected (%s)", fields[0]),
			}
		}
	}

	return Status{
		Installed: true,
		Connected: false,
		Message:   "No Android device detected over USB. Enable USB Debugging if using a cable.",
	}
}

// SetupReverse configures ADB reverse port forwarding.
func SetupReverse(port int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	portStr := fmt.Sprintf("tcp:%d", port)
	cmd := exec.CommandContext(ctx, "adb", "reverse", portStr, portStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb reverse %s failed: %s (%w)", portStr, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// RemoveReverse removes all reverse port forwards.
func RemoveReverse() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "adb", "reverse", "--remove-all")
	_ = cmd.Run()
}
