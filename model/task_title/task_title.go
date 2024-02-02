package task_title

import (
	"log"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	textinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	focused           bool
	editing           bool
	label             string
	spacingStyle      lipgloss.Style
	labelStyle        lipgloss.Style
	focusedLabelStyle lipgloss.Style
	textinput         textinput.Model
	promptStyle       lipgloss.Style
	textStyle         lipgloss.Style
	help              help.Model
	keys              keyMap
}

type keyMap struct {
	Edit        key.Binding
	StopEditing key.Binding
	Help        key.Binding
}

var defaultKeyMap = keyMap{
	Edit: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "edit"),
	),
	StopEditing: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "stop editing"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Edit, k.StopEditing},
		{k.Help},
	}
}

type ModelOption func(*Model)

// Create a new title model
// Accepts a number of options to set labels,
// initial content, styles, etc
func New(opts ...ModelOption) *Model {
	const (
		defaultLabel   = "Title"
		defaultFocused = false
		defaultEditing = false
	)

	defaultLabelStyle := lipgloss.NewStyle().
		Underline(true).
		MarginBottom(1)

	defaultFocusedLabelStyle := defaultLabelStyle.
		Copy().
		Italic(true).
		Bold(true)

	defaultSpacingStyle := lipgloss.NewStyle().
		MarginBottom(1)

	textinput := textinput.New()

	model := &Model{
		focused:           defaultFocused,
		editing:           defaultEditing,
		label:             defaultLabel,
		labelStyle:        defaultLabelStyle,
		focusedLabelStyle: defaultFocusedLabelStyle,
		spacingStyle:      defaultSpacingStyle,
		textinput:         textinput,
		textStyle:         textinput.TextStyle,
		promptStyle:       textinput.PromptStyle,
		help:              help.New(),
		keys:              defaultKeyMap,
	}

	for _, opt := range opts {
		opt(model)
	}
	return model
}

func WithLabel(label string) ModelOption {
	return func(m *Model) {
		m.label = label
	}
}

func WithLabelStyle(labelStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.labelStyle = labelStyle
	}
}

func WithFocusedLabelStyle(labelStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.focusedLabelStyle = labelStyle
	}
}

func WithSpacingStyle(labelStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.spacingStyle = labelStyle
	}
}

func WithContent(content string) ModelOption {
	return func(m *Model) {
		m.textinput.SetValue(content)
	}
}

func WithTextStyle(style lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.textinput.TextStyle = style
	}
}

func WithPromptStyle(style lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.textinput.PromptStyle = style
	}
}

func (m *Model) Focused() bool {
	return m.focused
}

func (m *Model) Editing() bool {
	return m.focused && m.textinput.Focused()
}

func (m *Model) SetEditing(editing bool) {
	m.editing = editing
	m.updateTextInputState()
}

func (m *Model) SetHelp(help bool) {
	m.help.ShowAll = help
}

func (m *Model) updateTextInputState() {
	if m.editing {
		m.textinput.Focus()
	} else {
		m.textinput.Blur()
	}
}

func (m *Model) FocusOff() {
	m.textinput.Blur()
	m.focused = false
}

func (m *Model) FocusOn() {
	m.focused = true
}

func (m Model) Help() help.Model {
	return m.help
}

func (m Model) HelpKeys() help.KeyMap {
	return m.keys
}

func (m *Model) GetContent() string {
	return m.textinput.Value()
}

func (m *Model) SetContent(content string) {
	m.textinput.SetValue(content)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help) && !m.Editing():
			log.Printf("Setting showall title")
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keys.Edit) && !m.Editing():
			m.SetEditing(true)
			return m, nil
		case key.Matches(msg, m.keys.StopEditing):
			m.SetEditing(false)
			return m, nil
		}

	}
	if m.Editing() {
		m.textinput, cmd = m.textinput.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	sections := []string{}
	switch {
	case m.Editing():
		sections = append(sections, m.focusedLabelStyle.Render(m.label))
		sections = append(sections, m.textinput.View())
	case !m.Editing():
		sections = append(sections, m.focusedLabelStyle.Render(m.label))
		sections = append(sections, m.textinput.Value())
	default:
		sections = append(sections, m.labelStyle.Render(m.label))
		sections = append(sections, m.textinput.Value())
	}
	return m.spacingStyle.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}
