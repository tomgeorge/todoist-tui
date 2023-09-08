package cache

import "github.com/tomgeorge/todoist-tui/pkg/types"


type Cache interface {
  GetProjects() []types.Project
  GetTasks(project types.Project) []types.Task
}
