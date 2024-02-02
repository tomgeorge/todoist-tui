package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	Version          = "v0.0.1"
	defaultBaseURL   = "https://api.todoist.com/rest/v2/"
	defaultUserAgent = "todoist-tui" + "/" + Version
)

var errNonNilContext = errors.New("context must not be nil")

type Client struct {
	client    *http.Client
	BaseURL   *url.URL
	UserAgent string

	common service
	Tasks  *TasksService
}

type service struct {
	client *Client
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

func (c *Client) NewRequest(method, urlStr string, body interface{}, opts ...RequestOption) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	u, err := c.BaseURL.Parse(c.BaseURL.Path)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String()+urlStr, buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
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
	return errors.New(fmt.Sprintf("%s request %s: %s", r.Request.Method, r.Request.URL.String(), r.Status))
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
	c.common.client = c
	c.Tasks = (*TasksService)(&c.common)
}

// roundTripperFunc creates a RoundTripper (transport)
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}
