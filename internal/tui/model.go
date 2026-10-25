package tui

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/boke/cli-demo/internal/tui/components"
	"github.com/boke/cli-demo/internal/tui/theme"
	"github.com/boke/cli-demo/internal/tui/utils"
)

// Model is the Bubble Tea model for the UI.
type Model struct {
	header         *components.Header
	tips           *components.Tips
	footer         *components.Footer
	inputPrompt    *components.InputPrompt // ⭐ 使用新的专业级 InputPrompt
	theme          theme.ColorsTheme
	semanticTheme  theme.SemanticColors
	terminalWidth  int
	terminalHeight int
	version        string
	workingDir     string
	ready          bool
	mode           components.PromptMode
	history        []string // 输入历史
}

func New(width, height int) Model {
	t := theme.Default()
	semanticTheme := t.ToSemanticColors()

	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	// Get current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "~"
	}

	// Create header component
	header := components.NewHeader(components.HeaderConfig{
		CustomAsciiArt: "",
		Version:        "1.0.0",
		ShowVersion:    true,
		TerminalWidth:  width,
		Theme:          t,
	})

	// Create tips component
	tips := components.NewTips(components.TipsConfig{
		Theme: t,
		Items: []string{
			"💡 Type /help to see available commands",
			"📁 Use @path/to/file to reference files (they'll be highlighted!)",
			"🎨 Try Ctrl+S for shell mode, Ctrl+Y for YOLO mode",
			"✨ Professional InputPrompt with syntax highlighting is now active!",
		},
		Width: width,
	})

	// Create footer component
	footer := components.NewFooter(components.FooterConfig{
		Theme:          t,
		Width:          width,
		WorkingDir:     workingDir,
		Model:          "gemini-2.5-pro",
		ContextUsage:   "(100% context left)",
		SandboxStatus:  "",
		ShowWorkingDir: true,
		ShowSandbox:    true,
		ShowModel:      true,
		DebugMode:      false,
		DebugMessage:   "",
	})

	// ⭐ 创建专业级 InputPrompt 组件 with enhanced keyboard handling
	inputPrompt := components.NewInputPrompt(components.InputPromptConfig{
		Theme:               semanticTheme,
		Placeholder:         "Type your message or @path/to/file",
		TerminalWidth:       width,
		Mode:                components.PromptModeNormal,
		ShowCursor:          true,
		EnableKittyProtocol: true,  // Enable Kitty keyboard protocol for advanced key handling
		DebugKeystroke:      false, // Set to true for debugging keyboard input
	})

	// Set up paste handler for drag-and-drop file paths
	inputPrompt.SetPasteHandler(func(content string) {
		// Handle pasted content (e.g., from drag-and-drop)
		// This will automatically insert the content at the cursor position
		// You can customize this behavior if needed
	})

	m := Model{
		header:         header,
		tips:           tips,
		footer:         footer,
		inputPrompt:    inputPrompt,
		theme:          t,
		semanticTheme:  semanticTheme,
		terminalWidth:  width,
		terminalHeight: height,
		version:        "1.0.0",
		workingDir:     workingDir,
		ready:          false,
		mode:           components.PromptModeNormal,
		history:        []string{},
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return m.inputPrompt.Focus()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			// Esc 清空输入
			m.inputPrompt.Reset()
			return m, nil

		case "enter":
			// 提交输入
			value := m.inputPrompt.Value()
			if value != "" {
				m.history = append(m.history, value)

				// 展示输入内容和语法高亮分析
				m.processInput(value)

				m.inputPrompt.Reset()
			}
			return m, nil

		case "ctrl+s":
			// 切换 Shell 模式
			if m.mode == components.PromptModeShell {
				m.mode = components.PromptModeNormal
			} else {
				m.mode = components.PromptModeShell
			}
			m.updatePromptMode()
			return m, nil

		case "ctrl+y":
			// 切换 YOLO 模式
			if m.mode == components.PromptModeYolo {
				m.mode = components.PromptModeNormal
			} else {
				m.mode = components.PromptModeYolo
			}
			m.updatePromptMode()
			return m, nil

		case "ctrl+r":
			// 切换反向搜索模式
			if m.mode == components.PromptModeReverseSearch {
				m.mode = components.PromptModeNormal
			} else {
				m.mode = components.PromptModeReverseSearch
			}
			m.updatePromptMode()
			return m, nil

		case "tab":
			// 简单的自动补全演示
			m.handleAutocomplete()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		m.ready = true

		// Update all components with new dimensions
		m.header.Update(components.HeaderConfig{
			CustomAsciiArt: "",
			Version:        m.version,
			ShowVersion:    true,
			TerminalWidth:  msg.Width,
			Theme:          m.theme,
		})

		m.tips.Update(components.TipsConfig{
			Theme: m.theme,
			Items: m.tips.Config.Items,
			Width: msg.Width,
		})

		m.inputPrompt.UpdateConfig(components.InputPromptConfig{
			Theme:               m.semanticTheme,
			Placeholder:         "Type your message or @path/to/file",
			TerminalWidth:       msg.Width,
			Mode:                m.mode,
			ShellModeActive:     m.mode == components.PromptModeShell,
			YoloMode:            m.mode == components.PromptModeYolo,
			ShowCursor:          true,
			EnableKittyProtocol: true,
			DebugKeystroke:      false,
		})

		m.footer.Update(components.FooterConfig{
			Theme:          m.theme,
			Width:          msg.Width,
			WorkingDir:     m.workingDir,
			Model:          "gemini-2.5-pro",
			ContextUsage:   "(100% context left)",
			SandboxStatus:  "",
			ShowWorkingDir: true,
			ShowSandbox:    true,
			ShowModel:      true,
			DebugMode:      false,
			DebugMessage:   "",
		})
	}

	// Update input prompt
	cmd := m.inputPrompt.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// updatePromptMode 更新输入提示的模式
