package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/tomgeorge/todoist-tui/pkg/cache"
	"github.com/tomgeorge/todoist-tui/pkg/types"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const todoist = "https://api.todoist.com/rest/v2"

type View int

const (
	ProjectsView View = 0
	TasksView    View = 1
	TaskWindow   View = 2
)

var (
	colors map[string]lipgloss.Color = map[string]lipgloss.Color{
		"berry_red":   lipgloss.Color("#b8256f"),
		"red":         lipgloss.Color("#db4035"),
		"orange":      lipgloss.Color("#ff9933"),
		"yellow":      lipgloss.Color("#fad000"),
		"olive_green": lipgloss.Color("#afb83b"),
		"lime_green":  lipgloss.Color("#7ecc49"),
		"green":       lipgloss.Color("#299438"),
		"mint_green":  lipgloss.Color("#6accbc"),
		"teal":        lipgloss.Color("#158fad"),
		"sky_blue":    lipgloss.Color("#14aaf5"),
		"light_blue":  lipgloss.Color("#96c3eb"),
		"blue":        lipgloss.Color("#4073ff"),
		"grape":       lipgloss.Color("#884dff"),
		"violet":      lipgloss.Color("#af38eb"),
		"lavender":    lipgloss.Color("#eb96eb"),
		"magenta":     lipgloss.Color("#e05194"),
		"salmon":      lipgloss.Color("#ff8d85"),
		"charcoal":    lipgloss.Color("#808080"),
		"grey":        lipgloss.Color("#b8b8b8"),
		"taupe":       lipgloss.Color("#ccac93"),
	}
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	cache           cache.Cache
	choice          int
	projectCursor   int
	projects        []types.Project
	selectedProject types.Project
	selectedTask    types.Task
	tasks           []types.Task
	taskList        list.Model
	view            View
	projectModel    table.Model
	taskModel       table.Model
	help            help.Model
	fetchingTasks   bool
	taskView        *taskViewModel
	viewingTask     bool
}

type taskViewModel struct {
	task    *types.Task
	flexBox *flexbox.HorizontalFlexBox
}

func initializeTaskViewModel() *taskViewModel {
	return &taskViewModel{
		task:    &types.Task{},
		flexBox: flexbox.NewHorizontal(0, 0),
	}
}

func (t *taskViewModel) UpdateWindowSize(width, height int) {
	t.flexBox.SetWidth(width)
	t.flexBox.SetHeight(height)
}

func (t *taskViewModel) Init() tea.Cmd {
	return nil
}

func (t *taskViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		log.Println("TaskViewModel - Got windowsizeMessage")
		t.flexBox.SetWidth(msg.Width)
		t.flexBox.SetHeight(msg.Height)
	}
	return t, nil
}

var (
	style1  = lipgloss.NewStyle().Background(lipgloss.Color("#fc5c65"))
	style2  = lipgloss.NewStyle().Background(lipgloss.Color("#fd9644"))
	style3  = lipgloss.NewStyle().Background(lipgloss.Color("#fed330"))
	style4  = lipgloss.NewStyle().Background(lipgloss.Color("#26de81"))
	style5  = lipgloss.NewStyle().Background(lipgloss.Color("#2bcbba"))
	style6  = lipgloss.NewStyle().Background(lipgloss.Color("#eb3b5a"))
	style7  = lipgloss.NewStyle().Background(lipgloss.Color("#fa8231"))
	style8  = lipgloss.NewStyle().Background(lipgloss.Color("#f7b731"))
	style9  = lipgloss.NewStyle().Background(lipgloss.Color("#20bf6b"))
	style10 = lipgloss.NewStyle().Background(lipgloss.Color("#0fb9b1"))
	style11 = lipgloss.NewStyle().Background(lipgloss.Color("#45aaf2"))
	style12 = lipgloss.NewStyle().Background(lipgloss.Color("#4b7bec"))
	style13 = lipgloss.NewStyle().Background(lipgloss.Color("#a55eea"))
	style14 = lipgloss.NewStyle().Background(lipgloss.Color("#d1d8e0"))
	style15 = lipgloss.NewStyle().Background(lipgloss.Color("#778ca3"))
	style16 = lipgloss.NewStyle().Background(lipgloss.Color("#2d98da"))
	style17 = lipgloss.NewStyle().Background(lipgloss.Color("#3867d6"))
	style18 = lipgloss.NewStyle().Background(lipgloss.Color("#8854d0"))
	style19 = lipgloss.NewStyle().Background(lipgloss.Color("#a5b1c2"))
	style20 = lipgloss.NewStyle().Background(lipgloss.Color("#4b6584"))
)

