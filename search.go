package scout

import (
	"context"
	"net/url"
	"strconv"
)

// SearchService covers web search, agentic AI queries, and search-run history.
type SearchService struct{ client *Client }

// SearchParams are the inputs to Create. Optional fields use pointers; set them
// with scout.String / scout.Int.
type SearchParams struct {
	Queries        []string `json:"queries"`
	Objective      *string  `json:"objective,omitempty"`
	Depth          *string  `json:"depth,omitempty"`
	Mode           *string  `json:"mode,omitempty"`
	Category       *string  `json:"category,omitempty"`
	Limit          *int     `json:"limit,omitempty"`
	Country        *string  `json:"country,omitempty"`
	Location       *string  `json:"location,omitempty"`
	Language       *string  `json:"language,omitempty"`
	Freshness      *string  `json:"freshness,omitempty"`
	IncludeDomains []string `json:"include_domains,omitempty"`
	ExcludeDomains []string `json:"exclude_domains,omitempty"`
	SessionID      *string  `json:"session_id,omitempty"`
	Webhook        *string  `json:"webhook,omitempty"`
}

// Create runs a web search.
func (s *SearchService) Create(ctx context.Context, params *SearchParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/search", nil, params, &out)
	return out, err
}

// AIQueryParams are the inputs to AIQuery.
type AIQueryParams struct {
	URL      string `json:"url"`
	Question string `json:"question"`
	MaxPages *int   `json:"max_pages,omitempty"`
}

// AIQuery answers a natural-language question by reading a page (and its links).
func (s *SearchService) AIQuery(ctx context.Context, params *AIQueryParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/ai-query", nil, params, &out)
	return out, err
}

// List returns prior search runs (most recent first).
func (s *SearchService) List(ctx context.Context, limit, offset int) (Result, error) {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	var out Result
	err := s.client.do(ctx, "GET", "/v1/searches", q, nil, &out)
	return out, err
}

// Iterate returns an auto-paginating iterator over all search runs.
func (s *SearchService) Iterate(ctx context.Context) *Iterator {
	return newIterator(ctx, s.client, "/v1/searches", 50)
}

// Get fetches a single search run by id.
func (s *SearchService) Get(ctx context.Context, searchID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "GET", "/v1/searches/"+url.PathEscape(searchID), nil, nil, &out)
	return out, err
}

// Cancel cancels an in-flight search run.
func (s *SearchService) Cancel(ctx context.Context, searchID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/searches/"+url.PathEscape(searchID)+"/cancel", nil, nil, &out)
	return out, err
}

// Events fetches the event stream (as JSON) for a search run.
func (s *SearchService) Events(ctx context.Context, searchID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "GET", "/v1/searches/"+url.PathEscape(searchID)+"/events", nil, nil, &out)
	return out, err
}

// StreamEvents streams a deep-search run's progress events live (SSE).
func (s *SearchService) StreamEvents(ctx context.Context, searchID string) (*Stream, error) {
	return s.client.openStream(ctx, "GET", "/v1/searches/"+url.PathEscape(searchID)+"/events", nil)
}
