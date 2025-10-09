# cli-demo

A modern TUI CLI built with the Charmbracelet stack, featuring a **production-grade InputPrompt component** that precisely implements the Gemini CLI specification.

## Features

- 🎨 **Component-Based Architecture**: Modular, reusable UI components
- 🌈 **Theme System**: Comprehensive theming with semantic colors
- 📱 **Responsive Design**: Adapts to terminal size changes
- ⌨️ **Professional InputPrompt**: Spec-compliant input with syntax highlighting
- 🎯 **Clean Layout**: Header, tips, input, and status bar components
- 🧪 **100% Tested**: Comprehensive test suite with benchmarks

## Charmbracelet Stack

| Library | Purpose | Usage |
|---------|---------|-------|
| 🫧 **bubbletea** | Core TUI framework (Elm architecture) | State management & event handling |
| 🧱 **bubbles** | UI components (textinput) | Text input functionality |
| 💅 **lipgloss** | Styling system | Colors, borders, layouts |
| 🌈 **termenv** | Terminal feature detection | Color support |
| 🧾 **runewidth** | Text width calculation | Unicode-aware width |

## Project Structure

```
cli-demo/
├── cmd/
│   ├── cli/                     # Main application
│   │   └── main.go
│   └── inputprompt-demo/        # InputPrompt demo
│       └── main.go
├── internal/tui/                # TUI implementation
│   ├── components/              # Reusable UI components
│   │   ├── ascii.go            # ASCII art logos
│   │   ├── header.go           # Header with responsive logo
│   │   ├── footer.go           # Footer/status bar
│   │   ├── tips.go             # Tips/hints display
│   │   ├── inputbox.go         # Legacy input (deprecated)
│   │   ├── inputprompt.go      # ⭐ Professional InputPrompt
│   │   ├── prompt_width.go     # Width calculation algorithm
│   │   └── *_test.go           # Component tests
│   ├── theme/                  # Theme system
│   │   ├── theme.go            # Core theme definitions
│   │   └── semantic.go         # Semantic color mapping
│   ├── utils/                  # Utility functions
│   │   ├── textutils.go        # Text measurement & manipulation
│   │   ├── highlight.go        # ⭐ Syntax highlighting engine
│   │   └── *_test.go           # Utility tests
│   └── model.go                # Main Bubble Tea model
├── docs/
│   ├── INPUTPROMPT_IMPLEMENTATION.md  # 📚 Complete implementation guide
│   ├── QUICK_START.md                 # 🚀 Quick start guide
│   └── REFACTORING_LOG.md             # 📝 Change history
├── go.mod
└── README.md
```

## Quick Start

### Main Application

```bash
# Build and run
go build -o cli ./cmd/cli
./cli
```

### InputPrompt Demo

Try the **interactive InputPrompt demo** to see all features:

```bash
# Build the demo
go build -o inputprompt-demo ./cmd/inputprompt-demo

# Run it
./inputprompt-demo
```

**Demo Features:**
- 🎨 Syntax highlighting (`/commands`, `@files`)
- 🔄 Multiple modes (Normal, Shell, YOLO, Reverse Search)
- 👻 Ghost text autocomplete
- 📏 Live width calculations
- 🌈 Dynamic border colors

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./internal/tui/components/...

