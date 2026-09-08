package westmarches

import "context"

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

// CollectAll pages through a paginated list endpoint and returns every item
// across all pages. It fetches pages of the given pageSize until the last page
// has been reached, honoring the server's pagination metadata.
//
//	full, err := CollectAll(ctx, 500, client.ListCharacters)
func CollectAll[T any](ctx context.Context, pageSize int, fetch PageFetcher[T]) ([]T, error) {
	if pageSize <= 0 {
		pageSize = 500
	}
	var all []T
	page := 1
	for {
		p, err := fetch(ctx, page, pageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, p.Data...)
		totalPages := p.Pagination.TotalPages
		if totalPages <= 0 {
			// The server did not report pagination metadata; stop after one page.
			return all, nil
		}
		if page >= totalPages {
			return all, nil
		}
		page++
	}
}
