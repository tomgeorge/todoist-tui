package types

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
