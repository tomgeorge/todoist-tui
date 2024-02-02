package sync

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

const (
	baseURLPath = "/v9"
)

func setup() (client *Client, mux *http.ServeMux, serverUrl string, teardown func()) {
	mux = http.NewServeMux()

	apiHandler := http.NewServeMux()
	apiHandler.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, mux))
	apiHandler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(os.Stderr, "FAIL: Client.BaseURL path prefix is not preserved in the request URL:")
		fmt.Fprintln(os.Stderr, "\t"+req.URL.String())
		http.Error(w, "Client.BaseURL path prefix is not preserved in the request URL", http.StatusInternalServerError)
	})

	server := httptest.NewServer(apiHandler)

	client = NewClient(nil)
	url, _ := url.Parse(server.URL + baseURLPath + "/")
	client.BaseURL = url
	return client, mux, server.URL, server.Close
}

func TestAddProject(t *testing.T) {
	c, server, _, teardown := setup()
	defer teardown()

	server.HandleFunc("/sync", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, `{"projects": [{"name": "foo"}]}`)
	})

	requestArgs := AddProjectArgs{
		Name:  "foo",
		Color: "charcoal",
	}
	project, err := c.AddProject(requestArgs)
	if err != nil {
		t.Errorf("Didn't expect an error but got %v", err)
	}

	if project.Name != "foo" {
		t.Errorf("Expected foo got %v", project.Name)
	}
}
