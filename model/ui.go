package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/messages"
	"github.com/tomgeorge/todoist-tui/model/drawer"
	"github.com/tomgeorge/todoist-tui/model/events"
	"github.com/tomgeorge/todoist-tui/model/footer"
	"github.com/tomgeorge/todoist-tui/model/loading"
	"github.com/tomgeorge/todoist-tui/model/project_view"
	"github.com/tomgeorge/todoist-tui/model/task_create"
	"github.com/tomgeorge/todoist-tui/services/sync"
	"github.com/tomgeorge/todoist-tui/types"
)

const (
	FooterHeight = 1
	EventsHeight = 1
	// FIXME: I just pulled this out of my ass
	ExtendedHelpHeight = 10
)

type Component interface {
	ComponentName() string
	Model() tea.Model
	Resize(width, height int) tea.Model
}

type Stack interface {
	Push(StackItemV2)
	Pop() StackItemV2
	Top() StackItemV2
}

type StackItemV2 interface {
	Component
	tea.Model
}

type StackItem struct {
	Model tea.Model
	Id    string
}

type Model struct {
	events       events.Model
	ctx          ctx.Context
	footer       footer.Model
	project      *types.Project
	state        *sync.SyncResponse
	errorStyle   lipgloss.Style
	successStyle lipgloss.Style
	projects     []*types.Project
	tasks        []*types.Item
	labels       []*types.Label
	stack        []StackItemV2
	projectView  project_view.Model
	drawer       drawer.Model
	spinner      spinner.Model
	taskCreate   task_create.Model
	showSpinner  bool
	loading      bool
	vp           viewport.Model
	help         help.Model
}

func New(ctx ctx.Context) *Model {
	const (
		defaultLoading     = true
		defaultWidth       = 10
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
		footer:       footer.New(ctx),
		vp:           viewport.New(defaultWidth, defaultHeight),
		help:         help.New(),
	}
}

type SaveAndQuitMsg struct {
	Error error
}

func (m *Model) doViewportStuff() {
	top := m.Top()
	topView := top.View()
	if top.ComponentName() == "loading" {
		// m.ctx.Logger.Debugf("Loading component on top, vp height %d, width %d", m.vp.Height, m.vp.Width)
	}
	// m.ctx.Logger.Debugf("Top is %s, it's view function returned [%s]", top.ComponentName(), top.View())
	m.vp.SetContent(topView)
}

func (m *Model) setViewportDimensions(width, height int) {
	m.vp.Width = width
	m.vp.Height = height
}

func (m *Model) save() tea.Cmd {
	return func() tea.Msg {
		if _, err := os.Stat(m.ctx.Config.StateDir); errors.Is(err, os.ErrNotExist) {
			if err = os.Mkdir(m.ctx.Config.StateDir, 0755); err != nil {
				return SaveAndQuitMsg{Error: fmt.Errorf("saving state file: %v", err)}
			}
		}
		state, err := os.Create(filepath.Join(m.ctx.Config.StateDir, "state.json"))
		if err != nil {
			m.ctx.Logger.Debug("error creating file", err)
			return SaveAndQuitMsg{Error: fmt.Errorf("error saving state file: %v", err)}
		}
		defer state.Close()
		json, err := json.MarshalIndent(m.state, "", "  ")
		if err != nil {
			m.ctx.Logger.Debug("error marshalling json", err)
			return SaveAndQuitMsg{Error: fmt.Errorf("error saving state file: %v", err)}
		}
		_, err = state.Write(json)
		if err != nil {
			m.ctx.Logger.Debug("writing file", err)
			return SaveAndQuitMsg{Error: fmt.Errorf("error saving state file: %v", err)}
		}
		return SaveAndQuitMsg{Error: nil}
	}
}

type StartSpinnerMessage struct{}

func (m Model) Init() tea.Cmd {
	return messages.Push("loading", loading.New(m.ctx))
}

func onPush(m Model, msg messages.PushMessage) (Model, tea.Cmd) {
	newStackItem, ok := msg.Model.(StackItemV2)
	if !ok {
		m.ctx.Logger.Debugf("Failed to push on the stack: %T", msg.Model)
		return m, m.events.Publish("Failed to cast to StackItemV2", m.errorStyle, false, 0)
	}
	m.ctx.Logger.Debugf("Pushed a stack item successfully. Name %s, type %T", newStackItem.ComponentName(), msg.Model)
	m.stack = append(m.stack, newStackItem)
	return m, m.stack[len(m.stack)-1].Init()
}

