package handler

import (
	"net/http"
	"testing"

	"github.com/alaya-archive/backend-go/internal/repository"
)

// bookPayload is the request body shared across the duplicate-flow tests.
// title/media_type/isbn drive the duplicate fingerprint.
func bookPayload(extras map[string]interface{}) map[string]interface{} {
	body := map[string]interface{}{
		"media_type": "book",
		"title":      "Dune",
		"creator":    "Frank Herbert",
		"isbn":       "9780441172719",
	}
	for k, v := range extras {
		body[k] = v
	}
	return body
}

func TestCreateMedia_PersistsISBN(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("a@test.com", "alice", "hunter22hunter22")

	rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(nil), access)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create media: got %d, body=%s", rec.Code, rec.Body.String())
	}
	var created repository.Media
	decodeJSON(t, rec, &created)
	if created.ISBN != "9780441172719" {
		t.Fatalf("isbn not persisted, got %q", created.ISBN)
	}
}

func TestCheckDuplicate(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("c@test.com", "carol", "hunter22hunter22")

	// No collection yet — check returns empty.
	rec := env.do(http.MethodGet, "/api/v1/media/check?type=book&isbn=9780441172719", nil, access)
	if rec.Code != http.StatusOK {
		t.Fatalf("check empty: got %d", rec.Code)
	}
	var empty struct {
		Items []repository.Media `json:"items"`
		Count int                `json:"count"`
	}
	decodeJSON(t, rec, &empty)
	if empty.Count != 0 || len(empty.Items) != 0 {
		t.Fatalf("expected empty, got count=%d items=%d", empty.Count, len(empty.Items))
	}

	// Create the book.
	if rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(nil), access); rec.Code != http.StatusCreated {
		t.Fatalf("seed create: got %d, body=%s", rec.Code, rec.Body.String())
	}

	// Now check returns it.
	rec = env.do(http.MethodGet, "/api/v1/media/check?type=book&isbn=9780441172719", nil, access)
	if rec.Code != http.StatusOK {
		t.Fatalf("check after create: got %d", rec.Code)
	}
	var found struct {
		Items []repository.Media `json:"items"`
		Count int                `json:"count"`
	}
	decodeJSON(t, rec, &found)
	if found.Count != 1 || len(found.Items) != 1 {
		t.Fatalf("expected one match, got count=%d items=%d", found.Count, len(found.Items))
	}
	if found.Items[0].ISBN != "9780441172719" {
		t.Fatalf("wrong isbn returned: %q", found.Items[0].ISBN)
	}

	// Different media_type with the same ISBN doesn't match — fingerprint is
	// (user, type, isbn).
	rec = env.do(http.MethodGet, "/api/v1/media/check?type=manga&isbn=9780441172719", nil, access)
	var crossType struct {
		Count int `json:"count"`
	}
	decodeJSON(t, rec, &crossType)
	if crossType.Count != 0 {
		t.Fatalf("cross-type check should be empty, got %d", crossType.Count)
	}
}

func TestCheckDuplicate_ScopedToUser(t *testing.T) {
	env := newTestEnv(t)
	aliceAccess, _ := env.registerUser("a2@test.com", "alice2", "hunter22hunter22")
	bobAccess, _ := env.registerUser("b2@test.com", "bob2", "hunter22hunter22")

	// Alice owns the book.
	if rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(nil), aliceAccess); rec.Code != http.StatusCreated {
		t.Fatalf("alice create: got %d", rec.Code)
	}

	// Bob's check should return zero — duplicate detection is per-user.
	rec := env.do(http.MethodGet, "/api/v1/media/check?type=book&isbn=9780441172719", nil, bobAccess)
	var resp struct {
		Count int `json:"count"`
	}
	decodeJSON(t, rec, &resp)
	if resp.Count != 0 {
		t.Fatalf("bob should see no duplicates of alice's book, got %d", resp.Count)
	}
}

func TestCreateMedia_OnDuplicateError(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("e@test.com", "eve", "hunter22hunter22")

	if rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(nil), access); rec.Code != http.StatusCreated {
		t.Fatalf("seed: got %d", rec.Code)
	}

	rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(nil), access)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var conflict struct {
		Error    string             `json:"error"`
		Existing []repository.Media `json:"existing"`
	}
	decodeJSON(t, rec, &conflict)
	if conflict.Error != "duplicate" {
		t.Fatalf("expected error=duplicate, got %q", conflict.Error)
	}
	if len(conflict.Existing) != 1 {
		t.Fatalf("expected one existing record, got %d", len(conflict.Existing))
	}
}

