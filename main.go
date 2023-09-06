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

type model struct {
	choice          int
	cursor          int
	projects        []Project
	selectedProject Project
	tasks           []Task
}

type Task struct{}
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

// func getTasks() tea.Msg {
//
// }

func main() {
  initialModel := model{
    choice: 0,
    cursor: 0,
    tasks: []Task{},
    projects: []Project{},
    selectedProject: Project{},
  }
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
  return getProjects
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case projectMsg:
		m.projects = msg.projects
		return m, nil

	case tea.KeyMsg:
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
		}
	}
	return m, nil
}

func (m model) View() string {

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
}
