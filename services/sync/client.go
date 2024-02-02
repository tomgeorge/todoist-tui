package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
	"github.com/tomgeorge/todoist-tui/types"
)

const (
	Version          = "v0.0.1"
	defaultBaseURL   = "https://api.todoist.com/sync/v9/"
	defaultUserAgent = "todoist-tui" + "/" + Version
)

var errNonNilContext = errors.New("context must not be nil")

type Client struct {
	client    *http.Client
	BaseURL   *url.URL
	UserAgent string
	syncToken string
}

type RequestOption func(req *http.Request)

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	httpClient2 := *httpClient
	c := &Client{client: &httpClient2}
	c.initialize()
	return c
}

func (c *Client) NewRequest(method, urlStr string, body io.Reader, opts ...RequestOption) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	if ctx == nil {
		return nil, errNonNilContext
	}
	req = req.WithContext(ctx)
	resp, err := c.client.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return nil, err
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return resp, fmt.Errorf("Request to %s returned %s", resp.Request.URL, resp.Status)
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}

func (c *Client) FullSync(ctx context.Context) (*SyncResponse, error) {
	return c.Sync(ctx, SyncRequest{SyncToken: "*", ResourceTypes: []string{"all"}})
}

// WithAuthToken returns a copy of the client configured to use the provided token for the Authorization header.
func (c *Client) WithAuthToken(token string) *Client {
	c2 := c.copy()
	defer c2.initialize()
	transport := c2.client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	c2.client.Transport = roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			req = req.Clone(req.Context())
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			return transport.RoundTrip(req)
		},
	)
	return c2
}

type ErrorResponse struct {
	StatusCode int
	Status     string
}

func CheckResponse(r *http.Response) error {
	if code := r.StatusCode; code >= 200 && code <= 299 {
		return nil
	}
	return fmt.Errorf("%s request %s: %s", r.Request.Method, r.Request.URL.String(), r.Status)
}

func (c *Client) copy() *Client {
	// can't use *c here because that would copy mutexes by value.
	clone := Client{
		client:    c.client,
		UserAgent: c.UserAgent,
		BaseURL:   c.BaseURL,
	}
	if clone.client == nil {
		clone.client = &http.Client{}
	}
	return &clone
}

// initialize sets default values and initializes services
func (c *Client) initialize() {
	if c.client == nil {
		c.client = &http.Client{}
	}

	if c.BaseURL == nil {
		c.BaseURL, _ = url.Parse(defaultBaseURL)
	}

	if c.UserAgent == "" {
		c.UserAgent = defaultUserAgent
	}

	if c.syncToken == "" {
		c.syncToken = "*"
	}
}

// roundTripperFunc creates a RoundTripper (transport)
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

type ResourceTypes []string

func NewResourceTypes(types ...string) ResourceTypes {
	return types
}

type SyncRequest struct {
	SyncToken     string        `url:"sync_token,omitempty"`
	ResourceTypes ResourceTypes `url:"resource_types,omitempty"`
	Commands      CommandList   `url:"commands,omitempty"`
}

type Command struct {
	Type   string      `json:"type" url:"type"`
	TempId string      `json:"temp_id" url:"temp_id"`
	Uuid   string      `json:"uuid" url:"uuid"`
	Args   interface{} `json:"args" url:"args"`
}

type CommandOption func(c *Command)

func NewCommand(commandType string, opts ...CommandOption) *Command {
	command := &Command{
		Type:   commandType,
		Uuid:   uuid.NewString(),
		TempId: uuid.NewString(),
		Args:   nil,
	}
	for _, opt := range opts {
		opt(command)
	}
	return command
}

func WithType(commandType string) CommandOption {
	return func(c *Command) {
		c.Type = commandType
	}
}

func WithTempId(tempId string) CommandOption {
	return func(c *Command) {
		c.TempId = tempId
	}
}

