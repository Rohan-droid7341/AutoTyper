package state

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unicode"

	"autotyper/internal/keyboard"
)

// Engine orchestrates keyboard typing, settings, and multi-device coordination.
type Engine struct {
	mu           sync.RWMutex
	settings     Settings
	simulator    keyboard.KeySimulator
	broadcaster  func(msg WSMessage)
	cancelActive context.CancelFunc
	isPunching   bool
}

// NewEngine creates a new typing engine.
func NewEngine(broadcaster func(msg WSMessage)) *Engine {
	return &Engine{
		settings: Settings{
			Text:               "",
			WPM:                100,
			FocusDelaySec:      3,
			PreserveFormatting: true,
			SmartIndent:        true,
			PasteMode:          false,
			JitterMode:         false,
			OopsMode:           false,
		},
		simulator:   keyboard.New(),
		broadcaster: broadcaster,
	}
}

// GetSettings returns a copy of current shared settings.
func (e *Engine) GetSettings() Settings {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.settings
}

// UpdateSettings updates the shared state and broadcasts changes.
func (e *Engine) UpdateSettings(msg WSMessage, senderIP string) {
	e.mu.Lock()
	if msg.Text != "" || msg.Type == "update_code" {
		e.settings.Text = msg.Text
	}
	if msg.WPM > 0 {
		e.settings.WPM = msg.WPM
	}
	if msg.FocusDelaySec >= 0 {
		e.settings.FocusDelaySec = msg.FocusDelaySec
	}
	if msg.PreserveFormatting != nil {
		e.settings.PreserveFormatting = *msg.PreserveFormatting
	}
	if msg.SmartIndent != nil {
		e.settings.SmartIndent = *msg.SmartIndent
	}
	if msg.PasteMode != nil {
		e.settings.PasteMode = *msg.PasteMode
	}
	if msg.JitterMode != nil {
		e.settings.JitterMode = *msg.JitterMode
	}
	if msg.OopsMode != nil {
		e.settings.OopsMode = *msg.OopsMode
	}
	snapshot := e.settings
	e.mu.Unlock()

	log.Printf("🔄 Code updated (%d chars) from %s", len(snapshot.Text), senderIP)

	e.broadcaster(WSMessage{
		Type:               "code_updated",
		Text:               snapshot.Text,
		WPM:                snapshot.WPM,
		FocusDelaySec:      snapshot.FocusDelaySec,
		PreserveFormatting: &snapshot.PreserveFormatting,
		SmartIndent:        &snapshot.SmartIndent,
		PasteMode:          &snapshot.PasteMode,
		JitterMode:         &snapshot.JitterMode,
		OopsMode:           &snapshot.OopsMode,
		Sender:             senderIP,
	})
}

// Stop halts any active punching job.
func (e *Engine) Stop() {
	e.mu.Lock()
	if e.cancelActive != nil {
		e.cancelActive()
		e.cancelActive = nil
	}
	e.isPunching = false
	e.mu.Unlock()

	e.broadcaster(WSMessage{Type: "stopped"})
	log.Println("⏹ Punching stopped by user")
}

// StartPunch starts the typing job in background.
func (e *Engine) StartPunch(msg WSMessage, senderIP string) error {
	e.mu.Lock()
	if e.cancelActive != nil {
		e.cancelActive()
	}

	// Update current settings with the punch payload
	if msg.Text != "" {
		e.settings.Text = msg.Text
	}
	if msg.WPM > 0 {
		e.settings.WPM = msg.WPM
	}
	if msg.PreserveFormatting != nil {
		e.settings.PreserveFormatting = *msg.PreserveFormatting
	}
	if msg.SmartIndent != nil {
		e.settings.SmartIndent = *msg.SmartIndent
	}
	if msg.PasteMode != nil {
		e.settings.PasteMode = *msg.PasteMode
	}
	if msg.JitterMode != nil {
		e.settings.JitterMode = *msg.JitterMode
	}
	if msg.OopsMode != nil {
		e.settings.OopsMode = *msg.OopsMode
	}

	cfg := e.settings
	ctx, cancel := context.WithCancel(context.Background())
	e.cancelActive = cancel
	e.isPunching = true
	e.mu.Unlock()

	if strings.TrimSpace(cfg.Text) == "" {
		cancel()
		return fmt.Errorf("no text to punch")
	}

	go e.runPunch(ctx, cfg, senderIP)
	return nil
}

