package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tomgeorge/todoist-tui/pkg/cache"
	"github.com/tomgeorge/todoist-tui/pkg/types"

	tea "github.com/charmbracelet/bubbletea"
)

const todoist = "https://api.todoist.com/rest/v2"

type View int

const (
	ProjectsView View = 0
	TasksView    View = 1
)

type model struct {
	cache           cache.Cache
	choice          int
	cursor          int
	projects        []types.Project
	selectedProject types.Project
	tasks           []types.Task
	view            View
}

type projectMsg struct {
	projects []types.Project
}

func main() {
	initialModel := model{
		cache:           cache.NewInMemoryCache(&http.Client{Timeout: 10 * time.Second}),
		choice:          0,
		cursor:          0,
		tasks:           []types.Task{},
		projects:        []types.Project{},
		selectedProject: types.Project{},
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
		return TaskMsg{m.cache.GetTasks(m.selectedProject)}
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
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}
		case "enter":
			m.selectedProject = m.projects[m.cursor]
			m.tasks = []types.Task{}
			m.view++
			return m, Tasks(m)
		case "backspace":
			if m.view == 0 {
				m.view = 0
			} else {
				m.view--
			}
		}
	case TaskMsg:
		m.tasks = msg.tasks
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	log.Println("In View()")
	switch m.view {
	case ProjectsView:
		body := "Todoist Project List\n\n"
		for i, project := range m.projects {
			cursor := " "
			if m.cursor == i {
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
		if len(m.tasks) == 0 {
			body += "No tasks found ✨"
		}
		for _, task := range m.tasks {
			log.Println("Adding a task")
			body += fmt.Sprintf("%s\n", task.Content)
		}
		body += "\nPress q to quit"
		return body
	}
	return "Loading"
}