# Run benchmarks
go test -bench=. ./internal/tui/components/...
```

## Architecture

### Components

Each component follows Go best practices:

- **Exported Configuration Structs**: Clear, typed configuration
- **Constructor Functions**: `NewXxx()` pattern for initialization
- **Update Methods**: Allow dynamic reconfiguration
- **Render Methods**: Return styled strings for display
- **Dimension Methods**: `Width()` and `Height()` for layout calculations

### Theme System

Three built-in themes:
- `Dark()` - Default dark theme (Catppuccin-inspired)
- `Light()` - Light theme with vibrant colors
- `ANSI()` - Basic ANSI colors for limited terminals

Semantic colors organize theming into logical groups:
- Text (Primary, Secondary, Link, Accent)
- Background (Primary, Diff Added/Removed)
- Border (Default, Focused)
- UI (Comment, Symbol, Gradient)
- Status (Error, Success, Warning)

### Layout

The UI follows the Gemini CLI design pattern:
1. **Header**: Responsive ASCII logo with optional version
2. **Tips**: Helpful getting-started information
3. **Input Box**: Bordered text input with focus styling
4. **Spacer**: Flexible spacing to push status to bottom
5. **Status Bar**: Left and right status indicators

## ⭐ InputPrompt Component

The **crown jewel** of this project is the `InputPrompt` component - a professional-grade implementation that precisely matches the [Gemini CLI specification](docs/INPUTPROMPT_IMPLEMENTATION.md).

### Key Features

#### 🧮 Exact Width Calculation
```go
// Implements the TypeScript formula exactly:
// inputWidth = floor(terminalWidth × 0.9) - 6
widths := components.CalculatePromptWidths(terminalWidth)
// Returns: InputWidth, ContainerWidth, SuggestionsWidth, FrameOverhead
```

#### 🎨 Syntax Highlighting
- **Slash Commands**: `/help`, `/reset` (highlighted on first line only)
- **File References**: `@path/to/file` (highlighted anywhere)
- **Smart Parsing**: Uses regex `(^\/[a-zA-Z0-9_-]+|@(?:\\ |[a-zA-Z0-9_./-])+)`
- **Visual Slicing**: `BuildSegmentsForVisualSlice()` for scrolling/wrapping

#### 🎭 Multiple Modes
```go
PromptModeNormal         // "> "  - default
PromptModeShell          // "! "  - shell commands
PromptModeReverseSearch  // "(r:) " - search mode
PromptModeYolo           // "* "  - auto-accept mode
```

#### 🖱️ Professional Cursor
- ANSI inverse video rendering (`\x1b[7m`)
- Unicode-aware positioning
- Works with emojis and wide characters

#### 👻 Ghost Text (Autocomplete)
```go
inputPrompt.SetGhostText("/help - show available commands")
// Displays gray preview text after cursor
```

#### 🌈 Dynamic Border Colors
- **Default**: Gray border when unfocused
- **Focused**: Blue border
- **Shell Mode**: Cyan border
- **YOLO Mode**: Red border
- **Auto-Edit**: Yellow border

### Usage Example

```go
import (
    "github.com/boke/cli-demo/internal/tui/components"
    "github.com/boke/cli-demo/internal/tui/theme"
    "github.com/boke/cli-demo/internal/tui/utils"
)

func main() {
    theme := theme.Default().ToSemanticColors()
    
    // Create the input prompt
    prompt := components.NewInputPrompt(components.InputPromptConfig{
        Theme:         theme,
        Placeholder:   "Type your message or @path/to/file",
        TerminalWidth: 100,
        Mode:          components.PromptModeNormal,
        ShowCursor:    true,
    })
    
    // Parse input for syntax highlighting
    tokens := utils.ParseInputForHighlighting("/help @README.md", 0)
    
    // In your Bubble Tea Update:
    cmd := prompt.Update(msg)
    
    // In your View:
    return prompt.Render()
}
```

### Test Coverage

```bash
$ go test -v ./internal/tui/components/...
✓ 32 test cases passing
✓ Width calculations verified
✓ Syntax highlighting validated
✓ Edge cases covered
```

### Documentation

See [`docs/INPUTPROMPT_IMPLEMENTATION.md`](docs/INPUTPROMPT_IMPLEMENTATION.md) for:
- Complete algorithm breakdown
- Visual examples
- Architecture details
- Performance characteristics
- Comparison with TypeScript version

## Key Bindings

### Main Application
- `Ctrl+C` / `Esc` - Quit
- `Enter` - Submit input
- Standard editing shortcuts

### InputPrompt Demo
- `Ctrl+S` - Toggle Shell mode
- `Ctrl+R` - Toggle Reverse Search
- `Ctrl+Y` - Toggle YOLO mode
- `Tab` - Try autocomplete (type `/h` or `@R` first)
- `Enter` - Submit
- `Ctrl+C` / `Esc` - Quit

## Development

Built with Go 1.22+ and follows:
- Standard Go project layout
- Clean Architecture principles
- Component-based design
- Exported APIs for extensibility

## Notes

- Requires a TTY to run
- Uses alt-screen for clean TUI experience
- Automatically adapts to terminal resizing
