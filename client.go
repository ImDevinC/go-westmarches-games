// Package westmarches is a Go client library for the West Marches Community API
// (https://www.westmarches.games/api/v1/docs).
//
// The client is authenticated with a bearer API key, respects the documented
// rate limits, and provides typed helpers for every endpoint in the API.
package westmarches

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// DefaultBaseURL is the production base URL for the West Marches API.
const DefaultBaseURL = "https://www.westmarches.games/api/v1"

// DefaultUserAgent is used when no custom User-Agent is configured.
const DefaultUserAgent = "westmarches-go/1.0.0"

// DefaultRateLimit is the maximum number of requests per minute the server
// allows per API key, used by the built-in rate limiter.
const DefaultRateLimit = 100

const maxRateLimitRetries = 3

// Options holds the client configuration. Use the With* functional options to
// build a *Client.
type Options struct {
	HTTPClient        *http.Client
	BaseURL           string
	UserAgent         string
	RequestsPerMinute int // <= 0 disables the built-in rate limiter.
}

// Option configures a *Client. Functional options are applied in order.
type Option func(*Options)

// WithHTTPClient overrides the underlying *http.Client. This is useful for
// custom transports, proxies, TLS settings, and testing.
func WithHTTPClient(c *http.Client) Option {
	return func(o *Options) { o.HTTPClient = c }
}

// WithBaseURL overrides the default production base URL.
func WithBaseURL(u string) Option {
	return func(o *Options) { o.BaseURL = u }
}

// WithUserAgent overrides the default User-Agent header.
func WithUserAgent(ua string) Option {
	return func(o *Options) { o.UserAgent = ua }
}

// WithRateLimit configures the client-side rate limiter in requests per minute.
// Pass 0 to disable the built-in limiter entirely. The default is
// DefaultRateLimit (100), matching the server's documented limit.
func WithRateLimit(requestsPerMinute int) Option {
	return func(o *Options) { o.RequestsPerMinute = requestsPerMinute }
}

// Client is a West Marches API client. A client is safe for concurrent use by
// multiple goroutines.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	userAgent  string

	limiter *rateLimiter

	rlMu          sync.RWMutex
	lastRateLimit *RateLimit
}

// NewClient returns a new West Marches API client authenticated with apiKey.
//
//	client := westmarches.NewClient("wm_your_api_key_here")
func NewClient(apiKey string, opts ...Option) *Client {
	o := Options{
		HTTPClient:        &http.Client{},
		BaseURL:           DefaultBaseURL,
		UserAgent:         DefaultUserAgent,
		RequestsPerMinute: DefaultRateLimit,
	}
	for _, opt := range opts {
		opt(&o)
	}
	c := &Client{
		httpClient: o.HTTPClient,
		baseURL:    o.BaseURL,
		apiKey:     apiKey,
		userAgent:  o.UserAgent,
	}
	if o.RequestsPerMinute > 0 {
		c.limiter = newRateLimiter(o.RequestsPerMinute)
	}
	return c
}

// RateLimit returns the most recent rate-limit status reported by the server,
// parsed from the X-RateLimit-* response headers. It returns nil until the
// first request has completed, or if the server did not include the headers.
func (c *Client) RateLimit() *RateLimit {
	c.rlMu.RLock()
	defer c.rlMu.RUnlock()
	if c.lastRateLimit == nil {
		return nil
	}
	rl := *c.lastRateLimit
	return &rl
}

// do performs a request, applies rate limiting, parses response headers, and
// returns the raw response body. Non-2xx responses are returned as *APIError.
func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	var encodedBody []byte
	if body != nil {
		var err error
		encodedBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("westmarches: encoding request body: %w", err)
		}
	}

	for retry := 0; ; retry++ {
		if c.limiter != nil {
			if err := c.limiter.wait(ctx); err != nil {
				return nil, err
			}
		}

		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(encodedBody)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("westmarches: building request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if c.userAgent != "" {
			req.Header.Set("User-Agent", c.userAgent)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("westmarches: sending request: %w", err)
		}

		rl := parseRateLimit(resp.Header)
		c.setRateLimit(rl)

		data, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("westmarches: reading response: %w", readErr)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return data, nil
		}

		var e Error
		_ = json.Unmarshal(data, &e)
		retryAfter, hasRetryAfter := parseRetryAfter(resp.Header.Get("Retry-After"), time.Now())
		apiErr := &APIError{
			StatusCode:    resp.StatusCode,
			Success:       e.Success,
			Message:       e.Error,
			RateLimit:     rl,
			RetryAfter:    retryAfter,
			HasRetryAfter: hasRetryAfter,
		}
		if apiErr.IsRateLimited() && method == http.MethodGet && retry < maxRateLimitRetries {
			if err := waitForRetry(ctx, rateLimitRetryDelay(apiErr, retry)); err != nil {
				return nil, err
			}
			continue
		}
		return nil, apiErr
	}
}

func parseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second, true
	}
	if resetAt, err := http.ParseTime(value); err == nil {
		if delay := resetAt.Sub(now); delay > 0 {
			return delay, true
		}
		return 0, true
	}
	return 0, false
}

func rateLimitRetryDelay(apiErr *APIError, retry int) time.Duration {
	if apiErr.HasRetryAfter {
		return apiErr.RetryAfter
	}
	if apiErr.RateLimit != nil {
		if delay := time.Until(apiErr.RateLimit.Reset); delay > 0 {
			return delay + 100*time.Millisecond
		}
	}
	return time.Second << retry
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// setRateLimit stores the latest server-reported rate limit.
func (c *Client) setRateLimit(rl *RateLimit) {
	if rl == nil {
		return
	}
	c.rlMu.Lock()
	c.lastRateLimit = rl
	c.rlMu.Unlock()
}

// RateLimit is the server's reported rate-limit status for the current window,
// parsed from the X-RateLimit-* response headers.
type RateLimit struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// parseRateLimit reads the X-RateLimit-* headers. It returns nil if the headers
// are absent or unparseable.
func parseRateLimit(h http.Header) *RateLimit {
	limitStr := h.Get("X-RateLimit-Limit")
	remainingStr := h.Get("X-RateLimit-Remaining")
	resetStr := h.Get("X-RateLimit-Reset")
	if limitStr == "" && remainingStr == "" && resetStr == "" {
		return nil
	}
	rl := &RateLimit{}
	if v, err := strconv.Atoi(limitStr); err == nil {
		rl.Limit = v
	}
	if v, err := strconv.Atoi(remainingStr); err == nil {
		rl.Remaining = v
	}
	if v, err := strconv.ParseInt(resetStr, 10, 64); err == nil {
		rl.Reset = time.Unix(v, 0)
	}
	return rl
}

// decodeData unmarshals the "data" field of a success envelope into out.
// The pointer out must not be nil.
func decodeData(data []byte, out any) error {
	var env struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
		Error   string          `json:"error"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("westmarches: decoding response: %w", err)
	}
	if len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("westmarches: decoding response data: %w", err)
	}
	return nil
}

// queryFromOptions encodes a ListOptions (and any extra query parameters) into
// a URL query string.
func queryFromOptions(o ListOptions, extra url.Values) string {
	if o.Page <= 0 && o.PageSize <= 0 && len(extra) == 0 {
		return ""
	}
	q := url.Values{}
	if o.Page > 0 {
		q.Set("page", strconv.Itoa(o.Page))
	}
	if o.PageSize > 0 {
		q.Set("pageSize", strconv.Itoa(o.PageSize))
	}
	for k, vs := range extra {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return "?" + q.Encode()
}
