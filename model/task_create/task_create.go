package task_create

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/model/button"
	"github.com/tomgeorge/todoist-tui/model/date_picker"
	"github.com/tomgeorge/todoist-tui/model/events"
	"github.com/tomgeorge/todoist-tui/model/picker"
	"github.com/tomgeorge/todoist-tui/model/task_description"
	"github.com/tomgeorge/todoist-tui/model/task_title"
	"github.com/tomgeorge/todoist-tui/services/sync"
	"github.com/tomgeorge/todoist-tui/types"
)

type keyMap struct {
	ScrollUp   key.Binding
	ScrollDown key.Binding
	Help       key.Binding
	Quit       key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ScrollDown, k.ScrollUp, k.Quit},
		{k.Help},
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
		key.WithHelp("ctrl+?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
}

type Model struct {
	ctx           ctx.Context
	viewport      viewport.Model
	focusedStyle  lipgloss.Style
	width         int
	height        int
	project       picker.Model
	title         task_title.Model
	description   task_description.Model
	labels        picker.Model
	priority      picker.Model
	dueDate       date_picker.Model
	submit        button.Model
	events        events.Model
	focused       Section
	help          help.Model
	keys          keyMap
	task          *types.Item
	parentProject *types.Project
	taskLabels    []*types.Label
	showSpinner   bool
	spinner       spinner.Model
}

type Section int

const (
	titleSection Section = iota
	projectSection
	dueSection
	descriptionSection
	labelsSection
	prioritySection
	submitSection
	lastSection   = submitSection
	initialWidth  = 50
	initialHeight = 50
)

type ModelOption func(*Model)

type ItemUpdatedMsg struct {
	Task  *types.Item
	Error error
}

var validationStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#41424e")).
	Foreground(lipgloss.Color("#e5c891")).
	MarginBottom(1)

