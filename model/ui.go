package model

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/messages"
	"github.com/tomgeorge/todoist-tui/model/events"
	"github.com/tomgeorge/todoist-tui/model/loading"
	"github.com/tomgeorge/todoist-tui/model/project_view"
	"github.com/tomgeorge/todoist-tui/model/task_create"
	"github.com/tomgeorge/todoist-tui/services/sync"
	"github.com/tomgeorge/todoist-tui/types"
)

type ViewIndex int

const (
	// loading     ViewIndex = iota
	projectView ViewIndex = iota
	taskCreate  ViewIndex = iota
)

type StackItem struct {
	Id    string
	Model tea.Model
}

type Model struct {
	stack        []StackItem
	loading      bool
	loadingError string
	errorStyle   lipgloss.Style
	spinner      spinner.Model
	state        *sync.SyncResponse
	ctx          ctx.Context
	events       events.Model
	project      *types.Project
	tasks        []*types.Item
	labels       []*types.Label
	projects     []*types.Project
	projectView  project_view.Model
	taskCreate   task_create.Model
	index        ViewIndex
	width        int
	height       int
}

func New(ctx ctx.Context) *Model {
	const (
		defaultLoading = true
		defaultWidth   = 50
		defaultHeight  = 50
	)
	return &Model{
		loading:    defaultLoading,
		errorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796")),
		spinner:    spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#939ab7")))),

		state:       nil,
		ctx:         ctx,
		events:      *events.New(ctx),
		project:     nil,
		projects:    nil,
		tasks:       nil,
		labels:      nil,
		projectView: project_view.Model{},
		taskCreate:  task_create.Model{},
		width:       defaultWidth,
		height:      defaultHeight,
	}
}

type StateMessage struct {
	state *sync.SyncResponse
	err   error
}

func (m *Model) performSync() tea.Cmd {
	return func() tea.Msg {
		state, err := m.ctx.Client.FullSync(context.Background())
		return StateMessage{state, err}
	}
}

type StartSpinnerMessage struct{}

func startSpinner() tea.Cmd {
	return func() tea.Msg { return StartSpinnerMessage{} }
}

func (m Model) Init() tea.Cmd {
	return messages.Push("loading", loading.New(m.ctx))
}

func push(m Model, msg messages.PushMessage) (Model, tea.Cmd) {
	m.stack = append(m.stack, StackItem{Id: msg.Id, Model: msg.Model})
	return m, m.stack[len(m.stack)-1].Model.Init()
}

func pop(m Model, msg messages.PopMessage) (Model, tea.Cmd) {
	m.stack = m.stack[:len(m.stack)-1]
	return m, nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case messages.PushMessage:
		m.ctx.Logger.Debug("PushMessage")
		return push(m, msg)
	case messages.PopMessage:
		return pop(m, msg)
	case tea.WindowSizeMsg:
		m.ctx.Logger.Debug("WindowSizeMsg")
		m.width = msg.Width
		m.height = msg.Height
		return m, cmd
	case tea.KeyMsg:
		m.ctx.Logger.Debug("KeyMsg")
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	m.ctx.Logger.Infof("Updating %s", m.stack[len(m.stack)-1].Id)
	lastModel := len(m.stack) - 1
	m.stack[lastModel].Model, cmd = m.stack[lastModel].Model.Update(msg)
	cmds = append(cmds, cmd)
	m.events, cmd = m.events.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if len(m.stack) == 0 {
		return ""
	}
	var sections []string
	sections = append(sections, m.stack[len(m.stack)-1].Model.View())
	// switch m.index {
	// case taskCreate:
	// 	sections = append(sections, m.taskCreate.View())
	// case projectView:
	// 	sections = append(sections, m.projectView.View())
	// }
	// sections = append(sections, m.events.View())
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
