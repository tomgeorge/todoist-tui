package types

import (
	"time"
)

// Also known as Comments. Notes is an old term they used and will be amended
// in a future version of the sync API.
//
// Availability of comments functionality is dependent on the current user
// plan. This value is indicated by the comments property of the user plan
// limits object
type Note struct {
	// The ID of the note
	Id string `json:"id"`
	// The ID of the user that posted the note
	PostedUid string `json:"posted_uid"`
	// The item which the note is part of
	ItemId string `json:"item_id"`
	// The content of the note. This value may contain markdown-formatted text and
	// hyperlinks. Details on markdown support can be found in the Text Formatting
	// article in the help center
	Content string `json:"content"`
	// A file attached to the note
	FileAttachment FileAttachment `json:"file_attachment"`
	// A list of user IDS to notify
	UidsToNotify []string `json:"uids_to_notify"`
	//Whether the note is marked as deleted
	IsDeleted bool `json:"is_deleted"`
	// The date when the note was posted
	PostedAt time.Time `json:"posted_at"`
	// A list of emoji reactions and corresponding user IDs.
	Reactions map[string]string `json:"reactions"`
}

// A comment, but for a project
type ProjectNote struct {
	// The ID of the note
	Id string `json:"id"`
	// The ID of the user that posted the note
	PostedUid string `json:"posted_uid"`
	// The item which the note is part of
	ProjectId string `json:"project_id"`
	// The content of the note. This value may contain markdown-formatted text and
	// hyperlinks. Details on markdown support can be found in the Text Formatting
	// article in the help center
	Content string `json:"content"`
	// A file attached to the note
	FileAttachment FileAttachment `json:"file_attachment"`
	// A list of user IDS to notify
	UidsToNotify []string `json:"uids_to_notify"`
	//Whether the note is marked as deleted
	IsDeleted bool `json:"is_deleted"`
	// The date when the note was posted
	PostedAt time.Time `json:"posted_at"`
	// A list of emoji reactions and corresponding user IDs.
	Reactions map[string]string `json:"reactions"`
}
