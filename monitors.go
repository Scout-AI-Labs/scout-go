package scout

import (
	"context"
	"net/url"
	"strconv"
)

// MonitorsService covers scheduled searches ("monitors"): run a query on a
// cadence and receive new results via webhook.
type MonitorsService struct{ client *Client }

// MonitorCreateParams are the inputs to Create.
type MonitorCreateParams struct {
	Query        string         `json:"query"`
	Webhook      *string        `json:"webhook,omitempty"`
	Cadence      *string        `json:"cadence,omitempty"`
	Cron         *string        `json:"cron,omitempty"`
	Mode         *string        `json:"mode,omitempty"`
	FilterPrompt *string        `json:"filter_prompt,omitempty"`
	Country      *string        `json:"country,omitempty"`
	Language     *string        `json:"language,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// Create creates a monitor with a query and a cadence or cron schedule.
func (s *MonitorsService) Create(ctx context.Context, params *MonitorCreateParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/monitors", nil, params, &out)
	return out, err
}

// List returns monitors.
func (s *MonitorsService) List(ctx context.Context, limit, offset int) (Result, error) {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	var out Result
	err := s.client.do(ctx, "GET", "/v1/monitors", q, nil, &out)
	return out, err
}

// Iterate returns an auto-paginating iterator over all monitors.
func (s *MonitorsService) Iterate(ctx context.Context) *Iterator {
	return newIterator(ctx, s.client, "/v1/monitors", 50)
}

// Get fetches a monitor by id.
func (s *MonitorsService) Get(ctx context.Context, monitorID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "GET", "/v1/monitors/"+url.PathEscape(monitorID), nil, nil, &out)
	return out, err
}

// MonitorUpdateParams are the inputs to Update.
type MonitorUpdateParams struct {
	Query        *string `json:"query,omitempty"`
	Webhook      *string `json:"webhook,omitempty"`
	Cadence      *string `json:"cadence,omitempty"`
	Cron         *string `json:"cron,omitempty"`
	FilterPrompt *string `json:"filter_prompt,omitempty"`
}

// Update updates a monitor's query, schedule, or webhook.
func (s *MonitorsService) Update(ctx context.Context, monitorID string, params *MonitorUpdateParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "PATCH", "/v1/monitors/"+url.PathEscape(monitorID), nil, params, &out)
	return out, err
}

// Delete deletes a monitor.
func (s *MonitorsService) Delete(ctx context.Context, monitorID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "DELETE", "/v1/monitors/"+url.PathEscape(monitorID), nil, nil, &out)
	return out, err
}

// Pause pauses a monitor.
func (s *MonitorsService) Pause(ctx context.Context, monitorID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/monitors/"+url.PathEscape(monitorID)+"/pause", nil, nil, &out)
	return out, err
}

// Resume resumes a paused monitor.
func (s *MonitorsService) Resume(ctx context.Context, monitorID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/monitors/"+url.PathEscape(monitorID)+"/resume", nil, nil, &out)
	return out, err
}

// Run triggers a monitor run immediately.
func (s *MonitorsService) Run(ctx context.Context, monitorID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/monitors/"+url.PathEscape(monitorID)+"/run", nil, nil, &out)
	return out, err
}

// Events fetches a monitor's events.
func (s *MonitorsService) Events(ctx context.Context, monitorID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "GET", "/v1/monitors/"+url.PathEscape(monitorID)+"/events", nil, nil, &out)
	return out, err
}

// StreamEvents streams a monitor's events live (SSE).
func (s *MonitorsService) StreamEvents(ctx context.Context, monitorID string) (*Stream, error) {
	return s.client.openStream(ctx, "GET", "/v1/monitors/"+url.PathEscape(monitorID)+"/events", nil)
}
