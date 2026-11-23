//go:build darwin

package keyboard

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"github.com/ebitengine/purego"
)

const (
	kCGHIDEventTap          uint32 = 0
	kCGEventFlagMaskCommand uint64 = 0x00100000 // 1 << 20

	vkMacReturn    uint16 = 0x24 // 36
	vkMacTab       uint16 = 0x30 // 48
	vkMacBackspace uint16 = 0x33 // 51
	vkMacV         uint16 = 0x09 // 9
)

type CGPoint struct {
	X float64
	Y float64
}

type CGSize struct {
	Width  float64
	Height float64
}

type CGRect struct {
	Origin CGPoint
	Size   CGSize
}

type darwinSimulator struct {
	mu                         sync.Mutex
	initialized                bool
	initErr                    error
	cgHandle                   uintptr
	asHandle                   uintptr
	cfHandle                   uintptr
	cgEventCreateKeyboardEvent func(source uintptr, virtualKey uint16, keyDown bool) uintptr
	cgEventKeyboardSetUnicode  func(event uintptr, stringLength int, unicodeString *uint16)
	cgEventPost                func(tap uint32, event uintptr)
	cgEventSetFlags            func(event uintptr, flags uint64)
	cgEventCreate              func(source uintptr) uintptr
	cgEventGetLocation         func(event uintptr) CGPoint
	cgMainDisplayID            func() uint32
	cgDisplayBounds            func(displayID uint32) CGRect
	cfRelease                  func(ref uintptr)
	axIsProcessTrusted         func() bool
}

var (
	darwinSim     *darwinSimulator
	darwinSimOnce sync.Once
)

func getDarwinSimulator() *darwinSimulator {
	darwinSimOnce.Do(func() {
		darwinSim = &darwinSimulator{}
		darwinSim.init()
	})
	return darwinSim
}

func newPlatformSimulator() KeySimulator {
	return getDarwinSimulator()
}

func (d *darwinSimulator) init() {
	cg, err := purego.Dlopen("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics", purego.RTLD_GLOBAL|purego.RTLD_NOW)
	if err != nil {
		d.initErr = fmt.Errorf("failed to dlopen CoreGraphics: %w", err)
		return
	}
	d.cgHandle = cg

	as, err := purego.Dlopen("/System/Library/Frameworks/ApplicationServices.framework/ApplicationServices", purego.RTLD_GLOBAL|purego.RTLD_NOW)
	if err == nil {
		d.asHandle = as
		purego.RegisterLibFunc(&d.axIsProcessTrusted, as, "AXIsProcessTrusted")
	}

	cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_GLOBAL|purego.RTLD_NOW)
	if err == nil {
		d.cfHandle = cf
		purego.RegisterLibFunc(&d.cfRelease, cf, "CFRelease")
	}

	purego.RegisterLibFunc(&d.cgEventCreateKeyboardEvent, cg, "CGEventCreateKeyboardEvent")
	purego.RegisterLibFunc(&d.cgEventKeyboardSetUnicode, cg, "CGEventKeyboardSetUnicodeString")
	purego.RegisterLibFunc(&d.cgEventPost, cg, "CGEventPost")
	purego.RegisterLibFunc(&d.cgEventSetFlags, cg, "CGEventSetFlags")
	purego.RegisterLibFunc(&d.cgEventCreate, cg, "CGEventCreate")
	purego.RegisterLibFunc(&d.cgEventGetLocation, cg, "CGEventGetLocation")
	purego.RegisterLibFunc(&d.cgMainDisplayID, cg, "CGMainDisplayID")
	purego.RegisterLibFunc(&d.cgDisplayBounds, cg, "CGDisplayBounds")

	d.initialized = true
}

func (d *darwinSimulator) WriteRune(r rune) error {
	if !d.initialized {
		if d.initErr != nil {
			return d.initErr
		}
		return fmt.Errorf("macOS CoreGraphics simulator not initialized")
	}

	if r <= 0xFFFF {
		u := uint16(r)

		down := d.cgEventCreateKeyboardEvent(0, 0, true)
		if down == 0 {
			return fmt.Errorf("failed to create key down event")
		}
		d.cgEventKeyboardSetUnicode(down, 1, &u)
		d.cgEventPost(kCGHIDEventTap, down)
		if d.cfRelease != nil {
			d.cfRelease(down)
		}

		up := d.cgEventCreateKeyboardEvent(0, 0, false)
		if up == 0 {
			return fmt.Errorf("failed to create key up event")
		}
		d.cgEventKeyboardSetUnicode(up, 1, &u)
		d.cgEventPost(kCGHIDEventTap, up)
		if d.cfRelease != nil {
			d.cfRelease(up)
		}

		return nil
	}

	// Surrogate pair for runes > 0xFFFF (emojis, etc.)
	r1, r2 := utf16.EncodeRune(r)
	chars := []uint16{uint16(r1), uint16(r2)}

	down := d.cgEventCreateKeyboardEvent(0, 0, true)
	if down == 0 {
		return fmt.Errorf("failed to create key down event")
	}
	d.cgEventKeyboardSetUnicode(down, 2, &chars[0])
	d.cgEventPost(kCGHIDEventTap, down)
	if d.cfRelease != nil {
		d.cfRelease(down)
	}

	up := d.cgEventCreateKeyboardEvent(0, 0, false)
	if up == 0 {
		return fmt.Errorf("failed to create key up event")
	}
	d.cgEventKeyboardSetUnicode(up, 2, &chars[0])
	d.cgEventPost(kCGHIDEventTap, up)
	if d.cfRelease != nil {
		d.cfRelease(up)
	}

	return nil
}

