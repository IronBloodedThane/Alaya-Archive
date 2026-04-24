package main

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/alaya-archive/backend-go/internal/config"
	"github.com/alaya-archive/backend-go/internal/email"
	"github.com/alaya-archive/backend-go/internal/handler"
	"github.com/alaya-archive/backend-go/internal/middleware"
	"github.com/alaya-archive/backend-go/internal/repository"
)

func NewRouter(db *sql.DB, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.CORS(cfg.CORSOrigins))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	userRepo := repository.NewUserRepository(db)
	mediaRepo := repository.NewMediaRepository(db)
	socialRepo := repository.NewSocialRepository(db)

	mailer := email.NewMailer(cfg.EmailAPIKey, cfg.EmailFrom)

	authHandler := handler.NewAuthHandler(userRepo, mailer, cfg)
	userHandler := handler.NewUserHandler(userRepo, cfg)
	mediaHandler := handler.NewMediaHandler(mediaRepo, userRepo, cfg)
	socialHandler := handler.NewSocialHandler(socialRepo, userRepo, cfg)

	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/verify-email", authHandler.VerifyEmail)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(cfg.SecretKey))

			// Auth
			r.Post("/auth/change-password", authHandler.ChangePassword)
			r.Post("/auth/delete-account", authHandler.DeleteAccount)
			r.Post("/auth/resend-verification", authHandler.ResendVerification)

			// Users
			r.Get("/users/me", userHandler.GetCurrentUser)
			r.Patch("/users/me", userHandler.UpdateProfile)
			r.Post("/users/me/avatar", userHandler.UploadAvatar)
			r.Delete("/users/me/avatar", userHandler.DeleteAvatar)

			// Media collection
			r.Route("/media", func(r chi.Router) {
				r.Get("/", mediaHandler.ListMedia)
				r.Post("/", mediaHandler.CreateMedia)
				r.Get("/stats", mediaHandler.GetStats)
				r.Get("/search", mediaHandler.SearchMedia)
				r.Get("/{mediaID}", mediaHandler.GetMedia)
				r.Patch("/{mediaID}", mediaHandler.UpdateMedia)
				r.Delete("/{mediaID}", mediaHandler.DeleteMedia)
				r.Post("/{mediaID}/rating", mediaHandler.RateMedia)
				r.Post("/{mediaID}/tags", mediaHandler.AddTags)
			})

			// Social
			r.Route("/social", func(r chi.Router) {
				r.Post("/follow/{userID}", socialHandler.FollowUser)
				r.Delete("/follow/{userID}", socialHandler.UnfollowUser)
				r.Get("/followers", socialHandler.GetFollowers)
				r.Get("/following", socialHandler.GetFollowing)
				r.Get("/feed", socialHandler.GetFeed)
			})

			// Friend requests
			r.Route("/friends", func(r chi.Router) {
				r.Post("/request/{userID}", socialHandler.SendFriendRequest)
				r.Post("/accept/{requestID}", socialHandler.AcceptFriendRequest)
				r.Post("/reject/{requestID}", socialHandler.RejectFriendRequest)
				r.Get("/", socialHandler.GetFriends)
				r.Get("/requests", socialHandler.GetFriendRequests)
				r.Delete("/{friendID}", socialHandler.RemoveFriend)
			})
		})

		// Public avatar — no auth needed so <img> tags can load it
		r.Get("/users/{username}/avatar", userHandler.GetAvatar)

		// Public profile (optional auth for follow status)
		r.Group(func(r chi.Router) {
			r.Use(middleware.OptionalAuth(cfg.SecretKey))
			r.Get("/users/{username}", userHandler.GetPublicProfile)
			r.Get("/users/{username}/collection", mediaHandler.GetPublicCollection)
		})
	})

	return r
}