func (t *taskViewModel) View() string {
	headerStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Background(lipgloss.Color("#a5b1c2"))

	log.Println(t.flexBox.GetHeight())
	log.Println(t.flexBox.GetWidth())


	cols := []*flexbox.Column{
		t.flexBox.NewColumn().AddCells(
			flexbox.NewCell(8, 1).SetStyle(style1),
      flexbox.NewCell(8, 9).SetStyle(style2),
		),
		t.flexBox.NewColumn().AddCells(
			flexbox.NewCell(6, 1).SetStyle(style3),
		),
	}
	t.flexBox.AddColumns(cols)
	return t.flexBox.Render()
	contentStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder())

	sidebarStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder())

	content := lipgloss.JoinVertical(lipgloss.Top,
		headerStyle.Render(t.task.ProjectId),
		contentStyle.Render(t.task.Content),
		contentStyle.Render(t.task.Description))
	sidebar := lipgloss.JoinVertical(lipgloss.Right,
		sidebarStyle.Render(t.task.ProjectId),
		sidebarStyle.Render(t.task.Due.Date),
		sidebarStyle.Render(fmt.Sprint(t.task.Priority)),
		sidebarStyle.Render(t.task.Labels...))
	return lipgloss.JoinVertical(lipgloss.Center, t.flexBox.Render(), content, sidebar)
}

func (t *taskViewModel) SetTask(task *types.Task) {
	t.task = task
}

type projectMsg struct {
	projects []types.Project
}

type taskViewMsg struct {
	task types.Task
}

func InitializeModel() *model {
	m := &model{
		cache:           cache.NewInMemoryCache(&http.Client{Timeout: 10 * time.Second}),
		choice:          0,
		projectCursor:   0,
		tasks:           []types.Task{},
		projects:        []types.Project{},
		selectedProject: types.Project{},
		taskList:        list.New([]list.Item{}, list.NewDefaultDelegate(), 20, 20),
		view:            ProjectsView,
		projectModel:    buildTable(buildProjectColumns(), "Loading"),
		taskModel:       buildTable(buildTaskColumns(), "No Project Selected"),
		help:            help.New(),
		taskView:        initializeTaskViewModel(),
	}
	m.projectModel.Focus()
	m.taskModel.Blur()
	return m
}

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatal("Could not open log file", err)
	}
	defer f.Close()
	model := InitializeModel()
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	log.Printf("Init")
	return getProjects(m)
}

func getProjects(m model) tea.Cmd {
	return func() tea.Msg {
		return projectMsg{m.cache.GetProjects()}
	}
}

type TaskMsg struct {
	tasks []types.Task
}

func Tasks(m model) tea.Cmd {
	return func() tea.Msg {
		log.Println("Getting tasks from cache")
		log.Println("Returning a task msg")
		return TaskMsg{m.cache.GetTasks(m.selectedProject)}
	}
}

func Task(m model) tea.Cmd {
	return func() tea.Msg {
		log.Println("Task() - Fetching tasks")
		return taskViewMsg{m.cache.GetTasks(m.selectedProject)[m.taskList.Index()]}
	}
}

func (m model) getTaskView() string {
	m.taskModel.SetHeight(screenHeight - 15)
	if m.fetchingTasks {
		m.taskModel.SetRows([]table.Row{{"Fetching tasks..."}})
	}
	return lipgloss.JoinVertical(lipgloss.Center, m.taskModel.View())
}

func (m model) projectView() string {
	m.projectModel.SetHeight(screenHeight - 15)
	return lipgloss.JoinVertical(lipgloss.Center, m.projectModel.View())
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	log.Println("In Update()")
	log.Printf("project model rows are %v", len(m.projectModel.Rows()))
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		screenWidth = msg.Width
		screenHeight = msg.Height
		m.taskView.UpdateWindowSize(screenWidth, screenHeight)
	case projectMsg:
		log.Println("Update got a projects message")
		m.projects = msg.projects
		m.projectModel.SetRows(toRows(m.projects))
		m.setSelectedProject()
		m.projectModel.Focus()
		return m, nil

	case tea.KeyMsg:
		log.Println("Update got a key message")
		switch {
		case key.Matches(msg, Keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, Keys.Up):
			if m.projectModel.Focused() {
				m.projectModel.MoveUp(1)
				m.setSelectedProject()
			}
			if m.taskModel.Focused() {
				m.taskModel.MoveUp(1)
			}
		case key.Matches(msg, Keys.Down):
			if m.projectModel.Focused() {
				m.projectModel.MoveDown(1)
				m.setSelectedProject()
			}
			if m.taskModel.Focused() {
				m.taskModel.MoveDown(1)
			}
		case key.Matches(msg, Keys.Left):
			if m.taskModel.Focused() {
				m.taskModel.Blur()
				m.projectModel.Focus()
			}
		case key.Matches(msg, Keys.Right):
			if m.projectModel.Focused() {
				m.projectModel.Blur()
				m.taskModel.Focus()
			}
		case key.Matches(msg, Keys.Enter):
			if m.projectModel.Focused() {
				m.fetchingTasks = true
				m.projectModel.Blur()
				m.taskModel.Focus()
				return m, Tasks(m)
			} else if m.taskModel.Focused() {
				return m, Task(m)
			}
		case key.Matches(msg, Keys.Top):
			log.Println("Key top msg")
			if m.projectModel.Focused() {
				cursorPosition := m.projectModel.Cursor()
				m.projectModel.MoveUp(cursorPosition)
				m.setSelectedProject()
			}
		case key.Matches(msg, Keys.Bottom):
			log.Println("key bottom message")
			if m.projectModel.Focused() {
				rowCount := len(m.projectModel.Rows())
				cursorPosition := m.projectModel.Cursor()
				m.projectModel.MoveDown(rowCount - cursorPosition)
				m.setSelectedProject()
			}
		}
	case TaskMsg:
		log.Println("Got task msg")
		m.fetchingTasks = false
		m.tasks = msg.tasks
		m.taskModel.SetRows(tasksToRows(msg.tasks))
		m.taskModel.Focus()
	case taskViewMsg:
		m.viewingTask = true
		m.taskView.SetTask(&msg.task)
		m.selectedTask = msg.task
	}
	var cmd tea.Cmd
	m.taskList, cmd = m.taskList.Update(message)
	return m, cmd
}

