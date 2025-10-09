package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/boke/cli-demo/internal/tui/theme"
)

// FooterConfig holds configuration for the Footer component.
type FooterConfig struct {
	Theme          theme.ColorsTheme
	Width          int
	WorkingDir     string // Current working directory
	Model          string // Model name (e.g., "gemini-2.5-pro")
	ContextUsage   string // Context usage info (e.g., "100% context left")
	SandboxStatus  string // Sandbox status (e.g., "no sandbox")
	ShowWorkingDir bool   // Whether to show working directory
	ShowSandbox    bool   // Whether to show sandbox status
	ShowModel      bool   // Whether to show model info
	DebugMode      bool   // Debug mode flag
	DebugMessage   string // Debug message to display
}

// Footer represents a footer component with status information.
// This implements similar functionality to the TypeScript Footer component.
type Footer struct {
	config      FooterConfig
	leftStyle   lipgloss.Style
	centerStyle lipgloss.Style
	rightStyle  lipgloss.Style
}

// NewFooter creates a new Footer component with the given configuration.
func NewFooter(config FooterConfig) *Footer {
	f := &Footer{
		config: config,
	}
	f.updateStyles()
	return f
}

// updateStyles creates the styles for different sections of the footer.
func (f *Footer) updateStyles() {
	f.leftStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(f.config.Theme.AccentCyan))

	f.centerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(f.config.Theme.AccentRed))

	f.rightStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(f.config.Theme.Gray))
}

// formatWorkingDir formats the working directory for display.
// Converts absolute path to tilde path and optionally shortens it.
func (f *Footer) formatWorkingDir() string {
	if f.config.WorkingDir == "" {
		return ""
	}

	// Convert to tilde path
	homeDir, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(f.config.WorkingDir, homeDir) {
		return "~" + strings.TrimPrefix(f.config.WorkingDir, homeDir)
	}

	// For narrow terminals, just show basename
	if f.config.Width < 80 {
		return filepath.Base(f.config.WorkingDir)
	}

	return f.config.WorkingDir
}

// renderLeftSection renders the left section (working directory, debug info).
func (f *Footer) renderLeftSection() string {
	if !f.config.ShowWorkingDir && !f.config.DebugMode {
		return ""
	}

	var parts []string

	if f.config.ShowWorkingDir {
		dirPath := f.formatWorkingDir()
		styledPath := f.leftStyle.Render(dirPath)
		parts = append(parts, styledPath)
	}

	if f.config.DebugMode {
		debugText := f.config.DebugMessage
		if debugText == "" {
			debugText = "--debug"
		}
		styledDebug := lipgloss.NewStyle().
			Foreground(lipgloss.Color(f.config.Theme.AccentRed)).
			Render(" " + debugText)
		parts = append(parts, styledDebug)
	}

	return strings.Join(parts, "")
}

// renderCenterSection renders the center section (sandbox/trust status).
func (f *Footer) renderCenterSection() string {
	if !f.config.ShowSandbox {
		return ""
	}

	sandboxEnv := os.Getenv("SANDBOX")
	var statusText string
	var statusColor string

	switch {
	case sandboxEnv != "" && sandboxEnv != "sandbox-exec":
		// Running in sandbox
		statusText = strings.TrimPrefix(sandboxEnv, "gemini-")
		statusText = strings.TrimPrefix(statusText, "cli-")
		statusColor = f.config.Theme.AccentGreen
	case sandboxEnv == "sandbox-exec":
		// macOS Seatbelt
		profile := os.Getenv("SEATBELT_PROFILE")
		if profile != "" {
			statusText = fmt.Sprintf("macOS Seatbelt (%s)", profile)
		} else {
			statusText = "macOS Seatbelt"
		}
		statusColor = f.config.Theme.AccentYellow
	default:
		// No sandbox
		if f.config.SandboxStatus != "" {
			statusText = f.config.SandboxStatus
		} else {
			statusText = "no sandbox (see /docs)"
		}
		statusColor = f.config.Theme.AccentRed
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor)).
		Render(statusText)
}

// renderRightSection renders the right section (model info, context usage).
func (f *Footer) renderRightSection() string {
	if !f.config.ShowModel {
		return ""
	}

	var parts []string

	if f.config.Model != "" {
		parts = append(parts, f.config.Model)
	}

	if f.config.ContextUsage != "" {
		contextText := lipgloss.NewStyle().
			Foreground(lipgloss.Color(f.config.Theme.Gray)).
			Render(f.config.ContextUsage)
		parts = append(parts, contextText)
	}

	if len(parts) == 0 {
		return ""
	}

	return f.rightStyle.Render(strings.Join(parts, " "))
}

// Render returns the rendered footer as a string.
// Layout: [Left Section] [Center Section - flexible spacing] [Right Section]
func (f *Footer) Render() string {
	leftSection := f.renderLeftSection()
	centerSection := f.renderCenterSection()
	rightSection := f.renderRightSection()

	// Calculate widths (strip ANSI for accurate width)
	leftWidth := lipgloss.Width(leftSection)
	centerWidth := lipgloss.Width(centerSection)
	rightWidth := lipgloss.Width(rightSection)

	// For narrow terminals, use vertical layout
	if f.config.Width < 80 {
		var lines []string
		if leftSection != "" {
			lines = append(lines, leftSection)
		}
		if centerSection != "" {
			lines = append(lines, centerSection)
		}
		if rightSection != "" {
			lines = append(lines, rightSection)
		}
		return strings.Join(lines, "\n")
	}

	// Calculate spacing
	usedWidth := leftWidth + centerWidth + rightWidth
	availableSpace := f.config.Width - usedWidth

	// Create spacing
	var leftPadding, rightPadding int
	if availableSpace > 0 {
		// Distribute space: left section | center (centered) | right section
		if centerSection != "" {
			// Center the middle section
			totalSide := availableSpace / 2
			leftPadding = totalSide
			rightPadding = availableSpace - leftPadding
		} else {
			// Just space between left and right
			leftPadding = 0
			rightPadding = availableSpace
		}
	}

	// Build the footer line
	var result strings.Builder
	if leftSection != "" {
		result.WriteString(leftSection)
	}
	if centerSection != "" {
		result.WriteString(strings.Repeat(" ", leftPadding))
		result.WriteString(centerSection)
	}
	if rightSection != "" {
		result.WriteString(strings.Repeat(" ", rightPadding))
		result.WriteString(rightSection)
	}

	return result.String()
}

// Update allows updating the footer configuration.
func (f *Footer) Update(config FooterConfig) {
	f.config = config
	f.updateStyles()
}

// Height returns the display height of the footer.
func (f *Footer) Height() int {
	// For narrow terminals, might be multi-line
	if f.config.Width < 80 {
		lines := 0
		if f.config.ShowWorkingDir || f.config.DebugMode {
			lines++
		}
		if f.config.ShowSandbox {
			lines++
		}
		if f.config.ShowModel {
			lines++
		}
		return lines
	}
	return 1 // Single line for wide terminals
}

// Width returns the display width of the footer.
func (f *Footer) Width() int {
	return f.config.Width
}
