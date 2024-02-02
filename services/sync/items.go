package sync

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"
	"github.com/tomgeorge/todoist-tui/types"
)

type AddItemArgs struct {
	ProjectId string `json:"project_id,omitempty"`
	// The text of the task. This value may contain markdown-formatted text
	// and hyperlinks. Details on markdown support can be found in the
	// Text Formatting article in the todoist Help Center.
	Content string `json:"content"`
	// A description for the task. This value may contain markdown-formatted text
	// and hyperlinks. Details on markdown support can be found in the Text
	// Formatting article in the todoist Help Center
	Description string `json:"description,omitempty"`
	// The due date of the task
	Due *types.DueDate `json:"due,omitempty"`
	// The priority of the task (a number between 1 and 4, 4 for very urgent and
	// 1 for natural). Note that very urgent is the priority 1 on clients, so p1
	// will return 4 in the API
	Priority int `json:"priority,omitempty"`
	// The ID of the parent task. Set to nil for root tasks
	ParentId string `json:"parent_id,omitempty"`
	// The order of the task. Defines the position of the task among all the tasks
	// with the same parent.
	ChildOrder int `json:"child_order,omitempty"`
	// The ID of the parent section. Set to nil for tasks not belonging to
	// a section.
	SectionId any `json:"section_id,omitempty"`
	// The order of the task inside the Today or Next 7 days view (a number, where
	// the smallest value would place the task at the top)
	DayOrder int `json:"day_order,omitempty"`
	// Whether the task's sub-tasks are collapsed
	Collapsed bool `json:"collapsed,omitempty"`
	// The task's labels (may represent personal or shared labels)
	Labels []string `json:"labels,omitempty"`
	// The ID of the user who assigned the task. This makes sense for shared
	// projects only. Accepts any user ID from the list of project collaborators.
	// If this value is unset or invalid, it will automatically be set to the
	// requestor's UID.
	AssignedByUid any `json:"assigned_by_uid,omitempty"`
	// The ID of user who is responsible for accomplishing the current task. This
	// makes sense for shared projects only. Accepts any user ID from the list of
	// project collaborators or null or an empty string to unset.
	ResponsibleUid any `json:"responsible_uid,omitempty"`
	// When this option is enabled, the default reminder will be added to the new
	// item if it has a due date with a time set. See also the auto_reminder user
	// option
	AutoReminder bool `json:"auto_reminder,omitempty"`
	// When this option is enabled, the labels will be parsed from the task
	// content and added to the task. In case the label doesn't exist, a new one
	// will be created.
	AutoParseLabels bool `json:"auto_parse_labels,omitempty"`
	// Object representing a task's duration. Includes a positive integer for the
	// amount of time the task will take, and the unit of time that the amount
	// represents which must be either minute or day. Both the amount and unit
	// must be defined. The object will be nil if the task has no duration.
	Duration *types.ItemDuration `json:"duration,omitempty"`
}

// The documentation says that you don't need to send the sync token/resource
// types when writing, but I think you would want to in this case, or at least
// you would want to then merge this into the sync state

// Add a new task to a project.
func (c *Client) AddTask(context context.Context, args AddItemArgs, opts ...CommandOption) (*types.Item, error) {
	if args.Content == "" {
		return nil, errors.New("cannot create a task with no content")
	}
	opts = append(opts, WithArgs(args))
	command := NewCommand("item_add", opts...)

	syncRequest := SyncRequest{
		SyncToken:     c.syncToken,
		ResourceTypes: NewResourceTypes("items"),
		Commands:      []Command{*command},
	}

	resp, err := c.Sync(context, syncRequest)
	if err != nil {
		return nil, err
	}
	if syncStatus := resp.SyncStatus[command.Uuid]; !syncStatus.Ok {
		return nil, fmt.Errorf("sync error item_add: %s", syncStatus.ErrorCode.Error)
	}
	itemId := resp.TempIdMapping[command.TempId]
	if itemId == "" {
		return nil, fmt.Errorf("could not find item ID mapping for temporary id %s", command.TempId)
	}
	item, found := lo.Find(resp.Items, func(i *types.Item) bool {
		return i.Id == itemId
	})
	if !found {
		return nil, fmt.Errorf("could not find created item with ID %s", itemId)
	}
	return item, nil
}

// See AddItemArgs for docstrings
type UpdateItemArgs struct {
	Id             string              `json:"id"`
	Content        string              `json:"content,omitempty"`
	Description    string              `json:"description,omitempty"`
	Due            *types.DueDate      `json:"due,omitempty"`
	Priority       int                 `json:"priority,omitempty"`
	Collapsed      bool                `json:"collapsed,omitempty"`
	Labels         []string            `json:"labels,omitempty"`
	AssignedByUid  string              `json:"assigned_by_uid,omitempty"`
	ResponsibleUid string              `json:"responsible_uid,omitempty"`
	DayOrder       int                 `json:"day_order,omitempty"`
	Duration       *types.ItemDuration `json:"duration,omitempty"`
}

// Update task attributes. Please note that updating the parent, moving,
// completing, or uncompleting tasks is not supported by this function, more
// specific function have to be used instead
func (c *Client) UpdateTask(context context.Context, args UpdateItemArgs, opts ...CommandOption) (*types.Item, error) {
	if args.Id == "" {
		return nil, errors.New("task ID not specified when trying to update task")
	}
	opts = append(opts, WithArgs(args))
	command := NewCommand("item_update", opts...)

	syncRequest := SyncRequest{
		SyncToken:     c.syncToken,
		ResourceTypes: NewResourceTypes("items"),
		Commands:      []Command{*command},
	}

	resp, err := c.Sync(context, syncRequest)
	if err != nil {
		return nil, err
	}
	if syncStatus := resp.SyncStatus[command.Uuid]; !syncStatus.Ok {
		return nil, fmt.Errorf("item_update sync error: %s", syncStatus.ErrorCode.Error)
	}
	// Temp ID mapping won't exist because no new item was created
	item, found := lo.Find(resp.Items, func(i *types.Item) bool {
		return i.Id == args.Id
	})
	if !found {
		return nil, fmt.Errorf("could not find updated item with ID %s", args.Id)
	}
	return item, nil
}
