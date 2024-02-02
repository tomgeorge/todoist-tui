package sync

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-querystring/query"
)

func TestNewRequest(t *testing.T) {

	cli := NewClient(nil)

	inUrl, outUrl := "foo", defaultBaseURL+"foo"
	values, _ := query.Values(SyncRequest{SyncToken: cli.syncToken, ResourceTypes: []string{"all"}})

	inBody, outBody := strings.NewReader(values.Encode()), "resource_types=%5B%22all%22%5D&sync_token=%2A"
	req, err := cli.NewRequest("POST", inUrl, inBody)

	if err != nil {
		t.Errorf("Did not expect an error, got %v", err)
	}

	if got, want := req.URL.String(), outUrl; got != want {
		t.Errorf("NewRequest(%q) URL is %v, want %v", inUrl, got, want)
	}

	body, _ := io.ReadAll(req.Body)
	if got, want := string(body), outBody; got != want {
		t.Errorf("NewRequest(%q) Body is %v, want %v", inBody, got, want)
	}

	if got, want := req.Header.Get("User-Agent"), cli.UserAgent; got != want {
		t.Errorf("NewRequest() User-Agent is %v, want %v", got, want)
	}
}

func TestNewRequest_badURL(t *testing.T) {
	c := NewClient(nil)
	_, err := c.NewRequest("POST", ":", nil)
	if err == nil {
		t.Errorf("Expected erorr to be returned")
	}
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		t.Errorf("Expected URL parse error, got %v", err)
	}
}

func TestNewRequest_emptyUserAgent(t *testing.T) {
	c := NewClient(nil)
	c.UserAgent = ""
	req, err := c.NewRequest("GET", "", nil)
	if err != nil {
		t.Errorf("NewRequest returned unexpected error: %v", err)
	}
	if _, ok := req.Header["User-Agent"]; ok {
		t.Errorf("constructed request contains unexpected User-Agent header")
	}
}

func TestNewRequest_emptyBody(t *testing.T) {
	c := NewClient(nil)
	req, err := c.NewRequest("GET", ".", nil)
	if err != nil {
		t.Errorf("NewRequest returned unexpected error: %v", err)
	}
	if req.Body != nil {
		t.Errorf("Constructed request contains non-nil body")
	}
}

func TestEncodeSyncRequest(t *testing.T) {
	testCases := []struct {
		desc  string
		given SyncRequest
		want  url.Values
	}{
		{
			desc: "sync token",
			given: SyncRequest{
				SyncToken:     "*",
				ResourceTypes: []string{"projects", "labels"}},
			want: map[string][]string{
				"sync_token":     {"*"},
				"resource_types": {"[\"projects\",\"labels\"]"},
			},
		},
		{
			desc: "full command",
			given: SyncRequest{
				SyncToken:     "a1b2c3d4",
				ResourceTypes: []string{"projects"},
				Commands: []Command{
					{
						Type:   "project_add",
						TempId: "temp",
						Uuid:   "temp",
						Args: AddProjectArgs{
							Name: "test project",
						},
					},
				},
			},
			want: map[string][]string{
				"sync_token":     {"a1b2c3d4"},
				"resource_types": {"[\"projects\"]"},
				"commands":       {`[{"type":"project_add","temp_id":"temp","uuid":"temp","args":{"name":"test project"}}]`},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := query.Values(tc.given)
			if err != nil {
				t.Errorf("Didn't expect error but got %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Logf("%s", cmp.Diff(tc.want, got))
				t.Logf("%s", got.Encode())
				t.Errorf("Expected %v but got %v", tc.want, got)
			}
		})
	}
}

func TestSyncOperationUnmarshal(t *testing.T) {
	testCases := []struct {
		desc    string
		temp_id string
		got     string
		want    OperationResult
	}{
		{
			desc:    "success",
			temp_id: "temp_id",
			got:     `{"sync_status": {"temp_id": "ok"}}`,
			want:    OperationResult{Ok: true, ErrorCode: SyncError{}},
		},
		{
			desc:    "failure",
			temp_id: "temp_id",
			got:     `{"sync_status": {"temp_id": {"error_code": 15, "error": "Invalid temporary id"}}}`,
			want: OperationResult{Ok: false, ErrorCode: SyncError{
				ErrorCode: 15,
				Error:     "Invalid temporary id",
			}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			cli, mux, _, teardown := setup()
			defer teardown()

			mux.HandleFunc("/sync", func(w http.ResponseWriter, req *http.Request) {
				fmt.Fprintf(w, tc.got)
			})

			syncRequest := SyncRequest{
				SyncToken: "token",
				Commands: []Command{
					{
						Type:   "project_add",
						TempId: "temp_id",
						Args:   "Something that won't work",
					},
				},
			}
			syncResponse, err := cli.Sync(context.Background(), syncRequest)
			if err != nil {
				t.Errorf("Did not expect error, got %v", err)
			}
			if got := syncResponse.SyncStatus[tc.temp_id]; !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Expected %v got %v", tc.want, got)
			}
		})
	}
}
