package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/DevvGwardo/hermes-tui/internal/config"
	"github.com/DevvGwardo/hermes-tui/internal/gateway"
	tea "github.com/charmbracelet/bubbletea"
)

// Msg types for Bubble Tea.
type (
	ConnectedMsg    struct{}
	ReconnectingMsg struct{ Attempt int; Delay time.Duration }
	SendResultMsg   struct{ Err error }
	StatusResultMsg struct{ Content string; Err error }
	ThinkResultMsg  struct{ Content string }

	// SessionsLoadedMsg is returned after listing sessions and fetching history.
	SessionsLoadedMsg struct {
		SessionKey string
		Model      string
		History    []gateway.Message
	}

	// AssistantResponseMsg is returned after the gateway processes a message.
	AssistantResponseMsg struct {
		Content    string
		SessionKey string
	}

	// MessageStreamErrorMsg is returned when the gateway reports an error during message processing.
	MessageStreamErrorMsg struct {
		Content string
	}
)

// Model is the main Bubble Tea model.
type Model struct {
	gateway     *gateway.Client
	sessionKey  string
	cfg         config.Config

	header      HeaderModel
	chat        ChatModel
	input       InputModel
	statusBar   StatusBarModel
	theme       Theme

	width       int
	height      int
	connected   bool
	streaming   bool
	streamBuf   string
	lastCtrlC   time.Time
	quitting    bool
	thinking    bool
	tickCount   int
	err         error
}

// NewModel creates the main TUI model.
func NewModel(gw *gateway.Client, sessionKey string, themeName string, cfg config.Config) Model {
	theme := NewTheme(themeName)
	if sessionKey == "" {
		sessionKey = cfg.SessionID
	}

	m := Model{
		gateway:    gw,
		sessionKey: sessionKey,
		cfg:        cfg,
		theme:      theme,
		header:     NewHeaderModel(theme, sessionKey, gw.BaseURL()),
		chat:       NewChatModel(theme),
		input:      NewInputModel(theme),
		statusBar:  NewStatusBarModel(theme),
		thinking:   cfg.Thinking,
		connected: false,
	}
	m.statusBar.SetThemeName(themeName)
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.connectCmd(),
	)
}

func (m Model) connectCmd() tea.Cmd {
	return func() tea.Msg {
		if err := m.gateway.Health(); err != nil {
			return StatusResultMsg{Err: fmt.Errorf("gateway unreachable: %w", err)}
		}
		return ConnectedMsg{}
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ConnectedMsg:
		m.connected = true
		m.header.SetConnected(true)
		m.statusBar.SetConnected(true)
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   "Connected to Hermes gateway.",
			Timestamp: time.Now(),
		})
		// Try to load sessions
		cmds = append(cmds, m.loadSessionsCmd())
		return m, tea.Batch(cmds...)

	case StatusResultMsg:
		if msg.Err != nil {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleError,
				Content:   fmt.Sprintf("Error: %v", msg.Err),
				Timestamp: time.Now(),
			})
		} else if msg.Content != "" {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleSystem,
				Content:   msg.Content,
				Timestamp: time.Now(),
			})
		}
		return m, nil

	case SendResultMsg:
		if msg.Err != nil {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleError,
				Content:   fmt.Sprintf("Failed to send: %v", msg.Err),
				Timestamp: time.Now(),
			})
			m.streaming = false
		}
		return m, nil

	case SessionsLoadedMsg:
		if msg.SessionKey != "" {
			m.sessionKey = msg.SessionKey
			m.statusBar.SetSession(msg.SessionKey)
			m.header.SetSession(msg.SessionKey)
		}
		if msg.Model != "" {
			m.statusBar.SetModel(msg.Model)
		}
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   fmt.Sprintf("Session: %s | Model: %s", msg.SessionKey, msg.Model),
			Timestamp: time.Now(),
		})
		for _, histMsg := range msg.History {
			role := RoleAssistant
			if histMsg.Role == "user" {
				role = RoleUser
			}
			m.chat.AddMessage(ChatMsg{
				Role:      role,
				Content:   histMsg.Content,
				Timestamp: time.Now(),
			})
		}
		m.statusBar.SetMessageCount(len(m.chat.GetMessages()))
		return m, nil

	case AssistantResponseMsg:
		if msg.SessionKey != "" && m.sessionKey == "" {
			m.sessionKey = msg.SessionKey
			m.statusBar.SetSession(msg.SessionKey)
			m.header.SetSession(msg.SessionKey)
		}
		if msg.Content != "" {
			m.chat.AddMessage(ChatMsg{
				Role:      RoleAssistant,
				Content:   msg.Content,
				Timestamp: time.Now(),
			})
		}
		m.statusBar.SetMessageCount(len(m.chat.GetMessages()))
		m.streaming = false
		return m, nil

	case MessageStreamErrorMsg:
		m.chat.AddMessage(ChatMsg{
			Role:      RoleError,
			Content:   msg.Content,
			Timestamp: time.Now(),
		})
		m.streaming = false
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout()
		return m, nil

	default:
		// Pass to input
		var inputCmd tea.Cmd
		m.input, inputCmd = m.input.Update(msg)
		cmds = append(cmds, inputCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		now := time.Now()
		if now.Sub(m.lastCtrlC) < 500*time.Millisecond {
			m.quitting = true
			return m, tea.Quit
		}
		m.lastCtrlC = now
		if m.input.Value() != "" {
			m.input.Reset()
			return m, nil
		}
		if m.streaming {
			return m, nil // TODO: abort
		}
		m.chat.AddMessage(ChatMsg{
			Role:      RoleSystem,
			Content:   "Press Ctrl+C again to exit.",
			Timestamp: time.Now(),
		})
		return m, nil

	case tea.KeyCtrlL:
		m.chat.Clear()
		return m, nil

	case tea.KeyPgUp:
		m.chat.ScrollUp(m.chat.Height() / 2)
		return m, nil

	case tea.KeyPgDown:
		m.chat.ScrollDown(m.chat.Height() / 2)
		return m, nil

	case tea.KeyHome:
		m.chat.ScrollToTop()
		return m, nil

	case tea.KeyEnd:
		m.chat.ScrollToBottom()
		return m, nil

	case tea.KeyEnter:
		return m.handleSubmit()
	}

	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	return m, inputCmd
}

