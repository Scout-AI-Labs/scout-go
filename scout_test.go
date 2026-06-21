package scout

import (
	"context"
	"encoding/json"
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
