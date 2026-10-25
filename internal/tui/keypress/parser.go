package keypress

import (
	"regexp"
	"strconv"
	"strings"
)

// ParsedKey represents a parsed keyboard sequence with the number of bytes consumed
type ParsedKey struct {
	Key    Key
	Length int // Number of bytes consumed from the input buffer
}

// KittyParser parses Kitty keyboard protocol sequences
type KittyParser struct {
	// Regex patterns for various sequence types
	revTabLegacy  *regexp.Regexp
	revTabParam   *regexp.Regexp
	arrowPrefix   *regexp.Regexp
	csiUPrefix    *regexp.Regexp
	legacyFuncKey *regexp.Regexp
}

// NewKittyParser creates a new Kitty protocol parser
func NewKittyParser() *KittyParser {
	return &KittyParser{
		// ESC [ Z - legacy reverse tab (Shift+Tab)
		revTabLegacy: regexp.MustCompile(`^` + regexp.QuoteMeta(ESC) + `\[Z`),

		// ESC [ 1 ; <mods> Z - parameterized reverse tab
		revTabParam: regexp.MustCompile(`^` + regexp.QuoteMeta(ESC) + `\[1;(\d+)Z`),

		// ESC [ 1 ; <mods> (A|B|C|D|H|F|P|Q|R|S) - arrows, home/end, F1-F4
		arrowPrefix: regexp.MustCompile(`^` + regexp.QuoteMeta(ESC) + `\[1;(\d+)([ABCDHFPQRS])`),

		// ESC [ <code> ; <mods> (u|~) - CSI-u or tilde-coded functional keys
		csiUPrefix: regexp.MustCompile(`^` + regexp.QuoteMeta(ESC) + `\[(\d+)(;(\d+))?([u~])`),

		// ESC [ (A|B|C|D|H|F) - legacy arrow keys without modifiers
		legacyFuncKey: regexp.MustCompile(`^` + regexp.QuoteMeta(ESC) + `\[([ABCDHF])`),
	}
}

