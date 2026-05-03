// Package lookup fetches metadata for collection items from external
// providers (Google Books, TMDB, IGDB, AniList). Each provider implements
// Provider so the handler layer can fan out by media type.
package lookup

import (
	"context"
	"errors"
)

// Result is the normalized shape returned to the API client regardless of
// which provider produced it. Fields not supplied by a given provider are
// left zero-valued; the client decides what to surface.
type Result struct {
	Provider     string   `json:"provider"`
	ProviderID   string   `json:"provider_id,omitempty"`
	Title        string   `json:"title"`
	Subtitle     string   `json:"subtitle,omitempty"`
	Authors      []string `json:"authors,omitempty"`
	Description  string   `json:"description,omitempty"`
	CoverImage   string   `json:"cover_image,omitempty"`
	Year         int      `json:"year,omitempty"`
	Categories   []string `json:"categories,omitempty"`
	ISBN10       string   `json:"isbn_10,omitempty"`
	ISBN13       string   `json:"isbn_13,omitempty"`
	PageCount    int      `json:"page_count,omitempty"`
	Language     string   `json:"language,omitempty"`
	Publisher    string   `json:"publisher,omitempty"`
	InfoURL      string   `json:"info_url,omitempty"`
	// Series and SeriesPosition are parsed from Title — multi-volume works
	// like manga ("Berserk Vol. 1") let the UI group volumes together.
	// Empty when the title doesn't look like part of a series.
	Series         string `json:"series,omitempty"`
	SeriesPosition int    `json:"series_position,omitempty"`
}

// Provider is implemented by each external metadata source. Implementations
// must be safe for concurrent use.
type Provider interface {
	// Name identifies the provider in responses and logs (e.g. "google_books").
	Name() string

	// ByISBN looks up a single volume by ISBN-10 or ISBN-13. Returns
	// ErrNotFound if the provider has no match.
	ByISBN(ctx context.Context, isbn string) (*Result, error)

	// Search performs a free-text query and returns up to limit results.
	// Order is provider-defined (typically relevance). limit <= 0 means
	// the provider's default.
	Search(ctx context.Context, query string, limit int) ([]*Result, error)
}

// ErrNotFound is returned by providers when a lookup yields no results.
// Handlers translate this into 404.
var ErrNotFound = errors.New("lookup: not found")
