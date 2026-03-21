package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// ChatModel manages the scrollable chat viewport.
type ChatModel struct {
	viewport    viewport.Model
	messages    []ChatMsg
	theme       Theme
	width       int
	height      int
	streamBuf   string
	streamTick  int
}

// NewChatModel creates a new chat viewport.
func NewChatModel(theme Theme) ChatModel {
	vp := viewport.New(80, 20)
	vp.SetContent("")

	return ChatModel{
		viewport: vp,
		messages: []ChatMsg{},
		theme:    theme,
		width:    80,
		height:   20,
	}
}

// SetTheme updates the chat's theme.
func (c *ChatModel) SetTheme(theme Theme) {
	c.theme = theme
}

// SetSize sets the viewport dimensions.
func (c *ChatModel) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.viewport.Width = width
	c.viewport.Height = height
	c.renderContent()
}

// SetWidth sets the content width.
func (c *ChatModel) SetWidth(width int) {
	c.width = width
}

// Height returns the viewport height.
func (c ChatModel) Height() int {
	return c.height
}

// AddMessage appends a new message to the chat.
func (c *ChatModel) AddMessage(msg ChatMsg) {
	c.messages = append(c.messages, msg)
	c.renderContent()
}

// UpdateLastAssistant updates the content of the most recent assistant message.
func (c *ChatModel) UpdateLastAssistant(content string, streaming bool) {
	if len(c.messages) == 0 {
		c.AddMessage(ChatMsg{Role: RoleAssistant, Content: content, Streaming: streaming})
		return
	}
	last := &c.messages[len(c.messages)-1]
	last.Content = content
	last.Streaming = streaming
	c.renderContent()
}

// ScrollUp scrolls the viewport up.
func (c *ChatModel) ScrollUp(n int) {
	c.viewport.ScrollUp(n)
}

// ScrollDown scrolls the viewport down.
func (c *ChatModel) ScrollDown(n int) {
	c.viewport.ScrollDown(n)
}

// ScrollToTop scrolls to the top.
func (c *ChatModel) ScrollToTop() {
	c.viewport.GotoTop()
}

// ScrollToBottom scrolls to the bottom.
func (c *ChatModel) ScrollToBottom() {
	c.viewport.GotoBottom()
}

// Clear removes all messages.
func (c *ChatModel) Clear() {
	c.messages = nil
	c.streamBuf = ""
	c.renderContent()
}

// GetMessages returns all messages.
func (c ChatModel) GetMessages() []ChatMsg {
	return c.messages
}

// Tick advances the thinking animation.
func (c *ChatModel) Tick() {
	c.streamTick++
	if c.streamTick > 3 {
		c.streamTick = 0
	}
}

func (c *ChatModel) renderContent() {
	var sb strings.Builder
	for _, msg := range c.messages {
		sb.WriteString(RenderMessage(msg, c.theme, c.width))
		sb.WriteString("\n")
	}
	c.viewport.SetContent(sb.String())
	// Auto-scroll to bottom on new content
	c.viewport.GotoBottom()
}

// View renders the chat viewport.
func (c ChatModel) View() string {
	p := c.theme.Palette
	content := c.viewport.View()
	return lipgloss.NewStyle().
		Background(p.Bg).
		Foreground(p.Fg).
		Render(content)
}
