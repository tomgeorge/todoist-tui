package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"

	"github.com/tomgeorge/todoist-tui/cmd"
	"github.com/tomgeorge/todoist-tui/types"
)

var testProjects = []types.Project{
	{
		Name:  "Inbox",
		Color: "red",
		Id:    "1",
	},
	{
		Name:  "Work tasks",
		Color: "charcoal",
		Id:    "2",
	},
}

const (
	ApiToken = "57c7e276c2251e2661a79f678020fdd202cdc97b"
)

var testLabels = []types.Label{
	{
		Name:  "Home",
		Color: "berry_red",
	},
	{
		Name:  "Work",
		Color: "violet",
	},
	{
		Name:  "Basketball",
		Color: "blue",
	},
}

var componentLabelStyle = lipgloss.NewStyle().
	Background(lipgloss.AdaptiveColor{Dark: "#eff1f5", Light: "#7287fd"}).
	Foreground(lipgloss.AdaptiveColor{Dark: "#303446", Light: "#eff1f5"}).
	Bold(true).
	MarginBottom(1).
	Padding(0, 1, 0, 1)

var focusedComponentLabelStyle = lipgloss.NewStyle().
	Background(lipgloss.AdaptiveColor{Dark: "#ca9ee6", Light: "#f38ba8"}).
	Foreground(lipgloss.AdaptiveColor{Dark: "#303446", Light: "#eff1f5"}).
	Bold(true).
	MarginBottom(1).
	Padding(0, 1, 0, 1)

var validationStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#41424e")).
	Foreground(lipgloss.Color("#e5c891")).
	MarginBottom(1)

var getTasks bool

func main() {
	err := cmd.NewRootCmd().Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start: %v", err)
		os.Exit(1)
	}
}

// flag.BoolVar(&getTasks, "get-tasks", false, "Get tasks")
//
// flag.Parse()
//
// if getTasks {
// 	// cli := services.NewClient(nil).WithAuthToken(os.Getenv("TODOIST_API_TOKEN"))
// 	// tasks, err := cli.Tasks.GetActiveTasks(context.Background())
// 	// if err != nil {
// 	// 	fmt.Printf("Error getting tasks: %v", err)
// 	// }
// 	// for _, task := range tasks {
// 	// 	fmt.Println(task.Content)
// 	// }
// 	// client := sync.NewClient(nil).WithAuthToken(os.Getenv("TODOIST_API_TOKEN"))
// 	// var body string
// 	// // resp, err := client.Sync([]string{"projects"}, body)
// 	// if err != nil {
// 	// 	fmt.Printf("Error %v response %v", err, resp)
// 	// }
// 	// fmt.Println(body)
// } else {
// 	f, err := tea.LogToFile("debug.log", "debug")
// 	if err != nil {
// 		log.Fatal(err)
// 		os.Exit(1)
// 	}
// 	defer f.Close()
//
// 	cli := sync.NewClient(nil).WithAuthToken(os.Getenv("TODOIST_API_TOKEN"))
//
// 	state, err := cli.Sync(context.Background(), sync.SyncRequest{
// 		SyncToken:     "*",
// 		ResourceTypes: sync.NewResourceTypes("all"),
// 	})
//
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	// item, _ := lo.Find(state.Items, func(i *types.Item) bool {
// 	// 	return i.Content == "My testing task"
// 	// })
// 	// parentProject, _ := lo.Find(state.Projects, func(p *types.Project) bool {
// 	// 	return p.Id == item.ProjectId
// 	// })
// 	// labelNonPtr := []types.Label{}
// 	// for _, label := range labels {
// 	// 	labelNonPtr = append(labelNonPtr, *label)
// 	// }
// 	// projects := state.Projects
// 	// projectsNonPtr := []types.Project{}
// 	// for _, project := range projects {
// 	// 	projectsNonPtr = append(projectsNonPtr, *project)
// 	// }
//
// 	// focusedStyle := lipgloss.NewStyle().MarginBottom(1).Border(lipgloss.NormalBorder(), true)
//
// 	// model := task_create.New(
// 	// 	task_create.WithPossibleLabels(labelNonPtr),
// 	// 	task_create.WithProjects(projectsNonPtr),
// 	// 	task_create.WithPriorities(types.Priorities),
// 	// 	task_create.WithFocusedStyle(focusedStyle),
// 	// 	task_create.WithTask(item),
// 	// 	task_create.WithParentProject(parentProject),
// 	// 	task_create.WithLabels(state.Labels),
// 	// )
//
// 	tuiApp, _ := lo.Find(state.Projects, func(p *types.Project) bool {
// 		return p.Name == "TUI App"
// 	})
// 	tasks := lo.Filter(state.Items, func(i *types.Item, _ int) bool {
// 		return i.ProjectId == tuiApp.Id
// 	})
// 	projects := state.Projects
// 	labels := state.Labels
// 	model := model.New(tuiApp, tasks, projects, labels)
// 	// projectView := project_view.New(parentProject, tasks)
// 	p := tea.NewProgram(model)
//
// 	if _, err := p.Run(); err != nil {
// 		log.Fatal(err)
// 	}
