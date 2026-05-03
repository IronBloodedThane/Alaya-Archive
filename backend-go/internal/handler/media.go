package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/alaya-archive/backend-go/internal/config"
	"github.com/alaya-archive/backend-go/internal/middleware"
	"github.com/alaya-archive/backend-go/internal/repository"
)

type MediaHandler struct {
	media *repository.MediaRepository
	users *repository.UserRepository
	cfg   *config.Config
}

func NewMediaHandler(media *repository.MediaRepository, users *repository.UserRepository, cfg *config.Config) *MediaHandler {
	return &MediaHandler{media: media, users: users, cfg: cfg}
}

type createMediaRequest struct {
	MediaType     string   `json:"media_type"`
	Title         string   `json:"title"`
	TitleOriginal string   `json:"title_original"`
	Description   string   `json:"description"`
	CoverImage    string   `json:"cover_image"`
	Status        string   `json:"status"`
	Rating        *int     `json:"rating"`
	Notes         string   `json:"notes"`
	YearReleased  *int     `json:"year_released"`
	Creator       string   `json:"creator"`
	Genre         string   `json:"genre"`
	VolumesTotal  *int     `json:"volumes_total"`
	VolumesOwned  *int     `json:"volumes_owned"`
	EpisodesTotal *int     `json:"episodes_total"`
	EpisodesWatched *int   `json:"episodes_watched"`
	ChaptersTotal *int     `json:"chapters_total"`
	ChaptersRead   *int     `json:"chapters_read"`
	ISBN           string   `json:"isbn"`
	Series         string   `json:"series"`
	SeriesPosition *int     `json:"series_position"`
	ListType       string   `json:"list_type"`
	IsPublic       *bool    `json:"is_public"`
	Tags          []string `json:"tags"`
	// OnDuplicate controls behavior when an item with the same
	// (user, media_type, isbn) already exists. One of:
	//   "error"     — default; respond 409 with the existing record
	//   "overwrite" — update the existing record from this payload
	//   "skip"      — no-op, return the existing record with status 200
	//   "allow"     — insert anyway as a second copy
	// Empty isbn always behaves as "allow" since we don't fingerprint blanks.
	OnDuplicate string `json:"on_duplicate"`
}

func (h *MediaHandler) CreateMedia(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req createMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" || req.MediaType == "" {
		writeError(w, http.StatusBadRequest, "title and media_type are required")
		return
	}

	status := req.Status
	if status == "" {
		status = "planned"
	}

	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	policy := req.OnDuplicate
	if policy == "" {
		policy = "error"
	}
	switch policy {
	case "error", "overwrite", "skip", "allow":
	default:
		writeError(w, http.StatusBadRequest, "on_duplicate must be one of: error, overwrite, skip, allow")
		return
	}

	listType := req.ListType
	if listType == "" {
		listType = "owned"
	}
	if listType != "owned" && listType != "wishlist" {
		writeError(w, http.StatusBadRequest, "list_type must be one of: owned, wishlist")
		return
	}

	if req.ISBN != "" {
		existing, err := h.media.FindDuplicates(userID, req.MediaType, req.ISBN)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to check duplicates")
			return
		}
		if len(existing) > 0 {
			switch policy {
			case "error":
				writeJSON(w, http.StatusConflict, map[string]interface{}{
					"error":    "duplicate",
					"existing": existing,
				})
				return
			case "skip":
				writeJSON(w, http.StatusOK, existing[0])
				return
			case "overwrite":
				target := existing[0]
				applyCreateRequest(target, &req, status, listType, isPublic)
				if err := h.media.Update(target); err != nil {
					writeError(w, http.StatusInternalServerError, "failed to update media")
					return
				}
				if req.Tags != nil {
					h.media.SetTags(target.ID, req.Tags)
				}
				updated, _ := h.media.GetByID(target.ID)
				writeJSON(w, http.StatusOK, updated)
				return
			case "allow":
				// fall through to insert a new copy
			}
		}
	}

	m := &repository.Media{
		ID:              newID(),
		UserID:          userID,
		MediaType:       req.MediaType,
		Title:           req.Title,
		TitleOriginal:   req.TitleOriginal,
		Description:     req.Description,
		CoverImage:      req.CoverImage,
		Status:          status,
		Rating:          req.Rating,
		Notes:           req.Notes,
		YearReleased:    req.YearReleased,
		Creator:         req.Creator,
		Genre:           req.Genre,
		VolumesTotal:    req.VolumesTotal,
		VolumesOwned:    req.VolumesOwned,
		EpisodesTotal:   req.EpisodesTotal,
		EpisodesWatched: req.EpisodesWatched,
		ChaptersTotal:   req.ChaptersTotal,
		ChaptersRead:    req.ChaptersRead,
		ISBN:            req.ISBN,
		Series:          req.Series,
		SeriesPosition:  req.SeriesPosition,
		ListType:        listType,
		IsPublic:        isPublic,
	}

	if err := h.media.Create(m); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create media")
		return
	}

	if len(req.Tags) > 0 {
		h.media.SetTags(m.ID, req.Tags)
	}

	created, _ := h.media.GetByID(m.ID)
	writeJSON(w, http.StatusCreated, created)
}

