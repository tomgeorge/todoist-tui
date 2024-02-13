package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/samber/lo"
	"github.com/tomgeorge/todoist-tui/types"
)

func TestUnmarshal(t *testing.T) {
	str := `{"content": "hey"}`
	var item types.Item
	err := json.NewDecoder(strings.NewReader(str)).Decode(&item)
	if err != nil {
		t.Fatalf("Didn't expect an error but got %v", err)
	}
	if item.Content != "hey" {
		t.Fatalf("Expected hey but got %s", item.Content)
	}

	list := `{"items": [{"content": "hey"},{"content":"hello"}]}`
	var response SyncResponse
	err = json.NewDecoder(strings.NewReader(list)).Decode(&response)
	if err != nil {
		t.Fatalf("Didn't expect an error but got %v", err)
	}
	if response.Items[0].Content != "hey" {
		t.Fatalf("Expected hey but got %s", response.Items[0].Content)
	}
}

func badResponse(uuid string) string {
	return fmt.Sprintf(`{"items": [], "sync_status": {"%s": {"error": "some error", "error_code": 1}}}`, uuid)
}

func TestAddItem(t *testing.T) {
	testCases := []struct {
		desc string
		in   AddItemArgs
		// function that takes a UUID and temp ID and returns a handler function :/
		handleFunc    func(string, string) http.HandlerFunc
		shouldSucceed bool
	}{
		{
			desc: "Add item success",
			in:   AddItemArgs{Content: "test item"},
			handleFunc: func(uuid, tempId string) http.HandlerFunc {
				return func(w http.ResponseWriter, req *http.Request) {
					fmt.Fprintf(w, `{"items": [{"id": "1","content": "test item"}], "temp_id_mapping": {"%s": "1"},"sync_status": {"%s": "ok"}}`, tempId, uuid)
				}
			},
			shouldSucceed: true,
		},
		{
			desc: "failure",
			in:   AddItemArgs{Content: "test item"},
			handleFunc: func(uuid, tempId string) http.HandlerFunc {
				return func(w http.ResponseWriter, req *http.Request) {
					fmt.Fprint(w, badResponse(uuid))
				}
			},
			shouldSucceed: false,
		},
		{
			desc: "task with no content",
			in:   AddItemArgs{},
			handleFunc: func(uuid, tempId string) http.HandlerFunc {
				return func(w http.ResponseWriter, req *http.Request) {
					fmt.Fprint(w, badResponse(uuid))
					w.WriteHeader(http.StatusBadRequest)
				}
			},
			shouldSucceed: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			c, mux, _, teardown := setup()
			defer teardown()
			in, handler := tc.in, tc.handleFunc
			tempId, uuid := "test-tempId", "test-uuid"
			mux.HandleFunc("/sync", handler(uuid, tempId))
			item, err := c.AddTask(context.Background(), in, WithTempId(tempId), WithUuid(uuid))
			switch tc.shouldSucceed {
			case true:
				if err != nil {
					t.Fatalf("Did not expect an error but got %v", err)
				}
				if item == nil {
					t.Fatalf("Got a nil item but didn't expect one")
				}
				if item.Content != in.Content {
					t.Fatalf("Expected %s but got %s", in.Content, item.Content)
				}
			default:
				if err == nil {
					t.Fatalf("Expected an error")
				}
			}
		})
	}
}

func TestReduce(t *testing.T) {
	type testy struct {
		Id    string
		count int
	}
	items := []testy{{Id: "1", count: 1}, {Id: "2", count: 2}, {Id: "3", count: 3}}
	expected := map[string]*testy{"1": &testy{Id: "1", count: 1}, "2": &testy{Id: "2", count: 2}, "3": &testy{Id: "3", count: 3}}
	got := lo.Reduce(items, func(m map[string]*testy, item testy, _ int) map[string]*testy {
		ret := m
		ret[item.Id] = &item
		return ret
	}, make(map[string]*testy))
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("%s", cmp.Diff(expected, got))
	}
}
