# Scout Go SDK

Official Go SDK for the [Scout](https://usescout.sh) web-intelligence API: search, scrape, screenshot, extract, crawl, and company enrichment.

- Built on the standard library (`net/http`, `encoding/json`).
- `context.Context`-first methods, functional options, and a single error type.
- Automatic retries with backoff and jitter, configurable timeouts, and idempotency keys on writes.

## Requirements

- Go 1.21+

## Installation

```sh
go get github.com/Scout-AI-Labs/scout-go
```

## Authentication

Generate an API key at [platform.usescout.sh/settings](https://platform.usescout.sh/settings). The client reads `SCOUT_API_KEY` from the environment by default:

```go
client := scout.NewClient()                         // uses SCOUT_API_KEY
client := scout.NewClient(scout.WithAPIKey("sk_...")) // or pass it explicitly
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"

	scout "github.com/Scout-AI-Labs/scout-go"
)

func main() {
	client := scout.NewClient()
	res, err := client.Search.Create(context.Background(), &scout.SearchParams{
		Queries: []string{"best climate tech startups 2026"},
		Depth:   scout.String("standard"),
		Country: scout.String("us"),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
}
```

## Examples

```go
ctx := context.Background()

// Scrape a page to Markdown
page, _ := client.Page.Markdown(ctx, &scout.PageMarkdownParams{URL: "https://example.com"})

// Screenshot
shot, _ := client.Page.Screenshot(ctx, &scout.PageScreenshotParams{
	URL: "https://example.com", FullPage: scout.Bool(true), Format: scout.String("png"),
})

// Structured extraction
data, _ := client.Extract.Create(ctx, &scout.ExtractParams{
	URLs:         []string{"https://example.com/pricing"},
	OutputSchema: map[string]any{"type": "object", "properties": map[string]any{"plans": map[string]any{"type": "array"}}},
})

// Company enrichment + logo
company, _ := client.Company.Enrich(ctx, &scout.DomainParams{Domain: "stripe.com"})
logo, _ := client.Company.Logo(ctx, &scout.LogoParams{Domain: "stripe.com", Format: scout.String("svg")})

// Crawl a site
crawl, _ := client.Site.Crawl(ctx, &scout.SiteCrawlParams{StartURL: "https://example.com", MaxPages: scout.Int(50)})
```

Optional fields are pointers; set them with the `scout.String`/`scout.Int`/`scout.Bool`/`scout.Float64` helpers, or leave them nil to omit.

## Error handling

Non-2xx responses return a `*scout.Error` carrying `StatusCode`, `Code`, `Message`, `RequestID`, and the parsed `Body`. Use the `IsX` predicates or `errors.As`:

```go
res, err := client.Search.Create(ctx, params)
if err != nil {
	switch {
	case scout.IsRateLimited(err):
		log.Println("slow down")
	case scout.IsAuthentication(err):
		log.Println("check your API key")
	default:
		var apiErr *scout.Error
		if errors.As(err, &apiErr) {
			log.Printf("HTTP %d (req %s): %s", apiErr.StatusCode, apiErr.RequestID, apiErr.Message)
		}
	}
}
```

| Status | Predicate |
|--------|-----------|
| 400 | `IsBadRequest` |
| 401 | `IsAuthentication` |
| 402 | `IsInsufficientCredits` |
| 403 | `IsPermissionDenied` |
| 404 | `IsNotFound` |
| 429 | `IsRateLimited` |
| ≥500 | `IsServerError` |

## Retries & timeouts

Transient failures (connection errors, 408/409/429/5xx) are retried automatically, **2 times by default**, with exponential backoff and jitter, honoring `Retry-After`. Write methods send an auto-generated `Idempotency-Key`.

```go
client := scout.NewClient(
	scout.WithTimeout(30*time.Second),
	scout.WithMaxRetries(4),
)
```

Per-request cancellation/deadline is handled through the `context.Context` you pass to each method.

## Auto-pagination

```go
it := client.Search.Iterate(ctx)
for it.Next() {
	item := it.Item()
	fmt.Println(item)
}
if err := it.Err(); err != nil {
	log.Fatal(err)
}
```

## Streaming

Stream chat completions and live run progress (search, jobs, find-all, monitors) with a `*scout.Stream`:

```go
st, err := client.Chat.Completions.CreateStream(ctx, &scout.ChatParams{
	Messages: []scout.ChatMessage{{Role: "user", Content: "Summarize EU AI regulation."}},
	WebSearch: scout.Bool(true),
})
if err != nil {
	log.Fatal(err)
}
defer st.Close()
for st.Next() {
	chunk := st.Current() // a scout.Result; read choices[0].delta.content
	fmt.Println(chunk)
}
if err := st.Err(); err != nil {
	log.Fatal(err)
}

// Live progress events:
es, _ := client.Search.StreamEvents(ctx, searchID)
defer es.Close()
for es.Next() {
	fmt.Println(es.Current()["type"])
}
```

`StreamEvents` is also on `Jobs`, `Lists.Runs`, and `Monitors`. Cancel a stream via the `context.Context` you pass in.

## Versioning

This SDK follows [SemVer](https://semver.org/) and sends the targeted Scout API version on every request; see [`CHANGELOG.md`](./CHANGELOG.md). API reference renders on [pkg.go.dev](https://pkg.go.dev/github.com/Scout-AI-Labs/scout-go).

## Contributing

Issues and pull requests are welcome at [Scout-AI-Labs/scout-go](https://github.com/Scout-AI-Labs/scout-go).

## License

[MIT](./LICENSE)