// ParsePrefix attempts to parse a single complete Kitty/parameterized/legacy sequence
// from the start of the buffer. Returns nil if no valid sequence is found.
func (p *KittyParser) ParsePrefix(buffer string) *ParsedKey {
	// Try reverse tab (legacy): ESC [ Z
	if m := p.revTabLegacy.FindStringSubmatch(buffer); m != nil {
		return &ParsedKey{
			Key: Key{
				Name:          "tab",
				Ctrl:          false,
				Alt:           false,
				Shift:         true,
				Paste:         false,
				Sequence:      m[0],
				KittyProtocol: true,
			},
			Length: len(m[0]),
		}
	}

	// Try reverse tab (parameterized): ESC [ 1 ; <mods> Z
	if m := p.revTabParam.FindStringSubmatch(buffer); m != nil {
		mods, _ := strconv.Atoi(m[1])
		if mods >= KittyModifierEventTypesOffset {
			mods -= KittyModifierEventTypesOffset
		}
		bits := mods - KittyModifierBase
		alt := (bits & ModifierAltBit) == ModifierAltBit
		ctrl := (bits & ModifierCtrlBit) == ModifierCtrlBit

		return &ParsedKey{
			Key: Key{
				Name:          "tab",
				Ctrl:          ctrl,
				Alt:           alt,
				Shift:         true, // Reverse tab always implies shift
				Paste:         false,
				Sequence:      m[0],
				KittyProtocol: true,
			},
			Length: len(m[0]),
		}
	}

	// Try parameterized functional keys: ESC [ 1 ; <mods> (A|B|C|D|H|F|P|Q|R|S)
	if m := p.arrowPrefix.FindStringSubmatch(buffer); m != nil {
		mods, _ := strconv.Atoi(m[1])
		if mods >= KittyModifierEventTypesOffset {
			mods -= KittyModifierEventTypesOffset
		}
		bits := mods - KittyModifierBase
		shift := (bits & ModifierShiftBit) == ModifierShiftBit
		alt := (bits & ModifierAltBit) == ModifierAltBit
		ctrl := (bits & ModifierCtrlBit) == ModifierCtrlBit

		sym := m[2]
		symbolToName := map[string]string{
			"A": "up",
			"B": "down",
			"C": "right",
			"D": "left",
			"H": "home",
			"F": "end",
			"P": "f1",
			"Q": "f2",
			"R": "f3",
			"S": "f4",
		}

		if name, ok := symbolToName[sym]; ok {
			return &ParsedKey{
				Key: Key{
					Name:          name,
					Ctrl:          ctrl,
					Alt:           alt,
					Shift:         shift,
					Paste:         false,
					Sequence:      m[0],
					KittyProtocol: true,
				},
				Length: len(m[0]),
			}
		}
	}

	// Try CSI-u and tilde-coded functional keys: ESC [ <code> ; <mods> (u|~)
	if m := p.csiUPrefix.FindStringSubmatch(buffer); m != nil {
		keyCode, _ := strconv.Atoi(m[1])
		modifiers := KittyModifierBase
		if m[3] != "" {
			modifiers, _ = strconv.Atoi(m[3])
		}
		if modifiers >= KittyModifierEventTypesOffset {
			modifiers -= KittyModifierEventTypesOffset
		}
		modifierBits := modifiers - KittyModifierBase
		shift := (modifierBits & ModifierShiftBit) == ModifierShiftBit
		alt := (modifierBits & ModifierAltBit) == ModifierAltBit
		ctrl := (modifierBits & ModifierCtrlBit) == ModifierCtrlBit

		terminator := m[4]

		// Tilde-coded functional keys (Delete, Insert, PageUp/Down, Home/End)
		if terminator == "~" {
			nameMap := map[int]string{
				1: "home",
				2: "insert",
				3: "delete",
				4: "end",
				5: "pageup",
				6: "pagedown",
			}

			if name, ok := nameMap[keyCode]; ok {
				return &ParsedKey{
					Key: Key{
						Name:          name,
						Ctrl:          ctrl,
						Alt:           alt,
						Shift:         shift,
						Paste:         false,
						Sequence:      m[0],
						KittyProtocol: true,
					},
					Length: len(m[0]),
				}
			}
		}

		// Standard Kitty keycodes
		kittyKeyCodeToName := map[int]string{
			CharCodeESC:           "escape",
			KittyKeycodeTab:       "tab",
			KittyKeycodeBackspace: "backspace",
			KittyKeycodeEnter:     "return",
		}

		if name, ok := kittyKeyCodeToName[keyCode]; ok {
			return &ParsedKey{
				Key: Key{
					Name:          name,
					Ctrl:          ctrl,
					Alt:           alt,
					Shift:         shift,
					Paste:         false,
					Sequence:      m[0],
					KittyProtocol: true,
				},
				Length: len(m[0]),
			}
		}

		// Ctrl+letters (a-z)
		if ctrl && keyCode >= int('a') && keyCode <= int('z') {
			letter := string(rune(keyCode))
			return &ParsedKey{
				Key: Key{
					Name:          letter,
					Ctrl:          true,
					Alt:           alt,
					Shift:         shift,
					Paste:         false,
					Sequence:      m[0],
					KittyProtocol: true,
				},
				Length: len(m[0]),
			}
		}
	}

	// Try legacy function keys without modifiers: ESC [ (A|B|C|D|H|F)
	if m := p.legacyFuncKey.FindStringSubmatch(buffer); m != nil {
		sym := m[1]
		nameMap := map[string]string{
			"A": "up",
			"B": "down",
			"C": "right",
			"D": "left",
			"H": "home",
			"F": "end",
		}

		if name, ok := nameMap[sym]; ok {
			return &ParsedKey{
				Key: Key{
					Name:          name,
					Ctrl:          false,
					Alt:           false,
					Shift:         false,
					Paste:         false,
					Sequence:      m[0],
					KittyProtocol: true,
				},
				Length: len(m[0]),
			}
		}
	}

	return nil
}

// ParseBuffer attempts to parse and consume as many complete sequences as possible
// from the buffer. Returns the parsed keys and the remaining unparsed buffer.
func (p *KittyParser) ParseBuffer(buffer string) ([]Key, string) {
	var keys []Key

	for len(buffer) > 0 {
		parsed := p.ParsePrefix(buffer)
		if parsed == nil {
			// Look for the next potential CSI start beyond index 0
			csiStart := ESC + "["
			nextStart := strings.Index(buffer[1:], csiStart)
			if nextStart >= 0 {
				// Skip the incomplete/invalid sequence
				buffer = buffer[nextStart+1:]
				continue
			}
			// No more sequences found
			break
		}

		// Consume the parsed prefix
		keys = append(keys, parsed.Key)
		buffer = buffer[parsed.Length:]
	}

	return keys, buffer
}
