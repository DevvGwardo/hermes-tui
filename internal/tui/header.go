package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HeaderModel is the top bar of the TUI.
type HeaderModel struct {
	theme       Theme
	width      int
	connected  bool
	sessionName string
	url        string
}

// NewHeaderModel creates a header with a theme.
func NewHeaderModel(theme Theme, sessionName, url string) HeaderModel {
	return HeaderModel{
		theme:       theme,
		width:       80,
		connected:   false,
		sessionName: sessionName,
		url:        url,
	}
}

// SetTheme updates the header's theme.
func (h *HeaderModel) SetTheme(theme Theme) {
	h.theme = theme
}

// SetWidth sets the content width.
func (h *HeaderModel) SetWidth(width int) {
	h.width = width
}

// SetConnected updates the connection indicator.
func (h *HeaderModel) SetConnected(connected bool) {
	h.connected = connected
}

// SetSession updates the displayed session name.
func (h *HeaderModel) SetSession(name string) {
	h.sessionName = name
}

// View renders the header bar.
func (h HeaderModel) View() string {
	p := h.theme.Palette

	connected := "○ disconnected"
	if h.connected {
		connected = "● connected"
	}

	connStyle := lipgloss.NewStyle().Foreground(p.Success)
	if !h.connected {
		connStyle = lipgloss.NewStyle().Foreground(p.FgMuted)
	}

	title := lipgloss.NewStyle().
		Foreground(p.HeaderFg).
		Bold(true).
		Render("Hermes TUI")

	sep := lipgloss.NewStyle().
		Foreground(p.Muted).
		Render(" │ ")

	sessionDisplay := h.sessionName
	if sessionDisplay == "" {
		sessionDisplay = "no session"
	}
	sessionStyle := lipgloss.NewStyle().Foreground(p.FgMuted).Render(sessionDisplay)

	conn := connStyle.Render(connected)

	// Build the line
	line := title + sep + sessionStyle + sep + conn

	// Pad to width
	padding := h.width - lipgloss.Width(line)
	if padding > 0 {
		line += strings.Repeat(" ", padding)
	}

	bgStyle := lipgloss.NewStyle().
		Background(p.HeaderBg).
		Width(h.width).
		Render(line)

	return lipgloss.NewStyle().
		Background(p.HeaderBg).
		Width(h.width).
		Render(bgStyle)
}
