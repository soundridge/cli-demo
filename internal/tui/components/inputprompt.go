package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/boke/cli-demo/internal/tui/keypress"
	"github.com/boke/cli-demo/internal/tui/theme"
	"github.com/boke/cli-demo/internal/tui/utils"
)

// PromptMode represents the different input modes for the prompt
type PromptMode int

const (
	// PromptModeNormal is the default input mode with '> ' prefix
	PromptModeNormal PromptMode = iota

	// PromptModeShell is the shell command mode with '! ' prefix
	PromptModeShell

	// PromptModeReverseSearch is the reverse search mode with '(r:) ' prefix
	PromptModeReverseSearch

	// PromptModeYolo is the YOLO mode with '* ' prefix
	PromptModeYolo
)

// InputPromptConfig holds configuration for the InputPrompt component
type InputPromptConfig struct {
	// Theme is the semantic color theme to use
	Theme theme.SemanticColors

	// Placeholder text to show when input is empty
	Placeholder string

	// TerminalWidth is the raw terminal width (will be adjusted internally)
	TerminalWidth int

	// Mode determines the prompt prefix and behavior
	Mode PromptMode

	// YoloMode enables YOLO mode styling
	YoloMode bool

	// AutoEditMode enables auto-edit mode styling
	AutoEditMode bool

	// ShellModeActive indicates if shell mode is active
	ShellModeActive bool

	// GhostText is the autocomplete suggestion text
	GhostText string

	// ShowCursor controls cursor visibility
	ShowCursor bool

	// EnableKittyProtocol enables Kitty keyboard protocol support
	EnableKittyProtocol bool

	// DebugKeystroke enables debug logging for keystrokes
	DebugKeystroke bool
}

// InputPrompt is a professional-grade input component matching the Gemini CLI specs.
// It implements the exact width calculation, highlighting, and rendering logic
// from the TypeScript/Ink version.
type InputPrompt struct {
	config InputPromptConfig

	// Input state
	value       string
	cursorPos   int
	cursorBlink bool

	// Calculated dimensions
	widths PromptWidths

	// Styles (cached for performance)
	borderStyle      lipgloss.Style
	prefixStyle      lipgloss.Style
	textStyle        lipgloss.Style
	accentStyle      lipgloss.Style
	secondaryStyle   lipgloss.Style
	placeholderStyle lipgloss.Style

	// Cursor model from bubbles
	cursor cursor.Model

	// Focus state
	focused bool

	// Enhanced keypress handler
	keypressHandler *keypress.Handler

	// Paste event handler
	onPasteReceived func(string)
}

// NewInputPrompt creates a new InputPrompt component with the given configuration.
// This implements the complete InputPrompt specification from the docs.
func NewInputPrompt(config InputPromptConfig) *InputPrompt {
	// Calculate widths using the exact algorithm from docs
	widths := CalculatePromptWidthsFromRaw(config.TerminalWidth)

	// Create cursor
	cur := cursor.New()
	cur.Style = lipgloss.NewStyle() // Cursor is rendered manually
	cur.SetMode(cursor.CursorBlink)

	ip := &InputPrompt{
		config:      config,
		value:       "",
		cursorPos:   0,
		cursorBlink: true,
		widths:      widths,
		cursor:      cur,
		focused:     true,
	}

	// Create enhanced keypress handler
	ip.keypressHandler = keypress.NewHandler(keypress.HandlerConfig{
		KittyEnabled:   config.EnableKittyProtocol,
		DebugKeystroke: config.DebugKeystroke,
		OnPaste: func(content string) {
			// Handle paste events
			if ip.onPasteReceived != nil {
				ip.onPasteReceived(content)
			} else {
				// Default: insert pasted content at cursor position
				ip.insertText(content)
			}
		},
		OnKey: func(key keypress.Key) bool {
			// This callback is used for secondary keys in batch sequences
			// Process them through the normal key handling
			if msg := key.Original; msg != nil {
				ip.handleKeyMsg(*msg)
			}
			return true
		},
	})

	ip.updateStyles()
	return ip
}

// updateStyles creates and caches all styles based on current config
func (ip *InputPrompt) updateStyles() {
	theme := ip.config.Theme

	// Determine border color based on state
	borderColor := ip.getBorderColor()

	// Border style with rounded corners and padding
	// Width should include: prefix + input content + padding (left + right)
	prefixWidth := len([]rune(ip.getPromptPrefix()))
	totalContentWidth := prefixWidth + ip.widths.InputWidth
	ip.borderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).               // paddingX = 1 (left and right)
		Width(totalContentWidth + 2) // +2 for left and right padding

	// Prefix style (changes based on mode)
	prefixColor := theme.Text.Accent
	if ip.config.ShellModeActive {
		prefixColor = theme.UI.Symbol
	}
	ip.prefixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(prefixColor))

	// Text styles
	ip.textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text.Primary))
	ip.accentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text.Accent))
	ip.secondaryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text.Secondary))
	ip.placeholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text.Secondary))
}

