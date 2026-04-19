package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/alaya-archive/backend-go/internal/config"
	"github.com/alaya-archive/backend-go/internal/middleware"
	"github.com/alaya-archive/backend-go/internal/repository"
)

type SocialHandler struct {
	social *repository.SocialRepository
	users  *repository.UserRepository
	cfg    *config.Config
}

func NewSocialHandler(social *repository.SocialRepository, users *repository.UserRepository, cfg *config.Config) *SocialHandler {
	return &SocialHandler{social: social, users: users, cfg: cfg}
}

// Follow/Unfollow

func (h *SocialHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	targetID := chi.URLParam(r, "userID")

	if userID == targetID {
		writeError(w, http.StatusBadRequest, "cannot follow yourself")
		return
	}

	if _, err := h.users.GetByID(targetID); err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := h.social.Follow(userID, targetID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to follow user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "followed"})
}

func (h *SocialHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	targetID := chi.URLParam(r, "userID")

	if err := h.social.Unfollow(userID, targetID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unfollow user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "unfollowed"})
}

func (h *SocialHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	followers, err := h.social.GetFollowers(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get followers")
		return
	}

	writeJSON(w, http.StatusOK, followers)
}

func (h *SocialHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	following, err := h.social.GetFollowing(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get following")
		return
	}

	writeJSON(w, http.StatusOK, following)
}

// Friend Requests

func (h *SocialHandler) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	targetID := chi.URLParam(r, "userID")

	if userID == targetID {
		writeError(w, http.StatusBadRequest, "cannot send friend request to yourself")
		return
	}

	if _, err := h.users.GetByID(targetID); err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	// Check if already friends
	areFriends, _ := h.social.AreFriends(userID, targetID)
	if areFriends {
		writeError(w, http.StatusConflict, "already friends")
		return
	}

	// Check for existing pending request
	if existing, _ := h.social.GetPendingRequest(userID, targetID); existing != nil {
		writeError(w, http.StatusConflict, "friend request already sent")
		return
	}

	// Check if the other user already sent us a request - auto-accept
	if incoming, _ := h.social.GetPendingRequest(targetID, userID); incoming != nil {
		if err := h.social.AcceptFriendRequest(incoming.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to accept friend request")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "friend request accepted (mutual)"})
		return
	}

	id := newID()
	if err := h.social.CreateFriendRequest(id, userID, targetID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to send friend request")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "friend request sent", "id": id})
}

func (h *SocialHandler) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := chi.URLParam(r, "requestID")

	req, err := h.social.GetFriendRequest(requestID)
	if err != nil {
		writeError(w, http.StatusNotFound, "friend request not found")
		return
	}

	if req.ToUserID != userID {
		writeError(w, http.StatusForbidden, "not your friend request")
		return
	}

	if err := h.social.AcceptFriendRequest(requestID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to accept friend request")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "friend request accepted"})
}

func (h *SocialHandler) RejectFriendRequest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := chi.URLParam(r, "requestID")

	req, err := h.social.GetFriendRequest(requestID)
	if err != nil {
		writeError(w, http.StatusNotFound, "friend request not found")
		return
	}

	if req.ToUserID != userID {
		writeError(w, http.StatusForbidden, "not your friend request")
		return
	}

	if err := h.social.RejectFriendRequest(requestID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reject friend request")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "friend request rejected"})
}

func (h *SocialHandler) GetFriends(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	friends, err := h.social.GetFriends(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get friends")
		return
	}

	if friends == nil {
		friends = []*repository.User{}
	}

	writeJSON(w, http.StatusOK, friends)
}

func (h *SocialHandler) GetFriendRequests(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	requests, err := h.social.GetFriendRequests(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get friend requests")
		return
	}

	if requests == nil {
		requests = []*repository.FriendRequest{}
	}

	writeJSON(w, http.StatusOK, requests)
}

func (h *SocialHandler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	friendID := chi.URLParam(r, "friendID")

	if err := h.social.RemoveFriend(userID, friendID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove friend")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "friend removed"})
}

// Feed

func (h *SocialHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	items, err := h.social.GetFeed(userID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get feed")
		return
	}

	if items == nil {
		items = []*repository.FeedItem{}
	}

	writeJSON(w, http.StatusOK, items)
}
