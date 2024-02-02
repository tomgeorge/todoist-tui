package types

type DueDate struct {
	// Due date in the format of YYYY-MM-DD. For recurring dates, the date of the
	// current iteration
	// TODO: Could this be a time.Time
	Date string `json:"date,omitempty"`
	// Timezone
	Timezone *string `json:"timezone,omitempty"`
	// Is the task recurring
	IsRecurring bool `json:"is_recurring,omitempty"`
	// String representation of the date (e.g. "Thursday", or "Every wednesday",
	// or "Tomorrow at 11")
	String string `json:"string,omitempty"`
	// Language (I'm just supporting en for now)
	Lang string `json:"lang,omitempty"`
}
