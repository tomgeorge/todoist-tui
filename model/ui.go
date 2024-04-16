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
	"github.com/tomgeorge/todoist-tui/model/events"
	"github.com/tomgeorge/todoist-tui/model/footer"
	"github.com/tomgeorge/todoist-tui/model/fuzzyfinder"
	"github.com/tomgeorge/todoist-tui/model/loading"
	"github.com/tomgeorge/todoist-tui/model/project_view"
	"github.com/tomgeorge/todoist-tui/model/stack_item"
	"github.com/tomgeorge/todoist-tui/model/task_create"
	"github.com/tomgeorge/todoist-tui/services/sync"
	"github.com/tomgeorge/todoist-tui/types"
)

const (
	FooterHeight  = 1
	EventsHeight  = 1
	SpinnerHeight = 1
	// FIXME: I just pulled this out of my ass
	ExtendedHelpHeight = 10
)

type Model struct {
	project         *types.Project
	state           *sync.SyncResponse
	help            help.Model
	successStyle    lipgloss.Style
	errorStyle      lipgloss.Style
	vp              viewport.Model
	stack           []stack_item.StackItem
	tasks           []*types.Item
	labels          []*types.Label
	projects        []*types.Project
	projectView     project_view.Model
	spinner         spinner.Model
	events          events.Model
	ctx             ctx.Context
	taskCreate      task_create.Model
	footer          footer.Model
	fuzzyfinder     fuzzyfinder.Model
	fuzzyfinderOpen bool
	showSpinner     bool
	loading         bool
}

// Calculates the viewport height minus the other UI components
func (m Model) ContentHeight(screenHeight int) int {
	if m.fuzzyfinderOpen {
		return screenHeight - lipgloss.Height(m.fuzzyfinder.View()) - EventsHeight - SpinnerHeight
	}
	height := screenHeight - FooterHeight - EventsHeight - SpinnerHeight
	if m.footer.ShowExtendedHelp() {
		height = screenHeight - ExtendedHelpHeight - EventsHeight - SpinnerHeight
	}
	return height
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
	newStackItem, ok := msg.Model.(stack_item.StackItem)
	if !ok {
		m.ctx.Logger.Debugf("Failed to push on the stack: %T", msg.Model)
		return m, m.events.Publish("Failed to cast to StackItem", m.errorStyle, false, 0)
	}
	m.ctx.Logger.Debugf("Pushed a stack item successfully. Name %s, type %T", newStackItem.ComponentName(), msg.Model)
	m.stack = append(m.stack, newStackItem)
	return m, m.stack[len(m.stack)-1].Init()
}

func pop(m Model, _ messages.PopMessage) (Model, tea.Cmd) {
	m.stack = m.stack[:len(m.stack)-1]
	return m, func() tea.Msg {
		return true
	}
}

