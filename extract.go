package scout

import "context"

// ExtractService covers multi-URL structured extraction.
type ExtractService struct{ client *Client }

// ExtractParams are the inputs to Create.
type ExtractParams struct {
	URLs          []string       `json:"urls"`
	Objective     *string        `json:"objective,omitempty"`
	SearchQueries []string       `json:"search_queries,omitempty"`
	FindViaSearch *bool          `json:"find_via_search,omitempty"`
	MaxCharsTotal *int           `json:"max_chars_total,omitempty"`
	MaxChars      *int           `json:"max_chars,omitempty"`
	OutputSchema  map[string]any `json:"output_schema,omitempty"`
}

// Create extracts structured data from one or more URLs. Provide an Objective
// or an OutputSchema (JSON Schema) to shape the result.
func (s *ExtractService) Create(ctx context.Context, params *ExtractParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/extract", nil, params, &out)
	return out, err
}
