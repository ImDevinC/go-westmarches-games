# westmarches-go

A Go client library for the [West Marches Community API](https://www.westmarches.games/api/v1/docs).

The client is fully typed, supports every documented endpoint, respects the
API's rate limits (both client-side and via the server's `X-RateLimit-*`
headers), and includes helpers for paginating through large collections.

## Features

- **All endpoints** — every operation in the OpenAPI spec (characters,
  adventures, wiki, rewards, marketplaces, currencies).
- **Rate limiting** — a built-in token-bucket limiter (100 requests/minute by
  default, matching the server) plus parsed `X-RateLimit-*` headers exposed on
  every response and on the client.
- **Pagination** — typed `Page[T]` results with `Walk` (an idiomatic,
  callback-driven pagination helper) and a `CollectAll` convenience wrapper
  that materializes the full result set.
- **Typed errors** — `*APIError` with the HTTP status code, the server's error
  message, and helpers like `IsNotFound()` / `IsRateLimited()`.
- **Zero dependencies** — only the Go standard library.

## Installation

```sh
go get github.com/ImDevinC/go-westmarches-games
```

Requires Go 1.21 or later.

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ImDevinC/go-westmarches-games"
)

func main() {
	ctx := context.Background()
	client := westmarches.NewClient("wm_your_api_key_here")

	// Page through every character, processing results as they arrive.
	err := westmarches.Walk(ctx, 500, func(ctx context.Context, page, pageSize int) (*westmarches.Page[westmarches.CharacterSummary], error) {
		return client.ListCharacters(ctx, westmarches.ListOptions{Page: page, PageSize: pageSize})
	}, func(page *westmarches.Page[westmarches.CharacterSummary]) error {
		for _, c := range page.Data {
			fmt.Printf("%s (level %d)\n", c.Name, c.Level)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

## Authentication

API keys are created in your community settings under **Settings → API Keys**
and require Legendary tier (boost level 8). Pass your key to `NewClient`; it
is sent as a `Bearer` token on every request.

```go
client := westmarches.NewClient("wm_xxxxx")
```

The community is automatically determined from the API key — no community slug
is needed.

## Rate limiting

The server limits requests to **100 per minute** per API key. The client:

1. **Proactively throttles** with a built-in token-bucket limiter configured to
   the same 100 requests/minute by default (disable with `WithRateLimit(0)`,
   adjust with `WithRateLimit(n)`).
2. **Tracks the server's reported limits** from the `X-RateLimit-Limit`,
   `X-RateLimit-Remaining`, and `X-RateLimit-Reset` response headers. Read the
   most recent status at any time:

```go
if rl := client.RateLimit(); rl != nil {
	fmt.Printf("%d/%d requests remaining, resets at %s\n", rl.Remaining, rl.Limit, rl.Reset)
}
```

If the server returns `429 Too Many Requests`, the error is an `*APIError` with
`IsRateLimited() == true`.

## Pagination

List endpoints (`ListCharacters`, `ListAdventures`, `ListArticles`) return a
`*Page[T]`:

```go
page, err := client.ListCharacters(ctx, westmarches.ListOptions{Page: 2, PageSize: 100})
// page.Data       []CharacterSummary
// page.Pagination Total, TotalPages, Page, PageSize
```

To page through the entire result set, use `Walk`. It fetches each page in
order and invokes your callback as pages arrive, so results are processed
incrementally instead of being buffered all at once. Return
`westmarches.ErrStopIteration` from the callback to stop early:

```go
err := westmarches.Walk(ctx, 500, func(ctx context.Context, page, pageSize int) (*westmarches.Page[westmarches.CharacterSummary], error) {
	return client.ListCharacters(ctx, westmarches.ListOptions{Page: page, PageSize: pageSize})
}, func(page *westmarches.Page[westmarches.CharacterSummary]) error {
	for _, c := range page.Data {
		fmt.Printf("%s (level %d)\n", c.Name, c.Level)
	}
	return nil // or westmarches.ErrStopIteration to stop early
})
if err != nil {
	log.Fatal(err)
}
```

If you want the complete result set materialized as a slice, use the
`CollectAll` convenience wrapper, which is built on top of `Walk`:

```go
allAdventures, err := westmarches.CollectAll(ctx, 500, func(ctx context.Context, page, pageSize int) (*westmarches.Page[westmarches.AdventureSummary], error) {
	return client.ListAdventures(ctx, westmarches.ListOptions{Page: page, PageSize: pageSize})
})
```

> Note: `pageSize` is capped at 500 by the server. The wiki articles endpoint
> defaults to 100 items per page.

## Error handling

Non-2xx responses are returned as `*APIError`:

```go
_, err := client.GetCharacter(ctx, "missing")
if err != nil {
	var apiErr *westmarches.APIError
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.IsNotFound():
			// 404
		case apiErr.IsUnauthorized():
			// 401
		case apiErr.IsForbidden():
			// 403
		case apiErr.IsRateLimited():
			// 429
		}
	}
}
```

## Endpoint coverage

All operations from the OpenAPI spec are implemented:

| Method | Path | Client method |
| --- | --- | --- |
| GET | `/characters` | `ListCharacters` |
| GET | `/characters/{characterId}` | `GetCharacter` |
| GET | `/characters/{characterId}/stats` | `GetCharacterStats` |
| POST | `/characters/{characterId}/rewards` | `DistributeReward` |
| PATCH | `/characters/{characterId}/status` | `UpdateCharacterStatus` |
| POST | `/characters/{characterId}/approve` | `ApproveCharacter` |
| PATCH | `/characters/{characterId}/inventory/{itemId}` | `UpdateInventoryItem` |
| POST | `/characters/{characterId}/inventory/{itemId}/consume` | `ConsumeInventoryItem` |
| POST | `/characters/{characterId}/inventory/{itemId}/sell` | `SellInventoryItem` |
| POST | `/characters/{characterId}/inventory/{itemId}/transfer` | `TransferInventoryItem` |
| PUT | `/characters/{characterId}/data-tables/{dataTableId}` | `UpdateDataTableRows` |
| GET | `/adventures` | `ListAdventures` |
| GET | `/adventures/{adventureId}` | `GetAdventure` |
| GET | `/wiki/articles` | `ListArticles` |
| GET | `/wiki/articles/{articleId}` | `GetArticle` |
| POST | `/rewards` | `DistributeBulkRewards` |
| GET | `/marketplaces` | `ListMarketplaces` |
| GET | `/currencies` | `ListCurrencies` |

## Examples

### Reward a character

```go
resp, err := client.DistributeReward(ctx, "char_abc", westmarches.RewardRequest{
	Experience: 100,
	Currencies: map[string]float64{"cl_currency_id": 50},
	Reason:     "Quest completion bonus",
	Items: []westmarches.RewardItem{
		{Name: "Healing Potion", Quantity: 2, IsConsumable: true},
	},
})
```

### Bulk rewards

```go
resp, err := client.DistributeBulkRewards(ctx, westmarches.BulkRewardsRequest{
	Rewards: []westmarches.BulkRewardEntry{
		{CharacterID: "char_1", Experience: 100, Currencies: map[string]float64{"cl1": 50}},
		{CharacterID: "char_2", Experience: 100, Currencies: map[string]float64{"cl1": 50}},
	},
})
```

### Update a character's data table

```go
resp, err := client.UpdateDataTableRows(ctx, "char_abc", "dt_id", []westmarches.DataTableRowInput{
	{Values: map[string]any{"column-uuid": "Longsword", "column-uuid-2": 15}, RowIndex: 0},
})
```

Column IDs come from `dataTables[].columns[].id` on `GetCharacter`.

### Search the wiki

```go
page, err := client.ListArticles(ctx, westmarches.ListArticlesOptions{
	Search:   "Riverrun",
	Category: westmarches.ArticleCategoryLocation,
})
```

## Configuration

```go
client := westmarches.NewClient(
	"wm_your_api_key_here",
	westmarches.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
	westmarches.WithRateLimit(100),          // requests/minute; 0 disables
	westmarches.WithUserAgent("myapp/1.0"),  // default: westmarches-go/1.0.0
	westmarches.WithBaseURL("https://www.westmarches.games/api/v1"), // default
)
```

## Development

```sh
go build ./...
go vet ./...
go test ./...
```

## Releases

Releases are fully automated with [semantic-release](https://semantic-release.gitbook.io/).
Merges to `main` run the `Release` GitHub Actions workflow, which builds and
tests the library, then creates a `vX.Y.Z` git tag and a matching GitHub
release with auto-generated release notes. No container images are published.

The next version is derived from [conventional commits](https://www.conventionalcommits.org/):

- `feat` → minor bump
- `fix`, `perf`, `refactor` → patch bump
- `docs`, `style`, `test`, `chore`, `ci`, `build`, `revert` → no release

The `Commitlint` workflow enforces the conventional-commit format on every pull
request, so the release analysis always has reliable input. Commit messages
must use a lowercase type and one of the types listed above.

CI tooling lives in `package.json` (`npm ci` installs it). To preview the next
release locally:

```sh
GITHUB_TOKEN=<token> npx semantic-release --dry-run
```

## Documentation

See `AGENTS.md` for maintainer and agent conventions. The upstream API spec is
published at <https://www.westmarches.games/api/openapi.json> and rendered at
<https://www.westmarches.games/api/v1/docs>.