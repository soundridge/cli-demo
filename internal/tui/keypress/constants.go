package keypress

import "time"

// Terminal escape sequences and special characters
const (
	ESC = "\u001B"

	// Paste mode sequences
	PastePrefix = ESC + "[200~"
	PasteSuffix = ESC + "[201~"

	// Focus sequences
	FocusIn  = ESC + "[I"
	FocusOut = ESC + "[O"

	// Quote characters for drag detection
	SingleQuote = "'"
	DoubleQuote = "\""
)

// Character codes
const (
	CharCodeESC   = 27
	CharCodeSpace = 32
	CharCodeTab   = 9
)

// Kitty keyboard protocol constants
const (
	// Kitty key codes
	KittyKeycodeTab         = 9
	KittyKeycodeBackspace   = 127
	KittyKeycodeEnter       = 13
	KittyKeycodeNumpadEnter = 13

	// Kitty special sequences
	KittyCtrlC = "[99;5u" // Ctrl+C in Kitty protocol

	// Kitty modifier constants
	KittyModifierBase             = 1
	KittyModifierEventTypesOffset = 128
	ModifierShiftBit              = 1
	ModifierAltBit                = 2
	ModifierCtrlBit               = 4
	ModifierSuperBit              = 8
	ModifierHyperBit              = 16
	ModifierMetaBit               = 32
	ModifierCapsLockBit           = 64
	ModifierNumLockBit            = 128

	// Maximum kitty sequence length before overflow
	MaxKittySequenceLength = 256
)

// Timeout constants
const (
	// Timeout for detecting backslash+enter for line continuation
	BackslashEnterDetectionWindowMs = 50 * time.Millisecond

	// Timeout for detecting end of drag-and-drop file paths
	DragCompletionTimeoutMs = 100 * time.Millisecond
)
