package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// InputModel wraps a textarea for user message input.
type InputModel struct {
	textarea textarea.Model
	theme    Theme
}

// NewInputModel creates a new input field.
func NewInputModel(theme Theme) InputModel {
	ta := textarea.New()
	ta.Placeholder = "Type a message... (Enter to send, Ctrl+C to quit)"
	ta.Prompt = "❯ "
	ta.CharLimit = 8192
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.Focus()

	return InputModel{
		textarea: ta,
		theme:    theme,
	}
}

// SetTheme updates the input's theme.
func (m *InputModel) SetTheme(theme Theme) {
	m.theme = theme
}

// SetWidth sets the content width.
func (m *InputModel) SetWidth(width int) {
	m.textarea.SetWidth(width)
}

// Focus gives the input focus.
func (m *InputModel) Focus() {
	m.textarea.Focus()
}

// Value returns the current input text.
func (m InputModel) Value() string {
	return m.textarea.Value()
}

// Reset clears the input.
func (m *InputModel) Reset() {
	m.textarea.Reset()
}

// InsertString inserts text at the cursor.
func (m *InputModel) InsertString(s string) {
	m.textarea.InsertString(s)
}

// InsertNewline inserts a newline at the cursor.
func (m *InputModel) InsertNewline() {
	m.textarea.InsertString("\n")
}

// Update handles tea.Msg and returns the updated model and any command.
func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	ta, cmd := m.textarea.Update(msg)
	m.textarea = ta
	return m, cmd
}

// View renders the input field.
func (m InputModel) View() string {
	p := m.theme.Palette
	ta := m.textarea
	// Style the textarea using the styling methods available in this version
	_ = p // kept for future styled version
	_ = ta
	return m.textarea.View()
}

