package scout

import "context"

// CompanyService covers company enrichment: profiles, logos, fonts, industry
// codes, and styleguide.
type CompanyService struct{ client *Client }

// DomainParams are the inputs to the domain-based company endpoints.
type DomainParams struct {
	Domain string `json:"domain"`
}

// Enrich returns a full company profile from a domain.
func (s *CompanyService) Enrich(ctx context.Context, params *DomainParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/company", nil, params, &out)
	return out, err
}

// ByEmailParams are the inputs to ByEmail.
type ByEmailParams struct {
	Email string `json:"email"`
}

// ByEmail resolves a company from a work email address.
func (s *CompanyService) ByEmail(ctx context.Context, params *ByEmailParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/company/by-email", nil, params, &out)
	return out, err
}

// ByNameParams are the inputs to ByName.
type ByNameParams struct {
	Name string `json:"name"`
}

// ByName resolves a company from its name.
func (s *CompanyService) ByName(ctx context.Context, params *ByNameParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/company/by-name", nil, params, &out)
	return out, err
}

// ByTickerParams are the inputs to ByTicker.
type ByTickerParams struct {
	Ticker string `json:"ticker"`
}

// ByTicker resolves a company from a stock ticker.
func (s *CompanyService) ByTicker(ctx context.Context, params *ByTickerParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/company/by-ticker", nil, params, &out)
	return out, err
}

// Simple returns a condensed company profile (faster, fewer fields).
func (s *CompanyService) Simple(ctx context.Context, params *DomainParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/company/simple", nil, params, &out)
	return out, err
}

// Fonts returns the brand fonts detected on the company's site.
func (s *CompanyService) Fonts(ctx context.Context, params *DomainParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/company/fonts", nil, params, &out)
	return out, err
}

// Styleguide returns the brand styleguide (colors, typography, logos).
func (s *CompanyService) Styleguide(ctx context.Context, params *DomainParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/company/styleguide", nil, params, &out)
	return out, err
}

// LogoParams are the inputs to Logo.
type LogoParams struct {
	Domain  string  `json:"domain"`
	Mode    *string `json:"mode,omitempty"`
	Format  *string `json:"format,omitempty"`
	Variant *string `json:"variant,omitempty"`
}

// Logo returns company logo metadata.
func (s *CompanyService) Logo(ctx context.Context, params *LogoParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/company/logo", nil, params, &out)
	return out, err
}
