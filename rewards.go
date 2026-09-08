package westmarches

import "context"

// DistributeBulkRewards awards experience and/or currencies to multiple
// characters in a single request. Rewards with identical values are merged
// into a single Discord notification. Requires write permission.
func (c *Client) DistributeBulkRewards(ctx context.Context, req BulkRewardsRequest) ([]RewardResponse, error) {
	data, err := c.do(ctx, "POST", "/rewards", req)
	if err != nil {
		return nil, err
	}
	var out []RewardResponse
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}
