package project_view

import (
	"slices"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/messages"
	"github.com/tomgeorge/todoist-tui/model/button"
	"github.com/tomgeorge/todoist-tui/model/task_create"
	"github.com/tomgeorge/todoist-tui/services/sync"
	"github.com/tomgeorge/todoist-tui/types"
)

const (
	HeaderHeight = 3
)

func (m Model) Resize(width, height int) tea.Model {
	m.ctx.Logger.Debugf("Project View Component - Resize Requested: Width %d Height %d", width, height)
	m.list.SetHeight(height - HeaderHeight)
	m.ctx.Logger.Debugf("Project View Component - Setting height to %d", height-HeaderHeight)
	return m
}

func (m Model) ComponentName() string {
	return "project_view"
}

func (m Model) Model() tea.Model {
	return m
}

type keyMap struct {
	ScrollUp        key.Binding
	ScrollDown      key.Binding
	NextProject     key.Binding
	PreviousProject key.Binding
	Help            key.Binding
	Confirm         key.Binding
	Choose          key.Binding
	Quit            key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ScrollUp, k.ScrollDown, k.Confirm, k.Choose},
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
	NextProject: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "next project"),
	),
	PreviousProject: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "previous project"),
	),
	Help: key.NewBinding(
		key.WithKeys("ctrl+_"),
		key.WithKeys("ctrl+?", "help"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Choose: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "Quick choose project"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithKeys("ctrl+c", "quit"),
	),
}

type Model struct {
	list        list.Model
	ctx         ctx.Context
	project     *types.Project
	create      *button.Model
	help        help.Model
	titleStyle  lipgloss.Style
	keys        keyMap
	tasks       []*types.Item
	allLabels   []*types.Label
	allProjects []*types.Project
	focused     bool
}

type ModelOption func(m *Model)

func NewListDelegate(ctx ctx.Context) list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = ctx.Theme.Focused.Title.Copy().PaddingLeft(1).BorderStyle(lipgloss.ThickBorder()).BorderLeft(true)
	delegate.Styles.SelectedDesc = ctx.Theme.Focused.Base
	delegate.Styles.NormalTitle = ctx.Theme.Blurred.Base
	delegate.Styles.NormalDesc = ctx.Theme.Blurred.Base
	return delegate
}

func NewList(ctx ctx.Context, items []list.Item, title string, width int, height int, delegate list.DefaultDelegate) list.Model {
	list := list.New(items, delegate, width, height-HeaderHeight)
	list.Title = title
	list.Styles.Title = ctx.Theme.Focused.Title
	list.SetShowHelp(false)
	list.SetShowTitle(false)
	return list
}

func NewFromState(ctx ctx.Context, state *sync.SyncResponse) *Model {
	tasks := lo.Filter(state.Items, func(i *types.Item, _ int) bool {
		return i.ProjectId == state.Projects[0].Id
	})
	return New(ctx, state.Projects[0], tasks, state.Labels, state.Projects)
}

func New(ctx ctx.Context, project *types.Project, tasks []*types.Item, labels []*types.Label, projects []*types.Project, opts ...ModelOption) *Model {
	var (
		defaultTitleStyle = lipgloss.NewStyle().Bold(true).Underline(true)
	)
	items := []list.Item{}
	for _, task := range tasks {
		items = append(items, task)
	}
	delegate := NewListDelegate(ctx)
	list := NewList(ctx, items, project.Name, ctx.MainContentWidth, ctx.MainContentHeight, delegate)
	m := &Model{
		ctx:         ctx,
		help:        help.New(),
		keys:        defaultKeys,
		tasks:       tasks,
		project:     project,
		allLabels:   labels,
		allProjects: projects,
		titleStyle:  defaultTitleStyle,
		list:        list,
		create: button.New(
			button.WithText("Create Task"),
			button.WithEnabled(true),
			button.WithFocusedStyle(ctx.Theme.Focused.FocusedButton.MarginTop(1)),
			button.WithBlurredStyle(ctx.Theme.Focused.BlurredButton.MarginTop(1)),
		),
	}
	return m
}

func WithListHeight(height int) ModelOption {
	return func(m *Model) {
		m.list.SetHeight(height)
	}
}

func WithListWidth(width int) ModelOption {
	return func(m *Model) {
		m.list.SetHeight(width)
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.ScrollUp):
			if m.list.Index() == -1 {
				return m, nil
			}
			m.list.Select(-1)
			m.create.SetFocus(true)
		case key.Matches(msg, m.keys.ScrollDown):
			if m.list.Index() == -1 {
				m.create.SetFocus(false)
				m.list.Select(0)
			}
		case msg.String() == "n":
			m.list.Select(-1)
			m.create.SetFocus(true)
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Confirm):
			if m.list.Index() == -1 {
				taskModel := task_create.New(
					m.ctx,
					task_create.WithParentProject(m.project),
					task_create.WithPossibleLabels(m.allLabels),
					task_create.WithProjects(m.allProjects),
				)
				return m, messages.Push("task_create", taskModel)
			}
			selected, ok := m.list.SelectedItem().(*types.Item)
			if !ok {
				return m, nil
			}
			labels := lo.Filter(m.allLabels, func(l *types.Label, _ int) bool {
				return slices.Contains(selected.Labels, l.Name)
			})
			taskModel := task_create.New(
				m.ctx,
				task_create.WithTask(selected),
				task_create.WithParentProject(m.project),
				task_create.WithLabels(labels),
				task_create.WithPossibleLabels(m.allLabels),
				task_create.WithProjects(m.allProjects),
			)
			return m, messages.Push("task_create", taskModel)
		case key.Matches(msg, m.keys.NextProject):
			return m, changeProject(m.project, messages.NextProject)
		case key.Matches(msg, m.keys.PreviousProject):
			return m, changeProject(m.project, messages.PreviousProject)
		}
	case tea.WindowSizeMsg:
		m.Resize(msg.Width, msg.Height)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func changeProject(project *types.Project, direction messages.ChangeProjectDirection) tea.Cmd {
	return func() tea.Msg {
		return messages.ChangeProjectMessage{
			From:      project,
			Direction: direction,
		}
	}
}

func (m Model) View() string {
	sections := []string{}
	sections = append(sections, m.create.View())
	sections = append(sections, m.list.View())
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

func (m Model) Project() *types.Project {
	return m.project
}
