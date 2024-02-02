package types

import "github.com/charmbracelet/lipgloss"

type ViewStyle string

const (
	BoardStyle ViewStyle = "board"
	ListStyle  ViewStyle = "list"
)

// This is from the rest api and includes the CommentCount property
type restProject struct {
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	Color          string    `json:"color"`
	ParentId       string    `json:"parent_id"`
	Order          int       `json:"order"`
	CommentCount   int       `json:"comment_count"`
	IsShared       bool      `json:"is_shared"`
	IsFavorite     bool      `jsin:"is_favorite"`
	IsInboxProject bool      `json:"is_inbox_project"`
	IsTeamInbox    bool      `json:"is_team_inbox"`
	ViewStyle      ViewStyle `json:"view_style"`
	Url            string    `json:"url"`
}

// A Todoist project
type Project struct {
	// The ID of the note
	Id string `json:"id"`
	// The name of the project
	Name string `json:"name"`
	// The color of the project icon. Refer to the name column in the todoist Colors
	// guide for more info
	Color string `json:"color"`
	// The ID of the parent project. Set to nil for root projects.
	ParentId string `json:"parent_id"`
	// The order of the project. Defines the position of the project among all the
	// projects with the same ParentId
	ChildOrder int `json:"child_order"`
	// Whether the project's sub-projects are collapsed
	Collapsed bool `json:"collapsed"`
	// Whether the project is shared
	Shared bool `json:"shared"`
	// Whether the project is marked as deleted
	IsDeleted bool `json:"is_deleted"`
	// Whether the project is marked as archived
	IsArchived bool `json:"is_archived"`
	// Whether the project is a favorite
	IsFavorite bool `json:"is_favorite"`
	// Identifier to find the match between different copies of shared projects.
	// When you share a project, its copy has a different ID for your
	// collaborators. To find a project in a different account that matches yours,
	// you can use SyncId. For non-shared projects the attribute is set to nil.
	SyncId string `json:"sync_id"`
	// Whether this project is Inbox (true or otherwise not set)
	InboxProject bool `json:"inbox_project"`
	// Whether the project is a Team inbox (true or otherwise not set)
	TeamInbox bool `json:"team_inbox"`
	// A string value (either list or board). Determines the way the project is
	// displayed within the todoist clients
	ViewStyle string `json:"view_style"`
}

func (p Project) Render() string {
	return p.Name
}

func (p Project) Style() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(Colors[p.Color])).
		MarginRight(1).
		MarginTop(1)
}

func (p Project) GetFormData() interface{} {
	return p.Id
}
