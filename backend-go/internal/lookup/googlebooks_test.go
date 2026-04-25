package lookup

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// stubGoogleBooks returns an httptest.Server that serves the given body
// (status 200) for any /volumes request. It records the last query string
// it received so tests can assert on it.
func stubGoogleBooks(t *testing.T, body string) (*httptest.Server, *url.Values) {
	t.Helper()
	last := &url.Values{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/volumes") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		*last = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, last
}

const sampleISBNHit = `{
  "totalItems": 1,
  "items": [{
    "id": "abc123",
    "volumeInfo": {
      "title": "Berserk Volume 1",
      "subtitle": "The Black Swordsman",
      "authors": ["Kentaro Miura"],
      "publisher": "Dark Horse",
      "publishedDate": "2003-09-16",
      "description": "Guts wanders.",
      "industryIdentifiers": [
        {"type": "ISBN_10", "identifier": "1593070209"},
        {"type": "ISBN_13", "identifier": "9781593070205"}
      ],
      "pageCount": 232,
      "categories": ["Comics & Graphic Novels"],
      "language": "en",
      "imageLinks": {"thumbnail": "https://example/cover.jpg", "smallThumbnail": "https://example/small.jpg"},
      "infoLink": "https://books.google/abc123"
    }
  }]
}`

const sampleEmpty = `{"totalItems": 0}`

const sampleSearchTwo = `{
  "totalItems": 2,
  "items": [
    {"id": "id1", "volumeInfo": {"title": "First", "publishedDate": "2020"}},
    {"id": "id2", "volumeInfo": {"title": "Second", "publishedDate": "bad-date", "imageLinks": {"smallThumbnail": "https://example/sm.jpg"}}}
  ]
}`

func TestGoogleBooks_ByISBN_Hit(t *testing.T) {
	srv, last := stubGoogleBooks(t, sampleISBNHit)
	gb := NewGoogleBooks("test-key").WithBaseURL(srv.URL)

	res, err := gb.ByISBN(context.Background(), "9781593070205")
	if err != nil {
		t.Fatalf("ByISBN: %v", err)
	}

	if got, want := last.Get("q"), "isbn:9781593070205"; got != want {
		t.Errorf("query: got %q, want %q", got, want)
	}
	if got, want := last.Get("key"), "test-key"; got != want {
		t.Errorf("api key not forwarded: got %q, want %q", got, want)
	}
	if got, want := last.Get("maxResults"), "1"; got != want {
		t.Errorf("maxResults: got %q, want %q", got, want)
	}

	if res.Title != "Berserk Volume 1" {
		t.Errorf("title: got %q", res.Title)
	}
	if res.Subtitle != "The Black Swordsman" {
		t.Errorf("subtitle: got %q", res.Subtitle)
	}
	if len(res.Authors) != 1 || res.Authors[0] != "Kentaro Miura" {
		t.Errorf("authors: got %v", res.Authors)
	}
	if res.ISBN10 != "1593070209" || res.ISBN13 != "9781593070205" {
		t.Errorf("isbns: got %q / %q", res.ISBN10, res.ISBN13)
	}
	if res.Year != 2003 {
		t.Errorf("year: got %d", res.Year)
	}
	if res.CoverImage != "https://example/cover.jpg" {
		t.Errorf("cover: got %q (should prefer thumbnail over smallThumbnail)", res.CoverImage)
	}
	if res.PageCount != 232 {
		t.Errorf("pageCount: got %d", res.PageCount)
	}
	if res.Provider != "google_books" {
		t.Errorf("provider: got %q", res.Provider)
	}
	if res.ProviderID != "abc123" {
		t.Errorf("provider id: got %q", res.ProviderID)
	}
}

func TestGoogleBooks_ByISBN_NotFound(t *testing.T) {
	srv, _ := stubGoogleBooks(t, sampleEmpty)
	gb := NewGoogleBooks("").WithBaseURL(srv.URL)

	_, err := gb.ByISBN(context.Background(), "0000000000000")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestGoogleBooks_ByISBN_EmptyInput(t *testing.T) {
	gb := NewGoogleBooks("")
	if _, err := gb.ByISBN(context.Background(), "   "); err == nil {
		t.Fatal("expected error for empty isbn")
	}
}

func TestGoogleBooks_ByISBN_NoAPIKey_Omitted(t *testing.T) {
	srv, last := stubGoogleBooks(t, sampleISBNHit)
	gb := NewGoogleBooks("").WithBaseURL(srv.URL)

	if _, err := gb.ByISBN(context.Background(), "9781593070205"); err != nil {
		t.Fatalf("ByISBN: %v", err)
	}
	if last.Get("key") != "" {
		t.Errorf("key should not be sent when unset, got %q", last.Get("key"))
	}
}

func TestGoogleBooks_Search_Hit(t *testing.T) {
	srv, last := stubGoogleBooks(t, sampleSearchTwo)
	gb := NewGoogleBooks("k").WithBaseURL(srv.URL)

	results, err := gb.Search(context.Background(), "berserk", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if last.Get("q") != "berserk" {
		t.Errorf("query: got %q", last.Get("q"))
	}
	if last.Get("maxResults") != "5" {
		t.Errorf("maxResults: got %q", last.Get("maxResults"))
	}
	if results[0].Title != "First" || results[0].Year != 2020 {
		t.Errorf("first result: %+v", results[0])
	}
	// Second item has a non-numeric publishedDate -> Year stays 0; cover
	// falls back to smallThumbnail.
	if results[1].Year != 0 {
		t.Errorf("expected year=0 for bad date, got %d", results[1].Year)
	}
	if results[1].CoverImage != "https://example/sm.jpg" {
		t.Errorf("cover fallback: got %q", results[1].CoverImage)
	}
}

func TestGoogleBooks_Search_DefaultLimit(t *testing.T) {
	srv, last := stubGoogleBooks(t, sampleEmpty)
	gb := NewGoogleBooks("").WithBaseURL(srv.URL)

	if _, err := gb.Search(context.Background(), "anything", 0); err != nil {
		t.Fatalf("Search: %v", err)
	}
	if last.Get("maxResults") != "10" {
		t.Errorf("default limit: got %q, want 10", last.Get("maxResults"))
	}
}

func TestGoogleBooks_Search_EmptyQuery(t *testing.T) {
	gb := NewGoogleBooks("")
	if _, err := gb.Search(context.Background(), "", 0); err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestGoogleBooks_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	gb := NewGoogleBooks("").WithBaseURL(srv.URL)
	if _, err := gb.ByISBN(context.Background(), "123"); err == nil {
		t.Fatal("expected error on 500")
	}
}

func TestGoogleBooks_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	t.Cleanup(srv.Close)

	gb := NewGoogleBooks("").WithBaseURL(srv.URL)
	if _, err := gb.ByISBN(context.Background(), "123"); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestGoogleBooks_TransportError(t *testing.T) {
	// Point at an address that won't accept connections to force a
	// transport-level error from http.Client.Do.
	gb := NewGoogleBooks("").WithBaseURL("http://127.0.0.1:1")
	if _, err := gb.ByISBN(context.Background(), "123"); err == nil {
		t.Fatal("expected http error")
	}
}

func TestGoogleBooks_Name(t *testing.T) {
	if NewGoogleBooks("").Name() != "google_books" {
		t.Fatal("provider name mismatch")
	}
}

func TestGoogleBooks_WithHTTPClient(t *testing.T) {
	srv, _ := stubGoogleBooks(t, sampleISBNHit)
	custom := &http.Client{}
	gb := NewGoogleBooks("").WithBaseURL(srv.URL).WithHTTPClient(custom)
	if gb.client != custom {
		t.Fatal("WithHTTPClient did not replace the client")
	}
	if _, err := gb.ByISBN(context.Background(), "9781593070205"); err != nil {
		t.Fatalf("ByISBN with custom client: %v", err)
	}
}

func TestParseYear(t *testing.T) {
	cases := map[string]int{
		"":            0,
		"abc":         0,
		"2024":        2024,
		"2024-05":     2024,
		"2024-05-12":  2024,
		"bad":         0,
		"19xx":        0,
	}
	for in, want := range cases {
		if got := parseYear(in); got != want {
			t.Errorf("parseYear(%q): got %d, want %d", in, got, want)
		}
	}
}