func WithUuid(uuid string) CommandOption {
	return func(c *Command) {
		c.Uuid = uuid
	}
}

func WithArgs(args interface{}) CommandOption {
	return func(c *Command) {
		c.Args = args
	}
}

func (r ResourceTypes) EncodeValues(key string, v *url.Values) error {
	if len(r) == 0 {
		return errors.New("resource_types must not be empty")
	}
	buf, err := json.Marshal(r)
	if err != nil {
		return err
	}
	v.Add("resource_types", string(buf))
	return nil
}

type CommandList []Command

func (c CommandList) EncodeValues(key string, v *url.Values) error {
	buf, err := json.Marshal(c)
	if err != nil {
		return err
	}
	v.Add("commands", string(buf))
	return nil
}

// Struct defined to handle Todoist sending back either "ok" or an object in the
// sync status
type OperationResult struct {
	Ok        bool
	ErrorCode SyncError
}

type SyncError struct {
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

func (o *OperationResult) UnmarshalJSON(data []byte) error {
	if string(data) == `"ok"` {
		o.Ok = true
		return nil
	}
	if data[0] == '{' && data[len(data)-1] == '}' {
		o.Ok = false
		return json.Unmarshal(data, &o.ErrorCode)
	}
	return fmt.Errorf("Failed to unmarshal sync status to operational result")
}

type SyncOption func(syncRequest SyncRequest)

func (c *Client) Sync(ctx context.Context, syncRequest SyncRequest, opt ...SyncOption) (*SyncResponse, error) {
	bodyParams, _ := query.Values(syncRequest)
	body := strings.NewReader(bodyParams.Encode())
	req, err := c.NewRequest("POST", "sync", body)
	if err != nil {
		return nil, err
	}

	var syncResponse *SyncResponse
	_, err = c.Do(ctx, req, &syncResponse)
	if err != nil {
		return nil, err
	}
	// Update the sync token
	c.syncToken = syncResponse.SyncToken
	return syncResponse, err
}

type CollaboratorState struct {
	// The shared project ID of the user.
	ProjectId string `json:"project_id"`
	//The user ID of the collaborator.
	UserId string `json:"user_id"`
	//The status of the collaborator state, either active or invited
	State string `json:"state"`
	// Set to true when the collaborator leaves the shared project
	IsDeleted bool `json:"is_deleted"`
}

type Collaborator struct {
	// The user ID of the collaborator.
	Id string `json:"id"`
	// The email of the collaborator.
	Email string `json:"email"`
	// The full name of the collaborator.
	FullName string `json:"full_name"`
	// The timezone of the collaborator.
	Timezone string `json:"timezone"`
	// The image ID for the collaborator's avatar, which can be used to get an avatar from a specific URL. Specifically the https://dcff1xvirvpfp.cloudfront.net/<image_id>_big.jpg can be used for a big (195x195 pixels) avatar, https://dcff1xvirvpfp.cloudfront.net/<image_id>_medium.jpg for a medium (60x60 pixels) avatar, and https://dcff1xvirvpfp.cloudfront.net/<image_id>_small.jpg for a small (35x35 pixels) avatar.
	ImageId string `json:"image_id"`
}

type CompletedInfo struct {
	// The ID of the project containing the completed items or archived sections
	ProjectId string `json:"project_id"`
	// This comes back from the API too
	V2ProjectId string `json:"v2_project_id"`
	// The ID of the section containing the completed items
	SectionId string `json:"section_id"`
	// This comes back from the API too
	V2SectionID string `json:"v2_section_id,omitempty"`
	// The ID of the item containing completed child items
	ItemId string `json:"item_id"`
	// I *think* This comes back from the API too
	V2ItemID string `json:"v2_item_id,omitempty"`
	// The number of completed items within the project
	CompletedItems int `json:"completed_items"`
	// The number of archived sections within the project
	ArchivedSections int `json:"archived_sections"`
}

type Filter struct {
	// The ID of the filter.
	Id string `json:"id"`
	// The name of the filter.
	Name string `json:"name"`
	// The query to search for. Examples of searches can be found in the Todoist help page.
	Query string `json:"query"`
	// The color of the filter icon. Refer to the name column in the todoist
	// colors guide for more info
	Color string `json:"color"`
	// Filter’s order in the filter list (where the smallest value should place the filter at the top).
	ItemOrder int `json:"item_order"`
	// Whether the filter is marked as deleted
	IsDeleted bool `json:"is_deleted"`
	// Whether the filter is a favorite
	IsFavorite bool `json:"is_favorite"`
}

type Folder struct {
	// The ID of the folder
	Folder string `json:"id"`
	// The name of the folder
	Name string `json:"name"`
	// The workspace ID of the folder
	WorkspaceId string `json:"workspace_id"`
	// Whether the filter is marked for delete
	IsDeleted bool `json:"is_deleted"`
}

// A file attachment is represented as a JSON object.
// The file attachment may point to a document previously uploaded by the
// uploads/add API call, or by any external resource.
//
// TODO see https://developer.todoist.com/sync/v9/#file-attachments, it looks
// like there are some extra properties for images and audio files
type FileAttachment struct {
	// The name of the file
	FileName string `json:"file_name"`
	// The size of the file in bytes
	FileSize int `json:"file_size"`
	// MIME type
	FileType string `json:"file_type"`
	// The URL where the file is located. Todoist doesn't cache the remote content
	// on their servers, nor does it stream or expose files directly from
	// third-party resources. Avoid providing links to non HTTPS resources, as
	// exposing them in todoist may issue a browser warning
	FileUrl string `json:"file_url"`
	// Upload completion state (e.g. pending, completed)
	UploadState string `json:"upload_state"`
}

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

// Reminder for a task. Availabilty is dependent on the current user plan,
// indicated by reminders, max_reminders_time, and max_reminders_location in the
// user plan limits object
type Reminder struct {
	// The ID of the reminder
	Id string `json:"id"`
	// The user ID which should be notified of the reminder, typically the current
	// user ID creating the reminder
	NotifyUid string `json:"notify_uid"`
	// The item ID for which the reminder is about
	ItemId string `json:"item_id"`
	// The type of the reminder:
	// 1. relative: for a time-based reminder specified in minutes from now
	// 2. absolute: for a time-based reminder with a specific time and date in the
	// future
	// 3. location: for a location-based-reminder
	Type string `json:"type"`
	// The due date of the reminder. See DueDate. Note that reminders only support
	// due dates with time, since full-day reminders don't make sense
	// TODO: operations on this should take the above into account
	Due *types.DueDate `json:"due"`
	// The relative time in minutes before the due date of the item, in which the
	// reminder should be triggered. Note that the item should have a due date
	// with time set in order to add a relative reminder
	MmOffset int `json:"mm_offset"`
	// An alias name for the location
	LocationName string `json:"name"`
	// The location latitude
	LocationLatitude string `json:"loc_lat"`
	// The location longitude
	LocationLongitude string `json:"loc_long"`
	// What should trigger the reminder: on_enter for entering a location, or
	// on_leave for leaving the location
	LocationTrigger string `json:"loc_trigger"`
	// The radius around the location that is still considered as part of the
	// location (in meters)
	Radius int `json:"radius"`
	// Whether the reminder is marked as deleted
	IsDeleted bool `json:"is_deleted"`
}

type Section struct {
	// The ID of the section
	Id string `json:"id"`
	// The name of the section
	Name string `json:"name"`
	// Project that the section resides in
	ProjectId string `json:"project_id"`
	// The order of the section. Defines the position of the section among all the
	// sections in the project
	SectionOrder int `json:"section_order"`
	// Whether the section's tasks are collapsed
	Collapsed bool `json:"collapsed"`
	// A special ID for shared sections. Used internally and can be ignored
	SyncId string `json:"sync_id"`
	// Whether the section is marked as deleted
	IsDeleted bool `json:"is_deleted"`
	// Whether the section is marked as archived
	IsArchived bool `json:"is_archived"`
	// The date when the section was archived
	ArchivedAt time.Time `json:"archived_at"`
	// The date when the section was created
	AddedAt time.Time `json:"added_at"`
}

type KarmaTrendDirection string

const (
	KarmaTrendUp   KarmaTrendDirection = "up"
	KarmaTrendDown KarmaTrendDirection = "down"
)

type DaysItem struct {
	Date  time.Time `json:"date"`
	Items []struct {
		Completed int    `json:"completed"`
		ProjectId string `json:"id"`
	} `json:"items"`
	TotalCompleted int `json:"total_completed"`
}

type WeekItem struct {
	// This is a date range, like "2014-11-03\/2014-11-09"
	Date  string `json:"date"`
	Items []struct {
		Completed int    `json:"completed"`
		ProjectId string `json:"id"`
	} `json:"items"`
	TotalCompleted int `json:"total_completed"`
}

var KarmaReasons = map[int]string{
	1:  "You added tasks.",
	2:  "You completed tasks.",
	3:  "Usage of advanced features.",
	4:  "You are using Todoist. Thanks!",
	5:  "Signed up for Todoist Beta!",
	6:  "Used Todoist Support section!",
	7:  "For using Todoist Pro - thanks for supporting us!",
	8:  "Getting Started Guide task completed!",
	9:  "Dail Goal reached!",
	10: "Weekly Goal reached!",
	50: "You have task that are over x days overdue",
	52: "Inactive for a long period of time",
}

// UNDOCUMENTED
type GoalStreak struct {
	Count int `json:"count"`
	// UNDOCUMENTED I hope these are times
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// UNDOCUMENTED
type Goals struct {
	KarmaDisabled       int        `json:"karma_disabled"`
	UserId              string     `json:"user_id"`
	MaxWeeklyStreak     GoalStreak `json:"max_weekly_streak"`
	IgnoreDays          []int      `json:"ignore_days"`
	VacationMode        int        `json:"vacation_mode"`
	CurrentWeeklyStreak GoalStreak `json:"curent_weekly_streak"`
	CurrentDailyStreak  GoalStreak `json:"current_daily_streak"`
	WeeklyGoal          int        `json:"weekly_goal"`
	MaxDailyStreak      GoalStreak `json:"max_daily_streak"`
}

// User productivity stats, these aren't related to SyncResponse.Stats
type ProductivityStats struct {
	// The karma delta on the last update (what does this mean, though?)
	KarmaLastUpdate float32 `json:"karma_last_update"`
	// Karma trend direction
	KarmaTrend KarmaTrendDirection `json:"karma_trend"`
	// Items completed in the last 7 days. Objects inside items are composed by an
	// id (project ID) and the number of completed tasks for items
	// Meta-comment: their writing is kinda bad sometimes
	DaysItems DaysItem `json:"days_items"`
	// Total completed tasks count
	CompletedCount int `json:"completed_count"`
	// Log of the last karma updates
	KarmaUpdateReasons struct {
		PositiveKarmaReasons []int `json:"positive_karma_reasons"`
		NewKarma             int   `json:"new_karma"`
		NegativeKarma        int   `json:"negative_karma"`
		PositiveKarms        int   `json:"positive_karma"`
		NegativeKarmaReasons []int `json:"negative_karma_reasons"`
		// Idk if this will actually marshall into a time
		Time time.Time `json:"time"`
	} `json:"karma_update_reasons"`
	// Karma score
	Karma float32 `json:"karma"`
	// Items completed in the last 4 weeks
	WeekItems []WeekItem `json:"week_items"`
	// Projects color mapping
	ProjectColors map[string]string
	// Goals definition. The same settings and stats shown in the interface (lol
	// what does this even mean)
	Goals
}

// UNDOCUMENTED
type Stats struct {
	CompletedCount int `json:"completed_count"`
	DaysItems      []struct {
		Date           string `json:"date"`
		TotalCompleted int    `json:"total_completed"`
	} `json:"days_items"`
	WeekItems []struct {
		From           string `json:"from"`
		To             string `json:"to"`
		TotalCompleted int    `json:"total_completed"`
	} `json:"week_items"`
}

// UNDOCUMENTED but kinda also documented
// Used internally for any special features that apply to the user. Current
// special features include whether the user has enabled beta, whether
// dateist_inline_disabled that is inline date parsing support is disabled,
// whether the dateist_lang is set which overrides the date parsing language,
// whether the gold_theme has been awarded to the user,
// whether the user has_push_reminders enabled,
// whether the user has karma_disabled,
// whether the user has karma_vacation mode enabled,
// and whether any special restriction applies to the user.
type Features struct {
	Beta                  int  `json:"beta"`
	DateistInlineDisabled bool `json:"dateist_inline_disabled"`
	DateistLang           any  `json:"dateist_lang"`
	GlobalTeams           bool `json:"global.teams"`
	HasPushReminders      bool `json:"has_push_reminders"`
	KarmaDisabled         bool `json:"karma_disabled"`
	KarmaVacation         bool `json:"karma_vacation"`
	Restriction           int  `json:"restriction"`
}

type PremiumStatusReason string

const (
	PremiumStatusReasonNotPremium            PremiumStatusReason = "not_premium"
	PremiumStatusReasonCurrentPersonalPlan   PremiumStatusReason = "current_personal_plan"
	PremiumStatusReasonActiveBusinessAccount PremiumStatusReason = "active_business_account"
	PremiumStatusReasonTeamsBusinessMember   PremiumStatusReason = "teams_business_member"
)

type UserTimeZoneInfo struct {
	GmtString string `json:"gmt_string"`
	Hours     int    `json:"hours"`
	IsDst     int    `json:"is_dst"`
	Minutes   int    `json:"minutes"`
	Timezone  string `json:"timezone"`
}

// A todoist user
type User struct {
	// The default time in minutes for the automatic reminders sets, whenever
	// a due date has been specified for a task ( TODO I don't understand what this
	// means)
	AutoReminder int `json:"auto_reminder"`
	// The link to a 195x195 pixel image of the user's avatar
	AvatarBig string `json:"avatar_big"`
	// The link to a 60x60 pixel image of the user's avatar
	AvatarMedium string `json:"avatar_medium"`
	// The link to a 640x640 pixel image of the user's avatar
	AvatarS640 string `json:"avatar_s640"`
	// The link to a 35x35 pixel image of the user's avatar
	AvatarSmall string `json:"avatar_small"`
	// The ID of the user's business account
	BusinessAccountID string `json:"business_account_id"`
	// The daily goal number of completed tasks for karma
	DailyGoal int `json:"daily_goal"`
	// 0 - Use DD-MM-YYYY date formats
	// 1 - Use MM-DD-YYYY date formats
	DateFormat int `json:"date_format"`
	// The language expected for date recognition of the user's lang, or null if
	// the user's lang determines it. One of da, de, en, es, fi, fr, it, ja, ko,
	// nl, pl, pt_BR, ru, sv, tr, zh_CN, zh_TW
	DateistLang string `json:"dateist_lang"`
	// Array of integers representing the user's days off, between 1 and 7, where
	// 1 is Monday and 7 is Sunday
	DaysOff []int `json:"days_off"`
	// The user's email address
	Email string `json:"email"`
	// An opaque ID used internally to handle features for the user
	FeatureIdentifier string `json:"feature_identifier"`
	// Internal use, features available to the user, see Features struct
	Features Features `json:"features"`
	// The user's full name
	FullName string `json:"full_name"`
	// Whether the user has a password set on the account.
	HasPassword bool `json:"has_password"`
	// The user's ID
	Id string `json:"id"`
	// The ID of the user's avatar
	ImageId string `json:"image_id"`
	// The ID of the user's Inbox project
	InboxProjectId string `json:"inbox_project_id"`
	// Whether the user is a business account administrator
	IsBizAdmin bool `json:"is_biz_admin"`
	// Whether the user has a Todoist Pro subscription
	IsPremium bool `json:"is_premium"`
	// The registration date of the user on Todoist, may be null for users from
	// the early days, TODO for now we'll treat this as beginning of the epoch
	JoinedAt time.Time
	// The user's karma score (the API says this is an integer but I got a float
	// back)
	Karma float32 `json:"karma"`
	// The user's karma trend
	KarmaTrend KarmaTrendDirection `json:"karma_trend"`
	// The user's language, which can take one of the following values:
	// da, de, en, es, fi, fr, it, ja, ko, nl, pl, pt_BR, ru, sv, tr, zh_CN, zh_TW.
	Lang string `json:"lang"`
	// The next day of the week that tasks will be posponed to (between 1 and 7,
	// where 1 is Monday and 7 is Sunday)
	NextWeek int `json:"next_week"`
	// Outlines why a user is premium
	PremiumStatusReason PremiumStatusReason `json:"premium_status"`
	// The date when a user's Todoist Pro subscription ends. This should be used
	// for information purposes only as this does not include the grace period
	// upon expiration. As a result, avoid using this to determin whether someone
	// has a Todoist Pro subscription and use IsPremium instead
	PremiumUntil time.Time `json:"premium_until"`
	// Whether to show projects by oldest dates first (0) or oldest dates last (1)
	SortOrder int `json:"sort_order"`
	// The first day of the week (between 1 and 7, where 1 is Monday and 7 is
	// Sunday)
	StartDay int `json:"start_day"`
	// The user's default view on Todoist. The start page can be one of the
	// following: inbox, teaminbox, today, next7days, project?id=1234 to open
	// a project, label?name=abc, to open a label, filter?id=1234 to open a filter
	StartPage string `json:"start_page"`
	// The ID of the Team Inbox project
	TeamInboxId string `json:"team_inbox_id"`
	// The currently selected Todoist theme (a number between 0 and 10)
	// I love that it's a string
	ThemeId string `json:"theme_id"`
	// The user's token that should be used to call the other API methods
	// TODO: should we just not keep this?
	Token string `json:"token"`
	// The user's timezone
	TzInfo UserTimeZoneInfo `json:"tz_info"`
	// The day used when a user chooses to schedule a task for the Weekend
	// (between 1 and 7, where 1 is Monday and 7 is Sunday)
	WeekendStartDay int `json:"weekend_start_day"`
	// Describes if the user has verified their email address or not
	VerificationStatus VerificationStatus `json:"verification_status"`
}

type VerificationStatus string

const (
	VerificationStatusUnverified VerificationStatus = "unverified"
	VerificationStatusVerified   VerificationStatus = "verified"
	VerificationStatusBlocked    VerificationStatus = "blocked"
	VerificationStatusLegacy     VerificationStatus = "leagcy"
)

// Describes the availability of features and any limitations applied for
// a given user plan
type UserPlanInfo struct {
	// The name of the plan
	PlanName string `json:"plan_name"`
	// Whether the user can view the
	// activity log, see https://developer.todoist.com/sync/v9/#activity
	ActivityLog bool `json:"activity_log"`
	// The number of days of history that will be displayed within the activity
	// log. If there is no limit, the value will be -1
	ActivityLogLimit int `json:"activity_log_limit"`
	// Whether backups will be automatically created for the user's account and
	// available for download, see https://developer.todoist.com/sync/v9/#backups
	AutomaticBackups bool `json:"automatic_backups"`
	// Whether calendar feeds can be enabled for the user's projects.
	CalendarFeeds bool `json:"calendar_feeds"`
	// Whether the user can add comments
	Comments bool `json:"comments"`
	// Whether the user can search in the completed tasks archive or access the
	// completed tasks overview
	CompletedTasks bool `json:"completed_tasks"`
	// Whether the user can use special themes or other visual customization such
	// as custom app icons
	CustomizationColor bool `json:"customization_color"`
	// Whether the user can add tasks or comments via email
	EmailForwarding bool `json:"email_forwarding"`
	// Whether the user can add and update filters
	Filters bool `json:"filters"`
	// The maximum number of filters a user can have.
	MaxFilters int `json:"max_filters"`
	// Whether the user can add labels
	Labels bool `json:"labels"`
	// The maximum number of labels a user can have
	MaxLabels int `json:"max_labels"`
	// Whether the user can add reminders
	Reminders bool `json:"reminders"`
	// The maximum number of location reminders a user can have
	MaxRemindersLocation int `json:"max_reminders_location"`
	// The maximum number of time-based reminders a user can have
	MaxRemindersTime int `json:"max_reminders_time"`
	// Whether the user can import and export project templates
	Templates bool `json:"templates"`
	// Whether the user can upload attachments
	Uploads bool `json:"uploads"`
	// The maximum size of an individual file the user can upload
	UploadLimitMb int `json:"upload_limit_mb"`
	// Whether the user can view productivity stats
	WeeklyTrends bool `json:"weekly_trends"`
	// The maximum number of active projects a user can have
	MaxProjects int `json:"max_projects"`
	// The maximum number of active sections a user can have
	MaxSections int `json:"max_sections"`
	// The maximum number of active tasks a user can have
	MaxTasks int `json:"max_tasks"`
	// The maximum number of collaborators a user can add to a project
	MaxCollaborators int `json:"max_collaborators"`
	// UNDOCUMENTED
	AdvancedPermissions bool `json:"advanced_permissions"`
	// UNDOCUMENTED
	CalendarLayout bool `json:"calendar_layout"`
	// UNDOCUMENTED
	Durations bool `json:"durations"`
	// UNDOCUMENTED
	MaxFoldersPerWorkspace int `json:"max_folders_per_workspace"`
	// UNDOCUMENTED
	MaxFreeWorkspacesCreated int `json:"max_free_workspaces_created"`
	// UNDOCUMENTED
	MaxGuestsPerWorkspace int `json:"max_guests_per_workspace"`
	// UNDOCUMENTED
	MaxProjectsJoined int `json:"max_projects_joined"`
}

type SyncResponse struct {
	CollaboratorStates []CollaboratorState `json:"collaborator_states,omitempty"`
	Collaborators      []Collaborator      `json:"collaborators,omitempty"`
	CompletedInfo      []CompletedInfo     `json:"completed_info,omitempty"`
	DayOrders          map[string]int      `json:"day_orders,omitempty"`
	// This appears to be an epoch timestamp
	DayOrdersTimestamp string `json:"day_orders_timestamp,omitempty"`
	// I have no idea what this does, I couldn't find any information on it
	DueExceptions []any    `json:"due_exceptions,omitempty"`
	Filters       []Filter `json:"filters,omitempty"`
	// This is a team feature
	Folders  []Folder `json:"folders,omitempty"`
	FullSync bool     `json:"full_sync"`
	// I'm assuming these are strings
	IncompleteItemIds []string `json:"incomplete_item_ids,omitempty"`
	// I'm assuming these are strings
	IncompleteProjectIds []string `json:"incomplete_project_ids,omitempty"`
	// Also known as Tasks. Items is an old term they used and will be amended in
	// a future version of the sync API
	Items  []*types.Item  `json:"items,omitempty"`
	Labels []*types.Label `json:"labels,omitempty"`
	// I think I'm just going to ignore live notifications for now, see
	// developer.todoist.com/sync/v9/#live-notifications
	LiveNotifications           []any  `json:"live_notifications,omitempty"`
	LiveNotificationsLastReadID string `json:"live_notifications_last_read_id,omitempty"`
	// Locations are a top-level entity in the sync model. They contain a list
	// of all locations that are used within user's current location reminders.
	//
	// The location object is specific, it is an ordered array
	// 0 - Name of the Location
	// 1 - Latitude
	// 2 - Longitude
	Locations [][]string `json:"locations,omitempty"`
	Notes     []Note     `json:"notes,omitempty"`
	// Apparently you can comment on projects, who knew?!?!
	ProjectNotes []ProjectNote    `json:"project_notes,omitempty"`
	Projects     []*types.Project `json:"projects,omitempty"`
	Reminders    []Reminder       `json:"reminders,omitempty"`
	Sections     []Section        `json:"sections,omitempty"`
	Stats        Stats            `json:"stats,omitempty"`
	// The returned sync token
	SyncToken string `json:"sync_token"`
	// UNDOCUMENTED
	Tooltips struct {
		Scheduled []string `json:"scheduled"`
		Seen      []string `json:"seen"`
	} `json:"tooltips,omitempty"`
	// TODO I feel like there are missing fields here that aren't documented
	User           User `json:"user,omitempty"`
	UserPlanLimits struct {
		Current UserPlanInfo `json:"current"`
		Next    UserPlanInfo `json:"next"`
	} `json:"user_plan_limits,omitempty"`
	// UNDOCUMENTED essentially
	// TODO break this out but I don't know if I'll ever add ways to modify this
	UserSettings struct {
		CompletedSoundDesktop  bool `json:"completed_sound_desktop"`
		CompletedSoundMobile   bool `json:"completed_sound_mobile"`
		HabitPushNotifications struct {
			Features []struct {
				Enabled bool   `json:"enabled"`
				Name    string `json:"name"`
				SendAt  string `json:"send_at"`
			} `json:"features"`
		} `json:"habit_push_notifications,omitempty"`
		LegacyPricing int `json:"legacy_pricing,omitempty"`
		Navigation    struct {
			CountsShown bool `json:"counts_shown"`
			Features    []struct {
				Name  string `json:"name"`
				Shown bool   `json:"shown"`
			} `json:"features"`
		} `json:"navigation"`
		QuickAdd struct {
			Features []struct {
				Name  string `json:"name"`
				Shown bool   `json:"shown"`
			} `json:"features"`
			LabelsShown bool `json:"labels_shown"`
		} `json:"quick_add"`
		ReminderDesktop        bool `json:"reminder_desktop"`
		ReminderEmail          bool `json:"reminder_email"`
		ReminderPush           bool `json:"reminder_push"`
		ResetRecurringSubtasks bool `json:"reset_recurring_subtasks"`
	} `json:"user_settings,omitempty"`
	// UNDOCUMENTED
	ViewOptions []struct {
		FilteredBy         any    `json:"filtered_by"`
		GroupedBy          any    `json:"grouped_by"`
		ID                 string `json:"id"`
		IsDeleted          bool   `json:"is_deleted"`
		ObjectID           string `json:"object_id"`
		ShowCompletedTasks bool   `json:"show_completed_tasks"`
		SortOrder          string `json:"sort_order"`
		SortedBy           string `json:"sorted_by"`
		V2ObjectID         string `json:"v2_object_id,omitempty"`
		ViewMode           string `json:"view_mode"`
		ViewType           string `json:"view_type"`
	} `json:"view_options,omitempty"`
	// UNDOCUMENTED
	Workspaces []any `json:"workspaces,omitempty"`
	// Mapping of temp IDs to Todoist IDs
	TempIdMapping map[string]string `json:"temp_id_mapping"`
	// Mapping of sync operations by uuid to their status
	SyncStatus map[string]OperationResult `json:"sync_status"`
}
