package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBarModel is the bottom status bar.
type StatusBarModel struct {
	theme       Theme
	width      int
	model      string
	sessionID  string
	msgCount   int
	themeName  string
	connected  bool
}

// NewStatusBarModel creates a status bar with a theme.
func NewStatusBarModel(theme Theme) StatusBarModel {
	return StatusBarModel{
		theme:      theme,
		width:      80,
		model:     "unknown",
		sessionID: "—",
		msgCount:  0,
		themeName: "ocean",
		connected: false,
	}
}

// SetTheme updates the status bar's theme.
func (s *StatusBarModel) SetTheme(theme Theme) {
	s.theme = theme
}

// SetWidth sets the content width.
func (s *StatusBarModel) SetWidth(width int) {
	s.width = width
}

// SetModel updates the displayed model name.
func (s *StatusBarModel) SetModel(model string) {
	s.model = model
}

// SetSession updates the displayed session ID.
func (s *StatusBarModel) SetSession(session string) {
	s.sessionID = session
}

// SetMessageCount updates the message count.
func (s *StatusBarModel) SetMessageCount(n int) {
	s.msgCount = n
}

// SetThemeName updates the displayed theme name.
func (s *StatusBarModel) SetThemeName(name string) {
	s.themeName = name
}

// SetConnected updates the connection indicator.
func (s *StatusBarModel) SetConnected(connected bool) {
	s.connected = connected
}

// Model returns the current model name.
func (s StatusBarModel) Model() string {
	return s.model
}

// View renders the status bar.
func (s StatusBarModel) View() string {
	p := s.theme.Palette

	// Shorten session ID for display
	sessionDisplay := s.sessionID
	if len(sessionDisplay) > 12 {
		sessionDisplay = sessionDisplay[:12] + "…"
	}
	if sessionDisplay == "" {
		sessionDisplay = "—"
	}

	parts := []string{
		lipgloss.NewStyle().Foreground(p.FgMuted).Render("model:"),
		lipgloss.NewStyle().Foreground(p.Primary).Render(s.model),
		lipgloss.NewStyle().Foreground(p.Muted).Render("│"),
		lipgloss.NewStyle().Foreground(p.FgMuted).Render("session:"),
		lipgloss.NewStyle().Foreground(p.Accent).Render(sessionDisplay),
		lipgloss.NewStyle().Foreground(p.Muted).Render("│"),
		lipgloss.NewStyle().Foreground(p.FgMuted).Render("msgs:"),
		lipgloss.NewStyle().Foreground(p.Success).Render(fmt.Sprintf("%d", s.msgCount)),
		lipgloss.NewStyle().Foreground(p.Muted).Render("│"),
		lipgloss.NewStyle().Foreground(p.FgMuted).Render("theme:"),
		lipgloss.NewStyle().Foreground(p.Warning).Render(s.themeName),
	}

	line := strings.Join(parts, " ")

	// Pad to width
	padding := s.width - lipgloss.Width(line)
	if padding > 0 {
		line += strings.Repeat(" ", padding)
	}

	return lipgloss.NewStyle().
		Background(p.StatusBg).
		Foreground(p.StatusFg).
		Width(s.width).
		Render(line)
}
