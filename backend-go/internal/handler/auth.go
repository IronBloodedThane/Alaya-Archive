package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/alaya-archive/backend-go/internal/auth"
	"github.com/alaya-archive/backend-go/internal/config"
	"github.com/alaya-archive/backend-go/internal/middleware"
	"github.com/alaya-archive/backend-go/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	users *repository.UserRepository
	cfg   *config.Config
}

func NewAuthHandler(users *repository.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{users: users, cfg: cfg}
}

type registerRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))

	if req.Email == "" || req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email, username, and password are required")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if len(req.Username) < 3 {
		writeError(w, http.StatusBadRequest, "username must be at least 3 characters")
		return
	}

	if existing, _ := h.users.GetByEmail(req.Email); existing != nil {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}
	if existing, _ := h.users.GetByUsername(req.Username); existing != nil {
		writeError(w, http.StatusConflict, "username already taken")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := &repository.User{
		ID:             newID(),
		Email:          req.Email,
		Username:       req.Username,
		HashedPassword: string(hashedPassword),
		DisplayName:    req.DisplayName,
	}
	if user.DisplayName == "" {
		user.DisplayName = req.Username
	}

	if err := h.users.Create(user); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Generate verification token (in production, send via email)
	_, _ = auth.CreateToken(user.ID, auth.EmailVerification, h.cfg.SecretKey, 24*time.Hour)

	tokens, err := h.createTokenPair(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create tokens")
		return
	}

	writeJSON(w, http.StatusCreated, tokens)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.users.GetByEmailOrUsername(strings.TrimSpace(strings.ToLower(req.Login)))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tokens, err := h.createTokenPair(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create tokens")
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims, err := auth.ValidateToken(req.RefreshToken, auth.RefreshToken, h.cfg.SecretKey)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	tokens, err := h.createTokenPair(claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create tokens")
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims, err := auth.ValidateToken(req.Token, auth.EmailVerification, h.cfg.SecretKey)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid or expired verification token")
		return
	}

	if err := h.users.VerifyEmail(claims.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to verify email")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "email verified"})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Always return success to avoid email enumeration
	writeJSON(w, http.StatusOK, map[string]string{"message": "if that email exists, a reset link has been sent"})

	user, err := h.users.GetByEmail(strings.TrimSpace(strings.ToLower(req.Email)))
	if err != nil || user == nil {
		return
	}

	// Generate reset token (in production, send via email)
	_, _ = auth.CreateToken(user.ID, auth.PasswordReset, h.cfg.SecretKey, time.Hour)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	claims, err := auth.ValidateToken(req.Token, auth.PasswordReset, h.cfg.SecretKey)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid or expired reset token")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	if err := h.users.UpdatePassword(claims.UserID, string(hashed)); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update password")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successfully"})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	user, err := h.users.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.CurrentPassword)); err != nil {
		writeError(w, http.StatusUnauthorized, "current password is incorrect")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	if err := h.users.UpdatePassword(userID, string(hashed)); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update password")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password changed"})
}

func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Confirmation string `json:"confirmation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Confirmation != "DELETE" {
		writeError(w, http.StatusBadRequest, "confirmation must be 'DELETE'")
		return
	}

	if err := h.users.Delete(userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete account")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "account deleted"})
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.users.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.EmailVerified {
		writeError(w, http.StatusBadRequest, "email already verified")
		return
	}

	_, _ = auth.CreateToken(user.ID, auth.EmailVerification, h.cfg.SecretKey, 24*time.Hour)

	writeJSON(w, http.StatusOK, map[string]string{"message": "verification email sent"})
}

func (h *AuthHandler) createTokenPair(userID string) (*tokenResponse, error) {
	accessToken, err := auth.CreateToken(userID, auth.AccessToken, h.cfg.SecretKey, time.Duration(h.cfg.AccessTokenExpireMin)*time.Minute)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.CreateToken(userID, auth.RefreshToken, h.cfg.SecretKey, time.Duration(h.cfg.RefreshTokenExpireDays)*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &tokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