func (m *Model) SelectProject(project types.Project) tea.Cmd {
	tasks := lo.Filter(m.state.Items, func(i *types.Item, _ int) bool {
		return i.ProjectId == project.Id
	})
	isProjectInState := func(item *types.Project) bool {
		return item.Id == project.Id
	}
	projectInState, _ := lo.Find(m.state.Projects, isProjectInState)
	view := project_view.New(
		m.ctx,
		projectInState,
		tasks,
		m.state.Labels,
		m.state.Projects,
	)
	return tea.Sequence(messages.Pop(), messages.Push("project_view", view))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if m.fuzzyfinderOpen {
		finderName := m.fuzzyfinder.Name()
		if msg, ok := msg.(tea.KeyMsg); ok {
			if finderName == "all" && msg.String() == "ctrl+f" {
				m.fuzzyfinderOpen = false
				m.onFuzzyFinderClose()
			}
			if finderName == "projects" && msg.String() == "ctrl+p" {
				m.fuzzyfinderOpen = false
				m.onFuzzyFinderClose()
			}
			// FIXME : Handle showing the help when you previously were in the
			// fuzzyfinder
			if msg.String() == "ctrl+_" {
				m.fuzzyfinderOpen = false
			}
		}
		if msg, ok := msg.(fuzzyfinder.SelectionMade); ok {
			if project, ok := msg.Selection.(*types.Project); ok {
				m.ctx.Logger.Debug("Got something %s", project.Name)
				m.fuzzyfinderOpen = false
				m.onFuzzyFinderClose()
				return m, m.SelectProject(*project)
			}
		}
		m.fuzzyfinder, cmd = m.fuzzyfinder.Update(msg)
		return m, cmd
	}
	switch msg := msg.(type) {
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
		m.ctx.MainContentHeight = m.ContentHeight(msg.Height)
		m.setViewportDimensions(m.ctx.MainContentWidth, m.ctx.MainContentHeight)
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
				newStackItem := stack_item.StackItem(project_view.NewFromState(m.ctx, m.state))
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
			if !m.fuzzyfinderOpen {
				m.fuzzyfinder = m.makeFuzzyFinder("projects")
				m.fuzzyfinderOpen = true
				return m.onFuzzyFinderOpen()
			}
			m.fuzzyfinderOpen = false
			m.onFuzzyFinderClose()
			return m, nil
		case "ctrl+c":
			return m, m.save()
		case "ctrl+_":
			m.footer.SetShowExtendedHelp(!m.footer.ShowExtendedHelp())
			m.ctx.MainContentHeight = m.ContentHeight(m.ctx.ScreenHeight)
			top := m.Top().Resize(m.ctx.MainContentWidth, m.ctx.MainContentHeight)
			m.setViewportDimensions(m.ctx.MainContentWidth, m.ctx.MainContentHeight)
			stackItem, ok := top.(stack_item.StackItem)
			if !ok {
				m.ctx.Logger.Error("failed to coerce stack item to StackItem")
			}
			m.stack[len(m.stack)-1] = stackItem
		case "ctrl+f":
			if !m.fuzzyfinderOpen {
				m.fuzzyfinder = m.makeFuzzyFinder("all")
				m.fuzzyfinderOpen = true
				return m.onFuzzyFinderOpen()
			}
			m.fuzzyfinderOpen = false
			m.onFuzzyFinderClose()
			return m, nil
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

	stackItem, ok := updated.(stack_item.StackItem)
	if !ok {
		m.ctx.Logger.Errorf("Failed to push on the stack %T", updated)
		return m, m.events.Publish("Failed to cast to StackItem", m.errorStyle, false, 0)
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

func (m Model) Top() stack_item.StackItem {
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
	sections = append(sections, mainContent)
	if m.showSpinner {
		sections = append(sections, m.spinner.View())
	}
	sections = append(sections, m.events.View())
	if !m.loading && !m.fuzzyfinderOpen {
		sections = append(sections, m.footer.View(m.Top(), m.ctx.ScreenWidth))
	}
	if m.fuzzyfinderOpen {
		sections = append(sections, m.fuzzyfinder.View())
	}
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return content
}

func (m Model) makeFuzzyFinder(searchName string) fuzzyfinder.Model {
	sources := make(fuzzyfinder.Sources, 0)
	switch searchName {
	case "projects":
		for _, project := range m.state.Projects {
			sources = append(sources, project)
		}
	default:
		for _, project := range m.state.Projects {
			sources = append(sources, project)
		}
		for _, label := range m.state.Labels {
			sources = append(sources, label)
		}
	}
	return fuzzyfinder.New(m.ctx, sources, 10,
		fuzzyfinder.WithStyles(m.ctx.Theme.FuzzyFinder),
		fuzzyfinder.WithSearchName(searchName))
}

func (m *Model) onFuzzyFinderOpen() (tea.Model, tea.Cmd) {
	m.ctx.MainContentHeight = m.ContentHeight(m.ctx.ScreenHeight)
	top := m.Top().Resize(m.ctx.MainContentWidth, m.ctx.MainContentHeight)
	m.setViewportDimensions(m.ctx.MainContentWidth, m.ctx.MainContentHeight)
	stackItem, ok := top.(stack_item.StackItem)
	if !ok {
		m.ctx.Logger.Error("failed to coerce stack item to StackItem")
	}
	m.stack[len(m.stack)-1] = stackItem
	return m, m.fuzzyfinder.Init()
}

func (m *Model) onFuzzyFinderClose() {
	m.ctx.MainContentHeight = m.ContentHeight(m.ctx.ScreenHeight)
	m.setViewportDimensions(m.ctx.MainContentWidth, m.ctx.MainContentHeight)
	top := m.Top().Resize(m.ctx.MainContentWidth, m.ctx.MainContentHeight)
	stackItem, ok := top.(stack_item.StackItem)
	if !ok {
		m.ctx.Logger.Error("failed to coerce stack item to StackItem")
	}
	m.stack[len(m.stack)-1] = stackItem
}

func (m Model) Events() events.Model {
	return m.events
}
