package model

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
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
	successStyle lipgloss.Style
	showSpinner  bool
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
		defaultLoading     = true
		defaultWidth       = 50
		defaultHeight      = 50
		defaultShowSpinner = false
	)
	return &Model{
		loading: defaultLoading,
		//FIXME: These need to go in the theme
		errorStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796")),
		successStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#a6d189")),
		spinner:      spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#939ab7")))),
		showSpinner:  defaultShowSpinner,
		state:        nil,
		ctx:          ctx,
		events:       *events.New(ctx),
		project:      nil,
		projects:     nil,
		tasks:        nil,
		labels:       nil,
		projectView:  project_view.Model{},
		taskCreate:   task_create.Model{},
		width:        defaultWidth,
		height:       defaultHeight,
	}
}

type SaveAndQuitMsg struct {
	Error error
}

func (m *Model) save() tea.Cmd {
	return func() tea.Msg {
		m.ctx.Logger.Info("Hey I'm here")
		state, err := os.Create(filepath.Join(m.ctx.Config.StateDir, "state.json"))
		if err != nil {
			m.ctx.Logger.Debug("error creating file", err)
			return SaveAndQuitMsg{Error: err}
		}
		defer state.Close()
		json, err := json.MarshalIndent(m.state, "", "  ")
		if err != nil {
			m.ctx.Logger.Debug("error marshalling json", err)
			return SaveAndQuitMsg{Error: err}
		}
		_, err = state.Write(json)
		if err != nil {
			m.ctx.Logger.Debug("writing file", err)
			return SaveAndQuitMsg{Error: err}
		}
		return SaveAndQuitMsg{Error: nil}
	}
}

type StartSpinnerMessage struct{}

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
	case messages.OperationPendingMessage:
		m.showSpinner = true
		cmds = append(cmds, m.spinner.Tick)
	case messages.StateMessage:
		if msg.Error == nil {
			m.state = msg.State
		}
	case messages.OperationResponse:
		m.showSpinner = false
		if msg.Error != nil {
			return m, m.events.Publish(msg.Error.Error(), m.errorStyle, true, 3*time.Second)
		} else {
			m.state = sync.Merge(m.state, msg.State)
			types := lo.Map(msg.Commands, func(c sync.Command, _ int) string { return c.Type })
			return m, m.events.Publish(
				fmt.Sprintf("Command(s) succeeded: %s", types),
				m.successStyle,
				true,
				3*time.Second)
		}
	case tea.KeyMsg:
		m.ctx.Logger.Debug("KeyMsg")
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Sequence(m.save(), tea.Quit)
		}
	}
	m.ctx.Logger.Infof("Updating %s", m.stack[len(m.stack)-1].Id)
	lastModel := len(m.stack) - 1
	m.stack[lastModel].Model, cmd = m.stack[lastModel].Model.Update(msg)
	cmds = append(cmds, cmd)
	m.events, cmd = m.events.Update(msg)
	cmds = append(cmds, cmd)
	m.spinner, cmd = m.spinner.Update(msg)
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
	if m.showSpinner {
		sections = append(sections, m.spinner.View())
	}
	sections = append(sections, m.events.View())
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
