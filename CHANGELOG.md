# Changelog

All notable changes to this project are documented here. This project adheres
to [Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-06-21

Initial release.

- Client built on the standard library (`net/http`, `encoding/json`).
- Full coverage of the Scout REST API across `Search`, `Page`, `Extract`, `Company`, `Lists`, `Products`, `Site`, `Jobs`, `Monitors`, `Chat`.
- `context.Context`-first methods, functional options (`WithAPIKey`, `WithBaseURL`, `WithTimeout`, `WithMaxRetries`, `WithHTTPClient`, `WithHeader`).
- Single `*scout.Error` type plus `IsRateLimited`/`IsAuthentication`/... predicates.
- Automatic retries with exponential backoff + jitter, honoring `Retry-After`.
- Auto-pagination via `Iterator`.
