package westmarches

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestNewClientDefaults(t *testing.T) {
	c := NewClient("key")
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
	if c.userAgent != DefaultUserAgent {
		t.Errorf("userAgent = %q, want %q", c.userAgent, DefaultUserAgent)
	}
	if c.limiter == nil {
		t.Error("expected built-in rate limiter to be enabled by default")
	}
	if c.RateLimit() != nil {
		t.Error("expected RateLimit() to be nil before any request")
	}
}

func TestNewClientOptions(t *testing.T) {
	c := NewClient("key",
		WithBaseURL("https://example.com"),
		WithUserAgent("custom/1.0"),
		WithRateLimit(0),
	)
	if c.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
	if c.userAgent != "custom/1.0" {
		t.Errorf("userAgent = %q", c.userAgent)
	}
	if c.limiter != nil {
		t.Error("expected rate limiter to be disabled")
	}
}

func TestClientDoAuthAndHeaders(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data":    map[string]any{"hello": "world"},
		})
	})

	var out struct {
		Hello string `json:"hello"`
	}
	data, err := c.do(context.Background(), "GET", "/currencies", nil)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	if err := decodeData(data, &out); err != nil {
		t.Fatalf("decodeData: %v", err)
	}
	if out.Hello != "world" {
		t.Errorf("decoded hello = %q", out.Hello)
	}

	if len(*reqs) != 1 {
		t.Fatalf("got %d requests, want 1", len(*reqs))
	}
	req := (*reqs)[0]
	if got := req.Header.Get("Authorization"); got != "Bearer test-api-key" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer test-api-key")
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q", got)
	}
	if got := req.Header.Get("User-Agent"); got != DefaultUserAgent {
		t.Errorf("User-Agent = %q", got)
	}
}

func TestClientDoPOSTContentType(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 201, map[string]any{"success": true, "data": map[string]any{}})
	})
	if _, err := c.do(context.Background(), "POST", "/rewards", map[string]any{"rewards": []any{}}); err != nil {
		t.Fatalf("do: %v", err)
	}
	req := (*reqs)[0]
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q", got)
	}
}

func TestClientDoAPIError(t *testing.T) {
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"success": false, "error": "Character not found"}`))
	})

	_, err := c.do(context.Background(), "GET", "/characters/nope", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Character not found" {
		t.Errorf("Message = %q", apiErr.Message)
	}
	if !apiErr.IsNotFound() {
		t.Error("expected IsNotFound() to be true")
	}
	if apiErr.IsUnauthorized() || apiErr.IsForbidden() || apiErr.IsRateLimited() {
		t.Error("unexpected status helper result")
	}
}

func TestClientDoAPIErrorWithoutBody(t *testing.T) {
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	_, err := c.do(context.Background(), "GET", "/currencies", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsRateLimited() {
		t.Error("expected IsRateLimited() to be true")
	}
}

func TestClientRateLimitParsing(t *testing.T) {
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		rateLimitHeaders(w)
		writeJSON(t, w, 200, map[string]any{"success": true, "data": []any{}})
	})

	if _, err := c.ListCurrencies(context.Background()); err != nil {
		t.Fatalf("ListCurrencies: %v", err)
	}
	rl := c.RateLimit()
	if rl == nil {
		t.Fatal("expected RateLimit to be set")
	}
	if rl.Limit != 100 {
		t.Errorf("Limit = %d", rl.Limit)
	}
	if rl.Remaining != 87 {
		t.Errorf("Remaining = %d", rl.Remaining)
	}
	if rl.Reset.Unix() != 1720000000 {
		t.Errorf("Reset = %d", rl.Reset.Unix())
	}
}

func TestRateLimitHeaderAbsent(t *testing.T) {
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 200, map[string]any{"success": true, "data": []any{}})
	})
	if _, err := c.ListCurrencies(context.Background()); err != nil {
		t.Fatalf("ListCurrencies: %v", err)
	}
	if rl := c.RateLimit(); rl != nil {
		t.Errorf("expected nil RateLimit, got %+v", rl)
	}
}

func TestQueryFromOptions(t *testing.T) {
	got := queryFromOptions(ListOptions{Page: 2, PageSize: 50}, nil)
	want := "?page=2&pageSize=50"
	if got != want {
		t.Errorf("queryFromOptions = %q, want %q", got, want)
	}
	if got := queryFromOptions(ListOptions{}, nil); got != "" {
		t.Errorf("empty options should produce no query, got %q", got)
	}
}

func TestDefaultBaseURLReachable(t *testing.T) {
	// This guards against accidental changes to the well-known production URL.
	if DefaultBaseURL != "https://www.westmarches.games/api/v1" {
		t.Errorf("DefaultBaseURL = %q", DefaultBaseURL)
	}
}
