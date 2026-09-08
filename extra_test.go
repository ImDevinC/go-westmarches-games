package westmarches

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestWithHTTPClientOption(t *testing.T) {
	hc := &http.Client{}
	c := NewClient("key", WithHTTPClient(hc))
	if c.httpClient != hc {
		t.Error("expected custom http client to be used")
	}
}

func TestAPIErrorError(t *testing.T) {
	err := &APIError{StatusCode: 404, Success: false, Message: "Character not found"}
	if got := err.Error(); got == "" {
		t.Error("expected non-empty error string")
	}
	// Without a message, the status text should be used.
	err2 := &APIError{StatusCode: 429}
	if got := err2.Error(); got == "" {
		t.Error("expected non-empty error string")
	}
	if err2.IsRateLimited() != true {
		t.Error("expected IsRateLimited")
	}
	if err2.IsForbidden() || err2.IsUnauthorized() || err2.IsNotFound() {
		t.Error("unexpected status helper result")
	}
}

func TestAPIErrorUnknownStatus(t *testing.T) {
	err := &APIError{StatusCode: 599}
	if got := err.Error(); got == "" {
		t.Error("expected non-empty error string")
	}
}

func TestCollectAllStopsWhenNoPagination(t *testing.T) {
	calls := 0
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": []map[string]any{
				{"id": "c1", "name": "A", "level": 1, "experience": 0, "status": "ACTIVE", "isApproved": true, "user": map[string]any{"id": "u1"}},
			},
			// Intentionally no pagination object.
		})
	})

	all, err := CollectAll(context.Background(), 500, func(ctx context.Context, page, pageSize int) (*Page[CharacterSummary], error) {
		return c.ListCharacters(ctx, ListOptions{Page: page, PageSize: pageSize})
	})
	if err != nil {
		t.Fatalf("CollectAll: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("len(all) = %d", len(all))
	}
	if calls != 1 {
		t.Errorf("expected 1 call when pagination metadata is missing, got %d", calls)
	}
}

func TestCollectAllErrorPropagates(t *testing.T) {
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 500, map[string]any{"success": false, "error": "boom"})
	})
	_, err := CollectAll(context.Background(), 500, func(ctx context.Context, page, pageSize int) (*Page[CharacterSummary], error) {
		return c.ListCharacters(ctx, ListOptions{Page: page, PageSize: pageSize})
	})
	if err == nil {
		t.Fatal("expected error to propagate from CollectAll")
	}
}

func TestWalkVisitsEveryPage(t *testing.T) {
	pages := map[int][]CharacterSummary{
		1: {{ID: "c1", Name: "A"}, {ID: "c2", Name: "B"}},
		2: {{ID: "c3", Name: "C"}},
	}
	var visited []string
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		page := atoiOr(r.URL.Query().Get("page"), 1)
		writeJSON(t, w, 200, map[string]any{
			"success":    true,
			"data":       pages[page],
			"pagination": map[string]any{"page": page, "pageSize": 2, "total": 3, "totalPages": 2},
		})
	})

	err := Walk(context.Background(), 2, func(ctx context.Context, page, pageSize int) (*Page[CharacterSummary], error) {
		return c.ListCharacters(ctx, ListOptions{Page: page, PageSize: pageSize})
	}, func(page *Page[CharacterSummary]) error {
		for _, item := range page.Data {
			visited = append(visited, item.ID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(visited) != 3 || visited[0] != "c1" || visited[2] != "c3" {
		t.Errorf("unexpected visited order: %v", visited)
	}
}

func TestWalkStopsEarlyOnErrStopIteration(t *testing.T) {
	calls := 0
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		page := atoiOr(r.URL.Query().Get("page"), 1)
		writeJSON(t, w, 200, map[string]any{
			"success":    true,
			"data":       []map[string]any{{"id": "c1", "name": "A", "level": 1, "experience": 0, "status": "ACTIVE", "isApproved": true, "user": map[string]any{"id": "u1"}}},
			"pagination": map[string]any{"page": page, "pageSize": 1, "total": 5, "totalPages": 5},
		})
	})

	seen := 0
	err := Walk(context.Background(), 1, func(ctx context.Context, page, pageSize int) (*Page[CharacterSummary], error) {
		return c.ListCharacters(ctx, ListOptions{Page: page, PageSize: pageSize})
	}, func(page *Page[CharacterSummary]) error {
		seen += len(page.Data)
		if seen >= 1 {
			return ErrStopIteration
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if seen != 1 {
		t.Errorf("seen = %d, want 1", seen)
	}
	if calls != 1 {
		t.Errorf("expected 1 fetch when stopping early, got %d", calls)
	}
}

func TestWalkCallbackErrorPropagates(t *testing.T) {
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 200, map[string]any{
			"success":    true,
			"data":       []map[string]any{{"id": "c1", "name": "A", "level": 1, "experience": 0, "status": "ACTIVE", "isApproved": true, "user": map[string]any{"id": "u1"}}},
			"pagination": map[string]any{"page": 1, "pageSize": 1, "total": 5, "totalPages": 5},
		})
	})

	sentinel := errors.New("stop")
	err := Walk(context.Background(), 1, func(ctx context.Context, page, pageSize int) (*Page[CharacterSummary], error) {
		return c.ListCharacters(ctx, ListOptions{Page: page, PageSize: pageSize})
	}, func(page *Page[CharacterSummary]) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected callback error to propagate, got %v", err)
	}
}

func TestGetCharacterStatsBadGateway(t *testing.T) {
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 502, map[string]any{"success": false, "error": "D&D Beyond fetch failed"})
	})
	_, err := c.GetCharacterStats(context.Background(), "char_1")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 502 {
		t.Errorf("status = %d", apiErr.StatusCode)
	}
}
