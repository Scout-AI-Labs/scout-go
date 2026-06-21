package scout

import "context"

// SiteService covers whole-site operations: crawl and sitemap discovery.
type SiteService struct{ client *Client }

// SiteCrawlParams are the inputs to Crawl.
type SiteCrawlParams struct {
	StartURL         string   `json:"start_url"`
	MaxPages         *int     `json:"max_pages,omitempty"`
	MaxDepth         *int     `json:"max_depth,omitempty"`
	SameHostOnly     *bool    `json:"same_host_only,omitempty"`
	IncludePatterns  []string `json:"include_patterns,omitempty"`
	ExcludePatterns  []string `json:"exclude_patterns,omitempty"`
	FollowSubdomains *bool    `json:"followSubdomains,omitempty"`
}

// Crawl crawls a site from StartURL.
func (s *SiteService) Crawl(ctx context.Context, params *SiteCrawlParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/site/crawl", nil, params, &out)
	return out, err
}

// SiteMapParams are the inputs to Map.
type SiteMapParams struct {
	StartURL string `json:"start_url"`
	MaxPages *int   `json:"max_pages,omitempty"`
}

// Map discovers a site's URLs (sitemap) from StartURL.
func (s *SiteService) Map(ctx context.Context, params *SiteMapParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/site/map", nil, params, &out)
	return out, err
}
