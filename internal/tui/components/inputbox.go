package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/boke/cli-demo/internal/tui/theme"
)

// InputBoxConfig holds configuration for the InputBox component.
type InputBoxConfig struct {
	Theme       theme.ColorsTheme
	Placeholder string
	Width       int
	Prompt      string
}

// InputBox represents an input field with border styling.
type InputBox struct {
	config    InputBoxConfig
	textInput textinput.Model
	boxStyle  lipgloss.Style
	focused   bool
}

// NewInputBox creates a new InputBox component with the given configuration.
func NewInputBox(config InputBoxConfig) *InputBox {
	ti := textinput.New()
	ti.Placeholder = config.Placeholder
	ti.CharLimit = 500
	ti.Width = config.Width - 6 // Account for border and padding
	ti.Prompt = "> "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(config.Theme.AccentCyan))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(config.Theme.AccentCyan))
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(config.Theme.AccentBlue))
	ti.Focus()

	ib := &InputBox{
		config:    config,
		textInput: ti,
		focused:   true,
	}
	ib.updateStyle()
	return ib
}

// updateStyle creates the border style for the input box.
func (ib *InputBox) updateStyle() {
	borderColor := ib.config.Theme.Gray
	if ib.focused {
		borderColor = ib.config.Theme.AccentBlue
	}

	ib.boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Width(ib.config.Width - 2).
		MarginTop(1)
}

// Update handles Bubble Tea messages for the input box.
func (ib *InputBox) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	ib.textInput, cmd = ib.textInput.Update(msg)
	return cmd
}

// Render returns the rendered input box as a string.
func (ib *InputBox) Render() string {
	return ib.boxStyle.Render(ib.textInput.View())
}

// Focus sets focus on the input box.
func (ib *InputBox) Focus() tea.Cmd {
	ib.focused = true
	ib.updateStyle()
	return ib.textInput.Focus()
}

// Blur removes focus from the input box.
func (ib *InputBox) Blur() {
	ib.focused = false
	ib.updateStyle()
	ib.textInput.Blur()
}

// Value returns the current value of the input box.
func (ib *InputBox) Value() string {
	return ib.textInput.Value()
}

// SetValue sets the value of the input box.
func (ib *InputBox) SetValue(value string) {
	ib.textInput.SetValue(value)
}

// Reset clears the input box value.
func (ib *InputBox) Reset() {
	ib.textInput.Reset()
}

// UpdateConfig updates the input box configuration.
func (ib *InputBox) UpdateConfig(config InputBoxConfig) {
	ib.config = config
	ib.textInput.Width = config.Width - 6
	ib.updateStyle()
}

// Height returns the display height of the input box.
func (ib *InputBox) Height() int {
	return 3 // Border top + content + border bottom + margin
}

// Width returns the display width of the input box.
func (ib *InputBox) Width() int {
	return ib.config.Width
}