func New(ctx ctx.Context, opts ...ModelOption) *Model {
	defaultProjects := []picker.PickerItem{}
	defaultLabels := []picker.PickerItem{}
	defaultPriorities := []picker.PickerItem{}

	projects := picker.NewModel(
		picker.WithItems(defaultProjects),
		picker.WithLabel("Project"),
		picker.WithMultipleSelection(false),
		picker.WithRequiredSelection(1),
		// picker.WithSelected([]picker.PickerItem{defaultProjects[0]}),
		// picker.WithLabelStyle(componentLabelStyle),
		// picker.WithFocusedLabelStyle(focusedComponentLabelStyle),
		picker.WithValidationStyle(validationStyle),
		picker.WithPlaceholder("Select a project"),
	)
	titleModel := task_title.New(
		task_title.WithLabel("Title"),
		task_title.WithContent("Buy Bread"),
		// task_title.WithTextStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#a12477"))),
		// task_title.WithLabelStyle(componentLabelStyle),
		// task_title.WithFocusedLabelStyle(focusedComponentLabelStyle),
	)
	descriptionModel := task_description.NewModel(
		// task_description.WithLabelStyle(componentLabelStyle),
		// task_description.WithFocusedLabelStyle(focusedComponentLabelStyle),
		task_description.WithValue("I need some bread from the store"),
	)
	dueDate := date_picker.NewModel(
		date_picker.WithLabel("Due Date"),
		date_picker.WithDueDate(false),
		date_picker.WithShowHelpUnderComponent(false),
	)
	labels := picker.NewModel(
		picker.WithItems(defaultLabels),
		picker.WithLabel("Labels"),
		// picker.WithLabelStyle(componentLabelStyle),
		// picker.WithFocusedLabelStyle(focusedComponentLabelStyle),
		picker.WithPlaceholder("Enter a label"),
	)

	priority := picker.NewModel(
		picker.WithLabel("Priority"),
		picker.WithMultipleSelection(false),
		picker.WithRequiredSelection(0),
		picker.WithItems(defaultPriorities),
	)

	// onSubmit := func(payload interface{}) tea.Cmd {
	// 	// return func() tea.Msg {
	// 	// 	log.Printf("onSubmit defined by the parent")
	// 	// 	c := sync.NewClient(nil).WithAuthToken(os.Getenv("TODOIST_API_TOKEN"))
	// 	// 	task, err := c.UpdateTask(context.Background(), sync.UpdateItemArgs{
	// 	//       Id: model.
	// 	//     })
	// 	// 	return ItemUpdatedMsg{task, err}
	// 	// }
	// }
	submit := button.New(
		button.WithText("Create Task"),
		button.WithEnabled(true),
		button.WithFocusedStyle(lipgloss.NewStyle().Background(lipgloss.Color("#a6d189")).Foreground(lipgloss.Color("#414559"))),
	)
	events := events.New(ctx)

	viewport := viewport.New(initialWidth, initialHeight)
	model := &Model{
		ctx:      ctx,
		viewport: viewport,
		focusedStyle: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true),
		title:         *titleModel,
		project:       *projects,
		dueDate:       *dueDate,
		description:   *descriptionModel,
		priority:      *priority,
		labels:        *labels,
		submit:        *submit,
		focused:       titleSection,
		help:          help.New(),
		keys:          defaultKeys,
		task:          nil,
		parentProject: nil,
		taskLabels:    nil,
		events:        *events,
		showSpinner:   false,
		spinner:       spinner.New(spinner.WithSpinner(spinner.Dot)),
	}

	for _, opt := range opts {
		opt(model)
	}

	if model.task != nil {
		model.title.SetContent(model.task.Content)
		model.description.SetContent(model.task.TaskDescription)
		model.project.SetSelected([]picker.PickerItem{model.parentProject})
		labelItems := make([]picker.PickerItem, 0)
		for _, label := range model.taskLabels {
			if slices.Contains(model.task.Labels, label.Name) {
				log.Printf("Adding label %v", label)
				label := *label
				labelItems = append(labelItems, label)
			}
		}
		model.labels.SetSelected(labelItems)
		model.priority.SetSelected([]picker.PickerItem{types.Priority(model.task.Priority)})
		if model.task.Due != nil {
			log.Printf("Due date isn't nil")
			if dd, err := time.Parse(time.RFC3339, model.task.Due.Date); err != nil {
				log.Printf("Didn't parse me a due date")
				model.dueDate.SetNaturalLanguageDueDate(model.task.Due.String)
			} else {
				log.Printf("Parsed me a due date, setting to %v", dd)
				model.dueDate.SetAbsoluteDueDate(dd.In(time.Local))
			}
		}
	}

	return model
}

func WithTask(task *types.Item) ModelOption {
	return func(m *Model) {
		m.task = task
	}
}

func WithParentProject(project *types.Project) ModelOption {
	return func(m *Model) {
		m.parentProject = project
	}
}

func WithLabels(labels []*types.Label) ModelOption {
	return func(m *Model) {
		m.taskLabels = labels
	}
}

func WithFocusedStyle(style lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.focusedStyle = style
	}
}

func WithPossibleLabels(labels []*types.Label) ModelOption {
	return func(m *Model) {
		items := make([]picker.PickerItem, len(labels))
		for i, label := range labels {
			items[i] = label
		}
		m.labels.SetItems(items)
	}
}

func WithProjects(projects []*types.Project) ModelOption {
	return func(m *Model) {
		items := make([]picker.PickerItem, len(projects))
		for i, project := range projects {
			items[i] = project
		}
		m.project.SetItems(items)
	}
}

