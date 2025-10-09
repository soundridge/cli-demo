package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/boke/cli-demo/internal/tui/theme"
)

// TipsConfig holds configuration for the Tips component.
type TipsConfig struct {
	Theme theme.ColorsTheme
	Items []string // List of tip items to display
	Width int      // Maximum width for wrapping
}

// Tips represents a tips/hints component that displays helpful information.
type Tips struct {
	Config TipsConfig // Exported for access in model updates
	style  lipgloss.Style
}

// NewTips creates a new Tips component with the given configuration.
func NewTips(config TipsConfig) *Tips {
	t := &Tips{
		Config: config,
	}
	t.updateStyle()
	return t
}

// updateStyle creates the style for tips text.
func (t *Tips) updateStyle() {
	t.style = lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Config.Theme.Gray)).
		MarginTop(1).
		MarginBottom(1)
}

// Render returns the rendered tips as a string.
func (t *Tips) Render() string {
	if len(t.Config.Items) == 0 {
		return ""
	}

	var output strings.Builder
	output.WriteString(t.style.Render("Tips for getting started:"))
	output.WriteString("\n")

	for i, tip := range t.Config.Items {
		line := lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Config.Theme.Gray)).
			Render(tip)

		if i < len(t.Config.Items)-1 {
			output.WriteString(line + "\n")
		} else {
			output.WriteString(line)
		}
	}

	return output.String()
}

// Update allows updating the tips configuration.
func (t *Tips) Update(config TipsConfig) {
	t.Config = config
	t.updateStyle()
}

// Height returns the display height of the tips component.
func (t *Tips) Height() int {
	if len(t.Config.Items) == 0 {
		return 0
	}
	// Title line + items + margins
	return 1 + len(t.Config.Items) + 2
}

// Width returns the display width of the tips component.
func (t *Tips) Width() int {
	return t.Config.Width
}
