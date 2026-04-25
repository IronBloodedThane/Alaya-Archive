package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alaya-archive/backend-go/internal/lookup"
)

// fakeProvider lets us drive the LookupHandler through every branch
// without making real HTTP calls.
type fakeProvider struct {
	name     string
	byISBN   func(ctx context.Context, isbn string) (*lookup.Result, error)
	search   func(ctx context.Context, q string, limit int) ([]*lookup.Result, error)
	lastQ    string
	lastLim  int
	lastISBN string
}

func (f *fakeProvider) Name() string { return f.name }

func (f *fakeProvider) ByISBN(ctx context.Context, isbn string) (*lookup.Result, error) {
	f.lastISBN = isbn
	return f.byISBN(ctx, isbn)
}

func (f *fakeProvider) Search(ctx context.Context, q string, limit int) ([]*lookup.Result, error) {
	f.lastQ = q
	f.lastLim = limit
	return f.search(ctx, q, limit)
}

func newLookupTestHandler(p lookup.Provider) *LookupHandler {
	return NewLookupHandler(p)
}

func doLookup(h *LookupHandler, query string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lookup?"+query, nil)
	rec := httptest.NewRecorder()
	h.Lookup(rec, req)
	return rec
}

func TestLookup_MissingType(t *testing.T) {
	h := newLookupTestHandler(&fakeProvider{name: "google_books"})
	rec := doLookup(h, "isbn=123")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestLookup_UnsupportedType(t *testing.T) {
	h := newLookupTestHandler(&fakeProvider{name: "google_books"})
	rec := doLookup(h, "type=movie&isbn=123")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestLookup_MissingISBNAndQuery(t *testing.T) {
	h := newLookupTestHandler(&fakeProvider{name: "google_books"})
	rec := doLookup(h, "type=book")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d", rec.Code)
	}
}

func TestLookup_BothISBNAndQuery(t *testing.T) {
	h := newLookupTestHandler(&fakeProvider{name: "google_books"})
	rec := doLookup(h, "type=book&isbn=123&q=foo")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d", rec.Code)
	}
}

func TestLookup_ByISBN_Hit(t *testing.T) {
	want := &lookup.Result{Title: "Berserk Vol 1", ISBN13: "9781593070205"}
	fp := &fakeProvider{
		name: "google_books",
		byISBN: func(ctx context.Context, isbn string) (*lookup.Result, error) {
			return want, nil
		},
	}
	h := newLookupTestHandler(fp)

	rec := doLookup(h, "type=manga&isbn=9781593070205")
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, body=%s", rec.Code, rec.Body.String())
	}
	if fp.lastISBN != "9781593070205" {
		t.Errorf("provider got isbn %q", fp.lastISBN)
	}

	var body struct {
		Provider string         `json:"provider"`
		Result   *lookup.Result `json:"result"`
	}
	decodeJSON(t, rec, &body)
	if body.Provider != "google_books" {
		t.Errorf("provider in response: %q", body.Provider)
	}
	if body.Result == nil || body.Result.Title != "Berserk Vol 1" {
		t.Errorf("result: %+v", body.Result)
	}
}

func TestLookup_ByISBN_NotFound(t *testing.T) {
	fp := &fakeProvider{
		name: "google_books",
		byISBN: func(ctx context.Context, isbn string) (*lookup.Result, error) {
			return nil, lookup.ErrNotFound
		},
	}
	h := newLookupTestHandler(fp)

	rec := doLookup(h, "type=book&isbn=000")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d", rec.Code)
	}
}

func TestLookup_ByISBN_ProviderError(t *testing.T) {
	fp := &fakeProvider{
		name: "google_books",
		byISBN: func(ctx context.Context, isbn string) (*lookup.Result, error) {
			return nil, errors.New("boom")
		},
	}
	h := newLookupTestHandler(fp)

	rec := doLookup(h, "type=book&isbn=123")
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status: got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestLookup_Search_Hit(t *testing.T) {
	want := []*lookup.Result{{Title: "First"}, {Title: "Second"}}
	fp := &fakeProvider{
		name: "google_books",
		search: func(ctx context.Context, q string, limit int) ([]*lookup.Result, error) {
			return want, nil
		},
	}
	h := newLookupTestHandler(fp)

	rec := doLookup(h, "type=manga&q=berserk&limit=5")
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, body=%s", rec.Code, rec.Body.String())
	}
	if fp.lastQ != "berserk" || fp.lastLim != 5 {
		t.Errorf("provider got q=%q limit=%d", fp.lastQ, fp.lastLim)
	}

	var body struct {
		Provider string           `json:"provider"`
		Results  []*lookup.Result `json:"results"`
	}
	decodeJSON(t, rec, &body)
	if len(body.Results) != 2 {
		t.Fatalf("results: %+v", body.Results)
	}
}

func TestLookup_Search_DefaultLimit(t *testing.T) {
	fp := &fakeProvider{
		name: "google_books",
		search: func(ctx context.Context, q string, limit int) ([]*lookup.Result, error) {
			return nil, nil
		},
	}
	h := newLookupTestHandler(fp)

	rec := doLookup(h, "type=book&q=anything")
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	// limit not provided -> handler forwards 0 so the provider applies
	// its own default.
	if fp.lastLim != 0 {
		t.Errorf("default limit forwarded as %d, want 0", fp.lastLim)
	}
}

func TestLookup_Search_ProviderError(t *testing.T) {
	fp := &fakeProvider{
		name: "google_books",
		search: func(ctx context.Context, q string, limit int) ([]*lookup.Result, error) {
			return nil, errors.New("upstream down")
		},
	}
	h := newLookupTestHandler(fp)

	rec := doLookup(h, "type=book&q=foo")
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status: got %d", rec.Code)
	}
}

func TestLookup_TypeIsCaseInsensitive(t *testing.T) {
	fp := &fakeProvider{
		name: "google_books",
		byISBN: func(ctx context.Context, isbn string) (*lookup.Result, error) {
			return &lookup.Result{Title: "ok"}, nil
		},
	}
	h := newLookupTestHandler(fp)

	rec := doLookup(h, "type=BOOK&isbn=123")
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, body=%s", rec.Code, rec.Body.String())
	}
}
