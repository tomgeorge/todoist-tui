package types

import (
	"testing"
	"time"
)

// Example due dates that are possible
//
//	{
//	    "date": "2016-12-01",
//	    "timezone": null,
//	    "string": "every day",
//	    "lang": "en",
//	    "is_recurring": true
//	}
//
//	{
//	    "date": "2016-12-0T12:00:00.000000",
//	    "timezone": null,
//	    "string": "every day at 12",
//	    "lang": "en",
//	    "is_recurring": true
//	}
//
//	{
//	    "date": "2016-12-06T13:00:00.000000Z",
//	    "timezone": "Europe/Madrid",
//	    "string": "ev day at 2pm",
//	    "lang": "en",
//	    "is_recurring": true
//	}
//
// Input example "due": {"string":  "tomorrow"}
//
//	"due": {
//	    "date": "2018-11-15",
//	    "timezone": null,
//	    "is_recurring": false,
//	    "string": "tomorrow",
//	    "lang": "en"
//	}
//
//	Input Example "due": {"string":  "tomorrow at 12"}
//
//	"due": {
//	    "date": "2018-11-15T12:00:00.000000",
//	    "timezone": null,
//	    "is_recurring": false,
//	    "string": "tomorrow at 12",
//	    "lang": "en"
//	}
//
//	Input example. Timezone set explicitly
//	"due": {"string": "tomorrow at 12", "timezone": "Asia/Jakarta"}
//
//	"due": {
//	    "date": "2018-11-16T05:00:00.000000Z",
//	    "timezone": "Asia/Jakarta",
//	    "is_recurring": false,
//	    "string": "tomorrow at 12",
//	    "lang": "en"
//	}
func TestThatIUnderstandHowTimeDotTimeWorks(t *testing.T) {
	due := DueDate{
		Date: "2018-11-16T12:00:00.000000Z",
	}
	dateTime, err := time.Parse(time.RFC3339, due.Date)
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}
	if dateTime.Month() != 11 ||
		dateTime.Year() != 2018 ||
		dateTime.Day() != 16 ||
		dateTime.Hour() != 12 ||
		dateTime.Minute() != 0 ||
		dateTime.Second() != 0 {
		t.Fatalf("Parsed to unexpected time: %v", dateTime)
	}

	// Regular full day date
	due = DueDate{
		Date: "2018-11-16",
	}
	dateTime, err = time.Parse(time.RFC3339, due.Date)
	if err == nil {
		t.Fatalf("I would expect an error here")
	}

	// Floating due date with time, the docs say this isn't "quite compatible with
	// RFC 3339
	due = DueDate{
		Date:        "2016-12-0T12:00:00.000000",
		Timezone:    nil,
		String:      "every day at 12",
		Lang:        "en",
		IsRecurring: true,
	}
	dateTime, err = time.Parse(time.RFC3339, due.Date)
	if err == nil {
		t.Fatalf("I would also expect an error here")
	}
	if err != nil {
		t.Logf("TestThatIUnderstandHowTimeDotTimeWorks-Got error %v", err)
	}
}
func TestParsing(t *testing.T) {
	// type testcase struct {
	// 	got  string
	// 	want time.Time
	// }

	// testcases := []testcase{
	// 	{
	// 		got: `{"date": "2016-12-01", "timezone": null, "string": "every day"}`,
	// 	},
	// }
}
