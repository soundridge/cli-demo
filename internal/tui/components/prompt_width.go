package components

import "math"

// PromptWidths holds the calculated width values for the InputPrompt component.
// This matches the TypeScript implementation's PromptWidths interface.
type PromptWidths struct {
	// InputWidth is the actual usable width for input content
	InputWidth int

	// ContainerWidth is the total width including borders, padding, and prefix
	ContainerWidth int

	// SuggestionsWidth is the width for the suggestions dropdown
	SuggestionsWidth int

	// FrameOverhead is the total width consumed by borders, padding, and prefix
	FrameOverhead int
}

const (
	// WIDTH_FRACTION determines what percentage of terminal width to use (90%)
	widthFraction = 0.9

	// FRAME_PADDING_AND_BORDER accounts for:
	// - Left border: 1 char
	// - Right border: 1 char
	// - Left padding: 1 space
	// - Right padding: 1 space
	// Total: 4 chars
	framePaddingAndBorder = 4

	// PROMPT_PREFIX_WIDTH is the width of the prompt prefix (e.g., "> " or "! ")
	// This is the standard width for most modes
	promptPrefixWidth = 2

	// MIN_CONTENT_WIDTH ensures the input has at least this many characters of space
	minContentWidth = 2

	// TERMINAL_PADDING_X is the horizontal padding to account for terminal margins
	// This matches the TypeScript implementation
	terminalPaddingX = 8
)

// CalculatePromptWidths computes all width-related values for the InputPrompt.
// This implements the exact algorithm from the TypeScript version:
//
//	terminalWidth = actual terminal columns
//	baseWidth = floor(terminalWidth × 0.9)
//	innerContentWidth = baseWidth - 4 - 2
//	inputWidth = max(2, innerContentWidth)
//	containerWidth = inputWidth + 6
//	suggestionsWidth = max(20, floor(terminalWidth × 1.0))
//
// Example: terminalWidth = 100
//
//	baseWidth = floor(100 × 0.9) = 90
//	innerContentWidth = 90 - 4 - 2 = 84
//	inputWidth = max(2, 84) = 84
//	containerWidth = 84 + 6 = 90
//	suggestionsWidth = max(20, 100) = 100
func CalculatePromptWidths(terminalWidth int) PromptWidths {
	// Step 1: Calculate base width (terminal width × 0.9)
	baseWidth := int(math.Floor(float64(terminalWidth) * widthFraction))

	// Step 2: Subtract frame overhead to get inner content width
	innerContentWidth := baseWidth - framePaddingAndBorder - promptPrefixWidth

	// Step 3: Ensure minimum width
	inputWidth := max(minContentWidth, innerContentWidth)

	// Step 4: Calculate total container width
	frameOverhead := framePaddingAndBorder + promptPrefixWidth
	containerWidth := inputWidth + frameOverhead

	// Step 5: Calculate suggestions width (use full width, minimum 20)
	suggestionsWidth := max(20, int(math.Floor(float64(terminalWidth)*1.0)))

	return PromptWidths{
		InputWidth:       inputWidth,
		ContainerWidth:   containerWidth,
		SuggestionsWidth: suggestionsWidth,
		FrameOverhead:    frameOverhead,
	}
}

// GetEffectiveTerminalWidth applies the TERMINAL_PADDING_X correction.
// This matches the useTerminalSize hook from TypeScript:
//
//	effectiveColumns = max(60, rawColumns) - TERMINAL_PADDING_X
func GetEffectiveTerminalWidth(rawColumns int) int {
	effectiveWidth := max(60, rawColumns) - terminalPaddingX
	return max(1, effectiveWidth) // Ensure at least 1 column
}

// CalculatePromptWidthsFromRaw is a convenience function that applies
// the terminal padding correction before calculating widths.
func CalculatePromptWidthsFromRaw(rawTerminalWidth int) PromptWidths {
	effectiveWidth := GetEffectiveTerminalWidth(rawTerminalWidth)
	return CalculatePromptWidths(effectiveWidth)
}

// GetPrefixWidth returns the width of the prompt prefix based on mode.
// Most modes use 2 characters, but reverse search mode uses 4.
func GetPrefixWidth(mode PromptMode) int {
	if mode == PromptModeReverseSearch {
		return 4 // "(r:) "
	}
	return promptPrefixWidth // 2 for "> ", "! ", "* "
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