func pop(m Model, _ messages.PopMessage) (Model, tea.Cmd) {
	m.stack = m.stack[:len(m.stack)-1]
	return m, nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case messages.OpenDrawerMessage:
		pickables := make([]drawer.Pickable, len(m.state.Projects))
		for i, project := range m.state.Projects {
			pickables[i] = project
		}
		m.drawer = drawer.New(pickables)
		return m, nil
	case messages.PushMessage:
		m.ctx.Logger.Debugf("Got PushMessage for %s", msg.Id)
		return onPush(m, msg)
	case messages.LoadingMessage:
		m.loading = true
	case loading.LoadingCompleteMessage:
		m.loading = false
		if len(m.state.Projects) == 0 {
			// FIXME: See the comment in root cmd. I need to watch for like a Panic
			// message or something for these.
			// I should also validate that the state object looks roughly as I would
			// expect it to (e.g. nonzero number of projects)
			return m, tea.Sequence(
				m.events.Publish("No projects found", m.errorStyle, false, 10),
				tea.Quit,
			)
		}
		tasks := lo.Filter(m.state.Items, func(i *types.Item, _ int) bool {
			return i.ProjectId == m.state.Projects[0].Id
		})
		projects := project_view.New(
			m.ctx,
			m.state.Projects[0],
			tasks,
			m.state.Labels,
			m.state.Projects,
			project_view.WithListHeight(m.ctx.MainContentHeight),
			project_view.WithListWidth(m.ctx.MainContentWidth))
		return m, tea.Batch(tea.EnterAltScreen, messages.Push("project_view", projects))
	case messages.PopMessage:
		return pop(m, msg)
	case tea.WindowSizeMsg:
		if m.ctx.MainContentHeight == 0 && m.ctx.MainContentWidth == 0 {
			m.ctx.Logger.Debugf("Initial WindowSizeMsg, Width: %d, Height %d", msg.Width, msg.Height)
		}
		m.ctx.ScreenWidth = msg.Width
		m.ctx.ScreenHeight = msg.Height
		m.ctx.MainContentWidth = msg.Width
		m.ctx.MainContentHeight = msg.Height - FooterHeight
		m.setViewportDimensions(m.ctx.MainContentWidth, m.ctx.MainContentHeight)
		// m.ctx.Logger.Infof("Setting default screen size, Width %d Height %d", m.ctx.ScreenWidth, m.ctx.ScreenHeight)
		// m.ctx.Logger.Debugf("Content Height %d, Content width: %d", m.ctx.MainContentHeight, m.ctx.MainContentWidth)
		return m, cmd
	case messages.OperationPendingMessage:
		m.showSpinner = true
		cmds = append(cmds, m.spinner.Tick)
	case messages.SaveStateMessage:
		// m.ctx.Logger.Debug("Main UI Component - SaveStateMessage")
		if msg.Error != nil {
			return m, m.events.Publish(msg.Error.Error(), m.errorStyle, false, 0)
		}
	case messages.StateMessage:
		if msg.Error != nil {
			return m, tea.Sequence(
				m.events.Publish(msg.Error.Error(), m.errorStyle, false, 0),
				tea.Quit,
			)
		}
		m.state = msg.State
	case messages.OperationResponse:
		m.showSpinner = false
		if msg.Error != nil {
			return m, m.events.Publish(msg.Error.Error(), m.errorStyle, true, 3*time.Second)
		} else {
			m.state = sync.Merge(m.state, msg.State)
			types := lo.Map(msg.Commands, func(c sync.Command, _ int) string { return c.Type })
			eventCmd := m.events.Publish(
				fmt.Sprintf("Command(s) succeeded: %s", types),
				m.successStyle,
				true,
				3*time.Second)
			return m, tea.Batch(eventCmd, messages.StateUpdated)
		}
	case messages.StateUpdatedMessage:
		for i, stackItem := range m.stack {
			if stackItem.ComponentName() == "project_view" {
				newStackItem := StackItemV2(project_view.NewFromState(m.ctx, m.state))
				m.stack[i] = newStackItem
			}
		}

	case messages.ChangeProjectMessage:
		// FIXME: I actually have the whole list of projects in the project view
		// model so I could probably do this calculation in there when sending the
		// message and send back the destination index
		currentProjectIndex := lo.IndexOf(m.state.Projects, msg.From)
		var newProjectIndex int
		switch currentProjectIndex {
		// FIXME: this should never happen and I should maybe do something if it happens
		case -1:
			return m, nil
		case 0:
			if msg.Direction == messages.NextProject {
				newProjectIndex = currentProjectIndex + 1
			} else {
				newProjectIndex = len(m.state.Projects) - 1
			}
		case len(m.state.Projects) - 1:
			if msg.Direction == messages.NextProject {
				newProjectIndex = 0
			} else {
				newProjectIndex = currentProjectIndex - 1
			}
		default:
			if msg.Direction == messages.NextProject {
				newProjectIndex = currentProjectIndex + 1
			} else {
				newProjectIndex = currentProjectIndex - 1
			}
		}
		m.ctx.Logger.Debugf("length: %d, project: %s current index: %d direction: %s, new index: %d",
			len(m.state.Projects),
			msg.From.Name,
			currentProjectIndex,
			msg.Direction,
			newProjectIndex,
		)
		tasks := lo.Filter(m.state.Items, func(i *types.Item, _ int) bool {
			return i.ProjectId == m.state.Projects[newProjectIndex].Id
		})
		projects := project_view.New(
			m.ctx,
			m.state.Projects[newProjectIndex],
			tasks,
			m.state.Labels,
			m.state.Projects,
		)
		m.ctx.Logger.Debug("got project change message", "new index", newProjectIndex, "project name", m.state.Projects[newProjectIndex].Name)
		return m, tea.Sequence(messages.Pop(), messages.Push("project_view", projects))

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+p":
			return m, messages.OpenDrawer
		case "ctrl+c":
			return m, m.save()
		case "?":
			m.footer.SetShowExtendedHelp(!m.footer.ShowExtendedHelp())
			m.ctx.MainContentHeight = m.ctx.ScreenHeight - FooterHeight - EventsHeight
			if m.footer.ShowExtendedHelp() {
				m.ctx.MainContentHeight = m.ctx.ScreenHeight - ExtendedHelpHeight - EventsHeight
			}
			top := m.Top().Resize(m.ctx.MainContentWidth, m.ctx.MainContentHeight)
			stackItem, ok := top.(StackItemV2)
			if !ok {
				m.ctx.Logger.Error("failed to coerce stack item to StackItemV2")
			}
			m.stack[len(m.stack)-1] = stackItem
			m.doViewportStuff()
		}
	case SaveAndQuitMsg:
		m.ctx.Logger.Debug("caught a save and quit message")
		if msg.Error != nil {
			failed := m.events.Publish(msg.Error.Error(), m.errorStyle, false, 0)
			return m, tea.Sequence(failed, tea.Quit)
		}
		return m, tea.Sequence(tea.ExitAltScreen, tea.Quit)
	}

	// Update top model
	model := m.Top().Model()
	updated, cmd := model.Update(msg)
	cmds = append(cmds, cmd)

	stackItem, ok := updated.(StackItemV2)
	if !ok {
		m.ctx.Logger.Errorf("Failed to push on the stack %T", updated)
		return m, m.events.Publish("Failed to cast to StackItemV2", m.errorStyle, false, 0)
	}
	topIndex := len(m.stack) - 1
	m.stack[topIndex] = stackItem
	m.events, cmd = m.events.Update(msg)
	cmds = append(cmds, cmd)
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)
	m.footer, cmd = m.footer.Update(msg)
	cmds = append(cmds, cmd)
	m.vp, cmd = m.vp.Update(msg)
	m.vp.SetContent(m.Top().View())
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) Top() StackItemV2 {
	return m.stack[len(m.stack)-1]
}

func (m Model) View() string {
	if len(m.stack) == 0 {
		return "Nothing on the stack"
	}
	var sections []string
	m.ctx.Logger.Debugf("Main UI View - Content Dimensions %dx%d", m.ctx.MainContentWidth, m.ctx.MainContentHeight)
	mainContent := lipgloss.NewStyle().Render(
		m.vp.View(),
	)

	if m.loading && m.Top().ComponentName() == "loading" {
		mainContent = m.Top().View()
	}
	// m.ctx.Logger.Debug(mainContent)
	sections = append(sections, mainContent)
	if m.showSpinner {
		sections = append(sections, m.spinner.View())
	}
	if m.drawer.Enabled() {
		sections = append(sections, m.drawer.View())
	}
	sections = append(sections, m.events.View())
	if !m.loading {
		sections = append(sections, m.footer.View())
	}
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return content
}

func (m Model) Events() events.Model {
	return m.events
}
