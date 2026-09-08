package westmarches

import "context"

// ListMarketplaces returns all marketplaces for the community tied to the API
// key, each with all their items. Use the IsDisabled field to determine
// whether an item is currently available for purchase. Item UUIDs are
// intentionally omitted — they are regenerated whenever a marketplace is saved
// and must not be used as stable identifiers.
func (c *Client) ListMarketplaces(ctx context.Context) ([]Marketplace, error) {
	data, err := c.do(ctx, "GET", "/marketplaces", nil)
	if err != nil {
		return nil, err
	}
	var out []Marketplace
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}