func (e *Engine) runPunch(ctx context.Context, cfg Settings, senderIP string) {
	defer func() {
		e.mu.Lock()
		e.isPunching = false
		e.mu.Unlock()
	}()

	text := cfg.Text
	totalRunes := len([]rune(text))

	// Fast Paste Mode
	if cfg.PasteMode {
		log.Printf("📋 Fast pasting %d chars via clipboard (triggered by %s)", totalRunes, senderIP)
		if err := e.simulator.PasteText(text); err != nil {
			e.broadcaster(WSMessage{Type: "error", Message: err.Error()})
			return
		}
		e.broadcaster(WSMessage{Type: "progress", Current: totalRunes, Total: totalRunes})
		e.broadcaster(WSMessage{Type: "complete"})
		log.Println("✅ Paste complete!")
		return
	}

	baseDelayMs := 12000 / cfg.WPM
	if baseDelayMs < 1 {
		baseDelayMs = 1
	}

	var modes []string
	if cfg.SmartIndent {
		modes = append(modes, "SMART_INDENT")
	}
	if cfg.JitterMode {
		modes = append(modes, "JITTER")
	}
	if cfg.OopsMode {
		modes = append(modes, "OOPS")
	}
	modesStr := ""
	if len(modes) > 0 {
		modesStr = fmt.Sprintf(" [%s]", strings.Join(modes, "+"))
	}
	log.Printf("⌨️  Typing %d chars at %d WPM (base delay %dms)%s", totalRunes, cfg.WPM, baseDelayMs, modesStr)

	// Smart IDE Auto-Indent Mode
	if cfg.SmartIndent && cfg.PreserveFormatting {
		e.typeSmartIndent(ctx, cfg, baseDelayMs, totalRunes)
	} else {
		e.typeNormal(ctx, cfg, baseDelayMs, totalRunes)
	}
}

func (e *Engine) typeSmartIndent(ctx context.Context, cfg Settings, baseDelayMs, totalRunes int) {
	lines := strings.Split(strings.ReplaceAll(cfg.Text, "\r\n", "\n"), "\n")
	curIDEIndent := 0
	const tabSize = 4
	typedRunes := 0

	for lineIdx, line := range lines {
		select {
		case <-ctx.Done():
			e.broadcaster(WSMessage{Type: "stopped"})
			return
		default:
		}

		if e.simulator.CheckFailsafe() {
			log.Println("🛡️  Emergency stop: Mouse moved to screen corner!")
			e.broadcaster(WSMessage{Type: "error", Message: "Emergency failsafe triggered (mouse in corner)"})
			return
		}

		strippedLine := strings.TrimLeft(line, " \t")
		leadingWS := line[:len(line)-len(strippedLine)]
		targetIndent := 0
		for _, ch := range leadingWS {
			if ch == '\t' {
				targetIndent += tabSize
			} else {
				targetIndent++
			}
		}

		if lineIdx == 0 {
			if len(leadingWS) > 0 {
				for _, wsChar := range leadingWS {
					if wsChar == '\t' {
						_ = e.simulator.PressTab()
					} else {
						_ = e.simulator.WriteRune(wsChar)
					}
				}
				curIDEIndent = targetIndent
				typedRunes += len([]rune(leadingWS))
			}
		} else {
			_ = e.simulator.PressEnter()
			typedRunes++

			// Wait for IDE auto-indent calculation
			time.Sleep(35 * time.Millisecond)

			prevLine := lines[lineIdx-1]
			cleanPrev := strings.TrimRight(strings.Split(strings.Split(prevLine, "#")[0], "//")[0], " \t")
			if strings.HasSuffix(cleanPrev, ":") || strings.HasSuffix(cleanPrev, "{") {
				curIDEIndent += tabSize
			}

			if len(strippedLine) > 0 {
				if targetIndent < curIDEIndent {
					dedents := (curIDEIndent - targetIndent) / tabSize
					remSpaces := (curIDEIndent - targetIndent) % tabSize
					for i := 0; i < dedents; i++ {
						time.Sleep(25 * time.Millisecond)
						_ = e.simulator.PressBackspace()
					}
					for i := 0; i < remSpaces; i++ {
						time.Sleep(15 * time.Millisecond)
						_ = e.simulator.PressBackspace()
					}
					curIDEIndent = targetIndent
				} else if targetIndent > curIDEIndent {
					extra := targetIndent - curIDEIndent
					for i := 0; i < extra; i++ {
						_ = e.simulator.WriteRune(' ')
					}
					curIDEIndent = targetIndent
				}
			}

			typedRunes += len([]rune(leadingWS))
			e.broadcaster(WSMessage{Type: "progress", Current: min(totalRunes, typedRunes), Total: totalRunes})
		}

		// Type characters in the line
		for _, ch := range strippedLine {
			select {
			case <-ctx.Done():
				e.broadcaster(WSMessage{Type: "stopped"})
				return
			default:
			}

			if e.simulator.CheckFailsafe() {
				log.Println("🛡️  Emergency stop: Mouse in corner!")
				e.broadcaster(WSMessage{Type: "error", Message: "Emergency failsafe triggered"})
				return
			}

			e.typeSingleRune(ch, cfg.OopsMode)

			typedRunes++
			if typedRunes%5 == 0 || typedRunes == totalRunes {
				e.broadcaster(WSMessage{Type: "progress", Current: min(totalRunes, typedRunes), Total: totalRunes})
			}

			e.applyDelay(baseDelayMs, cfg.JitterMode)
		}
	}

	e.broadcaster(WSMessage{Type: "complete"})
	log.Println("✅ Typing complete!")
}

