package theme

// ThemeType represents the type of theme (light, dark, ansi, custom)
type ThemeType string

const (
	ThemeTypeLight  ThemeType = "light"
	ThemeTypeDark   ThemeType = "dark"
	ThemeTypeANSI   ThemeType = "ansi"
	ThemeTypeCustom ThemeType = "custom"
)

// ColorsTheme defines the core color palette for the TUI
type ColorsTheme struct {
	Type           ThemeType
	Background     string
	Foreground     string
	LightBlue      string
	AccentBlue     string
	AccentPurple   string
	AccentCyan     string
	AccentGreen    string
	AccentYellow   string
	AccentRed      string
	DiffAdded      string
	DiffRemoved    string
	Comment        string
	Gray           string
	GradientColors []string
}

// CustomTheme provides a flexible theme structure with modern organization
type CustomTheme struct {
	Type ThemeType
	Name string

	Text       *TextColors
	Background *BackgroundColors
	Border     *BorderColors
	UI         *UIColors
	Status     *StatusColors

	// Legacy properties (optional, for compatibility)
	LegacyBackground     string
	LegacyForeground     string
	LegacyLightBlue      string
	LegacyAccentBlue     string
	LegacyAccentPurple   string
	LegacyAccentCyan     string
	LegacyAccentGreen    string
	LegacyAccentYellow   string
	LegacyAccentRed      string
	LegacyDiffAdded      string
	LegacyDiffRemoved    string
	LegacyComment        string
	LegacyGray           string
	LegacyGradientColors []string
}

// TextColors defines text-related colors
type TextColors struct {
	Primary   string
	Secondary string
	Link      string
	Accent    string
}

// BackgroundColors defines background-related colors
type BackgroundColors struct {
	Primary string
	Diff    *DiffColors
}

// DiffColors defines diff-related colors
type DiffColors struct {
	Added   string
	Removed string
}

// BorderColors defines border-related colors
type BorderColors struct {
	Default string
	Focused string
}

// UIColors defines UI element colors
type UIColors struct {
	Comment  string
	Symbol   string
	Gradient []string
}

// StatusColors defines status indicator colors
type StatusColors struct {
	Error   string
	Success string
	Warning string
}

// Light returns the light theme
func Light() ColorsTheme {
	return ColorsTheme{
		Type:           ThemeTypeLight,
		Background:     "#FAFAFA",
		Foreground:     "",
		LightBlue:      "#89BDCD",
		AccentBlue:     "#3B82F6",
		AccentPurple:   "#8B5CF6",
		AccentCyan:     "#06B6D4",
		AccentGreen:    "#3CA84B",
		AccentYellow:   "#D5A40A",
		AccentRed:      "#DD4C4C",
		DiffAdded:      "#C6EAD8",
		DiffRemoved:    "#FFCCCC",
		Comment:        "#008000",
		Gray:           "#97a0b0",
		GradientColors: []string{"#4796E4", "#847ACE", "#C3677F"},
	}
}

// Dark returns the dark theme
func Dark() ColorsTheme {
	return ColorsTheme{
		Type:           ThemeTypeDark,
		Background:     "#1E1E2E",
		Foreground:     "",
		LightBlue:      "#ADD8E6",
		AccentBlue:     "#89B4FA",
		AccentPurple:   "#CBA6F7",
		AccentCyan:     "#89DCEB",
		AccentGreen:    "#A6E3A1",
		AccentYellow:   "#F9E2AF",
		AccentRed:      "#F38BA8",
		DiffAdded:      "#28350B",
		DiffRemoved:    "#430000",
		Comment:        "#6C7086",
		Gray:           "#6C7086",
		GradientColors: []string{"#4796E4", "#847ACE", "#C3677F"},
	}
}

// ANSI returns the ANSI color theme
func ANSI() ColorsTheme {
	return ColorsTheme{
		Type:           ThemeTypeANSI,
		Background:     "black",
		Foreground:     "white",
		LightBlue:      "blue",
		AccentBlue:     "blue",
		AccentPurple:   "magenta",
		AccentCyan:     "cyan",
		AccentGreen:    "green",
		AccentYellow:   "yellow",
		AccentRed:      "red",
		DiffAdded:      "green",
		DiffRemoved:    "red",
		Comment:        "gray",
		Gray:           "gray",
		GradientColors: nil,
	}
}

// Default returns the default theme (dark)
func Default() ColorsTheme {
	return Dark()
}