// getBorderColor determines the border color based on current state.
// This implements the GetBorderColor logic from the docs.
func (ip *InputPrompt) getBorderColor() string {
	theme := ip.config.Theme

	if !ip.focused {
		return theme.Border.Default
	}

	// Determine status color
	statusColor, _ := ip.getStatusColor()
	if statusColor != "" {
		return statusColor
	}

	return theme.Border.Focused
}

// getStatusColor returns the status color and description based on mode.
// This implements the GetStatusColor logic from the docs.
func (ip *InputPrompt) getStatusColor() (string, string) {
	theme := ip.config.Theme

	if ip.config.ShellModeActive {
		return theme.UI.Symbol, "Shell mode"
	}

	if ip.config.YoloMode {
		return theme.Status.Error, "YOLO mode"
	}

	if ip.config.AutoEditMode {
		return theme.Status.Warning, "Accepting edits"
	}

	return "", ""
}

// getPromptPrefix returns the prefix string based on mode.
// This implements the GetPromptPrefix logic from the docs.
func (ip *InputPrompt) getPromptPrefix() string {
	mode := ip.config.Mode

	if mode == PromptModeShell {
		return "! "
	}

	if mode == PromptModeReverseSearch {
		return "(r:) "
	}

	if mode == PromptModeYolo || ip.config.YoloMode {
		return "* "
	}

	return "> "
}

// Update handles Bubble Tea messages
func (ip *InputPrompt) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Process through enhanced keypress handler
		key := ip.keypressHandler.ProcessTeaKey(msg)
		if key == nil {
			// Event consumed by handler (e.g., paste marker)
			return nil
		}
		return ip.handleEnhancedKey(*key)

	case tea.WindowSizeMsg:
		// Recalculate widths on terminal resize
		ip.widths = CalculatePromptWidthsFromRaw(msg.Width)
		ip.updateStyles()
		return nil
	}

	ip.cursor, cmd = ip.cursor.Update(msg)
	return cmd
}

// handleEnhancedKey processes keyboard input with enhanced key information
func (ip *InputPrompt) handleEnhancedKey(key keypress.Key) tea.Cmd {
	// Handle special keys with modifiers
	switch key.Name {
	case "left":
		if key.Ctrl {
			// Ctrl+Left: Move to previous word
			ip.moveToPreviousWord()
		} else {
			// Regular left arrow
			if ip.cursorPos > 0 {
				ip.cursorPos--
			}
		}

	case "right":
		if key.Ctrl {
			// Ctrl+Right: Move to next word
			ip.moveToNextWord()
		} else {
			// Regular right arrow
			runes := []rune(ip.value)
			if ip.cursorPos < len(runes) {
				ip.cursorPos++
			}
		}

	case "home":
		ip.cursorPos = 0

	case "end":
		runes := []rune(ip.value)
		ip.cursorPos = len(runes)

	case "backspace":
		if key.Ctrl {
			// Ctrl+Backspace: Delete word backwards
			ip.deleteWordBackward()
		} else {
			// Regular backspace
			if ip.cursorPos > 0 {
				runes := []rune(ip.value)
				ip.value = string(runes[:ip.cursorPos-1]) + string(runes[ip.cursorPos:])
				ip.cursorPos--
			}
		}

	case "delete":
		if key.Ctrl {
			// Ctrl+Delete: Delete word forward
			ip.deleteWordForward()
		} else {
			// Regular delete
			runes := []rune(ip.value)
			if ip.cursorPos < len(runes) {
				ip.value = string(runes[:ip.cursorPos]) + string(runes[ip.cursorPos+1:])
			}
		}

	case "return", "enter":
		if key.Shift {
			// Shift+Enter: Insert newline
			ip.insertText("\n")
		} else {
			// Regular enter is handled by the parent component
			// We just pass it through via the original KeyMsg
			if key.Original != nil {
				return ip.handleKeyMsg(*key.Original)
			}
		}

	default:
		// Regular character input
		if key.Paste {
			// This is pasted content
			ip.insertText(key.Sequence)
		} else if len(key.Name) == 1 && !key.Ctrl {
			// Single character
			ip.insertText(key.Name)
		} else if key.Original != nil {
			// Fall back to original handler for complex cases
			return ip.handleKeyMsg(*key.Original)
		}
	}

	return nil
}

