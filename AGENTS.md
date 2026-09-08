# AGENTS.md

This file is a guide for AI coding agents (and humans) working in this
repository. Read it before making changes.

## Project overview

`westmarches-go` is a Go client library for the **West Marches Community API**
(<https://www.westmarches.games/api/v1/docs>). It is a thin, fully typed
wrapper over the REST API: every endpoint from the OpenAPI spec is exposed as a
method on `*Client`, request/response payloads are represented as Go structs,
and the client handles authentication, rate limiting, and pagination helpers.

**Source of truth:** the upstream OpenAPI spec is published at
<https://www.westmarches.games/api/openapi.json> (rendered at
<https://www.westmarches.games/api/v1/docs>). When the API changes, regenerate
types and methods from that spec.

## Repository layout

```
client.go           Client, Options/With* functional options, request plumbing (do),
                    rate-limit header parsing, base URL handling
errors.go           *APIError, sentinel errors
ratelimit.go        Client-side token-bucket rate limiter
types.go            All API schema structs (characters, adventures, wiki, rewards,
                    marketplace, currencies, D&D Beyond stats, etc.)
pagination.go       ListOptions, Page[T], Walk/CollectAll pagination helpers
characters.go       Character endpoints
adventures.go       Adventure endpoints
wiki.go             Wiki article endpoints
rewards.go          Bulk rewards endpoint
marketplace.go      Marketplace endpoints
currencies.go       Currency endpoints
*_test.go           Unit tests using httptest (see Testing)
.github/workflows/  CI: commitlint (PRs) and release (pushes to main); also
                    holds the CI tooling: package.json/package-lock.json
                    (commitlint + semantic-release deps), commitlint.config.js
                    (conventional-commit rules enforced on PRs), and
                    .releaserc.js (semantic-release config: vX.Y.Z tags +
                    GitHub releases)
README.md           User-facing documentation
AGENTS.md           This file
```

## Conventions

- **Language:** Go (module `github.com/ImDevinC/go-westmarches-games`,
  requires Go 1.21+). Standard library only — do not add external dependencies.
- **Package layout:** a single package `westmarches`. Split implementation by
  resource area (one file per endpoint group).
- **Method naming:** endpoint methods use the `operationId` from the OpenAPI
  spec converted to Go `CamelCase` (e.g. `listCharacters` → `ListCharacters`,
  `distributeBulkRewards` → `DistributeBulkRewards`).
- **Request methods:** every method takes `ctx context.Context` as the first
  argument. Body-bearing methods take a request struct; list methods take a
  `ListOptions` (or `ListArticlesOptions`) and return `*Page[T]`; single-item
  methods return a typed pointer and an error.
- **Response envelope:** the API wraps data in `{"success": true, "data": ...,
  "pagination": {...}}`. Use `decodeData` for single objects and the local
  `jsonUnmarshal` for envelope structs. Non-2xx responses are converted to
  `*APIError` by `Client.do` — endpoint methods never inspect raw bodies.
- **Path escaping:** always `url.PathEscape` user-supplied path segments.
- **Nil vs empty:** use pointers (`*string`, `*int`, `*bool`) for nullable
  fields exactly as the OpenAPI spec marks them nullable.
- **Formatting:** run `gofmt` on all changed files. `go vet ./...` must pass.
- **Commit style:** conventional commits (`feat:`, `fix:`, `docs:`, `test:`,
  `refactor:`, `chore:`).

## Adding or changing an endpoint

1. **Read the OpenAPI spec** at <https://www.westmarches.games/api/openapi.json>
   (or the rendered docs) and identify the operation, its schema, and its
   error responses.
2. **Add/update types** in `types.go` (and request/response structs in the
   relevant resource file if they are endpoint-specific).
3. **Add the method** in the matching resource file (e.g. a new adventure
   endpoint goes in `adventures.go`). Follow the patterns of existing methods:
   build the path, call `c.do`, decode with `decodeData`/`jsonUnmarshal`.
4. **Add tests** in the matching `*_test.go` file using `newTestClient` and
   `httptest`. Cover the happy path, query parameters, request bodies, error
   responses, and rate-limit headers where relevant.
5. **Update the endpoint coverage table** in `README.md` (Method / Path /
   Client method columns) and add a usage example if the endpoint is notable.
6. **Update this file** (`AGENTS.md`) if the change alters the repository
   layout, conventions, or testing instructions.

> **Always keep both `README.md` and `AGENTS.md` up to date.** If a change
> affects how the library is used or how the codebase is structured, update
> both documents in the same commit. Do not leave them stale.

## Rate limiting

The server limits requests to **100 per minute** per API key and reports
`X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset` headers.
The client:

- throttles client-side via a token-bucket limiter in `ratelimit.go`
  (configured with `WithRateLimit(n)`; `0` disables it);
- stores the latest server-reported status, readable via `Client.RateLimit()`.

If you change rate-limit behaviour, update `README.md` (Rate limiting section)
and keep tests in `ratelimit_test.go` and `client_test.go` in sync.

## Pagination

List endpoints accept `page` (default 1) and `pageSize` (default 500, max 500;
wiki articles default to 100) query parameters and return a `pagination`
object. `Page[T]`, `Walk`, and `CollectAll` in `pagination.go` provide typed
access and full-result walking. `Walk` is the idiomatic primitive: it drives
the page loop internally and invokes a callback per page, supporting early
termination via `ErrStopIteration`. `CollectAll` is a convenience wrapper that
materializes the full result set on top of `Walk`. If the API changes
pagination semantics, update `pagination.go`, its tests, and the README.

## Testing

```sh
go build ./...
go vet ./...
go test ./...
go test -cover ./...     # currently ~83% statement coverage
```

- Tests use `net/http/httptest` — no network access, no mocks library.
- `newTestClient(t, handler)` returns a client pointed at a test server that
  records requests; assert method, path, query, headers, and body.
- Rate limiting is disabled in tests (`WithRateLimit(0)`) so tests are fast;
  the limiter itself is tested directly in `ratelimit_test.go`.

## Definition of done

- All endpoints from the current OpenAPI spec are implemented.
- `go build ./...`, `go vet ./...`, and `go test ./...` pass.
- README documents new user-facing behaviour.
- AGENTS.md reflects any structural/convention changes.