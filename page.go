package scout

import "context"

// PageService covers single-page operations: markdown, html, screenshot,
// images, extract.
type PageService struct{ client *Client }

// PageMarkdownParams are the inputs to Markdown.
type PageMarkdownParams struct {
	URL      string `json:"url"`
	MaxChars *int   `json:"max_chars,omitempty"`
}

// Markdown fetches a page rendered to clean Markdown.
func (s *PageService) Markdown(ctx context.Context, params *PageMarkdownParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/page/markdown", nil, params, &out)
	return out, err
}

// PageHTMLParams are the inputs to HTML.
type PageHTMLParams struct {
	URL string `json:"url"`
}

// HTML fetches a page's HTML.
func (s *PageService) HTML(ctx context.Context, params *PageHTMLParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/page/html", nil, params, &out)
	return out, err
}

// PageScreenshotParams are the inputs to Screenshot.
type PageScreenshotParams struct {
	URL             string  `json:"url"`
	ViewportWidth   *int    `json:"viewport_width,omitempty"`
	ViewportHeight  *int    `json:"viewport_height,omitempty"`
	FullPage        *bool   `json:"full_page,omitempty"`
	Format          *string `json:"format,omitempty"`
	WaitMs          *int    `json:"wait_ms,omitempty"`
	Inline          *bool   `json:"inline,omitempty"`
	ElementSelector *string `json:"element_selector,omitempty"`
	DismissOverlays *bool   `json:"dismiss_overlays,omitempty"`
}

// Screenshot captures a screenshot of a page.
func (s *PageService) Screenshot(ctx context.Context, params *PageScreenshotParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/page/screenshot", nil, params, &out)
	return out, err
}

// PageImagesParams are the inputs to Images.
type PageImagesParams struct {
	URL             string  `json:"url"`
	MaxImages       *int    `json:"max_images,omitempty"`
	IncludeDataURIs *bool   `json:"include_data_uris,omitempty"`
	Mode            *string `json:"mode,omitempty"`
}

// Images extracts the images on a page.
func (s *PageService) Images(ctx context.Context, params *PageImagesParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/page/images", nil, params, &out)
	return out, err
}

// PageExtractParams are the inputs to Extract.
type PageExtractParams struct {
	URL string `json:"url"`
}

// Extract performs structured extraction scoped to a single page.
func (s *PageService) Extract(ctx context.Context, params *PageExtractParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/page/extract", nil, params, &out)
	return out, err
}
