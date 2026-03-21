package tui

import (
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// MsgRole identifies the sender of a message.
type MsgRole string

const (
	RoleUser      MsgRole = "user"
	RoleAssistant MsgRole = "assistant"
	RoleSystem    MsgRole = "system"
	RoleError     MsgRole = "error"
)

// ChatMsg represents a single chat message.
type ChatMsg struct {
	Role      MsgRole
	Content   string
	Timestamp time.Time
	Streaming bool
	RunID     string
}

// glamourRenderer caches a glamour renderer per width.
var (
	glamourMu     sync.Mutex
	glamourCache  *glamour.TermRenderer
	glamourCacheW int
)

func getGlamourRenderer(width int) *glamour.TermRenderer {
	glamourMu.Lock()
	defer glamourMu.Unlock()
	if glamourCache != nil && glamourCacheW == width {
		return glamourCache
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	glamourCache = r
	glamourCacheW = width
	return r
}

// RenderMessage renders a single message with the given theme and width.
func RenderMessage(msg ChatMsg, theme Theme, width int) string {
	switch msg.Role {
	case RoleUser:
		return renderUserMessage(msg, theme, width)
	case RoleAssistant:
		return renderAssistantMessage(msg, theme, width)
	case RoleSystem:
		return renderSystemMessage(msg, theme, width)
	case RoleError:
		return renderErrorMessage(msg, theme, width)
	default:
		return msg.Content
	}
}

func renderUserMessage(msg ChatMsg, theme Theme, width int) string {
	p := theme.Palette

	name := lipgloss.NewStyle().
		Foreground(p.Secondary).
		Bold(true).
		Render("You")

	ts := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render(msg.Timestamp.Format("15:04"))

	headerGap := width - lipgloss.Width(name) - lipgloss.Width(ts) - 4
	if headerGap < 1 {
		headerGap = 1
	}
	header := name + strings.Repeat(" ", headerGap) + ts

	contentWidth := width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	content := lipgloss.NewStyle().
		Foreground(p.Fg).
		Width(contentWidth).
		Render(msg.Content)

	body := lipgloss.NewStyle().
		Foreground(p.Fg).
		Padding(0, 2).
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: "┃"}).
		BorderForeground(p.UserBorder).
		Width(width - 2).
		Render(content)

	return "\n" + header + "\n" + body + "\n"
}

func renderAssistantMessage(msg ChatMsg, theme Theme, width int) string {
	p := theme.Palette

	name := lipgloss.NewStyle().
		Foreground(p.Primary).
		Bold(true).
		Render("🦞 Hermes")

	ts := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render(msg.Timestamp.Format("15:04"))

	headerGap := width - lipgloss.Width(name) - lipgloss.Width(ts) - 4
	if headerGap < 1 {
		headerGap = 1
	}
	header := name + strings.Repeat(" ", headerGap) + ts

	contentWidth := width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	var content string
	if msg.Streaming {
		content = lipgloss.NewStyle().
			Foreground(p.Fg).
			Width(contentWidth).
			Render(msg.Content + " ▌") // cursor blink
	} else {
		// Render markdown
		if r := getGlamourRenderer(contentWidth); r != nil {
			content, _ = r.Render(msg.Content)
			// Glamour adds its own newlines; re-trim to fit
			lines := strings.Split(content, "\n")
			for i, l := range lines {
				if lipgloss.Width(l) > contentWidth {
					lines[i] = l[:contentWidth]
				}
			}
			content = strings.Join(lines, "\n")
		} else {
			content = lipgloss.NewStyle().
				Foreground(p.Fg).
				Width(contentWidth).
				Render(msg.Content)
		}
	}

	body := lipgloss.NewStyle().
		Foreground(p.Fg).
		Padding(0, 2).
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: "┃"}).
		BorderForeground(p.AssistBorder).
		Width(width - 2).
		Render(content)

	return "\n" + header + "\n" + body + "\n"
}

func renderSystemMessage(msg ChatMsg, theme Theme, width int) string {
	p := theme.Palette

	content := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Italic(true).
		Width(width - 4).
		Render(msg.Content)

	align := (width - lipgloss.Width(content)) / 2
	if align < 0 {
		align = 0
	}

	return "\n" + strings.Repeat(" ", align) + content + "\n"
}

func renderErrorMessage(msg ChatMsg, theme Theme, width int) string {
	p := theme.Palette

	prefix := lipgloss.NewStyle().
		Foreground(p.Error).
		Bold(true).
		Render("✗ error")

	ts := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Render(msg.Timestamp.Format("15:04"))

	headerGap := width - lipgloss.Width(prefix) - lipgloss.Width(ts) - 4
	if headerGap < 1 {
		headerGap = 1
	}
	header := prefix + strings.Repeat(" ", headerGap) + ts

	contentWidth := width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	content := lipgloss.NewStyle().
		Foreground(p.Error).
		Width(contentWidth).
		Render(msg.Content)

	body := lipgloss.NewStyle().
		Foreground(p.Error).
		Padding(0, 2).
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: "┃"}).
		BorderForeground(p.Error).
		Width(width - 2).
		Render(content)

	return "\n" + header + "\n" + body + "\n"
}

// RenderThinking renders an animated thinking indicator.
func RenderThinking(theme Theme, width int, tick int) string {
	p := theme.Palette

	dots := []string{"   ", ".  ", ".. ", "..."}
	dot := dots[tick%4]

	thinking := lipgloss.NewStyle().
		Foreground(p.FgMuted).
		Italic(true).
		Render("thinking" + dot)

	padding := width - lipgloss.Width(thinking) - 4
	if padding < 0 {
		padding = 0
	}

	return "\n" + strings.Repeat(" ", padding/2) + thinking + "\n"
}
