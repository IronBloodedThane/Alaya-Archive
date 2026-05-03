package lookup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	googleBooksDefaultBase = "https://www.googleapis.com/books/v1"
	googleBooksDefaultLimit = 10
)

// GoogleBooks is the Provider implementation for the Google Books v1 API.
// The zero value is not usable; construct via NewGoogleBooks.
type GoogleBooks struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewGoogleBooks builds a Google Books provider. apiKey may be empty
// (Google Books allows unauthenticated reads at lower rate limits).
func NewGoogleBooks(apiKey string) *GoogleBooks {
	return &GoogleBooks{
		apiKey:  apiKey,
		baseURL: googleBooksDefaultBase,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// WithBaseURL overrides the API base — used by tests pointing at httptest.
func (g *GoogleBooks) WithBaseURL(u string) *GoogleBooks {
	g.baseURL = strings.TrimRight(u, "/")
	return g
}

// WithHTTPClient swaps the HTTP client (e.g. to inject a custom transport
// in tests).
func (g *GoogleBooks) WithHTTPClient(c *http.Client) *GoogleBooks {
	g.client = c
	return g
}

func (g *GoogleBooks) Name() string { return "google_books" }

func (g *GoogleBooks) ByISBN(ctx context.Context, isbn string) (*Result, error) {
	isbn = strings.TrimSpace(isbn)
	if isbn == "" {
		return nil, fmt.Errorf("google_books: empty isbn")
	}
	results, err := g.query(ctx, "isbn:"+isbn, 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrNotFound
	}
	return results[0], nil
}

func (g *GoogleBooks) Search(ctx context.Context, query string, limit int) ([]*Result, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("google_books: empty query")
	}
	if limit <= 0 {
		limit = googleBooksDefaultLimit
	}
	return g.query(ctx, query, limit)
}

// query is the shared HTTP path for ByISBN and Search.
func (g *GoogleBooks) query(ctx context.Context, q string, maxResults int) ([]*Result, error) {
	params := url.Values{}
	params.Set("q", q)
	params.Set("maxResults", strconv.Itoa(maxResults))
	if g.apiKey != "" {
		params.Set("key", g.apiKey)
	}

	endpoint := g.baseURL + "/volumes?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("google_books: build request: %w", err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google_books: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("google_books: status %d", resp.StatusCode)
	}

	var body googleBooksResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("google_books: decode: %w", err)
	}

	out := make([]*Result, 0, len(body.Items))
	for _, item := range body.Items {
		out = append(out, item.toResult())
	}
	return out, nil
}

// --- Google Books wire types (only the fields we consume) ---

type googleBooksResponse struct {
	TotalItems int                  `json:"totalItems"`
	Items      []googleBooksVolume `json:"items"`
}

type googleBooksVolume struct {
	ID         string                  `json:"id"`
	VolumeInfo googleBooksVolumeInfo `json:"volumeInfo"`
}

type googleBooksVolumeInfo struct {
	Title               string                       `json:"title"`
	Subtitle            string                       `json:"subtitle"`
	Authors             []string                     `json:"authors"`
	Publisher           string                       `json:"publisher"`
	PublishedDate       string                       `json:"publishedDate"`
	Description         string                       `json:"description"`
	IndustryIdentifiers []googleBooksIndustryID     `json:"industryIdentifiers"`
	PageCount           int                          `json:"pageCount"`
	Categories          []string                     `json:"categories"`
	Language            string                       `json:"language"`
	ImageLinks          googleBooksImageLinks       `json:"imageLinks"`
	InfoLink            string                       `json:"infoLink"`
}

type googleBooksIndustryID struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

type googleBooksImageLinks struct {
	Thumbnail      string `json:"thumbnail"`
	SmallThumbnail string `json:"smallThumbnail"`
}

func (v googleBooksVolume) toResult() *Result {
	info := v.VolumeInfo
	r := &Result{
		Provider:    "google_books",
		ProviderID:  v.ID,
		Title:       info.Title,
		Subtitle:    info.Subtitle,
		Authors:     info.Authors,
		Description: info.Description,
		CoverImage:  pickCover(info.ImageLinks),
		Year:        parseYear(info.PublishedDate),
		Categories:  info.Categories,
		PageCount:   info.PageCount,
		Language:    info.Language,
		Publisher:   info.Publisher,
		InfoURL:     info.InfoLink,
	}
	for _, id := range info.IndustryIdentifiers {
		switch id.Type {
		case "ISBN_10":
			r.ISBN10 = id.Identifier
		case "ISBN_13":
			r.ISBN13 = id.Identifier
		}
	}
	// Best-effort series detection — Google Books has no first-class
	// series field, so we parse the title. Empty result is fine; the user
	// can override in the form.
	r.Series, r.SeriesPosition = ParseSeries(info.Title)
	return r
}

func pickCover(l googleBooksImageLinks) string {
	raw := l.Thumbnail
	if raw == "" {
		raw = l.SmallThumbnail
	}
	if raw == "" {
		return ""
	}
	// Google Books returns thumbnail URLs as http://. Android blocks cleartext
	// image loads by default and most modern web pages run on https, so we
	// upgrade the scheme on the way out. The CDN serves the same path on https.
	if strings.HasPrefix(raw, "http://") {
		raw = "https://" + raw[len("http://"):]
	}
	return raw
}

// parseYear pulls the leading 4-digit year out of a Google Books
// publishedDate, which can be "2024", "2024-05", or "2024-05-12".
// Returns 0 when no valid year is present.
func parseYear(s string) int {
	if len(s) < 4 {
		return 0
	}
	y, err := strconv.Atoi(s[:4])
	if err != nil {
		return 0
	}
	return y
}
