package keypress

import (
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Handler processes keyboard input with advanced features including:
// - Kitty protocol support
// - Paste mode detection
// - Drag-and-drop file path detection
// - Backslash-enter line continuation
type Handler struct {
	// Configuration
	kittyEnabled   bool
	debugKeystroke bool
	onPaste        PasteHandler
	onKey          KeyHandler

	// State
	mu                            sync.Mutex
	pasteMode                     bool
	pasteBuffer                   strings.Builder
	dragging                      bool
	dragBuffer                    strings.Builder
	dragTimer                     *time.Timer
	kittySequenceBuffer           string
	waitingForEnterAfterBackslash bool
	backslashTimer                *time.Timer

	// Parser
	kittyParser *KittyParser
}

// HandlerConfig configures the keypress handler
type HandlerConfig struct {
	// Enable Kitty keyboard protocol parsing
	KittyEnabled bool

	// Enable debug logging for keystrokes
	DebugKeystroke bool

	// Callback for paste events
	OnPaste PasteHandler

	// Callback for key events (return true to consume the event)
	OnKey KeyHandler
}

// NewHandler creates a new enhanced keypress handler
func NewHandler(config HandlerConfig) *Handler {
	return &Handler{
		kittyEnabled:   config.KittyEnabled,
		debugKeystroke: config.DebugKeystroke,
		onPaste:        config.OnPaste,
		onKey:          config.OnKey,
		kittyParser:    NewKittyParser(),
	}
}

// ProcessTeaKey processes a bubbletea KeyMsg and returns an enhanced Key event
// Returns nil if the event should be ignored (e.g., paste markers, focus events)
func (h *Handler) ProcessTeaKey(msg tea.KeyMsg) *Key {
	h.mu.Lock()
	defer h.mu.Unlock()

	sequence := msg.String()

	// Ignore focus events
	if sequence == FocusIn || sequence == FocusOut {
		return nil
	}

	// Check for paste mode markers
	if strings.HasPrefix(sequence, PastePrefix) {
		h.enterPasteMode()
		return nil
	}

	if strings.HasPrefix(sequence, PasteSuffix) {
		h.exitPasteMode()
		return nil
	}

	// If in paste mode, accumulate to paste buffer
	if h.pasteMode {
		h.pasteBuffer.WriteString(sequence)
		return nil
	}

	// Check for drag start (quoted strings)
	if sequence == SingleQuote || sequence == DoubleQuote || h.dragging {
		h.handleDragInput(sequence)
		return nil
	}

	// Handle backslash-enter detection for line continuation
	if key := h.handleBackslashEnter(msg, sequence); key != nil {
		return key
	}

	// Check for Ctrl+C (handle both regular and Kitty protocol)
	if isCtrlC(msg) || sequence == ESC+KittyCtrlC {
		// Clear any Kitty buffer on Ctrl+C
		h.kittySequenceBuffer = ""

		if sequence == ESC+KittyCtrlC {
			return &Key{
				Name:          "c",
				Ctrl:          true,
				Alt:           false,
				Shift:         false,
				Paste:         false,
				Sequence:      sequence,
				KittyProtocol: true,
				Original:      &msg,
			}
		}

		return &Key{
			Name:     "c",
			Ctrl:     true,
			Sequence: sequence,
			Original: &msg,
		}
	}

	// Try Kitty protocol parsing
	if h.kittyEnabled {
		if key := h.handleKittySequence(msg, sequence); key != nil {
			return key
		}
	}

	// Convert regular bubbletea KeyMsg to our Key
	return h.teaKeyToKey(msg)
}

// enterPasteMode starts paste mode
func (h *Handler) enterPasteMode() {
	h.pasteMode = true
	h.pasteBuffer.Reset()
}

// exitPasteMode ends paste mode and triggers the paste callback
func (h *Handler) exitPasteMode() {
	h.pasteMode = false
	content := h.pasteBuffer.String()
	h.pasteBuffer.Reset()

	if h.onPaste != nil && content != "" {
		h.onPaste(content)
	}
}

// handleDragInput handles drag-and-drop file path detection
func (h *Handler) handleDragInput(sequence string) {
	h.dragging = true
	h.dragBuffer.WriteString(sequence)

	// Reset timer
	if h.dragTimer != nil {
		h.dragTimer.Stop()
	}

	h.dragTimer = time.AfterFunc(DragCompletionTimeoutMs, func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		h.dragging = false
		content := h.dragBuffer.String()
		h.dragBuffer.Reset()

		if h.onPaste != nil && content != "" {
			h.onPaste(content)
		}
	})
}

