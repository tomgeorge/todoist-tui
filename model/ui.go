package model

import (
	"context"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/model/events"
	"github.com/tomgeorge/todoist-tui/model/project_view"
	"github.com/tomgeorge/todoist-tui/model/task_create"
	"github.com/tomgeorge/todoist-tui/services/sync"
	"github.com/tomgeorge/todoist-tui/types"
)

type ViewIndex int

const (
	loading     ViewIndex = iota
	projectView ViewIndex = iota
	taskCreate  ViewIndex = iota
)

type Model struct {
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
		loading:     defaultLoading,
		errorStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796")),
		spinner:     spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#939ab7")))),
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
	m.ctx.Logger.Info("Hey")
	return tea.Batch(m.performSync(), startSpinner())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case StartSpinnerMessage:
		m.loading = true
		cmds = append(cmds, m.spinner.Tick)
		return m, tea.Batch(cmds...)
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	case StateMessage:
		if msg.err != nil {
			err := fmt.Sprintf("failed to get data from todoist: %s", msg.err.Error())
			m.loadingError = err
			return m, tea.Sequence(tea.Quit)
		}
		m.loading = false
		m.state = msg.state
		m.labels = msg.state.Labels
		m.projects = msg.state.Projects
		tasks := lo.Filter(m.state.Items, func(i *types.Item, _ int) bool {
			return i.ProjectId == m.state.Projects[0].Id
		})
		m.projectView = *project_view.New(m.state.Projects[0], tasks)
		m.index = projectView
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, cmd
	case project_view.UpdateTaskMsg:
		m.projectView.SetFocused(false)
		labels := lo.Filter(m.labels, func(l *types.Label, _ int) bool {
			return slices.Contains(msg.Task.Labels, l.Name)
		})
		newModel := task_create.New(
			m.ctx,
			task_create.WithTask(&msg.Task),
			task_create.WithParentProject(m.projectView.Project()),
			task_create.WithProjects(m.projects),
			task_create.WithLabels(labels),
			task_create.WithPossibleLabels(m.labels),
		)
		m.taskCreate = *newModel
		m.index = taskCreate
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	m.ctx.Logger.Info("index ", m.index)
	switch m.index {
	case taskCreate:
		m.taskCreate, cmd = m.taskCreate.Update(msg)
		cmds = append(cmds, cmd)
	case projectView:
		m.projectView, cmd = m.projectView.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.events, cmd = m.events.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var sections []string
	if m.loading {
		sections = append(sections, fmt.Sprintf("%s fetching data from todoist...%s",
			m.spinner.View(),
			m.errorStyle.Render(m.loadingError)))
	}
	switch m.index {
	case taskCreate:
		sections = append(sections, m.taskCreate.View())
	case projectView:
		sections = append(sections, m.projectView.View())
	}
	sections = append(sections, m.events.View())
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