type item struct {
	createdDate string
	desc        string
}

func (i item) Title() string       { return i.createdDate }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.createdDate }

func (m model) View() string {
	log.Println("In View()")
	log.Printf("project model rows are %v", len(m.projectModel.Rows()))
	if m.viewingTask {
		return m.taskView.View()
	}
	var projectView, taskView string
	if m.projectModel.Focused() {
		projectView = selectedBoxStyle.Render(m.projectView())
		taskView = unselectedBoxStyle.Render(m.getTaskView())
	} else if m.taskModel.Focused() {
		projectView = unselectedBoxStyle.Render(m.projectView())
		taskView = selectedBoxStyle.Render(m.getTaskView())
	}

	viewArr := []string{projectView}
	viewArr = append(viewArr, taskView)

	tables := lipgloss.JoinHorizontal(lipgloss.Center, viewArr...)
	tables += lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("\n%s\n", m.selectedTask.Content),
		m.help.View(Keys))
	return tables
}

func projectColor(project types.Project) string {
	return lipgloss.NewStyle().Background(colors[project.Color]).Render("  ")
}

func buildTable(columns []table.Column, defaultMessage string) table.Model {
	return table.New(
		table.WithHeight(projectTableHeight),
		table.WithColumns(columns),
		table.WithRows([]table.Row{{defaultMessage}}),
	)
}

func buildProjectColumns() []table.Column {
	return []table.Column{
		{Title: "Projects", Width: 25},
	}
}

func buildTaskColumns() []table.Column {
	return []table.Column{
		{Title: "Task", Width: 35},
		{Title: "Created At", Width: 10},
	}
}

func RenderTask(task types.Task) string {
	out := ""
	out += "Creator ID: " + task.CreatorId + "\n"
	out += "Created At: " + task.CreatedAt + "\n"
	out += "Assignee ID: " + task.AssigneeId + "\n"
	out += "Assigner ID: " + task.AssignerId + "\n"
	out += "Comment Count: " + strconv.Itoa(task.CommentCount) + "\n"
	out += fmt.Sprintf("Is Completed? %v\n", task.IsCompleted)
	out += "Content: " + task.Content + "\n"
	out += "Description: " + task.Description + "\n"
	out += "Due: " + task.Due.String + "\n"
	out += " Duration: " + task.Duration + "\n"
	out += "ID: " + task.Id + "\n"
	out += fmt.Sprintf("Labels: %v\n", task.Labels)
	out += "Order: " + strconv.Itoa(task.Order) + "\n"
	out += "Priority: " + strconv.Itoa(task.Priority) + "\n"
	out += "Project ID: " + task.ProjectId + "\n"
	out += "Section ID: " + task.SectionId + "\n"
	out += "Parent ID: " + task.ParentId + "\n"
	out += "Url: " + task.Url + "\n"
	return out
}

func toRows(projects []types.Project) []table.Row {
	log.Printf("projects length is %v", len(projects))
	rows := make([]table.Row, 0, len(projects))
	for _, project := range projects {
		rows = append(rows, []string{project.Name})
	}
	log.Printf("Returning a list of %v rows", len(rows))
	return rows
}

func tasksToRows(tasks []types.Task) []table.Row {
	if len(tasks) == 0 {
		return []table.Row{
			{"No tasks found ✨"},
		}
	}
	rows := make([]table.Row, 0, len(tasks))
	for _, task := range tasks {
		rows = append(rows, []string{task.Content, task.CreatedAt})
	}
	return rows
}

func (m *model) setSelectedProject() {
	cursorPosition := m.projectModel.Cursor()
	m.selectedProject = m.projects[cursorPosition]
}
