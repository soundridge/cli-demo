package utils

import (
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/mattn/go-runewidth"
)

// ansiRegex matches ANSI escape sequences.
var ansiRegex = regexp.MustCompile(`[\x1b\x9b][\[\]()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]`)

// stringWidthCache stores cached string width calculations for performance.
var (
	stringWidthCache = make(map[string]int)
	cacheMu          sync.RWMutex
	maxCacheSize     = 1000
)

// GetAsciiArtWidth calculates the maximum width of a multi-line ASCII art string.
// It returns the length of the longest line in the ASCII art.
// This matches the TypeScript implementation's getAsciiArtWidth function.
func GetAsciiArtWidth(art string) int {
	if art == "" {
		return 0
	}

	lines := strings.Split(art, "\n")
	maxWidth := 0

	for _, line := range lines {
		// Use simple length for ASCII art (no ANSI codes expected in logo)
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	return maxWidth
}

// GetCachedStringWidth calculates the display width of a string with caching.
// This is optimized for performance, similar to the TypeScript implementation.
// It handles multi-byte characters, emoji, and wide characters correctly.
func GetCachedStringWidth(str string) int {
	// Fast path for ASCII printable characters (0x20-0x7E)
	if isASCIIPrintable(str) {
		return len(str)
	}

	// Check cache
	cacheMu.RLock()
	if width, found := stringWidthCache[str]; found {
		cacheMu.RUnlock()
		return width
	}
	cacheMu.RUnlock()

	// Calculate width using runewidth
	width := runewidth.StringWidth(str)

	// Store in cache
	cacheMu.Lock()
	if len(stringWidthCache) < maxCacheSize {
		stringWidthCache[str] = width
	}
	cacheMu.Unlock()

	return width
}

// isASCIIPrintable checks if all characters are ASCII printable (0x20-0x7E).
func isASCIIPrintable(str string) bool {
	for i := 0; i < len(str); i++ {
		if str[i] < 0x20 || str[i] > 0x7E {
			return false
		}
	}
	return true
}

// ClearStringWidthCache clears the string width cache.
// Useful for testing or memory management.
func ClearStringWidthCache() {
	cacheMu.Lock()
	stringWidthCache = make(map[string]int)
	cacheMu.Unlock()
}

// StripANSI removes ANSI escape sequences from a string.
// This is equivalent to the strip-ansi npm package functionality.
func StripANSI(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

// StripUnsafeCharacters removes characters that can break terminal rendering.
// This includes:
// - ANSI escape sequences
// - VT control sequences
// - C0 control chars (0x00-0x1F) except CR/LF
// - C1 control chars (0x80-0x9F)
func StripUnsafeCharacters(str string) string {
	// First strip ANSI codes
	stripped := StripANSI(str)

	// Filter unsafe characters
	var builder strings.Builder
	builder.Grow(len(stripped))

	for _, r := range stripped {
		// Preserve CR/LF for line handling
		if r == '\n' || r == '\r' {
			builder.WriteRune(r)
			continue
		}

		// Remove C0 control chars (0x00-0x1F) except CR/LF
		// Examples: BELL(0x07), BS(0x08), VT(0x0B), FF(0x0C)
		if r >= 0x00 && r <= 0x1F {
			continue
		}

		// Remove C1 control chars (0x80-0x9F) - legacy 8-bit control codes
		if r >= 0x80 && r <= 0x9F {
			continue
		}

		// Preserve all other characters including Unicode/emojis and DEL(0x7F)
		builder.WriteRune(r)
	}

	return builder.String()
}

// ToCodePoints converts a string to a slice of code points (runes).
// This is equivalent to Array.from(str) in JavaScript.
// It handles surrogate pairs correctly, treating emoji as single units.
func ToCodePoints(str string) []string {
	// Fast path for ASCII
	if isASCII(str) {
		result := make([]string, len(str))
		for i := 0; i < len(str); i++ {
			result[i] = string(str[i])
		}
		return result
	}

	// Handle Unicode
	runes := []rune(str)
	result := make([]string, len(runes))
	for i, r := range runes {
		result[i] = string(r)
	}
	return result
}

// isASCII checks if all characters in string are ASCII (0-127).
func isASCII(str string) bool {
	for i := 0; i < len(str); i++ {
		if str[i] > 127 {
			return false
		}
	}
	return true
}

// CodePointLength returns the number of code points (not bytes) in a string.
// This is equivalent to Array.from(str).length in JavaScript.
func CodePointLength(str string) int {
	return len([]rune(str))
}

// CodePointSlice slices a string by code point indices (not byte indices).
// This is equivalent to Array.from(str).slice(start, end).join(”) in JavaScript.
func CodePointSlice(str string, start int, end int) string {
	runes := []rune(str)
	if start < 0 {
		start = 0
	}
	if end < 0 || end > len(runes) {
		end = len(runes)
	}
	if start >= end {
		return ""
	}
	return string(runes[start:end])
}

// EscapeANSIControlCodes escapes ANSI control characters in a string.
// This makes them safe to display without interpretation.
func EscapeANSIControlCodes(str string) string {
	if !ansiRegex.MatchString(str) {
		return str
	}

	return ansiRegex.ReplaceAllStringFunc(str, func(match string) string {
		// Convert to a JSON-escaped string and remove quotes
		escaped := strings.ReplaceAll(match, "\x1b", "\\x1b")
		escaped = strings.ReplaceAll(escaped, "\x9b", "\\x9b")
		return escaped
	})
}

// IsPrintable checks if a rune is a printable character.
func IsPrintable(r rune) bool {
	return unicode.IsPrint(r)
}

// GetTerminalColumns returns a safe terminal width value.
// Returns default of 80 if the provided width is invalid.
func GetTerminalColumns(width int) int {
	if width <= 0 {
		return 80 // Default fallback
	}
	return width
}
