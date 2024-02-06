package button

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Confirm key.Binding
}

var defaultKeyMap = keyMap{
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "Confirm"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Confirm},
		{},
	}
}

func (m Model) Help() help.Model {
	return m.help
}

func (m *Model) SetHelp(showHelp bool) {
	m.help.ShowAll = showHelp
}

func (m Model) HelpKeys() help.KeyMap {
	return m.keys
}

type SubmitFn func(payload interface{}) tea.Cmd

type SubmitMsg struct {
	payload interface{}
}

type Model struct {
	text          string
	enabled       bool
	blurredStyle  lipgloss.Style
	focusedStyle  lipgloss.Style
	disabledStyle lipgloss.Style
	onSubmit      SubmitFn
	help          help.Model
	keys          keyMap
	focused       bool
}

type ModelOption func(*Model)

func New(opts ...ModelOption) *Model {
	defaultStyle := lipgloss.NewStyle().Underline(true)
	defaultFocusedStyle := defaultStyle.Copy().Bold(true).Italic(true)
	defaultDisabledStyle := defaultStyle.Copy().Strikethrough(true)
	const (
		defaultText    = "submit"
		defaultEnabled = true
	)

	defaultOnSubmit := func(payload interface{}) tea.Cmd {
		return func() tea.Msg {
			return SubmitMsg{}
		}
	}

	model := &Model{
		text:          defaultText,
		enabled:       defaultEnabled,
		blurredStyle:  defaultStyle,
		focusedStyle:  defaultFocusedStyle,
		disabledStyle: defaultDisabledStyle,
		onSubmit:      defaultOnSubmit,
		help:          help.New(),
		keys:          defaultKeyMap,
	}

	for _, opt := range opts {
		opt(model)
	}
	return model
}

func WithText(text string) ModelOption {
	return func(m *Model) {
		m.text = text
	}
}

func WithEnabled(enabled bool) ModelOption {
	return func(m *Model) {
		m.enabled = enabled
	}
}

func WithBlurredStyle(blurredStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.blurredStyle = blurredStyle
	}
}

func WithFocusedStyle(focusedStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.focusedStyle = focusedStyle
	}
}

func WithOnSubmit(onSubmit SubmitFn) ModelOption {
	return func(m *Model) {
		m.onSubmit = onSubmit
	}
}

func WithDisabledStyle(disabledStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.disabledStyle = disabledStyle
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Confirm):
			return m, m.onSubmit("okey dokey")
		}
	}
	return m, nil
}

func (m *Model) SetFocus(focused bool) {
	m.focused = focused
}

func (m Model) View() string {
	switch {
	case m.focused:
		return lipgloss.NewStyle().PaddingLeft(2).Render(m.focusedStyle.Render(m.text))
	default:
		return lipgloss.NewStyle().PaddingLeft(2).Render(m.blurredStyle.Render(m.text))
	}
}
