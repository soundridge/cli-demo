package keypress

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Key represents an enhanced keyboard event with advanced features
// like Kitty protocol support, paste detection, and modifier keys
type Key struct {
	// Name of the key (e.g., "a", "enter", "up", "f1")
	Name string

	// Modifier states
	Ctrl  bool
	Alt   bool // Also known as Meta in some contexts
	Shift bool

	// Special states
	Paste bool // True if this was part of a paste operation

	// Raw sequence that triggered this key event
	Sequence string

	// KittyProtocol indicates if this was parsed from Kitty protocol
	KittyProtocol bool

	// Original bubbletea KeyMsg (if available)
	Original *tea.KeyMsg
}

// String returns the key name for display
func (k Key) String() string {
	return k.Name
}

// IsChar returns true if this is a regular character key
func (k Key) IsChar() bool {
	return len(k.Name) == 1 && !k.Ctrl && !k.Alt
}

// IsCtrlC returns true if this is Ctrl+C
func (k Key) IsCtrlC() bool {
	return k.Ctrl && k.Name == "c"
}

// IsEnter returns true if this is the Enter/Return key
func (k Key) IsEnter() bool {
	return k.Name == "enter" || k.Name == "return"
}

// IsEscape returns true if this is the Escape key
func (k Key) IsEscape() bool {
	return k.Name == "escape" || k.Name == "esc"
}

// IsArrow returns true if this is an arrow key
func (k Key) IsArrow() bool {
	return k.Name == "up" || k.Name == "down" || k.Name == "left" || k.Name == "right"
}

// ToTeaKey converts this Key to a bubbletea KeyMsg
func (k Key) ToTeaKey() tea.KeyMsg {
	if k.Original != nil {
		return *k.Original
	}

	msg := tea.KeyMsg{
		Type: tea.KeyRunes,
	}

	// Map special keys
	switch k.Name {
	case "enter", "return":
		msg.Type = tea.KeyEnter
	case "escape", "esc":
		msg.Type = tea.KeyEscape
	case "backspace":
		msg.Type = tea.KeyBackspace
	case "delete":
		msg.Type = tea.KeyDelete
	case "up":
		msg.Type = tea.KeyUp
	case "down":
		msg.Type = tea.KeyDown
	case "left":
		msg.Type = tea.KeyLeft
	case "right":
		msg.Type = tea.KeyRight
	case "home":
		msg.Type = tea.KeyHome
	case "end":
		msg.Type = tea.KeyEnd
	case "pageup":
		msg.Type = tea.KeyPgUp
	case "pagedown":
		msg.Type = tea.KeyPgDown
	case "tab":
		msg.Type = tea.KeyTab
	case "space":
		msg.Type = tea.KeySpace
	default:
		if len(k.Name) == 1 {
			msg.Runes = []rune(k.Name)
		}
	}

	// Handle modifiers
	if k.Alt {
		msg.Alt = true
	}

	return msg
}

// KeyHandler is a function that handles key events
type KeyHandler func(key Key) bool

// PasteHandler is a function that handles paste events
type PasteHandler func(content string)