func (e *Engine) typeNormal(ctx context.Context, cfg Settings, baseDelayMs, totalRunes int) {
	runes := []rune(cfg.Text)
	for i, r := range runes {
		select {
		case <-ctx.Done():
			e.broadcaster(WSMessage{Type: "stopped"})
			return
		default:
		}

		if e.simulator.CheckFailsafe() {
			log.Println("🛡️  Emergency stop: Mouse in corner!")
			e.broadcaster(WSMessage{Type: "error", Message: "Emergency failsafe triggered"})
			return
		}

		if r == '\r' {
			continue
		} else if r == '\n' {
			if cfg.PreserveFormatting {
				_ = e.simulator.PressEnter()
			} else {
				_ = e.simulator.WriteRune(' ')
			}
		} else if r == '\t' {
			if cfg.PreserveFormatting {
				_ = e.simulator.PressTab()
			} else {
				for s := 0; s < 4; s++ {
					_ = e.simulator.WriteRune(' ')
				}
			}
		} else {
			e.typeSingleRune(r, cfg.OopsMode)
		}

		current := i + 1
		if current%5 == 0 || current == totalRunes {
			e.broadcaster(WSMessage{Type: "progress", Current: current, Total: totalRunes})
		}

		e.applyDelay(baseDelayMs, cfg.JitterMode)
	}

	e.broadcaster(WSMessage{Type: "complete"})
	log.Println("✅ Typing complete!")
}

func (e *Engine) typeSingleRune(r rune, oopsMode bool) {
	if oopsMode && unicode.IsLetter(r) && rand.Float64() < 0.03 {
		wrong := rune('a' + rand.Intn(26))
		if unicode.IsUpper(r) {
			wrong = unicode.ToUpper(wrong)
		}
		if wrong == r {
			wrong = 'x'
		}

		_ = e.simulator.WriteRune(wrong)
		time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)
		_ = e.simulator.PressBackspace()
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
	}
	_ = e.simulator.WriteRune(r)
}

func (e *Engine) applyDelay(baseDelayMs int, jitterMode bool) {
	if baseDelayMs <= 0 {
		return
	}
	delay := float64(baseDelayMs)
	if jitterMode {
		stddev := delay * 0.4
		jittered := rand.NormFloat64()*stddev + delay
		if jittered < 1.0 {
			jittered = 1.0
		}
		delay = jittered
	}
	time.Sleep(time.Duration(delay * float64(time.Millisecond)))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
