package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

const (
	baseURLPath = "/v2"
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

func TestWithURLs(t *testing.T) {
	cli := NewClient(nil)
	if cli.BaseURL.String() != defaultBaseURL {
		t.Errorf("Wanted base URL of %s got %s", defaultBaseURL, cli.BaseURL.String())
	}

	if cli.UserAgent != defaultUserAgent {
		t.Errorf("Wanted User-Agent of %s got %s", defaultUserAgent, cli.UserAgent)
	}
}
