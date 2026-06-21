package scout

import "context"

// ChatService covers OpenAI-compatible chat completions, optionally grounded
// with web search.
type ChatService struct {
	client *Client
	// Completions creates chat completions.
	Completions *ChatCompletionsService
}

// ChatCompletionsService creates chat completions.
type ChatCompletionsService struct{ client *Client }

// ChatMessage is a single message in a chat completion request.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatParams are the inputs to Create. Shape mirrors the OpenAI Chat
// Completions API; set WebSearch to ground the answer in live results.
type ChatParams struct {
	Messages    []ChatMessage `json:"messages"`
	Model       *string       `json:"model,omitempty"`
	Stream      *bool         `json:"stream,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
	WebSearch   *bool         `json:"web_search,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
}

// Create creates a chat completion.
func (s *ChatCompletionsService) Create(ctx context.Context, params *ChatParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/chat/completions", nil, params, &out)
	return out, err
}
