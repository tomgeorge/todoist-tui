package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tomgeorge/todoist-tui/pkg/cache"
	"github.com/tomgeorge/todoist-tui/pkg/types"

	"github.com/charmbracelet/bubbles/list"
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
}

type projectMsg struct {
	projects []types.Project
}

type taskViewMsg struct {
	task types.Task
}

func main() {
	initialModel := model{
		cache:           cache.NewInMemoryCache(&http.Client{Timeout: 10 * time.Second}),
		choice:          0,
		projectCursor:   0,
		tasks:           []types.Task{},
		projects:        []types.Project{},
		selectedProject: types.Project{},
		taskList:        list.New([]list.Item{}, list.NewDefaultDelegate(), 20, 20),
		view:            ProjectsView,
	}
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatal("Could not open log file", err)
	}
	defer f.Close()
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
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
		return TaskMsg{m.cache.GetTasks(m.selectedProject)}
	}
}

func Task(m model) tea.Cmd {
	return func() tea.Msg {
		return taskViewMsg{m.cache.GetTasks(m.selectedProject)[m.taskList.Index()]}
	}
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	log.Println("In Update()")
	switch msg := message.(type) {
	case projectMsg:
		log.Println("Update got a projects message")
		m.projects = msg.projects
		return m, nil

	case tea.KeyMsg:
		log.Println("Update got a key message")
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.projectCursor > 0 {
				m.projectCursor--
			}
		case "down", "j":
			if m.projectCursor < len(m.projects)-1 {
				m.projectCursor++
			}
		case "enter":
			m.selectedProject = m.projects[m.projectCursor]
			m.tasks = []types.Task{}
			if m.view == ProjectsView {
				m.view++
				return m, Tasks(m)
			} else if m.view == TasksView {
				m.view++
				return m, Task(m)
			}
		case "backspace":
			if m.view == 0 {
				m.view = 0
			} else {
				m.view--
			}
		}
	case TaskMsg:
		var items = []list.Item{}
		for _, task := range msg.tasks {
			log.Println("Appending to task list")
			items = append(items, item{
				createdDate: task.CreatedAt,
				desc:        task.Content,
			})
		}
		m.taskList.Title = "Tasks"
		m.taskList.SetItems(items)
	case taskViewMsg:
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
	switch m.view {
	case ProjectsView:
		body := "Todoist Project List\n\n"
		for i, project := range m.projects {
			cursor := " "
			if m.projectCursor == i {
				cursor = ">"
			}
			body += fmt.Sprintf("%s [%s]\n", cursor, project.Name)
		}
		body += fmt.Sprintf("Selected Project: %s", m.selectedProject.Name)
		body += "\nPress q to quit"
		return body
	case TasksView:
		body := fmt.Sprintf("Tasks for %s", m.selectedProject.Name)
		body += fmt.Sprintf("\n\nTasks For Project %s\n\n", m.selectedProject.Name)
		if len(m.taskList.Items()) == 0 {
			body += "No tasks found ✨"
		}
		body += docStyle.Render(m.taskList.View())
		return body
	case TaskWindow:
		return fmt.Sprintf("%v", m.selectedTask)
	}
	return "Loading"
}
