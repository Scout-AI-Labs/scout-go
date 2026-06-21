package scout

import "context"

// ProductsService covers product extraction from storefronts.
type ProductsService struct{ client *Client }

// ProductsParams are the inputs to Extract.
type ProductsParams struct {
	URL              string  `json:"url"`
	MaxPages         *int    `json:"max_pages,omitempty"`
	MaxDepth         *int    `json:"max_depth,omitempty"`
	Instructions     *string `json:"instructions,omitempty"`
	FollowSubdomains *bool   `json:"followSubdomains,omitempty"`
}

// Extract crawls a store and extracts its products.
func (s *ProductsService) Extract(ctx context.Context, params *ProductsParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/products", nil, params, &out)
	return out, err
}

// ProductOneParams are the inputs to One.
type ProductOneParams struct {
	URL string `json:"url"`
}

// One extracts a single product from one product-detail URL.
func (s *ProductsService) One(ctx context.Context, params *ProductOneParams) (Result, error) {
	var out Result
	err := s.client.do(ctx, "POST", "/v1/products/one", nil, params, &out)
	return out, err
}