// applyCreateRequest copies fields from a create request onto an existing
// record. Used by the on_duplicate=overwrite path. Only non-empty / non-nil
// fields overwrite — matches the semantics of UpdateMedia.
func applyCreateRequest(m *repository.Media, req *createMediaRequest, status, listType string, isPublic bool) {
	m.Title = req.Title
	if req.TitleOriginal != "" {
		m.TitleOriginal = req.TitleOriginal
	}
	if req.Description != "" {
		m.Description = req.Description
	}
	if req.CoverImage != "" {
		m.CoverImage = req.CoverImage
	}
	m.Status = status
	if req.Rating != nil {
		m.Rating = req.Rating
	}
	if req.Notes != "" {
		m.Notes = req.Notes
	}
	if req.YearReleased != nil {
		m.YearReleased = req.YearReleased
	}
	if req.Creator != "" {
		m.Creator = req.Creator
	}
	if req.Genre != "" {
		m.Genre = req.Genre
	}
	if req.VolumesTotal != nil {
		m.VolumesTotal = req.VolumesTotal
	}
	if req.VolumesOwned != nil {
		m.VolumesOwned = req.VolumesOwned
	}
	if req.EpisodesTotal != nil {
		m.EpisodesTotal = req.EpisodesTotal
	}
	if req.EpisodesWatched != nil {
		m.EpisodesWatched = req.EpisodesWatched
	}
	if req.ChaptersTotal != nil {
		m.ChaptersTotal = req.ChaptersTotal
	}
	if req.ChaptersRead != nil {
		m.ChaptersRead = req.ChaptersRead
	}
	if req.ISBN != "" {
		m.ISBN = req.ISBN
	}
	if req.Series != "" {
		m.Series = req.Series
	}
	if req.SeriesPosition != nil {
		m.SeriesPosition = req.SeriesPosition
	}
	m.ListType = listType
	m.IsPublic = isPublic
}

// CheckDuplicate handles GET /api/v1/media/check?type=<media_type>&isbn=<isbn>.
// Returns existing items in the caller's collection that share the same
// (media_type, isbn). Empty isbn returns an empty list — we don't fingerprint
// blanks.
func (h *MediaHandler) CheckDuplicate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	mediaType := r.URL.Query().Get("type")
	isbn := r.URL.Query().Get("isbn")

	if mediaType == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}

	items, err := h.media.FindDuplicates(userID, mediaType, isbn)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check duplicates")
		return
	}
	if items == nil {
		items = []*repository.Media{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}

func (h *MediaHandler) GetMedia(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	mediaID := chi.URLParam(r, "mediaID")

	m, err := h.media.GetByID(mediaID)
	if err != nil {
		writeError(w, http.StatusNotFound, "media not found")
		return
	}

	if m.UserID != userID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, m)
}

func (h *MediaHandler) ListMedia(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	filter := repository.MediaFilter{
		MediaType: r.URL.Query().Get("type"),
		Status:    r.URL.Query().Get("status"),
		Search:    r.URL.Query().Get("search"),
		Tag:       r.URL.Query().Get("tag"),
		ListType:  r.URL.Query().Get("list_type"),
		Series:    r.URL.Query().Get("series"),
		Limit:     queryInt(r, "limit", 50),
		Offset:    queryInt(r, "offset", 0),
	}

	items, total, err := h.media.List(userID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list media")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": total,
	})
}

