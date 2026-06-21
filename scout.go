// Package scout is the official Go SDK for the Scout web-intelligence API:
// search, scrape, screenshot, extract, crawl, and company enrichment.
//
// It has zero third-party dependencies — built entirely on the standard
// library's net/http and encoding/json.
//
//	client := scout.NewClient(scout.WithAPIKey("sk_..."))
//	res, err := client.Search.Create(ctx, &scout.SearchParams{Queries: []string{"climate tech"}})
package scout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// Version is the SDK package version.
	Version = "0.1.0"
	// APIVersion is the Scout REST API version this SDK targets.
	APIVersion = "2026-06-21"

	defaultBaseURL    = "https://core.usescout.sh"
	defaultTimeout    = 60 * time.Second
	defaultMaxRetries = 2
)

// Result is a decoded JSON object response. Fields vary by endpoint.
type Result = map[string]any

// Client is the entry point to the Scout API. Construct it with NewClient.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	maxRetries int
	headers    map[string]string

	// Resource groups — a faithful 1:1 mirror of the REST API tags.
	Search   *SearchService
	Page     *PageService
	Extract  *ExtractService
	Company  *CompanyService
	Lists    *ListsService
	Products *ProductsService
	Site     *SiteService
	Jobs     *JobsService
	Monitors *MonitorsService
	Chat     *ChatService
}

// Option configures a Client.
type Option func(*Client)

// WithAPIKey sets the API key. Defaults to the SCOUT_API_KEY env var.
func WithAPIKey(key string) Option { return func(c *Client) { c.apiKey = key } }

// WithBaseURL overrides the API origin.
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") } }

// WithHTTPClient supplies a custom *http.Client (proxies, transports, timeouts).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// WithTimeout sets the per-request timeout (applied via the http.Client).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// WithMaxRetries sets the number of automatic retries for transient failures.
func WithMaxRetries(n int) Option { return func(c *Client) { c.maxRetries = n } }

// WithHeader adds a header sent on every request.
func WithHeader(key, value string) Option {
	return func(c *Client) { c.headers[key] = value }
}

// NewClient builds a Client. The API key falls back to the SCOUT_API_KEY
// environment variable when WithAPIKey is not supplied.
func NewClient(opts ...Option) *Client {
	c := &Client{
		apiKey:     os.Getenv("SCOUT_API_KEY"),
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
		maxRetries: defaultMaxRetries,
		headers:    map[string]string{},
	}
	for _, opt := range opts {
		opt(c)
	}
	c.Search = &SearchService{client: c}
	c.Page = &PageService{client: c}
	c.Extract = &ExtractService{client: c}
	c.Company = &CompanyService{client: c}
	c.Lists = &ListsService{client: c, Runs: &ListRunsService{client: c}}
	c.Products = &ProductsService{client: c}
	c.Site = &SiteService{client: c}
	c.Jobs = &JobsService{client: c}
	c.Monitors = &MonitorsService{client: c}
	c.Chat = &ChatService{client: c, Completions: &ChatCompletionsService{client: c}}
	return c
}

var retryStatuses = map[int]bool{408: true, 409: true, 429: true, 500: true, 502: true, 503: true, 504: true}

// do issues a request with retries and decodes the JSON response into out
// (which may be nil). The query map and body may both be nil.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	if c.apiKey == "" {
		return &Error{Message: "missing API key: use scout.WithAPIKey or set SCOUT_API_KEY"}
	}

	endpoint := c.baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var bodyBytes []byte
	if body != nil && method != http.MethodGet {
		b, err := json.Marshal(body)
		if err != nil {
			return &Error{Message: fmt.Sprintf("failed to encode request body: %v", err)}
		}
		bodyBytes = b
	}

	isWrite := method != http.MethodGet
	for attempt := 0; ; attempt++ {
		resp, err := c.attempt(ctx, method, endpoint, bodyBytes, isWrite)
		if err == nil {
			return decode(resp, out)
		}
		if !isRetriable(err) || attempt >= c.maxRetries {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff(attempt, err)):
		}
	}
}

func (c *Client) attempt(ctx context.Context, method, endpoint string, bodyBytes []byte, isWrite bool) (*http.Response, error) {
	var reader *bytes.Reader
	if bodyBytes != nil {
		reader = bytes.NewReader(bodyBytes)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "scout-go/"+Version)
	req.Header.Set("Scout-Version", APIVersion)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if isWrite {
		req.Header.Set("Idempotency-Key", newIdempotencyKey())
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Network/timeout/cancel: wrap as a retriable connection error.
		return nil, &connError{err: err}
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}
	return nil, errorFromResponse(resp)
}

func backoff(attempt int, err error) time.Duration {
	if apiErr, ok := err.(*Error); ok {
		if ra := apiErr.Header.Get("Retry-After"); ra != "" {
			if secs, e := strconv.Atoi(ra); e == nil {
				if d := time.Duration(secs) * time.Second; d < 60*time.Second {
					return d
				}
				return 60 * time.Second
			}
		}
	}
	base := min(500*(1<<attempt), 8000)
	jitter := 0.5 + rand.Float64()*0.5
	return time.Duration(float64(base)*jitter) * time.Millisecond
}

func newIdempotencyKey() string {
	const hex = "0123456789abcdef"
	b := make([]byte, 32)
	for i := range b {
		b[i] = hex[rand.Intn(16)]
	}
	return "idmp-" + string(b)
}
