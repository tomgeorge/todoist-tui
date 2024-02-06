package loading

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/messages"
	"github.com/tomgeorge/todoist-tui/model/project_view"
	"github.com/tomgeorge/todoist-tui/types"
)

type Model struct {
	ctx          ctx.Context
	loading      bool
	spinner      spinner.Model
	errorMessage string
	errorStyle   lipgloss.Style
}

type ModelOption func(m *Model)

func New(ctx ctx.Context, opts ...ModelOption) *Model {
	var (
		defaultLoading = true
		defaultSpinner = spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#939ab7"))),
		)
		defaultErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796"))
	)
	m := &Model{
		ctx:          ctx,
		loading:      defaultLoading,
		spinner:      defaultSpinner,
		errorMessage: "",
		errorStyle:   defaultErrorStyle,
	}

	for _, opt := range opts {
		opt(m)
	}
	return m
}

func WithSpinnerModel(spinnerModel spinner.Model) ModelOption {
	return func(m *Model) {
		m.spinner = spinnerModel
	}
}

func WithErrorStyle(errorStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.errorStyle = errorStyle
	}
}

func (m *Model) performSync() tea.Cmd {
	return func() tea.Msg {
		state, err := m.ctx.Client.FullSync(context.Background())
		return messages.StateMessage{State: state, Err: err}
	}
}

func load() tea.Cmd {
	return func() tea.Msg { return messages.LoadingMessage{} }
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.performSync(), load())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.ctx.Logger.Debug("loading model update")
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.ctx.Logger.Debug("got a tick msg")
	case messages.LoadingMessage:
		m.ctx.Logger.Debug("LoadingMessage")
		return m, m.spinner.Tick
	case messages.StateMessage:
		m.ctx.Logger.Debug("StateMessage")
		if msg.Err != nil {
			m.ctx.Logger.Debug("Got an error %s", msg.Err.Error())
			m.errorMessage = fmt.Sprintf("Error getting data from todoist: %s", msg.Err.Error())
			m.loading = false
			return m, tea.Sequence(tea.Quit)
		}
		state := msg.State
		tasks := lo.Filter(state.Items, func(i *types.Item, _ int) bool {
			return i.ProjectId == state.Projects[0].Id
		})
		projects := project_view.New(m.ctx, msg.State.Projects[0], tasks, msg.State.Labels, msg.State.Projects)
		return m, messages.Push("project_view", projects)
	}
	return m, cmd
}

func (m Model) View() string {
	m.ctx.Logger.Debugf("loading.view(), error message is %s", m.errorMessage)
	if m.loading {
		return fmt.Sprintf("%s fetching data from todoist...",
			m.spinner.View())
	} else {
		return lipgloss.JoinVertical(lipgloss.Left, m.errorStyle.Render(m.errorMessage, "\n"))
	}
}
