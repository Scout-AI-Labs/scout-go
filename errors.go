package scout

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Error is returned for any non-2xx HTTP response. Inspect StatusCode, or use
// the IsX predicates, to branch on the failure.
//
//	var apiErr *scout.Error
//	if errors.As(err, &apiErr) && apiErr.StatusCode == 429 { ... }
type Error struct {
	StatusCode int            // HTTP status, 0 for client-side errors
	Code       string         // machine-readable code from the body, if any
	Message    string         // human-readable message
	RequestID  string         // x-request-id header, for support
	Body       map[string]any // parsed JSON error body, if any
	Header     http.Header    // response headers
}

func (e *Error) Error() string {
	if e.StatusCode == 0 {
		return "scout: " + e.Message
	}
	return fmt.Sprintf("scout: HTTP %d: %s", e.StatusCode, e.Message)
}

// connError wraps a transport-level failure (DNS, refused, timeout, cancel).
type connError struct{ err error }

func (e *connError) Error() string { return "scout: connection error: " + e.err.Error() }
func (e *connError) Unwrap() error { return e.err }

func errorFromResponse(resp *http.Response) *Error {
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	e := &Error{
		StatusCode: resp.StatusCode,
		RequestID:  resp.Header.Get("X-Request-Id"),
		Header:     resp.Header,
		Message:    fmt.Sprintf("Scout API returned HTTP %d", resp.StatusCode),
	}
	var parsed map[string]any
	if json.Unmarshal(raw, &parsed) == nil {
		e.Body = parsed
		if msg := messageFromBody(parsed); msg != "" {
			e.Message = msg
		}
		if code, ok := parsed["code"].(string); ok {
			e.Code = code
		}
	} else if len(raw) > 0 {
		e.Message = string(raw)
	}
	return e
}

func messageFromBody(body map[string]any) string {
	for _, key := range []string{"detail", "error", "message"} {
		switch v := body[key].(type) {
		case string:
			return v
		case map[string]any:
			if msg, ok := v["message"].(string); ok {
				return msg
			}
		}
	}
	return ""
}

func decode(resp *http.Response, out any) error {
	defer resp.Body.Close()
	if out == nil || resp.StatusCode == http.StatusNoContent {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil && err != io.EOF {
		return &Error{StatusCode: resp.StatusCode, Message: "failed to decode response: " + err.Error()}
	}
	return nil
}

func isRetriable(err error) bool {
	var ce *connError
	if errors.As(err, &ce) {
		return true
	}
	var apiErr *Error
	if errors.As(err, &apiErr) && apiErr.StatusCode != 0 {
		return retryStatuses[apiErr.StatusCode]
	}
	return false
}

// StatusCode returns the HTTP status from a *scout.Error, or 0.
func StatusCode(err error) int {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode
	}
	return 0
}

// IsBadRequest reports whether err is a 400.
func IsBadRequest(err error) bool { return StatusCode(err) == 400 }

// IsAuthentication reports whether err is a 401.
func IsAuthentication(err error) bool { return StatusCode(err) == 401 }

// IsInsufficientCredits reports whether err is a 402.
func IsInsufficientCredits(err error) bool { return StatusCode(err) == 402 }

// IsPermissionDenied reports whether err is a 403.
func IsPermissionDenied(err error) bool { return StatusCode(err) == 403 }

// IsNotFound reports whether err is a 404.
func IsNotFound(err error) bool { return StatusCode(err) == 404 }

// IsRateLimited reports whether err is a 429.
func IsRateLimited(err error) bool { return StatusCode(err) == 429 }

// IsServerError reports whether err is a 5xx.
func IsServerError(err error) bool { return StatusCode(err) >= 500 }