func TestCreateMedia_OnDuplicateSkip(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("s@test.com", "steve", "hunter22hunter22")

	rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(nil), access)
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed: got %d", rec.Code)
	}
	var first repository.Media
	decodeJSON(t, rec, &first)

	rec = env.do(http.MethodPost, "/api/v1/media/", bookPayload(map[string]interface{}{
		"on_duplicate": "skip",
		"title":        "Dune (Reissue)",
	}), access)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var returned repository.Media
	decodeJSON(t, rec, &returned)
	if returned.ID != first.ID {
		t.Fatalf("skip should return existing id %q, got %q", first.ID, returned.ID)
	}
	if returned.Title != "Dune" {
		t.Fatalf("skip should not modify title; got %q", returned.Title)
	}

	// Only one row exists.
	listRec := env.do(http.MethodGet, "/api/v1/media/", nil, access)
	var list struct {
		Total int `json:"total"`
	}
	decodeJSON(t, listRec, &list)
	if list.Total != 1 {
		t.Fatalf("expected 1 row after skip, got %d", list.Total)
	}
}

func TestCreateMedia_OnDuplicateOverwrite(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("o@test.com", "olive", "hunter22hunter22")

	rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(nil), access)
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed: got %d", rec.Code)
	}
	var first repository.Media
	decodeJSON(t, rec, &first)

	rating := 9
	rec = env.do(http.MethodPost, "/api/v1/media/", bookPayload(map[string]interface{}{
		"on_duplicate": "overwrite",
		"title":        "Dune (Deluxe Edition)",
		"rating":       rating,
		"status":       "completed",
	}), access)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var updated repository.Media
	decodeJSON(t, rec, &updated)
	if updated.ID != first.ID {
		t.Fatalf("overwrite should reuse id %q, got %q", first.ID, updated.ID)
	}
	if updated.Title != "Dune (Deluxe Edition)" {
		t.Fatalf("title not updated, got %q", updated.Title)
	}
	if updated.Rating == nil || *updated.Rating != 9 {
		t.Fatalf("rating not updated: %v", updated.Rating)
	}
	if updated.Status != "completed" {
		t.Fatalf("status not updated: %q", updated.Status)
	}

	listRec := env.do(http.MethodGet, "/api/v1/media/", nil, access)
	var list struct {
		Total int `json:"total"`
	}
	decodeJSON(t, listRec, &list)
	if list.Total != 1 {
		t.Fatalf("expected 1 row after overwrite, got %d", list.Total)
	}
}

func TestCreateMedia_OnDuplicateAllow(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("d@test.com", "doug", "hunter22hunter22")

	if rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(nil), access); rec.Code != http.StatusCreated {
		t.Fatalf("seed: got %d", rec.Code)
	}

	rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(map[string]interface{}{
		"on_duplicate": "allow",
		"notes":        "second copy — gift from grandma",
	}), access)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", rec.Code, rec.Body.String())
	}

	listRec := env.do(http.MethodGet, "/api/v1/media/", nil, access)
	var list struct {
		Total int `json:"total"`
	}
	decodeJSON(t, listRec, &list)
	if list.Total != 2 {
		t.Fatalf("expected 2 rows after allow, got %d", list.Total)
	}
}

func TestCreateMedia_RejectsUnknownOnDuplicate(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("z@test.com", "zoe", "hunter22hunter22")

	rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(map[string]interface{}{
		"on_duplicate": "nuke",
	}), access)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown policy, got %d", rec.Code)
	}
}

func TestCreateMedia_DefaultsToOwned(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("lt1@test.com", "lt1", "hunter22hunter22")

	rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(nil), access)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d, body=%s", rec.Code, rec.Body.String())
	}
	var m repository.Media
	decodeJSON(t, rec, &m)
	if m.ListType != "owned" {
		t.Fatalf("expected list_type owned by default, got %q", m.ListType)
	}
}

