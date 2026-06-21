package scout

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	var flakyHits int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "req_abc123")
		switch r.URL.Path {
		case "/v1/search":
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			json.NewEncoder(w).Encode(map[string]any{
				"ok":   true,
				"auth": r.Header.Get("Authorization"),
				"idem": r.Header.Get("Idempotency-Key"),
				"echo": body,
			})
		case "/v1/flaky":
			if atomic.AddInt32(&flakyHits, 1) < 3 {
				w.WriteHeader(500)
				json.NewEncoder(w).Encode(map[string]any{"detail": "transient"})
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "tries": atomic.LoadInt32(&flakyHits)})
		case "/v1/nope":
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(map[string]any{"detail": "invalid api key"})
		case "/v1/searches":
			json.NewEncoder(w).Encode(map[string]any{"items": []any{map[string]any{"id": 1}}})
		case "/v1/chat/completions":
			w.Header().Set("Content-Type", "text/event-stream")
			io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hel\"}}]}\n\n")
			io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"lo\"}}]}\n\n")
			io.WriteString(w, "data: [DONE]\n\n")
		case "/v1/searches/abc/events":
			w.Header().Set("Content-Type", "text/event-stream")
			io.WriteString(w, ": keepalive\n\n")
			io.WriteString(w, "event: run.progress\ndata: {\"type\":\"run.progress\"}\n\n")
			io.WriteString(w, "event: run.completed\ndata: {\"type\":\"run.completed\"}\n\n")
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	c := NewClient(WithAPIKey("sk_live_xyz"), WithBaseURL(srv.URL), WithMaxRetries(3))

	// 1) POST round-trip + auth + idempotency
	res, err := c.Search.Create(ctx, &SearchParams{Queries: []string{"hello world"}, Depth: String("standard")})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if res["auth"] != "Bearer sk_live_xyz" {
		t.Fatalf("auth header: %v", res["auth"])
	}
	if res["idem"] == "" || res["idem"] == nil {
		t.Fatal("missing idempotency key")
	}
	echo := res["echo"].(map[string]any)
	if echo["depth"] != "standard" {
		t.Fatalf("body not encoded: %v", echo)
	}

	// 2) retry on 500 then succeed
	f, err := c.do2(ctx, "POST", "/v1/flaky")
	if err != nil {
		t.Fatalf("flaky: %v", err)
	}
	if f["ok"] != true {
		t.Fatalf("flaky not recovered: %v", f)
	}

	// 3) 401 -> typed error with request id
	_, err = c.do2(ctx, "POST", "/v1/nope")
	if err == nil {
		t.Fatal("expected 401 error")
	}
	if !IsAuthentication(err) {
		t.Fatalf("expected auth error, got %v", err)
	}
	var apiErr *Error
	if !asError(err, &apiErr) || apiErr.RequestID != "req_abc123" {
		t.Fatalf("missing request id: %v", err)
	}

	// 4) pagination iterator
	it := c.Search.Iterate(ctx)
	count := 0
	for it.Next() {
		count++
	}
	if it.Err() != nil {
		t.Fatalf("iterate: %v", it.Err())
	}
	if count != 1 {
		t.Fatalf("expected 1 item, got %d", count)
	}

	// 5) chat completion stream
	cs, err := c.Chat.Completions.CreateStream(ctx, &ChatParams{
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("create stream: %v", err)
	}
	var content string
	for cs.Next() {
		choices := cs.Current()["choices"].([]any)
		delta := choices[0].(map[string]any)["delta"].(map[string]any)
		content += delta["content"].(string)
	}
	cs.Close()
	if cs.Err() != nil {
		t.Fatalf("chat stream err: %v", cs.Err())
	}
	if content != "Hello" {
		t.Fatalf("chat stream content: %q", content)
	}

	// 6) events stream
	es, err := c.Search.StreamEvents(ctx, "abc")
	if err != nil {
		t.Fatalf("stream events: %v", err)
	}
	var types []string
	for es.Next() {
		types = append(types, es.Current()["type"].(string))
	}
	es.Close()
	if len(types) != 2 || types[0] != "run.progress" || types[1] != "run.completed" {
		t.Fatalf("events: %v", types)
	}
}

// helpers exposed only to the test
func (c *Client) do2(ctx context.Context, method, path string) (Result, error) {
	var out Result
	err := c.do(ctx, method, path, nil, map[string]any{}, &out)
	return out, err
}

func asError(err error, target **Error) bool {
	for err != nil {
		if e, ok := err.(*Error); ok {
			*target = e
			return true
		}
		type unwrapper interface{ Unwrap() error }
		if u, ok := err.(unwrapper); ok {
			err = u.Unwrap()
		} else {
			return false
		}
	}
	return false
}
