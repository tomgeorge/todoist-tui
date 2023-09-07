package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const todoist = "https://api.todoist.com/rest/v2"

type View int

const (
	ProjectsView View = 0
	TasksView    View = 1
)

type model struct {
	choice          int
	cursor          int
	projects        []Project
	selectedProject Project
	tasks           []Task
	view            View
}

type DueDate struct {
	Date        string `json:"date"`
	IsRecurring bool   `json:"is_recurring"`
	Datetime    string `json:"datetime"`
	String      string `json:"string"`
	Timezone    string `json:"timezone"`
}

type Task struct {
	CreatorId    string   `json:"creator_id"`
	CreatedAt    string   `json:"created_at"`
	AssigneeId   string   `json:"assignee_id"`
	AssignerId   string   `json:"assigner_id"`
	CommentCound int      `json:"comment_count"`
	IsCompleted  bool     `json:"is_completed"`
	Content      string   `json:"content"`
	Description  string   `json:"description"`
	Due          DueDate  `json:"due"`
	Duration     string   `json:"duration"`
	Id           string   `json:"id"`
	Labels       []string `json:"labels"`
	Order        int      `json:"order"`
	Priority     int      `json:"priority"`
	ProjectId    string   `json:"project_id"`
	SectionId    string   `json:"section_id"`
	ParentId     string   `json:"parent_id"`
	Url          string   `json:"url"`
}

type Project struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	CommentCount   int    `json:"comment_count"`
	Order          int    `json:"order"`
	Color          string `json:"color"`
	IsShared       bool   `json:"is_shared"`
	IsFavorite     bool   `json:"is_favorite"`
	IsInboxProject bool   `json:"is_inbox_project"`
	IsTeamInbox    bool   `json:"is_team_inbox"`
	ViewStyle      string `json:"view_style"`
	Url            string `json:"url"`
	ParentId       string `json:"parent_id"`
}

type projectMsg struct {
	projects []Project
}

func main() {
	initialModel := model{
		choice:          0,
		cursor:          0,
		tasks:           []Task{},
		projects:        []Project{},
		selectedProject: Project{},
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
	return getProjects
}

func getProjects() tea.Msg {
	c := &http.Client{Timeout: 10 * time.Second}
	url, _ := url.Parse(fmt.Sprintf("%s/projects", todoist))
	res, err := c.Do(&http.Request{
		URL:    url,
		Method: "GET",
		Header: map[string][]string{
			"Authorization": {fmt.Sprintf("Bearer %s", os.Getenv("TODOIST_API_TOKEN"))},
		},
	})
	if err != nil {
		log.Fatalf("An error occured %v", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}
	var projects []Project
	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	jsonErr := json.Unmarshal(body, &projects)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return projectMsg{projects}
}

type TaskMsg struct {
	tasks []Task
}

func Tasks(project Project) tea.Cmd {
	return func() tea.Msg {
		log.Println("Tasks")
		c := &http.Client{Timeout: 10 * time.Second}
		url, _ := url.Parse(fmt.Sprintf("%s/tasks?project_id=%s", todoist, project.Id))
		res, err := c.Do(&http.Request{
			URL:    url,
			Method: "GET",
			Header: map[string][]string{
				"Authorization": {fmt.Sprintf("Bearer %s", os.Getenv("TODOIST_API_TOKEN"))},
			},
		})
		if err != nil {
			log.Fatalf("An error occured %v", err)
		}

		if res.Body != nil {
			defer res.Body.Close()
		}
		var tasks []Task
		body, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			log.Fatal(readErr)
		}

		jsonErr := json.Unmarshal(body, &tasks)

		if jsonErr != nil {
			log.Fatal(jsonErr)
		}
		return TaskMsg{tasks}

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
			m.view++
			return m, Tasks(m.selectedProject)
    case "backspace":
      if m.view == 0 {
        m.view = 0
      }
      m.view--
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
