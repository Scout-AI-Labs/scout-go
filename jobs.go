package scout

import (
	"context"
	"net/url"
	"strconv"
)

// JobsService covers async tasks ("jobs"): submit a natural-language task, then
// poll the task or stream its events until it completes.
type JobsService struct{ client *Client }

// JobCreateParams are the inputs to Create.
type JobCreateParams struct {
	Task         string         `json:"task"`
	OutputSchema map[string]any `json:"output_schema,omitempty"`
	Processor    *string        `json:"processor,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Webhook      *string        `json:"webhook,omitempty"`
}

// Create submits a job. The result includes a task id to poll with Get.
func (s *JobsService) Create(ctx context.Context, params *JobCreateParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/jobs", nil, params, &out)
	return out, err
}

// List returns jobs (most recent first).
func (s *JobsService) List(ctx context.Context, limit, offset int) (Result, error) {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	var out Result
	err := s.client.do(ctx, "GET", "/v1/jobs", q, nil, &out)
	return out, err
}

// Iterate returns an auto-paginating iterator over all jobs.
func (s *JobsService) Iterate(ctx context.Context) *Iterator {
	return newIterator(ctx, s.client, "/v1/jobs", 50)
}

// Get fetches a job by task id.
func (s *JobsService) Get(ctx context.Context, taskID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "GET", "/v1/jobs/"+url.PathEscape(taskID), nil, nil, &out)
	return out, err
}

// Cancel cancels a running job.
func (s *JobsService) Cancel(ctx context.Context, taskID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/jobs/"+url.PathEscape(taskID)+"/cancel", nil, nil, &out)
	return out, err
}

// Events fetches a job's events.
func (s *JobsService) Events(ctx context.Context, taskID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "GET", "/v1/jobs/"+url.PathEscape(taskID)+"/events", nil, nil, &out)
	return out, err
}

// StreamEvents streams a job's progress events live (SSE).
func (s *JobsService) StreamEvents(ctx context.Context, taskID string) (*Stream, error) {
	return s.client.openStream(ctx, "GET", "/v1/jobs/"+url.PathEscape(taskID)+"/events", nil)
}

// StartRun starts a run for a job.
func (s *JobsService) StartRun(ctx context.Context, body map[string]any) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/jobs/runs", nil, body, &out)
	return out, err
}

// RunResult fetches the result of a completed run.
func (s *JobsService) RunResult(ctx context.Context, runID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "GET", "/v1/jobs/runs/"+url.PathEscape(runID), nil, nil, &out)
	return out, err
}

// RunEvents fetches a run's events.
func (s *JobsService) RunEvents(ctx context.Context, runID string) (Result, error) {
	var out Result
	err := s.client.do(ctx, "GET", "/v1/jobs/runs/"+url.PathEscape(runID)+"/events", nil, nil, &out)
	return out, err
}
