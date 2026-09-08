package westmarches

import (
	"context"
	"errors"
)

// ListOptions controls pagination for list endpoints.
//
// Page defaults to 1 and PageSize defaults to 500 (max 500) when left zero.
// Note that the wiki articles endpoint caps its default page size at 100;
// pass an explicit PageSize to override.
type ListOptions struct {
	Page     int
	PageSize int
}

// ListArticlesOptions controls the wiki articles list endpoint, which
// additionally supports search and category filters.
type ListArticlesOptions struct {
	ListOptions
	// Search filters by article title (case-insensitive partial match).
	Search string
	// Category filters by article category.
	Category ArticleCategory
}

// PageFetcher is a function that fetches one page of a paginated collection.
type PageFetcher[T any] func(ctx context.Context, page int, pageSize int) (*Page[T], error)

// ErrStopIteration stops Walk early without fetching the remaining pages. It
// is not returned as an error by Walk; the walk simply ends at the current
// page, mirroring the convention used by the standard library (e.g.
// filepath.SkipDir).
var ErrStopIteration = errors.New("westmarches: stop pagination iteration")

// Walk pages through a paginated list endpoint, invoking fn for each page in
// order. It fetches pages of the given pageSize until the last page has been
// reached, honoring the server's pagination metadata. Returning
// ErrStopIteration from fn stops the walk early without fetching the
// remaining pages.
//
// Walk is the idiomatic pagination primitive: callers process pages as they
// arrive instead of buffering the entire result set, and can stop early once
// they have what they need.
//
//	err := Walk(ctx, 500, func(ctx context.Context, page, pageSize int) (*Page[CharacterSummary], error) {
//		return client.ListCharacters(ctx, ListOptions{Page: page, PageSize: pageSize})
//	}, func(page *Page[CharacterSummary]) error {
//		fmt.Println(page.Data)
//		return nil
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
func Walk[T any](ctx context.Context, pageSize int, fetch PageFetcher[T], fn func(page *Page[T]) error) error {
	if pageSize <= 0 {
		pageSize = 500
	}
	page := 1
	for {
		p, err := fetch(ctx, page, pageSize)
		if err != nil {
			return err
		}
		if err := fn(p); err != nil {
			if errors.Is(err, ErrStopIteration) {
				return nil
			}
			return err
		}
		totalPages := p.Pagination.TotalPages
		if totalPages <= 0 {
			// The server did not report pagination metadata; stop after one page.
			return nil
		}
		if page >= totalPages {
			return nil
		}
		page++
	}
}

// CollectAll pages through a paginated list endpoint and returns every item
// across all pages. It is a convenience wrapper around Walk for callers that
// want the complete result set materialized in memory; use Walk directly when
// you want to process results as they arrive or stop early.
//
//	full, err := CollectAll(ctx, 500, func(ctx context.Context, page, pageSize int) (*Page[CharacterSummary], error) {
//		return client.ListCharacters(ctx, ListOptions{Page: page, PageSize: pageSize})
//	})
func CollectAll[T any](ctx context.Context, pageSize int, fetch PageFetcher[T]) ([]T, error) {
	var all []T
	err := Walk(ctx, pageSize, fetch, func(page *Page[T]) error {
		all = append(all, page.Data...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return all, nil
}
