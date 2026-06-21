package scout

// Pointer helpers for optional fields. A nil pointer is omitted from the
// request body (the structs use `omitempty`), so use these to set a value:
//
//	&scout.SearchParams{Queries: []string{"x"}, Country: scout.String("us")}

// String returns a pointer to v.
func String(v string) *string { return &v }

// Int returns a pointer to v.
func Int(v int) *int { return &v }

// Bool returns a pointer to v.
func Bool(v bool) *bool { return &v }

// Float64 returns a pointer to v.
func Float64(v float64) *float64 { return &v }