func (m *Model) updatePromptMode() {
	m.inputPrompt.SetMode(m.mode)
	m.inputPrompt.UpdateConfig(components.InputPromptConfig{
		Theme:               m.semanticTheme,
		Placeholder:         "Type your message or @path/to/file",
		TerminalWidth:       m.terminalWidth,
		Mode:                m.mode,
		ShellModeActive:     m.mode == components.PromptModeShell,
		YoloMode:            m.mode == components.PromptModeYolo,
		ShowCursor:          true,
		EnableKittyProtocol: true,
		DebugKeystroke:      false,
	})
}

// handleAutocomplete 处理自动补全
func (m *Model) handleAutocomplete() {
	currentValue := m.inputPrompt.Value()

	suggestions := map[string]string{
		"/h":   "/help - show available commands",
		"/he":  "/help - show available commands",
		"/hel": "/help - show available commands",
		"/r":   "/reset - reset the session",
		"/re":  "/reset - reset the session",
		"@R":   "@README.md",
		"@r":   "@README.md",
		"@d":   "@docs/",
		"@i":   "@internal/",
		"@c":   "@cmd/",
	}

	if suggestion, ok := suggestions[currentValue]; ok {
		m.inputPrompt.SetGhostText(suggestion)
	} else {
		m.inputPrompt.SetGhostText("")
	}
}

// processInput 处理输入并展示语法高亮分析
func (m *Model) processInput(value string) {
	// 这里可以添加实际的处理逻辑
	// 例如：执行命令、加载文件等

	// 演示语法高亮分析
	lines := utils.SplitInputLines(value)
	for lineIdx, line := range lines {
		tokens := utils.ParseInputForHighlighting(line, lineIdx)

		for _, token := range tokens {
			if token.IsCommandToken() {
				// 处理命令
				switch token.Text {
				case "/help":
					// 显示帮助
				case "/reset":
					// 重置会话
				}
			} else if token.IsFileToken() {
				// 处理文件引用
				// 可以尝试读取文件等
			}
		}
	}
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Render all components
	header := m.header.Render()
	tips := m.tips.Render()
	footer := m.footer.Render()
	inputPrompt := m.inputPrompt.Render()

	// 显示当前模式状态
	var modeIndicator string
	switch m.mode {
	case components.PromptModeShell:
		modeIndicator = "🐚 Shell Mode | "
	case components.PromptModeYolo:
		modeIndicator = "⚡ YOLO Mode | "
	case components.PromptModeReverseSearch:
		modeIndicator = "🔍 Search Mode | "
	default:
		modeIndicator = ""
	}

	// 显示最近的输入历史（最多3条）
	var historyView string
	if len(m.history) > 0 {
		historyView = "\n📜 Recent inputs:\n"
		start := len(m.history) - 3
		if start < 0 {
			start = 0
		}
		for i := start; i < len(m.history); i++ {
			historyView += "  " + m.history[i] + "\n"
		}
	}

	// Layout:
	// - Header (ASCII logo + version)
	// - Tips
	// - Mode indicator
	// - Input Prompt (⭐ 新的专业级组件)
	// - History
	// - Footer
	var output strings.Builder
	output.WriteString(header)
	output.WriteString("\n")
	output.WriteString(tips)
	output.WriteString("\n")
	if modeIndicator != "" {
		output.WriteString(modeIndicator)
		output.WriteString("Ctrl+S/Y/R to toggle modes\n")
	}
	output.WriteString(inputPrompt)
	if historyView != "" {
		output.WriteString(historyView)
	}
	output.WriteString("\n")
	output.WriteString(footer)

	return output.String()
}
