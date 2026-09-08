package westmarches

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient returns a client pointed at a test server. The server records
// every request it receives (into *[]http.Request) and delegates handling to
// the provided handler. Rate limiting is disabled so tests run fast.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server, *[]http.Request) {
	t.Helper()
	var requests []http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
		}
		r.Body.Close()
		clone := r.Clone(r.Context())
		clone.Body = io.NopCloser(bytes.NewReader(body))
		clone.ContentLength = int64(len(body))
		r.Body = io.NopCloser(bytes.NewReader(body))
		requests = append(requests, *clone)
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	c := NewClient("test-api-key", WithBaseURL(srv.URL), WithRateLimit(0))
	return c, srv, &requests
}

// writeJSON writes v as the JSON response body with the given status code.
func writeJSON(t *testing.T, w http.ResponseWriter, status int, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encoding response: %v", err)
	}
}

// rateLimitHeaders sets the documented X-RateLimit-* response headers.
func rateLimitHeaders(w http.ResponseWriter) {
	w.Header().Set("X-RateLimit-Limit", "100")
	w.Header().Set("X-RateLimit-Remaining", "87")
	w.Header().Set("X-RateLimit-Reset", "1720000000")
}

// decodeJSONBody decodes the request body into v.
func decodeJSONBody(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
