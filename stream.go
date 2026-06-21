package scout

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// Stream is a live Server-Sent-Events stream that decodes each event's JSON
// payload into a Result.
//
//	st, err := client.Chat.Completions.CreateStream(ctx, params)
//	if err != nil { ... }
//	defer st.Close()
//	for st.Next() {
//	    chunk := st.Current()
//	}
//	if err := st.Err(); err != nil { ... }
type Stream struct {
	resp   *http.Response
	reader *bufio.Reader
	cur    Result
	err    error
}

type sseEvent struct {
	event string
	data  string
}

func (c *Client) openStream(ctx context.Context, method, path string, body any) (*Stream, error) {
	if c.apiKey == "" {
		return nil, &Error{Message: "missing API key: use scout.WithAPIKey or set SCOUT_API_KEY"}
	}
	var reader *bytes.Reader
	if body != nil && method != http.MethodGet {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, &Error{Message: "failed to encode request body: " + err.Error()}
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("User-Agent", "scout-go/"+Version)
	req.Header.Set("Scout-Version", APIVersion)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	if body != nil && method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", newIdempotencyKey())
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &connError{err: err}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errorFromResponse(resp)
	}
	return &Stream{resp: resp, reader: bufio.NewReader(resp.Body)}, nil
}

// Next advances to the next event, returning false at end-of-stream, on a
// `[DONE]` sentinel, or on error (see Err).
func (s *Stream) Next() bool {
	if s.err != nil {
		return false
	}
	evt, ok := s.readEvent()
	if !ok {
		return false
	}
	if evt.data == "[DONE]" {
		return false
	}
	var r Result
	if err := json.Unmarshal([]byte(evt.data), &r); err != nil {
		s.err = &Error{Message: "failed to decode event: " + err.Error()}
		return false
	}
	s.cur = r
	return true
}

// Current returns the most recent decoded event. Call after Next returns true.
func (s *Stream) Current() Result { return s.cur }

// Err returns the first error encountered while streaming.
func (s *Stream) Err() error { return s.err }

// Close releases the underlying connection.
func (s *Stream) Close() error { return s.resp.Body.Close() }

func (s *Stream) readEvent() (sseEvent, bool) {
	var event string
	var data []string
	for {
		line, err := s.reader.ReadString('\n')
		if len(line) > 0 {
			trimmed := strings.TrimRight(line, "\r\n")
			switch {
			case trimmed == "":
				if len(data) > 0 {
					return sseEvent{event, strings.Join(data, "\n")}, true
				}
				event = ""
				data = nil
			case strings.HasPrefix(trimmed, ":"):
				// comment / keepalive
			default:
				field, value := trimmed, ""
				if i := strings.IndexByte(trimmed, ':'); i >= 0 {
					field, value = trimmed[:i], trimmed[i+1:]
					value = strings.TrimPrefix(value, " ")
				}
				if field == "event" {
					event = value
				} else if field == "data" {
					data = append(data, value)
				}
			}
		}
		if err != nil {
			if len(data) > 0 {
				return sseEvent{event, strings.Join(data, "\n")}, true
			}
			if err != io.EOF {
				s.err = &connError{err: err}
			}
			return sseEvent{}, false
		}
	}
}
