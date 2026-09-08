package westmarches

import "context"

// ListCurrencies returns all currencies defined for the community tied to the
// API key. Use these IDs when providing currencies in the rewards endpoints.
func (c *Client) ListCurrencies(ctx context.Context) ([]Currency, error) {
	data, err := c.do(ctx, "GET", "/currencies", nil)
	if err != nil {
		return nil, err
	}
	var out []Currency
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}
