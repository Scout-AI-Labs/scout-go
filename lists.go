package scout

import (
	"context"
	"net/url"
	"strconv"
)

// ListsService covers find-all ("lists"): build a list of entities matching a
// natural-language query, then enrich or extend the run.
type ListsService struct {
	client *Client
	// Runs holds operations on async find-all runs.
	Runs *ListRunsService
}

// ListsParams are the inputs to Create and Run.
type ListsParams struct {
	Query        string         `json:"query"`
	Fields       []string       `json:"fields,omitempty"`
	OutputSchema map[string]any `json:"output_schema,omitempty"`
	Limit        *int           `json:"limit,omitempty"`
}

// Create runs a find-all synchronously.
func (s *ListsService) Create(ctx context.Context, params *ListsParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/lists", nil, params, &out)
	return out, err
}

// Run starts an async find-all run; poll Runs.Get for progress.
func (s *ListsService) Run(ctx context.Context, params *ListsParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/lists/runs", nil, params, &out)
	return out, err
}

// ListRunsService covers async find-all runs.
type ListRunsService struct{ client *Client }

// List returns find-all runs.
func (s *ListRunsService) List(ctx context.Context, limit, offset int) (Result, error) {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	var out Result
	err := s.client.do(ctx, "GET", "/v1/lists/runs", q, nil, &out)
	return out, err
}

// Iterate returns an auto-paginating iterator over all find-all runs.
func (s *ListRunsService) Iterate(ctx context.Context) *Iterator {
	return newIterator(ctx, s.client, "/v1/lists/runs", 50)
}

// Get fetches a find-all run by id.
func (s *ListRunsService) Get(ctx context.Context, findallID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "GET", "/v1/lists/runs/"+url.PathEscape(findallID), nil, nil, &out)
	return out, err
}

// Cancel cancels a find-all run.
func (s *ListRunsService) Cancel(ctx context.Context, findallID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/lists/runs/"+url.PathEscape(findallID)+"/cancel", nil, nil, &out)
	return out, err
}

// Enrich enriches the run's entities with additional fields.
func (s *ListRunsService) Enrich(ctx context.Context, findallID string, body map[string]any) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/lists/runs/"+url.PathEscape(findallID)+"/enrich", nil, body, &out)
	return out, err
}

// Extend extends the run with more matching entities.
func (s *ListRunsService) Extend(ctx context.Context, findallID string, body map[string]any) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/lists/runs/"+url.PathEscape(findallID)+"/extend", nil, body, &out)
	return out, err
}

// Events fetches a find-all run's events.
func (s *ListRunsService) Events(ctx context.Context, findallID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "GET", "/v1/lists/runs/"+url.PathEscape(findallID)+"/events", nil, nil, &out)
	return out, err
}

// StreamEvents streams a find-all run's progress events live (SSE).
func (s *ListRunsService) StreamEvents(ctx context.Context, findallID string) (*Stream, error) {
	return s.client.openStream(ctx, "GET", "/v1/lists/runs/"+url.PathEscape(findallID)+"/events", nil)
}
