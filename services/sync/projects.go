package sync

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/tomgeorge/todoist-tui/types"
)

type AddProjectArgs struct {
	// The name of the project
	Name string `json:"name"`
	// The color of the project icon. Refer to the name columnn in the colors
	// guide
	Color string `json:"color,omitempty"`
	// The ID of the parent project
	ParentId string `json:"parent_id,omitempty"`
	// The order of the project. Defines the position of the project among all the
	// projects with teh same parent id
	ChildOrder int `json:"child_order,omitempty"`
	// Whether the project is a favorite
	IsFavorite bool `json:"is_favorite,omitempty"`
	// Determines the way the project is displayed within the Todoist clients
	ViewStyle string `json:"view_style,omitempty"`
}

func (c *Client) AddProject(args AddProjectArgs) (*types.Project, error) {
	command := Command{
		Type:   "project_add",
		TempId: uuid.NewString(),
		Uuid:   uuid.NewString(),
		Args:   args,
	}

	syncRequest := SyncRequest{
		SyncToken:     c.syncToken,
		ResourceTypes: []string{"projects"},
		Commands:      []Command{command},
	}

	resp, err := c.Sync(context.Background(), syncRequest)
	if err != nil {
		return nil, err
	}
	projectId := resp.TempIdMapping[command.TempId]
	project, _ := lo.Find(resp.Projects, func(p *types.Project) bool {
		return p.Id == projectId
	})
	return project, nil
}
