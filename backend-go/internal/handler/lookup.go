package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/alaya-archive/backend-go/internal/lookup"
)

// LookupHandler exposes external metadata lookups. Today only books and
// manga are supported (both via Google Books). Other media types will be
// added by registering more providers in the providers map.
type LookupHandler struct {
	// providers maps a media_type ("book", "manga", ...) to the provider
	// that can answer for it. Multiple types can share a provider.
	providers map[string]lookup.Provider
}

// NewLookupHandler wires the supported media types to their providers.
// books and manga both go through Google Books in this slice.
func NewLookupHandler(googleBooks lookup.Provider) *LookupHandler {
	return &LookupHandler{
		providers: map[string]lookup.Provider{
			"book":  googleBooks,
			"manga": googleBooks,
		},
	}
}

// Lookup handles GET /api/v1/lookup.
//
// Required query params:
//   - type: media type (e.g. "book", "manga")
//   - exactly one of: isbn=<code> for barcode scans, or q=<text> for OCR/text search
//
// Optional:
//   - limit: cap on results for text search (provider default applies otherwise)
func (h *LookupHandler) Lookup(w http.ResponseWriter, r *http.Request) {
	mediaType := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("type")))
	isbn := strings.TrimSpace(r.URL.Query().Get("isbn"))
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	if mediaType == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}
	provider, ok := h.providers[mediaType]
	if !ok {
		writeError(w, http.StatusBadRequest, "unsupported media type for lookup: "+mediaType)
		return
	}

	if isbn == "" && query == "" {
		writeError(w, http.StatusBadRequest, "isbn or q is required")
		return
	}
	if isbn != "" && query != "" {
		writeError(w, http.StatusBadRequest, "provide either isbn or q, not both")
		return
	}

	ctx := r.Context()

	if isbn != "" {
		result, err := provider.ByISBN(ctx, isbn)
		if errors.Is(err, lookup.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no metadata found for isbn")
			return
		}
		if err != nil {
			writeError(w, http.StatusBadGateway, "lookup failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"provider": provider.Name(),
			"result":   result,
		})
		return
	}

	limit := queryInt(r, "limit", 0)
	results, err := provider.Search(ctx, query, limit)
	if err != nil {
		writeError(w, http.StatusBadGateway, "lookup failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"provider": provider.Name(),
		"results":  results,
	})
}