// handleKeyMsg processes keyboard input (legacy, for compatibility)
func (ip *InputPrompt) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyLeft:
		if ip.cursorPos > 0 {
			ip.cursorPos--
		}

	case tea.KeyRight:
		runes := []rune(ip.value)
		if ip.cursorPos < len(runes) {
			ip.cursorPos++
		}

	case tea.KeyHome:
		ip.cursorPos = 0

	case tea.KeyEnd:
		runes := []rune(ip.value)
		ip.cursorPos = len(runes)

	case tea.KeyBackspace:
		if ip.cursorPos > 0 {
			runes := []rune(ip.value)
			ip.value = string(runes[:ip.cursorPos-1]) + string(runes[ip.cursorPos:])
			ip.cursorPos--
		}

	case tea.KeyDelete:
		runes := []rune(ip.value)
		if ip.cursorPos < len(runes) {
			ip.value = string(runes[:ip.cursorPos]) + string(runes[ip.cursorPos+1:])
		}

	case tea.KeyRunes:
		// Insert character at cursor position
		runes := []rune(ip.value)
		before := string(runes[:ip.cursorPos])
		after := string(runes[ip.cursorPos:])
		ip.value = before + string(msg.Runes) + after
		ip.cursorPos += len(msg.Runes)
	}

	return nil
}

// insertText inserts text at the current cursor position
func (ip *InputPrompt) insertText(text string) {
	if text == "" {
		return
	}

	// Limit total input length to prevent extremely long inputs
	const maxInputLength = 10000 // Reasonable limit for CLI input
	currentRunes := []rune(ip.value)
	if len(currentRunes) >= maxInputLength {
		return // Don't insert if already at max length
	}

	textRunes := []rune(text)
	remainingSpace := maxInputLength - len(currentRunes)
	if len(textRunes) > remainingSpace {
		textRunes = textRunes[:remainingSpace]
		text = string(textRunes)
	}

	before := string(currentRunes[:ip.cursorPos])
	after := string(currentRunes[ip.cursorPos:])
	ip.value = before + text + after
	ip.cursorPos += len(textRunes)
}

// moveToPreviousWord moves cursor to the start of the previous word
func (ip *InputPrompt) moveToPreviousWord() {
	runes := []rune(ip.value)
	if ip.cursorPos == 0 {
		return
	}

	// Skip whitespace
	pos := ip.cursorPos - 1
	for pos > 0 && isWhitespace(runes[pos]) {
		pos--
	}

	// Skip word characters
	for pos > 0 && !isWhitespace(runes[pos]) {
		pos--
	}

	if pos > 0 {
		ip.cursorPos = pos + 1
	} else {
		ip.cursorPos = 0
	}
}

// moveToNextWord moves cursor to the start of the next word
func (ip *InputPrompt) moveToNextWord() {
	runes := []rune(ip.value)
	if ip.cursorPos >= len(runes) {
		return
	}

	// Skip current word
	pos := ip.cursorPos
	for pos < len(runes) && !isWhitespace(runes[pos]) {
		pos++
	}

	// Skip whitespace
	for pos < len(runes) && isWhitespace(runes[pos]) {
		pos++
	}

	ip.cursorPos = pos
}

// deleteWordBackward deletes from cursor to start of current/previous word
func (ip *InputPrompt) deleteWordBackward() {
	if ip.cursorPos == 0 {
		return
	}

	runes := []rune(ip.value)
	oldPos := ip.cursorPos

	// Find start of word
	ip.moveToPreviousWord()

	// Delete from new position to old position
	ip.value = string(runes[:ip.cursorPos]) + string(runes[oldPos:])
}

// deleteWordForward deletes from cursor to end of current/next word
func (ip *InputPrompt) deleteWordForward() {
	runes := []rune(ip.value)
	if ip.cursorPos >= len(runes) {
		return
	}

	// Skip whitespace
	pos := ip.cursorPos
	for pos < len(runes) && isWhitespace(runes[pos]) {
		pos++
	}

	// Skip word characters
	for pos < len(runes) && !isWhitespace(runes[pos]) {
		pos++
	}

	// Delete from cursor to pos
	ip.value = string(runes[:ip.cursorPos]) + string(runes[pos:])
}

