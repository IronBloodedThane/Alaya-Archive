package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/alaya-archive/backend-go/internal/config"
	"github.com/alaya-archive/backend-go/internal/middleware"
	"github.com/alaya-archive/backend-go/internal/repository"
)

type UserHandler struct {
	users *repository.UserRepository
	cfg   *config.Config
}

func NewUserHandler(users *repository.UserRepository, cfg *config.Config) *UserHandler {
	return &UserHandler{users: users, cfg: cfg}
}

const maxAvatarBytes = 5 << 20 // 5 MiB

var allowedAvatarMime = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.users.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.users.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	var req struct {
		DisplayName   *string `json:"display_name"`
		Bio           *string `json:"bio"`
		ProfilePublic *bool   `json:"profile_public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.ProfilePublic != nil {
		user.ProfilePublic = *req.ProfilePublic
	}

	if err := h.users.Update(user); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	// Refresh to reflect updated_at
	refreshed, err := h.users.GetByID(userID)
	if err == nil {
		user = refreshed
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarBytes)

	if err := r.ParseMultipartForm(maxAvatarBytes); err != nil {
		writeError(w, http.StatusBadRequest, "file too large (max 5MB)")
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing avatar file")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	mime := http.DetectContentType(data)
	if !allowedAvatarMime[mime] {
		writeError(w, http.StatusBadRequest, "unsupported image type (must be JPEG, PNG, GIF, or WebP)")
		return
	}

	if err := h.users.SetAvatar(userID, data, mime); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save avatar")
		return
	}

	user, err := h.users.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if err := h.users.ClearAvatar(userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete avatar")
		return
	}

	user, err := h.users.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	data, mime, updatedAt, err := h.users.GetAvatar(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load avatar")
		return
	}
	if len(data) == 0 {
		http.NotFound(w, r)
		return
	}

	etag := fmt.Sprintf(`"%d"`, updatedAt.Unix())
	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	w.Write(data)
}

func (h *UserHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := h.users.GetByUsername(username)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if !user.ProfilePublic {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	user.Email = ""
	user.HashedPassword = ""

	writeJSON(w, http.StatusOK, user)
}
