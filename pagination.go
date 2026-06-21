package scout

import (
	"context"
	"net/url"
	"strconv"
)

// Iterator lazily walks every item across the pages of a list endpoint.
//
//	it := client.Search.Iterate(ctx)
//	for it.Next() {
//	    item := it.Item()
//	}
//	if err := it.Err(); err != nil { ... }
type Iterator struct {
	client *Client
	ctx    context.Context
	path   string
	limit  int
	offset int

	buf  []any
	pos  int
	done bool
	err  error
}

func newIterator(ctx context.Context, client *Client, path string, limit int) *Iterator {
	if limit <= 0 {
		limit = 50
	}
	return &Iterator{client: client, ctx: ctx, path: path, limit: limit}
}

// Next advances to the next item, fetching another page when needed. It
// returns false when the items are exhausted or an error occurs (see Err).
func (it *Iterator) Next() bool {
	if it.err != nil {
		return false
	}
	if it.pos < len(it.buf) {
		it.pos++
		return true
	}
	if it.done {
		return false
	}
	q := url.Values{}
	q.Set("limit", strconv.Itoa(it.limit))
	q.Set("offset", strconv.Itoa(it.offset))
	var page any
	if err := it.client.do(it.ctx, "GET", it.path, q, nil, &page); err != nil {
		it.err = err
		return false
	}
	items := extractItems(page)
	if len(items) < it.limit {
		it.done = true
	}
	it.offset += len(items)
	it.buf = items
	it.pos = 0
	if len(items) == 0 {
		return false
	}
	it.pos++
	return true
}

// Item returns the current item. Call after Next returns true.
func (it *Iterator) Item() any {
	if it.pos == 0 || it.pos > len(it.buf) {
		return nil
	}
	return it.buf[it.pos-1]
}

// Err returns the first error encountered while paging.
func (it *Iterator) Err() error { return it.err }

var commonItemKeys = []string{"items", "data", "results", "searches", "runs", "jobs", "monitors"}

func extractItems(payload any) []any {
	switch v := payload.(type) {
	case []any:
		return v
	case map[string]any:
		for _, key := range commonItemKeys {
			if arr, ok := v[key].([]any); ok {
				return arr
			}
		}
		for _, val := range v {
			if arr, ok := val.([]any); ok {
				return arr
			}
		}
	}
	return nil
}
