package keyboard

// KeySimulator defines the interface for typing and clipboard simulation.
type KeySimulator interface {
	WriteRune(r rune) error
	PressEnter() error
	PressTab() error
	PressBackspace() error
	PasteText(text string) error
	CheckFailsafe() bool
}

// New returns the platform-specific KeySimulator implementation.
func New() KeySimulator {
	return newPlatformSimulator()
}

// CheckAccessibility returns true if accessibility permission is granted.
// Always returns true on Windows. On macOS, checks AXIsProcessTrusted.
func CheckAccessibility() bool {
	return checkAccessibilityPlatform()
}

// RequestAccessibility prompts macOS accessibility dialog or opens System Settings.
// No-op on Windows.
func RequestAccessibility() {
	requestAccessibilityPlatform()
}
