package task_description

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
)

type Model struct {
	logger            *zap.SugaredLogger
	focused           bool
	label             string
	labelStyle        lipgloss.Style
	focusedLabelStyle lipgloss.Style
	focusedStyle      lipgloss.Style
	blurredStyle      lipgloss.Style
	input             textarea.Model
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

func NewModel(logger *zap.SugaredLogger, opts ...ModelOption) *Model {
	const (
		defaultLabel = "Description"
	)
	defaultLabelStyle := lipgloss.NewStyle().
		Underline(true).
		MarginBottom(1)

	defaultFocusedLabelStyle := defaultLabelStyle.
		Copy().
		Bold(true).
		Italic(true)

	defaultFocusedStyle := lipgloss.NewStyle().PaddingLeft(1)
	defaultBlurredStyle := defaultFocusedStyle.Copy()

	input := textarea.New()
	input.ShowLineNumbers = false
	input.Placeholder = "No description..."
	input.Prompt = ""
	model := &Model{
		logger:            logger,
		focused:           false,
		label:             defaultLabel,
		labelStyle:        defaultLabelStyle,
		focusedLabelStyle: defaultFocusedLabelStyle,
		focusedStyle:      defaultFocusedStyle,
		blurredStyle:      defaultBlurredStyle,
		input:             input,
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

func WithValue(value string) ModelOption {
	return func(m *Model) {
		m.input.SetValue(value)
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

func WithFocusedStyle(style lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.focusedStyle = style
	}
}

func WithBlurredStyle(style lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.blurredStyle = style
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	m.logger.Info("In description update")
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.logger.Info("Setting width")
		m.input.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.logger.Info("Setting showall title")
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keys.Edit) && !m.input.Focused():
			m.input.Focus()
			return m, nil
		case key.Matches(msg, m.keys.StopEditing):
			if m.input.Focused() {
				m.input.Blur()
			}
		}
	}
	if m.input.Focused() {
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

// FIXME: unsightly!
func (m Model) View() string {
	var content string
	switch {
	case m.focused && m.input.Focused():
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			m.focusedLabelStyle.Render(m.label),
			m.input.View(),
		)
	case m.focused && !m.input.Focused():
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			m.focusedLabelStyle.Render(m.label),
			m.input.View(),
		)
	default:
		// This would be super cool, maybe one day
		// output, _ := glamour.Render(m.input.Value(), "dark")
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			m.labelStyle.Render(m.label),
			m.input.View(),
		)
	}
	if m.focused {
		return m.focusedStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				content,
			),
		)
	}
	return m.blurredStyle.Render(lipgloss.JoinVertical(lipgloss.Left, content))
}

func (m *Model) Focused() bool {
	return m.focused
}

func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

func (m *Model) FocusOn() {
	m.focused = true
}

func (m *Model) FocusOff() {
	m.focused = false
	m.input.Blur()
}

func (m *Model) SetHelp(help bool) {
	m.help.ShowAll = help
}

func (m *Model) GetContent() string {
	return m.input.Value()
}

func (m *Model) SetContent(content string) {
	m.input.SetValue(content)
}

func (m Model) Help() help.Model {
	return m.help
}

func (m Model) HelpKeys() help.KeyMap {
	return m.keys
}
