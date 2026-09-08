package westmarches

import (
	"context"
	"net/url"
)

// ListAdventures returns all published (non-draft) adventures in the community
// tied to the API key, ordered by start time descending. Pass a closure
// wrapping this method to Walk or CollectAll to retrieve every page in one
// call.
func (c *Client) ListAdventures(ctx context.Context, opts ListOptions) (*Page[AdventureSummary], error) {
	data, err := c.do(ctx, "GET", "/adventures"+queryFromOptions(opts, nil), nil)
	if err != nil {
		return nil, err
	}
	var env struct {
		Success    bool               `json:"success"`
		Data       []AdventureSummary `json:"data"`
		Pagination Pagination         `json:"pagination"`
	}
	if err := jsonUnmarshal(data, &env); err != nil {
		return nil, err
	}
	return &Page[AdventureSummary]{Data: env.Data, Pagination: env.Pagination}, nil
}

// GetAdventure returns detailed information for a single adventure. Only notes
// with public content are included.
func (c *Client) GetAdventure(ctx context.Context, adventureID string) (*AdventureDetail, error) {
	data, err := c.do(ctx, "GET", "/adventures/"+url.PathEscape(adventureID), nil)
	if err != nil {
		return nil, err
	}
	var out AdventureDetail
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
