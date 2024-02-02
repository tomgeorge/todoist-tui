package project_view

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tomgeorge/todoist-tui/types"
)

type keyMap struct {
	ScrollUp   key.Binding
	ScrollDown key.Binding
	Help       key.Binding
	Confirm    key.Binding
	Quit       key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ScrollUp, k.ScrollUp, k.Confirm},
		{k.Help, k.Quit},
	}
}

var defaultKeys = keyMap{
	ScrollUp: key.NewBinding(
		key.WithKeys("ctrl+k", "up"),
		key.WithHelp("ctrl+k", "scroll up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("ctrl+j", "down"),
		key.WithHelp("ctrl+j", "scroll down"),
	),
	Help: key.NewBinding(
		key.WithKeys("ctrl+_"),
		key.WithKeys("ctrl+?", "help"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithKeys("ctrl+c", "quit"),
	),
}

type Model struct {
	help       help.Model
	keys       keyMap
	project    *types.Project
	tasks      []*types.Item
	titleStyle lipgloss.Style
	list       list.Model
	focused    bool
}

type ModelOption func(m *Model)

func New(project *types.Project, tasks []*types.Item, opts ...ModelOption) *Model {
	var (
		defaultTitleStyle = lipgloss.NewStyle().Bold(true).Underline(true)
	)
	items := []list.Item{}
	for _, task := range tasks {
		items = append(items, task)
	}
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).
		Padding(0, 0, 0, 2)
	list := list.New(items, delegate, 50, 50)
	list.Title = project.Name
	m := &Model{
		help:       help.New(),
		keys:       defaultKeys,
		tasks:      tasks,
		project:    project,
		titleStyle: defaultTitleStyle,
		list:       list,
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

type UpdateTaskMsg struct {
	Task types.Item
}

func UpdateTask(task *types.Item) tea.Cmd {
	return func() tea.Msg {
		return UpdateTaskMsg{*task}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Confirm):
			selected, ok := m.list.SelectedItem().(*types.Item)
			if !ok {
				return m, nil
			}
			return m, UpdateTask(selected)
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	sections := []string{}
	// title := lipgloss.JoinVertical(lipgloss.Center, fmt.Sprintf("%s\n", m.project.Style().Copy().Bold(true).Render(m.project.Name)))
	// sections = append(sections, title)
	// for _, task := range m.tasks {
	// 	sections = append(sections, task.Content)
	// }
	sections = append(sections, m.list.View())
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) SetFocused(focused bool) {
	m.focused = focused
}

func (m Model) Project() *types.Project {
	return m.project
}