// isWhitespace checks if a rune is whitespace
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// Render returns the rendered InputPrompt as a string.
// This implements the complete rendering logic including:
// - Border with rounded corners
// - Prompt prefix
// - Syntax-highlighted input text
// - Cursor rendering with inverse video
// - Ghost text (autocomplete preview)
func (ip *InputPrompt) Render() string {
	prefix := ip.getPromptPrefix()
	styledPrefix := ip.prefixStyle.Render(prefix)

	var content string

	if ip.value == "" && ip.config.Placeholder != "" {
		// Show placeholder
		content = ip.renderPlaceholder()
	} else {
		// Render actual input with highlighting and cursor
		content = ip.renderInputContent()
	}

	// Combine prefix and content
	innerContent := styledPrefix + content

	// Apply border
	return ip.borderStyle.Render(innerContent)
}

// renderPlaceholder renders the placeholder text with cursor
func (ip *InputPrompt) renderPlaceholder() string {
	if !ip.config.ShowCursor || !ip.focused {
		return ip.placeholderStyle.Render(ip.config.Placeholder)
	}

	// Show cursor at first position
	runes := []rune(ip.config.Placeholder)
	if len(runes) == 0 {
		return ip.renderCursor(" ")
	}

	firstChar := string(runes[0])
	rest := string(runes[1:])

	cursor := ip.renderCursor(firstChar)
	restStyled := ip.placeholderStyle.Render(rest)

	return cursor + restStyled
}

// renderInputContent renders the input text with syntax highlighting and cursor
func (ip *InputPrompt) renderInputContent() string {
	lines := utils.SplitInputLines(ip.value)
	var rendered strings.Builder

	// Track cursor position across lines
	currentPos := 0

	for lineIdx, line := range lines {
		if lineIdx > 0 {
			rendered.WriteString("\n")
			currentPos++ // newline counts as 1 position
		}

		lineRendered := ip.renderLine(line, lineIdx, currentPos)
		rendered.WriteString(lineRendered)

		currentPos += len([]rune(line))
	}

	// Add ghost text if available
	if ip.config.GhostText != "" && ip.focused {
		ghostText := ip.renderGhostText()
		rendered.WriteString(ghostText)
	}

	return rendered.String()
}

// renderLine renders a single line with syntax highlighting and cursor
func (ip *InputPrompt) renderLine(line string, lineIndex int, lineStartPos int) string {
	if line == "" {
		// Handle empty line with cursor
		if ip.cursorPos == lineStartPos && ip.config.ShowCursor && ip.focused {
			return ip.renderCursor(" ")
		}
		return ""
	}

	// Truncate line if it exceeds the input width
	lineWidth := runewidth.StringWidth(line)
	if lineWidth > ip.widths.InputWidth {
		// Truncate to fit within input width
		line = truncateToWidth(line, ip.widths.InputWidth)
	}

	tokens := utils.ParseInputForHighlighting(line, lineIndex)
	var rendered strings.Builder

	currentPos := lineStartPos

	for _, token := range tokens {
		tokenRunes := []rune(token.Text)
		tokenLen := len(tokenRunes)

		// Check if cursor is within this token
		if ip.config.ShowCursor && ip.focused &&
			ip.cursorPos >= currentPos && ip.cursorPos < currentPos+tokenLen {

			// Render token with cursor
			cursorOffset := ip.cursorPos - currentPos
			before := string(tokenRunes[:cursorOffset])
			cursorChar := string(tokenRunes[cursorOffset])
			after := string(tokenRunes[cursorOffset+1:])

			rendered.WriteString(ip.styleToken(before, token.Type))
			rendered.WriteString(ip.renderCursor(cursorChar))
			rendered.WriteString(ip.styleToken(after, token.Type))
		} else {
			// Render token normally
			rendered.WriteString(ip.styleToken(token.Text, token.Type))
		}

		currentPos += tokenLen
	}

	// Cursor at end of line
	if ip.config.ShowCursor && ip.focused && ip.cursorPos == currentPos {
		rendered.WriteString(ip.renderCursor(" "))
	}

	return rendered.String()
}

// styleToken applies the appropriate color to a token based on its type
func (ip *InputPrompt) styleToken(text string, tokenType utils.HighlightTokenType) string {
	if text == "" {
		return ""
	}

	switch tokenType {
	case utils.TokenTypeCommand, utils.TokenTypeFile:
		return ip.accentStyle.Render(text)
	default:
		return ip.textStyle.Render(text)
	}
}

// renderCursor renders a character with inverse video (ANSI reverse).
// This implements the cursor rendering logic from the docs.
func (ip *InputPrompt) renderCursor(char string) string {
	// ANSI escape codes:
	// \x1b[7m  = Enable reverse video
	// \x1b[27m = Disable reverse video
	return "\x1b[7m" + char + "\x1b[27m"
}