func (h *MediaHandler) UpdateMedia(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	mediaID := chi.URLParam(r, "mediaID")

	m, err := h.media.GetByID(mediaID)
	if err != nil {
		writeError(w, http.StatusNotFound, "media not found")
		return
	}
	if m.UserID != userID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	var req createMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title != "" {
		m.Title = req.Title
	}
	if req.TitleOriginal != "" {
		m.TitleOriginal = req.TitleOriginal
	}
	if req.Description != "" {
		m.Description = req.Description
	}
	if req.CoverImage != "" {
		m.CoverImage = req.CoverImage
	}
	if req.Status != "" {
		m.Status = req.Status
	}
	if req.Rating != nil {
		m.Rating = req.Rating
	}
	if req.Notes != "" {
		m.Notes = req.Notes
	}
	if req.YearReleased != nil {
		m.YearReleased = req.YearReleased
	}
	if req.Creator != "" {
		m.Creator = req.Creator
	}
	if req.Genre != "" {
		m.Genre = req.Genre
	}
	if req.VolumesTotal != nil {
		m.VolumesTotal = req.VolumesTotal
	}
	if req.VolumesOwned != nil {
		m.VolumesOwned = req.VolumesOwned
	}
	if req.EpisodesTotal != nil {
		m.EpisodesTotal = req.EpisodesTotal
	}
	if req.EpisodesWatched != nil {
		m.EpisodesWatched = req.EpisodesWatched
	}
	if req.ChaptersTotal != nil {
		m.ChaptersTotal = req.ChaptersTotal
	}
	if req.ChaptersRead != nil {
		m.ChaptersRead = req.ChaptersRead
	}
	if req.ISBN != "" {
		m.ISBN = req.ISBN
	}
	if req.Series != "" {
		m.Series = req.Series
	}
	if req.SeriesPosition != nil {
		m.SeriesPosition = req.SeriesPosition
	}
	if req.ListType != "" {
		if req.ListType != "owned" && req.ListType != "wishlist" {
			writeError(w, http.StatusBadRequest, "list_type must be one of: owned, wishlist")
			return
		}
		m.ListType = req.ListType
	}
	if req.IsPublic != nil {
		m.IsPublic = *req.IsPublic
	}

	if err := h.media.Update(m); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update media")
		return
	}

	if req.Tags != nil {
		h.media.SetTags(m.ID, req.Tags)
	}

	updated, _ := h.media.GetByID(m.ID)
	writeJSON(w, http.StatusOK, updated)
}

func (h *MediaHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	mediaID := chi.URLParam(r, "mediaID")

	m, err := h.media.GetByID(mediaID)
	if err != nil {
		writeError(w, http.StatusNotFound, "media not found")
		return
	}
	if m.UserID != userID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := h.media.Delete(mediaID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete media")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "media deleted"})
}

func (h *MediaHandler) RateMedia(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	mediaID := chi.URLParam(r, "mediaID")

	m, err := h.media.GetByID(mediaID)
	if err != nil {
		writeError(w, http.StatusNotFound, "media not found")
		return
	}
	if m.UserID != userID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	var req struct {
		Rating int `json:"rating"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Rating < 1 || req.Rating > 10 {
		writeError(w, http.StatusBadRequest, "rating must be between 1 and 10")
		return
	}

	m.Rating = &req.Rating
	if err := h.media.Update(m); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to rate media")
		return
	}

	writeJSON(w, http.StatusOK, m)
}

func (h *MediaHandler) AddTags(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	mediaID := chi.URLParam(r, "mediaID")

	m, err := h.media.GetByID(mediaID)
	if err != nil {
		writeError(w, http.StatusNotFound, "media not found")
		return
	}
	if m.UserID != userID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	var req struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.media.SetTags(mediaID, req.Tags); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to set tags")
		return
	}

	updated, _ := h.media.GetByID(mediaID)
	writeJSON(w, http.StatusOK, updated)
}

func (h *MediaHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	stats, err := h.media.GetStats(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (h *MediaHandler) SearchMedia(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "search query is required")
		return
	}

	filter := repository.MediaFilter{
		Search: query,
		Limit:  queryInt(r, "limit", 50),
		Offset: queryInt(r, "offset", 0),
	}

	items, total, err := h.media.List(userID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to search media")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": total,
	})
}

func (h *MediaHandler) GetPublicCollection(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := h.users.GetByUsername(username)
	if err != nil || !user.ProfilePublic {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	filter := repository.MediaFilter{
		MediaType: r.URL.Query().Get("type"),
		Limit:     queryInt(r, "limit", 50),
		Offset:    queryInt(r, "offset", 0),
	}

	items, total, err := h.media.ListPublicByUser(user.ID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load collection")
		return
	}

	if items == nil {
		items = []*repository.Media{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": total,
		"user": map[string]interface{}{
			"username":     user.Username,
			"display_name": user.DisplayName,
			"bio":          user.Bio,
			"has_avatar":   user.HasAvatar,
		},
	})
}

func queryInt(r *http.Request, key string, fallback int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