func WithPriorities(priorities []types.Priority) ModelOption {
	return func(m *Model) {
		items := make([]picker.PickerItem, len(priorities))
		for i, priority := range priorities {
			items[i] = priority
		}
		m.priority.SetItems(items)
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Order
// 1. Title
// 2. Project
// 3. Due Date
// 4. Description
// 5. Labels
// 6. Submit
func (m *Model) HandleScrollDown() {
	switch m.focused {
	case titleSection:
		m.title.FocusOff()
		m.project.FocusOn()
	case projectSection:
		m.project.FocusOff()
		m.dueDate.FocusOn()
	case dueSection:
		m.dueDate.FocusOff()
		m.description.FocusOn()
	case descriptionSection:
		m.description.FocusOff()
		m.labels.FocusOn()
	case labelsSection:
		m.labels.FocusOff()
		m.priority.FocusOn()
	case prioritySection:
		m.priority.FocusOff()
	}
	m.focused++
}

func (m *Model) HandleScrollUp() {
	switch m.focused {
	case projectSection:
		m.project.FocusOff()
		m.title.FocusOn()
	case dueSection:
		m.dueDate.FocusOff()
		m.project.FocusOn()
	case descriptionSection:
		m.description.FocusOff()
		m.dueDate.FocusOn()
	case labelsSection:
		m.labels.FocusOff()
		m.description.FocusOn()
	case prioritySection:
		m.priority.FocusOff()
		m.labels.FocusOn()
	case submitSection:
		m.priority.FocusOn()
	}
	m.focused--
}

// Update the help status of all children
// Whether or not they are rendered should be controlled in the parent's View()
func (m *Model) handleHelp() {
	m.title.SetHelp(m.help.ShowAll)
	m.project.SetHelp(m.help.ShowAll)
	m.dueDate.SetHelp(m.help.ShowAll)
	m.description.SetHelp(m.help.ShowAll)
	m.labels.SetHelp(m.help.ShowAll)
	m.priority.SetHelp(m.help.ShowAll)
	m.submit.SetHelp(m.help.ShowAll)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case *tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case button.SubmitMsg:
		log.Printf("Parent got the SubmitMsg from the child")
		m.showSpinner = true
		return m, m.UpdateTask()
	case ItemUpdatedMsg:
		log.Printf("Item updated msg %v %v", msg.Task, msg.Error)
		var event events.NewMessage
		if msg.Error == nil {
			event = events.NewMessage{
				Duration: 10 * time.Second,
				Message:  "Task updated successfully",
				Style:    lipgloss.NewStyle().Foreground(lipgloss.Color("#40a02b")),
			}
		} else {
			event = events.NewMessage{
				Duration: 10 * time.Second,
				Message:  msg.Error.Error(),
				Style:    lipgloss.NewStyle().Foreground(lipgloss.Color("#d20f39")),
			}
		}
		m.showSpinner = false
		events, cmd := m.events.Update(event)
		m.events = events
		return m, cmd
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		m.width = msg.Width
		m.height = msg.Height
	// x, y := m.focusedStyle.GetFrameSize()
	// 	return m, commands.NotifyResize(m.width, m.height, x, y)
	// case commands.ResizeChildMessage:
	// 	log.Println("Got resizechild message")
	// 	m, cmd = m.Resize(msg)
	// 	cmds = append(cmds, cmd)
	// case timer.TickMsg:
	// 	log.Printf("Task create tick msg")
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			log.Printf("Got help")
			m.help.ShowAll = !m.help.ShowAll
			// We are intercepting the child components help commands and not
			// delegating to them
			m.handleHelp()
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.ScrollDown):
			log.Printf("Scroll request, focused is %d", m.focused)
			if m.focused != lastSection {
				m.HandleScrollDown()
			}
		case key.Matches(msg, m.keys.ScrollUp):
			log.Printf("Scroll request, focused is %d", m.focused)
			if m.focused > 0 {
				m.HandleScrollUp()
			}
		}
	}

	log.Printf("In switch focused")
	switch m.focused {
	case titleSection:
		log.Printf("Title section is focused")
		m.title.FocusOn()
		m.title, cmd = m.title.Update(msg)
		cmds = append(cmds, cmd)
	case projectSection:
		m.project.FocusOn()
		log.Printf("projects are focused")
		m.project, cmd = m.project.Update(msg)
		cmds = append(cmds, cmd)
	case dueSection:
		m.dueDate.FocusOn()
		m.dueDate, cmd = m.dueDate.Update(msg)
		cmds = append(cmds, cmd)
	case descriptionSection:
		log.Println("Description is focused")
		m.description, cmd = m.description.Update(msg)
		cmds = append(cmds, cmd)
	case labelsSection:
		log.Println("Labels are focused")
		m.labels, cmd = m.labels.Update(msg)
		cmds = append(cmds, cmd)
	case prioritySection:
		m.priority, cmd = m.priority.Update(msg)
		cmds = append(cmds, cmd)
	case submitSection:
		log.Printf("Submit focused")
		m.submit, cmd = m.submit.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.events, cmd = m.events.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

type selectedFunc func(m *Model) bool

func titleSelected(m *Model) bool {
	return m.focused == titleSection
}

func descriptionSelected(m *Model) bool {
	return m.focused == descriptionSection
}

func dueDateSelected(m *Model) bool {
	return m.focused == dueSection
}

func labelsSelected(m *Model) bool {
	return m.focused == labelsSection
}

func prioritySelected(m *Model) bool {
	return m.focused == prioritySection
}

func submitSelected(m *Model) bool {
	return m.focused == submitSection
}

func projectsSelected(m *Model) bool {
	return m.focused == projectSection
}

func (m *Model) renderContent(isSelected selectedFunc, fullWidth bool, content string) string {
	if isSelected(m) {
		if fullWidth {
			return m.focusedStyle.Copy().Width(m.width - m.focusedStyle.GetHorizontalFrameSize()).Render(content)
		}
		return m.focusedStyle.Render(content)
	}
	return content
}

func (m *Model) getTaskContent() types.CreateTaskRequest {
	items := m.labels.GetSelectedItems()
	labels := make([]string, len(items))
	for i, item := range items {
		labels[i] = item.Render()
	}

	taskRequest := types.CreateTaskRequest{
		Content:     m.title.GetContent(),
		Description: m.description.GetContent(),
		Labels:      labels,
	}

	dueDate := m.dueDate.GetContent()
	switch {
	case dueDate.HasDueDate && dueDate.HumanInputDate != "":
		taskRequest.DueString = dueDate.HumanInputDate
	case dueDate.HasDueDate && dueDate.HumanInputDate == "" && dueDate.IncludeHoursAndMinutes:
		taskRequest.DueDateTime = dueDate.AbsoluteDate.Format(time.RFC3339)
	default:
		taskRequest.DueDate = dueDate.AbsoluteDate.Format("2006-01-02")
	}

	priorityItems := m.priority.GetSelectedItems()
	if len(priorityItems) == 1 {
		priority, err := strconv.Atoi(priorityItems[0].Render())
		if err != nil {
			//FIXME: handle error
		}
		taskRequest.Priority = priority
	}

	if len(m.project.GetSelectedItems()) == 1 {
		project, ok := m.project.GetSelectedItems()[0].(types.Project)
		if !ok {
			//FIXME handle error
		}
		taskRequest.ProjectId = project.Id
	}
	return taskRequest
}

func (m *Model) MakeTask() types.CreateTaskRequest {
	return m.getTaskContent()
}

func renderTaskContent(t types.CreateTaskRequest) string {
	json, _ := json.Marshal(t)
	return lipgloss.NewStyle().Render(string(json))
}

func (m Model) View() string {
	sections := []string{}
	sections = append(sections, m.renderContent(titleSelected, true, m.title.View()))
	sections = append(sections, m.renderContent(projectsSelected, true, m.project.View()))
	sections = append(sections, m.renderContent(dueDateSelected, true, m.dueDate.View()))
	sections = append(sections, m.renderContent(descriptionSelected, true, m.description.View()))
	sections = append(sections, m.renderContent(labelsSelected, true, m.labels.View()))
	sections = append(sections, m.renderContent(prioritySelected, true, m.priority.View()))
	sections = append(sections, m.renderContent(submitSelected, false, m.submit.View()))

	switch m.focused {
	case titleSection:
		if m.title.Help().ShowAll {
			sections = append(sections, m.title.Help().View(m.title.HelpKeys()))
		}
	case projectSection:
		if m.project.Help().ShowAll {
			sections = append(sections, m.project.Help().View(m.project.HelpKeys()))
		}
	case dueSection:
		if m.dueDate.Help().ShowAll {
			sections = append(sections, m.dueDate.Help().View(m.dueDate.HelpKeys()))
		}
	case descriptionSection:
		if m.description.Help().ShowAll {
			sections = append(sections, m.description.Help().View(m.description.HelpKeys()))
		}
	case labelsSection:
		if m.labels.Help().ShowAll {
			sections = append(sections, m.labels.Help().View(m.labels.HelpKeys()))
		}
	case prioritySection:
		if m.priority.Help().ShowAll {
			sections = append(sections, m.priority.Help().View(m.priority.HelpKeys()))
		}
	case submitSection:
		if m.submit.Help().ShowAll {
			sections = append(sections, m.submit.Help().View(m.submit.HelpKeys()))
		}
	}
	if m.showSpinner {
		sections = append(sections, m.spinner.View())
	}
	sections = append(sections, m.events.View())
	sections = append(sections, m.help.View(m.keys))
	diff, _ := m.diff()
	diffJson, _ := json.Marshal(diff)
	sections = append(sections, fmt.Sprintf("%s\n", string(diffJson)))
	// buf, _ := json.Marshal(m.task.Due)
	// parsed, _ := time.Parse("2006-01-02T15:04:04", m.task.Due.Date)
	// sections = append(sections, fmt.Sprintf("JSON of due date %s parsed time %s", string(buf), m.task.Due.Date))
	// If the form grows larger than the screensize we may need this
	// m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, sections...))
	// return m.viewport.View()
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) diff() (*sync.UpdateItemArgs, error) {
	args := &sync.UpdateItemArgs{}
	args.Id = m.task.Id
	if m.task.Content != m.title.GetContent() {
		args.Content = m.title.GetContent()
	}
	if m.task.TaskDescription != m.description.GetContent() {
		args.Description = m.description.GetContent()
	}
	dueDate := m.dueDate.GetContent()
	log.Printf("duedate %v\n", dueDate)
	switch {
	case dueDate.HasDueDate && dueDate.HumanInputDate != "":
		args.Due = &types.DueDate{}
		args.Due.String = dueDate.HumanInputDate
	case dueDate.HasDueDate && dueDate.HumanInputDate == "" && dueDate.IncludeHoursAndMinutes:
		args.Due = &types.DueDate{}
		// convert to UTC
		location, err := time.LoadLocation("Local")
		if err != nil {
			return nil, fmt.Errorf("updating task: couldn't get local time zone")
		}
		args.Due.Date = dueDate.AbsoluteDate.In(location).UTC().Format(time.RFC3339)
	case !dueDate.HasDueDate:
		break
	default:
		args.Due = &types.DueDate{}
		args.Due.Date = dueDate.AbsoluteDate.Format("2006-01-02")
	}
	selectedLabels := []string{}
	for _, label := range m.labels.GetSelectedItems() {
		selectedLabels = append(selectedLabels, label.Render())
	}

	if !equalIgnoreOrder(selectedLabels, m.task.Labels) {
		args.Labels = selectedLabels
	}

	priorityItems := m.priority.GetSelectedItems()
	if len(priorityItems) == 1 {
		priority, _ := strconv.Atoi(priorityItems[0].Render())
		if m.task.Priority != priority {
			args.Priority = priority
		}
	}
	return args, nil
}

// Compares two slices, ignoring order, like set equality
func equalIgnoreOrder[T comparable](s1 []T, s2 []T) bool {
	if len(s1) != len(s2) {
		return false
	}
	diff := make(map[T]int, len(s1))
	for _, e := range s1 {
		diff[e]++
	}
	for _, e := range s2 {
		if _, ok := diff[e]; !ok {
			return false
		}
		diff[e]--
		if diff[e] == 0 {
			delete(diff, e)
		}
	}
	return len(diff) == 0
}

func (m *Model) UpdateTask() tea.Cmd {
	return func() tea.Msg {
		c := sync.NewClient(nil).WithAuthToken(os.Getenv("TODOIST_API_TOKEN"))
		diff, _ := m.diff()
		task, err := c.UpdateTask(context.Background(), *diff)
		return ItemUpdatedMsg{task, err}
	}
}
