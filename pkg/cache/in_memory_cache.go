package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/tomgeorge/todoist-tui/pkg/types"
)

type store struct {
	projects []types.Project
	tasks    map[string][]types.Task
}

type InMemoryCache struct {
	client   *http.Client
	store    *store
}

func NewInMemoryCache(c *http.Client) *InMemoryCache {
	return &InMemoryCache{
		client: c,
		store: &store{
			projects: []types.Project{},
			tasks:    map[string][]types.Task{},
		},
	}
}

func (m *InMemoryCache) GetProjects() []types.Project {
	if len(m.store.projects) == 0 {
		log.Println("Not fetched yet, requesting from todoist")
		url, _ := url.Parse("https://api.todoist.com/rest/v2/projects")
		res, err := m.client.Do(&http.Request{
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
		var projects []types.Project
		body, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			log.Fatal(readErr)
		}

		jsonErr := json.Unmarshal(body, &projects)

		if jsonErr != nil {
			log.Fatal(jsonErr)
		}
		m.store.projects = projects
	}
	return m.store.projects
}

func (m *InMemoryCache) GetTasks(project types.Project) []types.Task {
	if len(m.store.tasks[project.Name]) == 0 {
		log.Println("Haven't fetched yet, going to the API")
		url, _ := url.Parse(fmt.Sprintf("https://api.todoist.com/rest/v2/tasks?project_id=%s", project.Id))
		res, err := m.client.Do(&http.Request{
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
		var tasks []types.Task
		body, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			log.Fatal(readErr)
		}

		jsonErr := json.Unmarshal(body, &tasks)

		if jsonErr != nil {
			log.Fatal(jsonErr)
		}
		m.store.tasks[project.Name] = tasks
    log.Println("InMemoryCache.GetTasks - I got some tasks back from the API")
	}
	return m.store.tasks[project.Name]
}
