package handler

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

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

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, 2<<20) // 2MB

	if err := r.ParseMultipartForm(2 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "file too large (max 2MB)")
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

	user, err := h.users.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	user.Avatar = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data)

	if err := h.users.Update(user); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update avatar")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.users.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	user.Avatar = ""
	if err := h.users.Update(user); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete avatar")
		return
	}

	writeJSON(w, http.StatusOK, user)
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
