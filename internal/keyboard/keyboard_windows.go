//go:build windows

package keyboard

import (
	"fmt"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procSendInput        = user32.NewProc("SendInput")
	procGetCursorPos     = user32.NewProc("GetCursorPos")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procOpenClipboard    = user32.NewProc("OpenClipboard")
	procEmptyClipboard   = user32.NewProc("EmptyClipboard")
	procSetClipboardData = user32.NewProc("SetClipboardData")
	procCloseClipboard   = user32.NewProc("CloseClipboard")

	procGlobalAlloc   = kernel32.NewProc("GlobalAlloc")
	procGlobalLock    = kernel32.NewProc("GlobalLock")
	procGlobalUnlock  = kernel32.NewProc("GlobalUnlock")
	procRtlMoveMemory = kernel32.NewProc("RtlMoveMemory")
)

const (
	inputKeyboard = 1

	keyeventfExtendedKey = 0x0001
	keyeventfKeyUp       = 0x0002
	keyeventfUnicode     = 0x0004

	vkBack    = 0x08
	vkTab     = 0x09
	vkReturn  = 0x0D
	vkControl = 0x11
	vkV       = 0x56

	smCxScreen = 0
	smCyScreen = 1

	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

type keybdInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type inputStruct struct {
	inputType uint32
	_         uint32 // Padding for 64-bit alignment
	ki        keybdInput
	_         [8]byte // Union padding
}

type point struct {
	x int32
	y int32
}

type windowsSimulator struct{}

func newPlatformSimulator() KeySimulator {
	return &windowsSimulator{}
}

func sendInputs(inputs []inputStruct) error {
	if len(inputs) == 0 {
		return nil
	}
	n := uint32(len(inputs))
	cbSize := int32(unsafe.Sizeof(inputs[0]))
	ret, _, err := procSendInput.Call(
		uintptr(n),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(cbSize),
	)
	if ret != uintptr(n) {
		return fmt.Errorf("SendInput failed (sent %d of %d): %v", ret, n, err)
	}
	return nil
}

func (w *windowsSimulator) WriteRune(r rune) error {
	if r <= 0xFFFF {
		inputs := []inputStruct{
			{
				inputType: inputKeyboard,
				ki: keybdInput{
					wScan:   uint16(r),
					dwFlags: keyeventfUnicode,
				},
			},
			{
				inputType: inputKeyboard,
				ki: keybdInput{
					wScan:   uint16(r),
					dwFlags: keyeventfUnicode | keyeventfKeyUp,
				},
			},
		}
		return sendInputs(inputs)
	}

	// Surrogate pair for runes > 0xFFFF (e.g. emojis)
	r1, r2 := utf16.EncodeRune(r)
	inputs := []inputStruct{
		{
			inputType: inputKeyboard,
			ki: keybdInput{
				wScan:   uint16(r1),
				dwFlags: keyeventfUnicode,
			},
		},
		{
			inputType: inputKeyboard,
			ki: keybdInput{
				wScan:   uint16(r1),
				dwFlags: keyeventfUnicode | keyeventfKeyUp,
			},
		},
		{
			inputType: inputKeyboard,
			ki: keybdInput{
				wScan:   uint16(r2),
				dwFlags: keyeventfUnicode,
			},
		},
		{
			inputType: inputKeyboard,
			ki: keybdInput{
				wScan:   uint16(r2),
				dwFlags: keyeventfUnicode | keyeventfKeyUp,
			},
		},
	}
	return sendInputs(inputs)
}

func (w *windowsSimulator) PressEnter() error {
	inputs := []inputStruct{
		{
			inputType: inputKeyboard,
			ki:        keybdInput{wVk: vkReturn},
		},
		{
			inputType: inputKeyboard,
			ki:        keybdInput{wVk: vkReturn, dwFlags: keyeventfKeyUp},
		},
	}
	return sendInputs(inputs)
}

func (w *windowsSimulator) PressTab() error {
	inputs := []inputStruct{
		{
			inputType: inputKeyboard,
			ki:        keybdInput{wVk: vkTab},
		},
		{
			inputType: inputKeyboard,
			ki:        keybdInput{wVk: vkTab, dwFlags: keyeventfKeyUp},
		},
	}
	return sendInputs(inputs)
}

func (w *windowsSimulator) PressBackspace() error {
	inputs := []inputStruct{
		{
			inputType: inputKeyboard,
			ki:        keybdInput{wVk: vkBack},
		},
		{
			inputType: inputKeyboard,
			ki:        keybdInput{wVk: vkBack, dwFlags: keyeventfKeyUp},
		},
	}
	return sendInputs(inputs)
}

func (w *windowsSimulator) PasteText(text string) error {
	// 1. Copy text to Windows Clipboard
	if err := setClipboardText(text); err != nil {
		return fmt.Errorf("clipboard copy error: %w", err)
	}

	time.Sleep(20 * time.Millisecond)

	// 2. Trigger Ctrl + V
	inputs := []inputStruct{
		{
			inputType: inputKeyboard,
			ki:        keybdInput{wVk: vkControl},
		},
		{
			inputType: inputKeyboard,
			ki:        keybdInput{wVk: vkV},
		},
		{
			inputType: inputKeyboard,
			ki:        keybdInput{wVk: vkV, dwFlags: keyeventfKeyUp},
		},
		{
			inputType: inputKeyboard,
			ki:        keybdInput{wVk: vkControl, dwFlags: keyeventfKeyUp},
		},
	}
	return sendInputs(inputs)
}

func (w *windowsSimulator) CheckFailsafe() bool {
	var pt point
	ret, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if ret == 0 {
		return false
	}

	wW, _, _ := procGetSystemMetrics.Call(uintptr(smCxScreen))
	hH, _, _ := procGetSystemMetrics.Call(uintptr(smCyScreen))
	width := int32(wW)
	height := int32(hH)

	const cornerThreshold = 10
	inTopLeft := pt.x <= cornerThreshold && pt.y <= cornerThreshold
	inTopRight := pt.x >= width-cornerThreshold && pt.y <= cornerThreshold
	inBottomLeft := pt.x <= cornerThreshold && pt.y >= height-cornerThreshold
	inBottomRight := pt.x >= width-cornerThreshold && pt.y >= height-cornerThreshold

	return inTopLeft || inTopRight || inBottomLeft || inBottomRight
}

func setClipboardText(text string) error {
	ret, _, _ := procOpenClipboard.Call(0)
	if ret == 0 {
		return fmt.Errorf("failed to open clipboard")
	}
	defer procCloseClipboard.Call()

	procEmptyClipboard.Call()

	utf16Data := utf16.Encode([]rune(text + "\x00"))
	size := uintptr(len(utf16Data) * 2)

	hMem, _, _ := procGlobalAlloc.Call(uintptr(gmemMoveable), size)
	if hMem == 0 {
		return fmt.Errorf("GlobalAlloc failed")
	}

	ptr, _, _ := procGlobalLock.Call(hMem)
	if ptr == 0 {
		return fmt.Errorf("GlobalLock failed")
	}

	procRtlMoveMemory.Call(ptr, uintptr(unsafe.Pointer(&utf16Data[0])), size)
	procGlobalUnlock.Call(hMem)

	res, _, err := procSetClipboardData.Call(uintptr(cfUnicodeText), hMem)
	if res == 0 {
		return fmt.Errorf("SetClipboardData failed: %v", err)
	}
	return nil
}

func checkAccessibilityPlatform() bool {
	return true
}

func requestAccessibilityPlatform() {
	// No-op on Windows
}