// handleBackslashEnter detects backslash followed by enter for line continuation
func (h *Handler) handleBackslashEnter(msg tea.KeyMsg, sequence string) *Key {
	// If waiting for enter after backslash
	if h.waitingForEnterAfterBackslash {
		if msg.Type == tea.KeyEnter {
			// Cancel timer
			if h.backslashTimer != nil {
				h.backslashTimer.Stop()
				h.backslashTimer = nil
			}
			h.waitingForEnterAfterBackslash = false

			// Return Shift+Enter to indicate line continuation
			return &Key{
				Name:     "return",
				Shift:    true,
				Sequence: "\r",
				Original: &msg,
			}
		}

		// Any other key cancels the backslash wait
		if h.backslashTimer != nil {
			h.backslashTimer.Stop()
			h.backslashTimer = nil
		}
		h.waitingForEnterAfterBackslash = false

		// Emit the backslash first
		if h.onKey != nil {
			h.onKey(Key{
				Name:     "",
				Sequence: "\\",
			})
		}

		// Then process the current key normally
		return h.teaKeyToKey(msg)
	}

	// Check if this is a backslash
	if sequence == "\\" && msg.Type == tea.KeyRunes {
		h.waitingForEnterAfterBackslash = true
		h.backslashTimer = time.AfterFunc(BackslashEnterDetectionWindowMs, func() {
			h.mu.Lock()
			defer h.mu.Unlock()

			h.waitingForEnterAfterBackslash = false
			h.backslashTimer = nil

			// Emit the backslash
			if h.onKey != nil {
				h.onKey(Key{
					Name:     "",
					Sequence: "\\",
				})
			}
		})

		return nil // Don't emit the backslash yet
	}

	return nil
}

// handleKittySequence processes Kitty protocol sequences
func (h *Handler) handleKittySequence(msg tea.KeyMsg, sequence string) *Key {
	// Check if this looks like a Kitty/CSI sequence
	csiStart := ESC + "["
	if !strings.HasPrefix(h.kittySequenceBuffer, csiStart) &&
		!strings.HasPrefix(sequence, csiStart) {
		return nil
	}

	// Don't treat paste/focus sequences as Kitty
	if strings.HasPrefix(sequence, PastePrefix) ||
		strings.HasPrefix(sequence, PasteSuffix) ||
		strings.HasPrefix(sequence, FocusIn) ||
		strings.HasPrefix(sequence, FocusOut) {
		return nil
	}

	// Accumulate to buffer
	h.kittySequenceBuffer += sequence

	// Try to parse the buffer
	keys, remaining := h.kittyParser.ParseBuffer(h.kittySequenceBuffer)
	h.kittySequenceBuffer = remaining

	// If we parsed any keys, return the first one
	if len(keys) > 0 {
		key := keys[0]
		key.Original = &msg

		// Emit remaining keys via callback
		for i := 1; i < len(keys); i++ {
			if h.onKey != nil {
				h.onKey(keys[i])
			}
		}

		return &key
	}

	// Check for buffer overflow
	if len(h.kittySequenceBuffer) > MaxKittySequenceLength {
		if h.debugKeystroke {
			// Log overflow (in real app, use proper logging)
		}
		h.kittySequenceBuffer = ""
		return nil
	}

	// Still accumulating, don't emit anything yet
	return nil
}

// teaKeyToKey converts a bubbletea KeyMsg to our Key type
func (h *Handler) teaKeyToKey(msg tea.KeyMsg) *Key {
	key := &Key{
		Alt:      msg.Alt,
		Sequence: msg.String(),
		Original: &msg,
	}

	// Map key type
	switch msg.Type {
	case tea.KeyEnter:
		key.Name = "return"
	case tea.KeyEscape:
		key.Name = "escape"
	case tea.KeyBackspace:
		key.Name = "backspace"
	case tea.KeyDelete:
		key.Name = "delete"
	case tea.KeyUp:
		key.Name = "up"
	case tea.KeyDown:
		key.Name = "down"
	case tea.KeyLeft:
		key.Name = "left"
	case tea.KeyRight:
		key.Name = "right"
	case tea.KeyHome:
		key.Name = "home"
	case tea.KeyEnd:
		key.Name = "end"
	case tea.KeyPgUp:
		key.Name = "pageup"
	case tea.KeyPgDown:
		key.Name = "pagedown"
	case tea.KeyTab:
		key.Name = "tab"
	case tea.KeySpace:
		key.Name = "space"
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			key.Name = string(msg.Runes[0])
		} else {
			key.Sequence = string(msg.Runes)
		}
	case tea.KeyCtrlC:
		key.Name = "c"
		key.Ctrl = true
	default:
		// Try to extract from string
		s := msg.String()
		if strings.HasPrefix(s, "ctrl+") {
			key.Ctrl = true
			key.Name = strings.TrimPrefix(s, "ctrl+")
		} else if strings.HasPrefix(s, "alt+") {
			key.Alt = true
			key.Name = strings.TrimPrefix(s, "alt+")
		} else {
			key.Name = s
		}
	}

	// Handle Meta+Enter (ESC followed by Enter)
	if msg.Type == tea.KeyEnter && msg.String() == ESC+"\r" {
		key.Alt = true
	}

	return key
}

// isCtrlC checks if a KeyMsg is Ctrl+C
func isCtrlC(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyCtrlC ||
		(msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'c' && msg.String() == "ctrl+c")
}

// Cleanup releases resources (timers, etc.)
func (h *Handler) Cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.dragTimer != nil {
		h.dragTimer.Stop()
		h.dragTimer = nil
	}

	if h.backslashTimer != nil {
		h.backslashTimer.Stop()
		h.backslashTimer = nil
	}

	// Flush any pending paste data
	if h.pasteMode {
		h.exitPasteMode()
	}

	// Flush any pending drag data
	if h.dragging {
		h.dragging = false
		content := h.dragBuffer.String()
		h.dragBuffer.Reset()

		if h.onPaste != nil && content != "" {
			h.onPaste(content)
		}
	}
}