func (m Model) handleSubmit() (tea.Model, tea.Cmd) {
	text := m.input.Value()
	if text == "" {
		return m, nil
	}
	m.input.Reset()
	m.streaming = true

	m.chat.AddMessage(ChatMsg{
		Role:      RoleUser,
		Content:   text,
		Timestamp: time.Now(),
	})

	return m, m.sendMessageCmd(text)
}

func (m Model) sendMessageCmd(text string) tea.Cmd {
	gw := m.gateway
	sessionKey := m.sessionKey
	return func() tea.Msg {
		// Resolve session if needed
		if sessionKey == "" {
			sessions, err := gw.ListSessions()
			if err != nil || len(sessions) == 0 {
				return SendResultMsg{Err: fmt.Errorf("no active sessions")}
			}
			sessionKey = sessions[0].Key
		}

		ch, err := gw.SendMessage(sessionKey, text)
		if err != nil {
			return SendResultMsg{Err: err}
		}

		// Drain the channel (blocks until gateway is done)
		for chunk := range ch {
			if strings.HasPrefix(chunk, "[error]") {
				return MessageStreamErrorMsg{
					Content: strings.TrimPrefix(chunk, "[error] "),
				}
			}
		}

		// Fetch history to get the assistant response
		msgs, err := gw.GetSessionHistory(sessionKey, 5)
		if err != nil {
			return SendResultMsg{Err: fmt.Errorf("fetch response: %w", err)}
		}

		var lastAssistant string
		for i := len(msgs) - 1; i >= 0; i-- {
			if msgs[i].Role == "assistant" {
				lastAssistant = msgs[i].Content
				break
			}
		}

		return AssistantResponseMsg{
			Content:    lastAssistant,
			SessionKey: sessionKey,
		}
	}
}

func (m Model) loadSessionsCmd() tea.Cmd {
	gw := m.gateway
	currentSessionKey := m.sessionKey
	return func() tea.Msg {
		sessions, err := gw.ListSessions()
		if err != nil {
			return StatusResultMsg{Err: fmt.Errorf("list sessions: %w", err)}
		}
		if len(sessions) == 0 {
			return StatusResultMsg{Content: "No active sessions found."}
		}

		resolvedKey := currentSessionKey
		if resolvedKey == "" {
			resolvedKey = sessions[0].Key
		}

		modelName := sessions[0].Model

		// Pre-fetch history
		var history []gateway.Message
		msgs, err := gw.GetSessionHistory(resolvedKey, 50)
		if err == nil {
			history = msgs
		}

		return SessionsLoadedMsg{
			SessionKey: resolvedKey,
			Model:      modelName,
			History:    history,
		}
	}
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return lipgloss.NewStyle().Foreground(m.theme.Palette.Muted).Render("Goodbye! \n")
	}

	margin := strings.Repeat(" ", sideMargin)
	inner := m.width - sideMargin*2
	if inner < 40 {
		inner = 40
	}

	header := addMargin(m.header.View(), margin)
	chat := addMargin(m.chat.View(), margin)
	sep := margin + lipgloss.NewStyle().Foreground(m.theme.Palette.Muted).Render(strings.Repeat("─", inner))
	input := addMargin(m.input.View(), margin)
	status := addMargin(m.statusBar.View(), margin)

	view := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", header, chat, sep, input, status)
	return view
}

func addMargin(block, margin string) string {
	lines := strings.Split(block, "\n")
	for i, line := range lines {
		lines[i] = margin + line
	}
	return strings.Join(lines, "\n")
}

const sideMargin = 2

func (m *Model) layout() {
	inner := m.width - sideMargin*2
	if inner < 40 {
		inner = 40
	}
	m.header.SetWidth(inner)
	m.statusBar.SetWidth(inner)
	m.input.SetWidth(inner)

	chatHeight := m.height - 9
	if chatHeight < 5 {
		chatHeight = 5
	}
	m.chat.SetSize(inner, chatHeight)
}