func (d *darwinSimulator) postKey(vk uint16) error {
	if !d.initialized {
		return fmt.Errorf("macOS simulator not initialized")
	}

	down := d.cgEventCreateKeyboardEvent(0, vk, true)
	if down == 0 {
		return fmt.Errorf("failed to create key down event for vk %d", vk)
	}
	d.cgEventPost(kCGHIDEventTap, down)
	if d.cfRelease != nil {
		d.cfRelease(down)
	}

	up := d.cgEventCreateKeyboardEvent(0, vk, false)
	if up == 0 {
		return fmt.Errorf("failed to create key up event for vk %d", vk)
	}
	d.cgEventPost(kCGHIDEventTap, up)
	if d.cfRelease != nil {
		d.cfRelease(up)
	}

	return nil
}

func (d *darwinSimulator) PressEnter() error {
	return d.postKey(vkMacReturn)
}

func (d *darwinSimulator) PressTab() error {
	return d.postKey(vkMacTab)
}

func (d *darwinSimulator) PressBackspace() error {
	return d.postKey(vkMacBackspace)
}

func (d *darwinSimulator) PasteText(text string) error {
	// 1. Copy to macOS clipboard using pbcopy
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy text to macOS clipboard (pbcopy): %w", err)
	}

	time.Sleep(20 * time.Millisecond)

	// 2. Press Cmd + V
	if !d.initialized {
		// Fallback to AppleScript keystroke if CoreGraphics unavailable
		return exec.Command("osascript", "-e", `tell application "System Events" to keystroke "v" using command down`).Run()
	}

	down := d.cgEventCreateKeyboardEvent(0, vkMacV, true)
	if down != 0 {
		d.cgEventSetFlags(down, kCGEventFlagMaskCommand)
		d.cgEventPost(kCGHIDEventTap, down)
		if d.cfRelease != nil {
			d.cfRelease(down)
		}
	}

	time.Sleep(10 * time.Millisecond)

	up := d.cgEventCreateKeyboardEvent(0, vkMacV, false)
	if up != 0 {
		d.cgEventSetFlags(up, kCGEventFlagMaskCommand)
		d.cgEventPost(kCGHIDEventTap, up)
		if d.cfRelease != nil {
			d.cfRelease(up)
		}
	}

	return nil
}

func (d *darwinSimulator) CheckFailsafe() bool {
	if !d.initialized || d.cgEventCreate == nil || d.cgEventGetLocation == nil || d.cgDisplayBounds == nil || d.cgMainDisplayID == nil {
		return false
	}

	evt := d.cgEventCreate(0)
	if evt == 0 {
		return false
	}
	if d.cfRelease != nil {
		defer d.cfRelease(evt)
	}

	loc := d.cgEventGetLocation(evt)
	disp := d.cgMainDisplayID()
	bounds := d.cgDisplayBounds(disp)

	const cornerThreshold = 10.0
	inTopLeft := loc.X <= (bounds.Origin.X+cornerThreshold) && loc.Y <= (bounds.Origin.Y+cornerThreshold)
	inTopRight := loc.X >= (bounds.Origin.X+bounds.Size.Width-cornerThreshold) && loc.Y <= (bounds.Origin.Y+cornerThreshold)
	inBottomLeft := loc.X <= (bounds.Origin.X+cornerThreshold) && loc.Y >= (bounds.Origin.Y+bounds.Size.Height-cornerThreshold)
	inBottomRight := loc.X >= (bounds.Origin.X+bounds.Size.Width-cornerThreshold) && loc.Y >= (bounds.Origin.Y+bounds.Size.Height-cornerThreshold)

	return inTopLeft || inTopRight || inBottomLeft || inBottomRight
}

func checkAccessibilityPlatform() bool {
	sim := getDarwinSimulator()
	if sim != nil && sim.axIsProcessTrusted != nil {
		return sim.axIsProcessTrusted()
	}
	return false
}

func requestAccessibilityPlatform() {
	// 1. Trigger AppleScript System Events to cause macOS to display the native Accessibility prompt
	go func() {
		_ = exec.Command("osascript", "-e", `tell application "System Events" to key code 0`).Run()
	}()

	// 2. Open macOS System Settings directly to Privacy & Security -> Accessibility
	go func() {
		time.Sleep(300 * time.Millisecond)
		_ = exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility").Start()
	}()

	log.Println("📢 Prompted macOS Accessibility permission dialog and opened System Settings.")
}
