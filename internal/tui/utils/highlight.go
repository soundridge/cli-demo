package utils

import (
	"regexp"
	"strings"
)

// HighlightToken represents a parsed token with its type for syntax highlighting.
// This matches the TypeScript HighlightToken interface.
type HighlightToken struct {
	Text string
	Type HighlightTokenType
}

// HighlightTokenType represents the type of syntax token
type HighlightTokenType string

const (
	// TokenTypeDefault represents normal text
	TokenTypeDefault HighlightTokenType = "default"

	// TokenTypeCommand represents slash commands (e.g., /help)
	TokenTypeCommand HighlightTokenType = "command"

	// TokenTypeFile represents file references (e.g., @path/to/file)
	TokenTypeFile HighlightTokenType = "file"
)

// highlightRegex matches slash commands and file references.
// Pattern from TypeScript: /(^\/[a-zA-Z0-9_-]+|@(?:\\ |[a-zA-Z0-9_./-])+)/g
//
// Matches:
// - Slash commands at line start: /help, /reset, /exit
// - File references anywhere: @path/to/file, @./relative, @../parent
// - Escaped spaces in file paths: @my\ file.txt
var highlightRegex = regexp.MustCompile(`(^\/[a-zA-Z0-9_-]+|@(?:\\ |[a-zA-Z0-9_./-])+)`)

// ParseInputForHighlighting tokenizes input text for syntax highlighting.
// This is a direct port of the TypeScript parseInputForHighlighting function.
//
// Rules:
// - Slash commands (/) are only highlighted at line index 0
// - File references (@) are highlighted on any line
// - All other text is marked as 'default'
//
// Parameters:
//   - text: The input text to parse
//   - lineIndex: The index of the current line (0-based, only line 0 highlights commands)
//
// Example:
//
//	tokens := ParseInputForHighlighting("/help show me @README.md", 0)
//	// Returns:
//	// [
//	//   {Text: "/help", Type: "command"},
//	//   {Text: " show me ", Type: "default"},
//	//   {Text: "@README.md", Type: "file"},
//	// ]
func ParseInputForHighlighting(text string, lineIndex int) []HighlightToken {
	if text == "" {
		return []HighlightToken{{Text: "", Type: TokenTypeDefault}}
	}

	tokens := []HighlightToken{}
	lastIndex := 0

	// Find all matches
	matches := highlightRegex.FindAllStringIndex(text, -1)

	for _, match := range matches {
		matchIndex := match[0]
		matchEnd := match[1]
		fullMatch := text[matchIndex:matchEnd]

		// Add the text before the match as a default token
		if matchIndex > lastIndex {
			tokens = append(tokens, HighlightToken{
				Text: text[lastIndex:matchIndex],
				Type: TokenTypeDefault,
			})
		}

		// Determine token type
		var tokenType HighlightTokenType
		if strings.HasPrefix(fullMatch, "/") {
			tokenType = TokenTypeCommand
		} else {
			tokenType = TokenTypeFile
		}

		// Only highlight slash commands if the line index is 0
		if tokenType == TokenTypeCommand && lineIndex != 0 {
			tokens = append(tokens, HighlightToken{
				Text: fullMatch,
				Type: TokenTypeDefault,
			})
		} else {
			tokens = append(tokens, HighlightToken{
				Text: fullMatch,
				Type: tokenType,
			})
		}

		lastIndex = matchEnd
	}

	// Add any remaining text after the last match
	if lastIndex < len(text) {
		tokens = append(tokens, HighlightToken{
			Text: text[lastIndex:],
			Type: TokenTypeDefault,
		})
	}

	return tokens
}

// BuildSegmentsForVisualSlice builds a slice of tokens within a visual range.
// This is a direct port of the TypeScript buildSegmentsForVisualSlice function.
//
// This function is useful for rendering only a portion of the input text
// (e.g., for horizontal scrolling or wrapping) while preserving syntax highlighting.
//
// Parameters:
//   - tokens: The tokens to slice
//   - sliceStart: Start position in code points (inclusive)
//   - sliceEnd: End position in code points (exclusive)
//
// Returns:
//   - A new slice of tokens that fall within the specified range
//   - Adjacent tokens of the same type are merged
//
// Example:
//
//	tokens := ParseInputForHighlighting("/help @file.txt", 0)
//	// tokens = [{"/help", "command"}, {" ", "default"}, {"@file.txt", "file"}]
//
//	// Get characters 3-8 (code points)
//	sliced := BuildSegmentsForVisualSlice(tokens, 3, 8)
//	// sliced = [{"lp ", "command"}, {"@f", "file"}]
func BuildSegmentsForVisualSlice(
	tokens []HighlightToken,
	sliceStart int,
	sliceEnd int,
) []HighlightToken {
	if sliceStart >= sliceEnd {
		return []HighlightToken{}
	}

	segments := []HighlightToken{}
	tokenCpStart := 0

	for _, token := range tokens {
		tokenLen := CodePointLength(token.Text)
		tokenStart := tokenCpStart
		tokenEnd := tokenStart + tokenLen

		// Calculate overlap between token and requested slice
		overlapStart := max(tokenStart, sliceStart)
		overlapEnd := min(tokenEnd, sliceEnd)

		if overlapStart < overlapEnd {
			// Extract the overlapping portion
			sliceStartInToken := overlapStart - tokenStart
			sliceEndInToken := overlapEnd - tokenStart
			rawSlice := CodePointSlice(token.Text, sliceStartInToken, sliceEndInToken)

			// Merge with previous segment if same type
			lastIdx := len(segments) - 1
			if lastIdx >= 0 && segments[lastIdx].Type == token.Type {
				segments[lastIdx].Text += rawSlice
			} else {
				segments = append(segments, HighlightToken{
					Type: token.Type,
					Text: rawSlice,
				})
			}
		}

		tokenCpStart += tokenLen
	}

	return segments
}

// IsCommandToken returns true if the token is a command token
func (t HighlightToken) IsCommandToken() bool {
	return t.Type == TokenTypeCommand
}

// IsFileToken returns true if the token is a file token
func (t HighlightToken) IsFileToken() bool {
	return t.Type == TokenTypeFile
}

// IsDefaultToken returns true if the token is a default text token
func (t HighlightToken) IsDefaultToken() bool {
	return t.Type == TokenTypeDefault
}

// SplitInputLines splits input text into lines while preserving structure.
// This is a helper function for processing multi-line input.
func SplitInputLines(text string) []string {
	if text == "" {
		return []string{""}
	}
	return strings.Split(text, "\n")
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
