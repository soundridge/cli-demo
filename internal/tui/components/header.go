package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/boke/cli-demo/internal/tui/theme"
	"github.com/boke/cli-demo/internal/tui/utils"
)

// HeaderConfig holds configuration for the Header component.
// This matches the HeaderProps interface from the TypeScript implementation.
type HeaderConfig struct {
	CustomAsciiArt string            // User-defined ASCII art (optional)
	Version        string            // Application version
	ShowVersion    bool              // Whether to display the version (similar to "nightly" flag)
	TerminalWidth  int               // Current terminal width
	Theme          theme.ColorsTheme // Theme for styling
}

// Header represents a header component with responsive ASCII art logo.
// This implements the same functionality as the React Header component.
type Header struct {
	config HeaderConfig
}

// NewHeader creates a new Header component with the given configuration.
func NewHeader(config HeaderConfig) *Header {
	return &Header{
		config: config,
	}
}

// selectLogo chooses the appropriate logo based on terminal width.
// This implements the same logic as the TypeScript Header component:
// - If customAsciiArt is provided, use it
// - Otherwise, select based on terminal width
func (h *Header) selectLogo() string {
	// Custom ASCII art takes precedence
	if h.config.CustomAsciiArt != "" {
		return h.config.CustomAsciiArt
	}

	termWidth := h.config.TerminalWidth
	widthOfLong := utils.GetAsciiArtWidth(LongAsciiLogo)
	widthOfShort := utils.GetAsciiArtWidth(ShortAsciiLogo)

	// Select logo size based on terminal width
	// This matches the TypeScript logic exactly
	switch {
	case termWidth >= widthOfLong:
		return LongAsciiLogo
	case termWidth >= widthOfShort:
		return ShortAsciiLogo
	default:
		return TinyAsciiLogo
	}
}

// applyGradient applies a gradient effect to text using lipgloss.
// This mimics the Gradient component from ink-gradient used in the TypeScript version.
// If gradient colors are not available, falls back to a single accent color.
func (h *Header) applyGradient(text string) string {
	gradientColors := h.config.Theme.GradientColors

	// If no gradient colors available, use accent purple as fallback
	if len(gradientColors) == 0 {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(h.config.Theme.AccentPurple))
		return style.Render(text)
	}

	// Single color - just apply it
	if len(gradientColors) == 1 {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(gradientColors[0]))
		return style.Render(text)
	}

	// Multi-line gradient: cycle through colors per line
	// This creates a vertical gradient effect for ASCII art
	lines := strings.Split(text, "\n")
	styledLines := make([]string, 0, len(lines))

	for i, line := range lines {
		// Cycle through gradient colors based on line index
		colorIndex := i % len(gradientColors)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(gradientColors[colorIndex]))
		styledLines = append(styledLines, style.Render(line))
	}

	return strings.Join(styledLines, "\n")
}

// Render returns the rendered header as a string.
// This implements the same rendering logic as the TypeScript Header component:
// 1. Select appropriate logo
// 2. Apply gradient styling
// 3. Optionally add version info right-aligned
func (h *Header) Render() string {
	displayTitle := h.selectLogo()
	artWidth := utils.GetAsciiArtWidth(displayTitle)

	// Apply gradient styling to the ASCII art
	// This mimics: <Gradient colors={theme.ui.gradient}><Text>{displayTitle}</Text></Gradient>
	styledTitle := h.applyGradient(displayTitle)

	// Build the header output
	var output strings.Builder
	output.WriteString(styledTitle)

	// Add version info if requested (similar to the "nightly" condition)
	// This mimics: {nightly && <Box>...</Box>}
	if h.config.ShowVersion && h.config.Version != "" {
		versionText := fmt.Sprintf("v%s", h.config.Version)

		// Style the version with gradient
		styledVersion := h.applyGradient(versionText)

		// Calculate padding to right-align version within art width
		// This mimics: <Box width="100%" justifyContent="flex-end">
		versionWidth := utils.CodePointLength(versionText) // Use code point length for accuracy
		padding := artWidth - versionWidth
		if padding < 0 {
			padding = 0
		}

		output.WriteString("\n")
		output.WriteString(strings.Repeat(" ", padding))
		output.WriteString(styledVersion)
	}

	return output.String()
}

// Update allows updating the header configuration (e.g., on terminal resize).
func (h *Header) Update(config HeaderConfig) {
	h.config = config
}

// Width returns the current display width of the header.
// This matches the artWidth calculation in the TypeScript implementation.
func (h *Header) Width() int {
	return utils.GetAsciiArtWidth(h.selectLogo())
}

// Height returns the current display height of the header.
// Calculates based on number of lines in the logo plus optional version line.
func (h *Header) Height() int {
	logo := h.selectLogo()
	lines := strings.Split(logo, "\n")
	height := len(lines)

	// Add one line for version if showing
	if h.config.ShowVersion && h.config.Version != "" {
		height++
	}

	return height
}
