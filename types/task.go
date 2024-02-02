package types

import "time"

type CreateTaskRequest struct {
	Content      string   `json:"content"`
	Description  string   `json:"description,omitempty"`
	ProjectId    string   `json:"project_id,omitempty"`
	SectionId    string   `json:"section_id,omitempty"`
	ParentId     string   `json:"parent_id,omitempty"`
	Order        int      `json:"order,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	Priority     int      `json:"priority,omitempty"`
	DueString    string   `json:"due_string,omitempty"`
	DueDate      string   `json:"due_date,omitempty"`
	DueDateTime  string   `json:"due_datetime,omitempty"`
	DueLang      string   `json:"due_lang,omitempty"`
	AssigneeId   string   `json:"assignee_id,omitempty"`
	Duration     int      `json:"duration,omitempty"`
	DurationUnit string   `json:"duration_unit,omitempty"`
}

type ItemDuration struct {
	Amount int    `json:"amount,omitempty"`
	Unit   string `json:"unit,omitempty"`
}

func (i *Item) Title() string       { return i.Content }
func (i *Item) Description() string { return i.TaskDescription }
func (i *Item) FilterValue() string { return i.Content }

// This is also known as a task
// I think I might want to just use the REST API to create tasks so I don't have
// to mess around with due dates
type Item struct {
	// The ID of the task
	Id string `json:"id,omitempty"`
	// The owner of the task
	UserID string `json:"user_id,omitempty"`
	// The ID of the parent project
	ProjectId string `json:"project_id,omitempty"`
	// The text of the task. This value may contain markdown-formatted text
	// and hyperlinks. Details on markdown support can be found in the
	// Text Formatting article in the todoist Help Center.
	Content string `json:"content"`
	// A description for the task. This value may contain markdown-formatted text
	// and hyperlinks. Details on markdown support can be found in the Text
	// Formatting article in the todoist Help Center
	TaskDescription string `json:"description,omitempty"`
	// The due date of the task
	Due *DueDate `json:"due,omitempty"`
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
	// The ID of the user who created the task. This makes sense for shared
	// projects only. Accepts any user ID from the list of project collaborators.
	// If this value is unset or invalid, it will automatically be set to the
	// requestor's UID
	AddedByUid any `json:"added_by_uid,omitempty"`
	// The ID of the user who assigned the task. This makes sense for shared
	// projects only. Accepts any user ID from the list of project collaborators.
	// If this value is unset or invalid, it will automatically be set to the
	// requestor's UID.
	AssignedByUid any `json:"assigned_by_uid,omitempty"`
	// The ID of user who is responsible for accomplishing the current task. This
	// makes sense for shared projects only. Accepts any user ID from the list of
	// project collaborators or null or an empty string to unset.
	ResponsibleUid any `json:"responsible_uid,omitempty"`
	// Whether this task is marked as completed
	Checked bool `json:"checked,omitempty"`
	// Whether the task is marked as deleted
	IsDeleted bool `json:"is_deleted,omitempty"`
	// Identifier to find the match between tasks in shared projects of different
	// collaborators. When you share a task, its copy has a different ID in the
	// projects of your collaborators. To find a task in another account that
	// matches yours, you can use this attribute. For non-shared tasks, the
	// attribute is nil.
	SyncId any `json:"sync_id,omitempty"`
	// The date when the task was completed, or nil if not completed
	CompletedAt time.Time `json:"completed_at,omitempty"`
	// The date when the task was created
	AddedAt time.Time `json:"added_at,omitempty"`
	// Object representing a task's duration. Includes a positive integer for the
	// amount of time the task will take, and the unit of time that the amount
	// represents which must be either minute or day. Both the amount and unit
	// must be defined. The object will be nil if the task has no duration.
	Duration ItemDuration `json:"duration,omitempty"`
	// UNDOCUMENTED This seems to be always set to the beginning of the epoch e.g.
	// 1970-01-01T00:00:00Z
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// UNDOCUMENTED. This seems to be a new version of task ID
	V2ID string `json:"v2_id,omitempty"`
	// UNDOCUMENTED. Seems to be a new version of the parent ID
	V2ParentID any `json:"v2_parent_id,omitempty"`
	// UNDOCUMENTED. Seems to be a new version of the project ID
	V2ProjectID string `json:"v2_project_id,omitempty"`
	// UNDOCUMENTED. Seems to be a new version of the section ID
	V2SectionID any `json:"v2_section_id,omitempty"`
}