func TestCreateMedia_AcceptsWishlist(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("lt2@test.com", "lt2", "hunter22hunter22")

	rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(map[string]interface{}{
		"list_type": "wishlist",
	}), access)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create wishlist: got %d, body=%s", rec.Code, rec.Body.String())
	}
	var m repository.Media
	decodeJSON(t, rec, &m)
	if m.ListType != "wishlist" {
		t.Fatalf("expected list_type wishlist, got %q", m.ListType)
	}
}

func TestCreateMedia_RejectsBadListType(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("lt3@test.com", "lt3", "hunter22hunter22")

	rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(map[string]interface{}{
		"list_type": "borrowed",
	}), access)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown list_type, got %d", rec.Code)
	}
}

func TestCreateMedia_DuplicateAcrossListTypes(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("lt4@test.com", "lt4", "hunter22hunter22")

	// Add to wishlist first.
	if rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(map[string]interface{}{
		"list_type": "wishlist",
	}), access); rec.Code != http.StatusCreated {
		t.Fatalf("seed wishlist: got %d, body=%s", rec.Code, rec.Body.String())
	}

	// Re-scanning the same ISBN as owned should still trip the duplicate
	// check — the fingerprint is (user, media_type, isbn), independent of
	// list_type. User chooses overwrite to "promote" it.
	rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(map[string]interface{}{
		"list_type": "owned",
	}), access)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 across list_types, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateMedia_OverwritePromotesListType(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("lt5@test.com", "lt5", "hunter22hunter22")

	if rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(map[string]interface{}{
		"list_type": "wishlist",
	}), access); rec.Code != http.StatusCreated {
		t.Fatalf("seed wishlist: got %d", rec.Code)
	}

	rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(map[string]interface{}{
		"list_type":    "owned",
		"on_duplicate": "overwrite",
	}), access)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on overwrite, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var m repository.Media
	decodeJSON(t, rec, &m)
	if m.ListType != "owned" {
		t.Fatalf("overwrite should promote list_type to owned, got %q", m.ListType)
	}
}

func TestListMedia_FiltersByListType(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("lt6@test.com", "lt6", "hunter22hunter22")

	// Two distinct ISBNs so they aren't duplicates of each other.
	if rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(nil), access); rec.Code != http.StatusCreated {
		t.Fatalf("seed owned: got %d", rec.Code)
	}
	if rec := env.do(http.MethodPost, "/api/v1/media/", bookPayload(map[string]interface{}{
		"isbn":      "9780062315007",
		"title":     "The Alchemist",
		"list_type": "wishlist",
	}), access); rec.Code != http.StatusCreated {
		t.Fatalf("seed wishlist: got %d", rec.Code)
	}

	// list_type=wishlist should return only the wishlist row.
	rec := env.do(http.MethodGet, "/api/v1/media/?list_type=wishlist", nil, access)
	var resp struct {
		Items []repository.Media `json:"items"`
		Total int                `json:"total"`
	}
	decodeJSON(t, rec, &resp)
	if resp.Total != 1 || len(resp.Items) != 1 || resp.Items[0].ListType != "wishlist" {
		t.Fatalf("wishlist filter wrong: total=%d items=%v", resp.Total, resp.Items)
	}

	// No filter returns both.
	rec = env.do(http.MethodGet, "/api/v1/media/", nil, access)
	decodeJSON(t, rec, &resp)
	if resp.Total != 2 {
		t.Fatalf("expected 2 unfiltered, got %d", resp.Total)
	}
}

func TestCreateMedia_BlankISBNAlwaysAllowed(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("n@test.com", "noisbn", "hunter22hunter22")

	body := map[string]interface{}{
		"media_type": "book",
		"title":      "Untitled Manuscript",
	}
	if rec := env.do(http.MethodPost, "/api/v1/media/", body, access); rec.Code != http.StatusCreated {
		t.Fatalf("first create: got %d", rec.Code)
	}
	// Same payload again — without ISBN, no fingerprint, so this must succeed
	// even with the default on_duplicate=error.
	if rec := env.do(http.MethodPost, "/api/v1/media/", body, access); rec.Code != http.StatusCreated {
		t.Fatalf("second create with blank isbn should succeed, got %d, body=%s", rec.Code, rec.Body.String())
	}
}
