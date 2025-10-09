# Keypress Package

A sophisticated keyboard input handling package for Go CLI applications, inspired by and ported from the Google Gemini CLI's TypeScript implementation.

## Features

- **Kitty Keyboard Protocol Support**: Advanced terminal keyboard protocol with better modifier key detection
- **Paste Mode Detection**: Automatically detects and handles paste operations (ESC[200~ / ESC[201~)
- **Drag-and-Drop File Paths**: Detects when users drag files into the terminal (quoted strings with timeout)
- **Backslash-Enter Line Continuation**: Detects `\` followed by `Enter` for multiline input
- **Advanced Escape Sequence Parsing**: Parses CSI-u sequences, tilde-coded function keys, and legacy sequences
- **Modifier Key Support**: Full support for Ctrl, Alt/Meta, and Shift modifiers
- **Thread-Safe**: Uses mutex locks for concurrent safety

## Architecture

### Components

1. **constants.go**: Constants for escape sequences, character codes, and timeouts
2. **key.go**: Key event structure and type definitions
3. **parser.go**: Kitty protocol parser for escape sequences
4. **handler.go**: Main handler that orchestrates all features

## Usage

### Basic Setup

```go
import "github.com/boke/cli-demo/internal/tui/keypress"

// Create a handler
handler := keypress.NewHandler(keypress.HandlerConfig{
    KittyEnabled:   true,  // Enable Kitty protocol parsing
    DebugKeystroke: false, // Enable debug logging
    OnPaste: func(content string) {
        // Handle pasted content
        fmt.Println("Pasted:", content)
    },
    OnKey: func(key keypress.Key) bool {
        // Handle individual key events
        fmt.Printf("Key: %s, Ctrl: %v, Alt: %v\n", key.Name, key.Ctrl, key.Alt)
        return true // Return true to consume the event
    },
})

// Clean up when done
defer handler.Cleanup()
```

### Processing Bubbletea Key Messages

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Process through enhanced keypress handler
        key := m.keypressHandler.ProcessTeaKey(msg)
        if key == nil {
            // Event consumed by handler (e.g., paste marker)
            return m, nil
        }
        
        // Handle the enhanced key event
        if key.Ctrl && key.Name == "c" {
            return m, tea.Quit
        }
        
        if key.Paste {
            // This is pasted content
            m.insertText(key.Sequence)
        }
    }
    return m, nil
}
```

### Key Event Structure

```go
type Key struct {
    Name          string    // Key name (e.g., "a", "enter", "up")
    Ctrl          bool      // Ctrl modifier
    Alt           bool      // Alt/Meta modifier
    Shift         bool      // Shift modifier
    Paste         bool      // True if from paste operation
    Sequence      string    // Raw sequence
    KittyProtocol bool      // Parsed from Kitty protocol
    Original      *tea.KeyMsg // Original bubbletea message
}
```

## Supported Features

### Kitty Protocol Sequences

The handler supports the following Kitty keyboard protocol features:

- **CSI-u sequences**: `ESC [ <code> ; <modifiers> u`
- **Parameterized functional keys**: `ESC [ 1 ; <mods> (A|B|C|D|H|F|P|Q|R|S)`
- **Reverse tab**: `ESC [ Z` and `ESC [ 1 ; <mods> Z`
- **Tilde-coded keys**: `ESC [ <code> ; <mods> ~` (Delete, Insert, PageUp/Down, etc.)
- **Legacy sequences**: `ESC [ (A|B|C|D|H|F)` (arrows, Home, End)

### Paste Detection

Automatically detects paste mode using terminal sequences:
- **Start**: `ESC[200~`
- **End**: `ESC[201~`

Content between these markers is buffered and delivered via the `OnPaste` callback.

### Drag-and-Drop Detection

Detects when users drag files into the terminal:
- Starts on single quote (`'`) or double quote (`"`)
- Accumulates input until 100ms timeout
- Delivers complete path via `OnPaste` callback

### Backslash-Enter Detection

Detects `\` followed by `Enter` within 50ms for line continuation:
- Returns `Key{Name: "return", Shift: true}` to indicate multiline input
- Falls back to regular backslash if no Enter follows

## Modifier Key Encoding

Kitty protocol uses a base-1 encoding for modifiers:

```
Base value: 1
+ Shift: 1
+ Alt:   2
+ Ctrl:  4
+ Super: 8
+ Hyper: 16
+ Meta:  32

Examples:
- No modifiers: 1
- Shift only: 2 (1 + 1)
- Ctrl only: 5 (1 + 4)
- Ctrl+Shift: 6 (1 + 1 + 4)
- Ctrl+Alt+Shift: 8 (1 + 1 + 2 + 4)
```

Some sequences also add 128 to indicate event types (keypress vs release).

## Key Name Mappings

### Arrow Keys
- `A` → `up`
- `B` → `down`
- `C` → `right`
- `D` → `left`

### Special Keys
- `H` → `home`
- `F` → `end`
- `P` → `f1`
- `Q` → `f2`
- `R` → `f3`
- `S` → `f4`

### Tilde-Coded Keys
- `1~` → `home`
- `2~` → `insert`
- `3~` → `delete`
- `4~` → `end`
- `5~` → `pageup`
- `6~` → `pagedown`

## Integration with InputPrompt

The `InputPrompt` component automatically integrates with the keypress handler:

```go
// In InputPrompt component
ip := &InputPrompt{...}
ip.keypressHandler = keypress.NewHandler(keypress.HandlerConfig{
    KittyEnabled: config.EnableKittyProtocol,
    OnPaste: func(content string) {
        ip.insertText(content) // Auto-insert pasted content
    },
})

// Process keys in Update
func (ip *InputPrompt) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        key := ip.keypressHandler.ProcessTeaKey(msg)
        if key != nil {
            return ip.handleEnhancedKey(*key)
        }
    }
}
```

## Advanced Features

### Word Navigation

The enhanced key handler supports:
- **Ctrl+Left/Right**: Navigate by words
- **Ctrl+Backspace**: Delete word backward
- **Ctrl+Delete**: Delete word forward

### Multiline Input

- **Shift+Enter**: Insert newline (detected from backslash-enter or Kitty protocol)

## Debugging

Enable debug keystroke logging:

```go
handler := keypress.NewHandler(keypress.HandlerConfig{
    DebugKeystroke: true,
})
```

This will log:
- Kitty buffer accumulation
- Parsed sequences
- Buffer overflows
- Invalid/incomplete sequences

## Buffer Overflow Protection

The handler protects against malformed sequences:
- Maximum buffer size: 256 bytes
- Automatically clears buffer on overflow
- Attempts to recover by finding next valid CSI sequence

## Thread Safety

All handler state is protected by a mutex (`sync.Mutex`):
- Safe to call `ProcessTeaKey` from multiple goroutines
- Timers are properly synchronized
- Cleanup is thread-safe

## Cleanup

Always call `Cleanup()` when disposing the handler:

```go
defer handler.Cleanup()
```

This will:
- Stop all timers
- Flush pending paste data
- Flush pending drag data
- Release resources

## Comparison with TypeScript Implementation

This Go implementation closely follows the TypeScript/Node.js version from Google Gemini CLI:

| Feature | TypeScript | Go |
|---------|-----------|-----|
| Kitty protocol | ✅ | ✅ |
| Paste detection | ✅ | ✅ |
| Drag-and-drop | ✅ | ✅ |
| Backslash-enter | ✅ | ✅ |
| Focus events | ✅ | ✅ (ignored) |
| Raw mode | readline/stdin | bubbletea |
| Buffer parsing | Regex | Regex |

## Performance Considerations

- Regex patterns are compiled once at initialization
- Minimal allocations for key events
- Efficient string buffer management for paste/drag operations
- Lazy timer creation (only when needed)

## License

Same as parent project.

