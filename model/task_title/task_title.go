package task_title

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	textinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tomgeorge/todoist-tui/ctx"
)

type Model struct {
	ctx               ctx.Context
	focused           bool
	editing           bool
	label             string
	labelStyle        lipgloss.Style
	focusedStyle      lipgloss.Style
	blurredStyle      lipgloss.Style
	focusedLabelStyle lipgloss.Style
	textinput         textinput.Model
	promptStyle       lipgloss.Style
	textStyle         lipgloss.Style
	help              help.Model
	debug             bool
	keys              keyMap
}

type keyMap struct {
	Edit        key.Binding
	StopEditing key.Binding
	Help        key.Binding
	Debug       key.Binding
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
		key.WithKeys("ctrl+_"),
		key.WithHelp("ctrl+?", "help"),
	),
	Debug: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "Show debug information of focused component"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Edit, k.StopEditing},
		{k.Help, k.Debug},
	}
}

type ModelOption func(*Model)

// Create a new title model
// Accepts a number of options to set labels,
// initial content, styles, etc
func New(ctx ctx.Context, opts ...ModelOption) *Model {
	const (
		defaultLabel   = "Title"
		defaultFocused = false
		defaultEditing = false
		defaultDebug   = false
	)

	defaultLabelStyle := lipgloss.NewStyle().
		Underline(true).
		MarginBottom(1)

	defaultFocusedLabelStyle := defaultLabelStyle.
		Copy().
		Italic(true).
		Bold(true)

	defaultFocusedStyle := lipgloss.NewStyle().
		MarginBottom(1)
	defaultBlurredStyle := defaultFocusedStyle.Copy()

	textinput := textinput.New()

	model := &Model{
		ctx:               ctx,
		focused:           defaultFocused,
		editing:           defaultEditing,
		label:             defaultLabel,
		labelStyle:        defaultLabelStyle,
		focusedLabelStyle: defaultFocusedLabelStyle,
		textinput:         textinput,
		textStyle:         textinput.TextStyle,
		promptStyle:       textinput.PromptStyle,
		blurredStyle:      defaultBlurredStyle,
		focusedStyle:      defaultFocusedStyle,
		help:              help.New(),
		keys:              defaultKeyMap,
		debug:             false,
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

func WithFocusedStyle(focusedStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.focusedStyle = focusedStyle
	}
}

func WithBlurredStyle(blurredStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.blurredStyle = blurredStyle
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

func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

func (m Model) Editing() bool {
	return m.focused && m.textinput.Focused()
}

func (m *Model) SetEditing(editing bool) {
	m.editing = editing
	m.updateTextInputState()
}

func (m *Model) Debug() bool {
	return m.debug
}

func (m *Model) SetDebug(debug bool) {
	m.debug = debug
}

func (m *Model) SetHelp(help bool) {
	m.help.ShowAll = help
}

func (m *Model) updateTextInputState() {
	if m.editing {
		m.textinput.Focus()
		m.textinput.CursorEnd()
	} else {
		m.textinput.Blur()
	}
}

func (m *Model) FocusOff() {
	m.textinput.Blur()
	m.focused = false
	m.editing = false
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
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keys.Edit) && !m.Editing():
			m.SetEditing(true)
			return m, nil
		case key.Matches(msg, m.keys.StopEditing):
			m.SetEditing(false)
			return m, nil
		case key.Matches(msg, m.keys.Debug):
			m.debug = !m.debug
			return m, nil
		}

	}
	if m.Editing() {
		m.textinput, cmd = m.textinput.Update(msg)
	}
	return m, cmd
}

func (m *Model) DebugInfo() string {
	return m.ctx.Theme.Help.FullDesc.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			fmt.Sprintf("focused %t\n", m.Focused()),
			fmt.Sprintf("editing %t\n", m.Editing()),
			fmt.Sprintf("textinput value [%s]\n", m.textinput.Value()),
		),
	)
}

func (m Model) View() string {
	sections := []string{}
	switch {
	case m.Focused() && m.Editing():
		sections = append(sections, m.focusedLabelStyle.Render(m.label))
		sections = append(sections, m.textinput.View())
	case m.Focused() && !m.Editing():
		sections = append(sections, m.focusedLabelStyle.Render(m.label))
		sections = append(sections, m.textinput.Value())
	default:
		sections = append(sections, m.labelStyle.Render(m.label))
		sections = append(sections, m.textinput.Value())
	}
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	style := m.blurredStyle
	if m.Focused() {
		style = m.focusedStyle
	}
	if m.Debug() {
		return lipgloss.JoinHorizontal(0, style.Render(content), m.DebugInfo())
	}
	return style.Render(content)
}
