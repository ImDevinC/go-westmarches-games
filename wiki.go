package westmarches

import (
	"context"
	"net/url"
)

// ListArticles returns all published, public wiki articles in the community.
// Excludes drafts, GM-only pages, and archived pages. Ordered alphabetically
// by title. Pass a closure wrapping this method to Walk or CollectAll to
// retrieve every page in one call.
func (c *Client) ListArticles(ctx context.Context, opts ListArticlesOptions) (*Page[ArticleSummary], error) {
	extra := url.Values{}
	if opts.Search != "" {
		extra.Set("search", opts.Search)
	}
	if opts.Category != "" {
		extra.Set("category", string(opts.Category))
	}
	data, err := c.do(ctx, "GET", "/wiki/articles"+queryFromOptions(opts.ListOptions, extra), nil)
	if err != nil {
		return nil, err
	}
	var env struct {
		Success    bool             `json:"success"`
		Data       []ArticleSummary `json:"data"`
		Pagination Pagination       `json:"pagination"`
	}
	if err := jsonUnmarshal(data, &env); err != nil {
		return nil, err
	}
	return &Page[ArticleSummary]{Data: env.Data, Pagination: env.Pagination}, nil
}

// GetArticle returns a single wiki article including its full content body.
// The Content field is a ProseMirror/TipTap JSON document; text can be
// extracted from the "text" fields in leaf nodes, which is suitable for
// passing to an AI model.
func (c *Client) GetArticle(ctx context.Context, articleID string) (*ArticleDetail, error) {
	data, err := c.do(ctx, "GET", "/wiki/articles/"+url.PathEscape(articleID), nil)
	if err != nil {
		return nil, err
	}
	var out ArticleDetail
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