// renderGhostText renders the autocomplete ghost text.
// This is a simplified version - full implementation would include
// word wrapping and multi-line support as per the docs.
func (ip *InputPrompt) renderGhostText() string {
	if ip.config.GhostText == "" {
		return ""
	}

	// Verify ghost text starts with current input
	if !strings.HasPrefix(ip.config.GhostText, ip.value) {
		return ""
	}

	// Get the suffix to display
	ghostSuffix := ip.config.GhostText[len(ip.value):]
	if ghostSuffix == "" {
		return ""
	}

	// Calculate remaining width on current line
	currentLineText := ip.getCurrentLineText()
	usedWidth := runewidth.StringWidth(currentLineText)
	remainingWidth := max(0, ip.widths.InputWidth-usedWidth)

	// Simple inline ghost text (first line only)
	ghostRunes := []rune(ghostSuffix)

	// Find first newline
	var inlineGhost string
	for i, r := range ghostRunes {
		if r == '\n' {
			inlineGhost = string(ghostRunes[:i])
			break
		}
	}

	if inlineGhost == "" {
		inlineGhost = ghostSuffix
	}

	// Truncate if too long
	if runewidth.StringWidth(inlineGhost) > remainingWidth {
		inlineGhost = truncateToWidth(inlineGhost, remainingWidth)
	}

	return ip.secondaryStyle.Render(inlineGhost)
}

// getCurrentLineText returns the text of the current line (where cursor is)
func (ip *InputPrompt) getCurrentLineText() string {
	lines := utils.SplitInputLines(ip.value)
	currentPos := 0

	for _, line := range lines {
		lineLen := len([]rune(line)) + 1 // +1 for newline
		if ip.cursorPos < currentPos+lineLen {
			return line
		}
		currentPos += lineLen
	}

	return ""
}

// truncateToWidth truncates a string to fit within the specified display width
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(s)
	currentWidth := 0

	for i, r := range runes {
		charWidth := runewidth.RuneWidth(r)
		if currentWidth+charWidth > width {
			return string(runes[:i])
		}
		currentWidth += charWidth
	}

	return s
}

// Focus sets focus on the input prompt
func (ip *InputPrompt) Focus() tea.Cmd {
	ip.focused = true
	ip.updateStyles()
	return ip.cursor.Focus()
}

// Blur removes focus from the input prompt
func (ip *InputPrompt) Blur() {
	ip.focused = false
	ip.updateStyles()
	ip.cursor.Blur()
}

// Value returns the current input value
func (ip *InputPrompt) Value() string {
	return ip.value
}

// SetValue sets the input value and moves cursor to end
func (ip *InputPrompt) SetValue(value string) {
	ip.value = value
	ip.cursorPos = len([]rune(value))
}

// Reset clears the input value and resets cursor
func (ip *InputPrompt) Reset() {
	ip.value = ""
	ip.cursorPos = 0
}

// CursorPosition returns the current cursor position
func (ip *InputPrompt) CursorPosition() int {
	return ip.cursorPos
}

// SetCursorPosition sets the cursor position (clamped to valid range)
func (ip *InputPrompt) SetCursorPosition(pos int) {
	runes := []rune(ip.value)
	if pos < 0 {
		ip.cursorPos = 0
	} else if pos > len(runes) {
		ip.cursorPos = len(runes)
	} else {
		ip.cursorPos = pos
	}
}

// UpdateConfig updates the configuration and recalculates styles
func (ip *InputPrompt) UpdateConfig(config InputPromptConfig) {
	ip.config = config
	ip.widths = CalculatePromptWidthsFromRaw(config.TerminalWidth)
	ip.updateStyles()
}

// Height returns the display height (including border)
func (ip *InputPrompt) Height() int {
	// Count lines in value
	lines := strings.Count(ip.value, "\n") + 1

	// Add border (top + bottom)
	return lines + 2
}

// Width returns the display width (including border)
func (ip *InputPrompt) Width() int {
	return ip.widths.ContainerWidth
}

// Widths returns the calculated width values
func (ip *InputPrompt) Widths() PromptWidths {
	return ip.widths
}

// SetMode changes the prompt mode
func (ip *InputPrompt) SetMode(mode PromptMode) {
	ip.config.Mode = mode
	ip.updateStyles()
}

// SetGhostText sets the autocomplete ghost text
func (ip *InputPrompt) SetGhostText(text string) {
	ip.config.GhostText = text
}

// SetPasteHandler sets a custom handler for paste events
func (ip *InputPrompt) SetPasteHandler(handler func(string)) {
	ip.onPasteReceived = handler
}

// Cleanup releases resources (timers, etc.) - call this when disposing the component
func (ip *InputPrompt) Cleanup() {
	if ip.keypressHandler != nil {
		ip.keypressHandler.Cleanup()
	}
}
