package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/alaya-archive/backend-go/internal/config"
	"github.com/alaya-archive/backend-go/internal/database"
	"github.com/alaya-archive/backend-go/internal/middleware"
	"github.com/alaya-archive/backend-go/internal/repository"

	_ "modernc.org/sqlite"
)

// fakeMailer captures SendVerification and SendPasswordReset calls so tests
// can extract the verification/reset URL (which contains the JWT token).
type fakeMailer struct {
	verifyCalls []mailCall
	resetCalls  []mailCall
}

type mailCall struct {
	To  string
	URL string
}

func (f *fakeMailer) SendVerification(to, verifyURL string) error {
	f.verifyCalls = append(f.verifyCalls, mailCall{To: to, URL: verifyURL})
	return nil
}

func (f *fakeMailer) SendPasswordReset(to, resetURL string) error {
	f.resetCalls = append(f.resetCalls, mailCall{To: to, URL: resetURL})
	return nil
}

// tokenFromURL pulls the ?token=... query parameter out of a verify/reset URL.
func tokenFromURL(t *testing.T, raw string) string {
	t.Helper()
	const needle = "token="
	idx := len(raw) - 1
	for i := 0; i+len(needle) <= len(raw); i++ {
		if raw[i:i+len(needle)] == needle {
			idx = i + len(needle)
			break
		}
	}
	return raw[idx:]
}

// testEnv is the minimal set of wiring every auth test needs.
type testEnv struct {
	t        *testing.T
	db       *sql.DB
	cfg      *config.Config
	mailer   *fakeMailer
	userRepo *repository.UserRepository
	auth     *AuthHandler
	user     *UserHandler
	router   *chi.Mux
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	// In-memory DB must not be dropped between connections.
	db.SetMaxOpenConns(1)

	// Mirror the production DSN's foreign-key enforcement so cascade tests
	// actually test real behavior.
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}
	var fk int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil || fk != 1 {
		t.Fatalf("foreign_keys not enabled (fk=%d, err=%v)", fk, err)
	}

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cfg := &config.Config{
		SecretKey:              "test-secret-key-0123456789abcdef",
		FrontendURL:            "http://localhost:5173",
		AccessTokenExpireMin:   15,
		RefreshTokenExpireDays: 30,
	}
	mailer := &fakeMailer{}
	userRepo := repository.NewUserRepository(db)

	authHandler := NewAuthHandler(userRepo, mailer, cfg)
	userHandler := NewUserHandler(userRepo, cfg)

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/verify-email", authHandler.VerifyEmail)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(cfg.SecretKey))
			r.Post("/auth/change-password", authHandler.ChangePassword)
			r.Post("/auth/delete-account", authHandler.DeleteAccount)
			r.Post("/auth/resend-verification", authHandler.ResendVerification)
			r.Get("/users/me", userHandler.GetCurrentUser)
		})
	})

	t.Cleanup(func() { db.Close() })

	return &testEnv{
		t:        t,
		db:       db,
		cfg:      cfg,
		mailer:   mailer,
		userRepo: userRepo,
		auth:     authHandler,
		user:     userHandler,
		router:   r,
	}
}

// do sends a request through the router and returns the response. Body can be
// nil for GETs or any JSON-marshalable value for POSTs.
func (e *testEnv) do(method, path string, body any, bearer string) *httptest.ResponseRecorder {
	e.t.Helper()

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			e.t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}

	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	return rec
}

// decodeJSON unmarshals the response body into v.
func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("decode response body %q: %v", rec.Body.String(), err)
	}
}

// registerUser is a convenience that creates a user through the real endpoint
// and returns the tokens plus the email used.
func (e *testEnv) registerUser(email, username, password string) (accessToken, refreshToken string) {
	e.t.Helper()

	rec := e.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":        email,
		"username":     username,
		"password":     password,
		"display_name": username,
	}, "")
	if rec.Code != http.StatusCreated {
		e.t.Fatalf("register %s: got %d, body=%s", email, rec.Code, rec.Body.String())
	}

	var tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	decodeJSON(e.t, rec, &tokens)
	return tokens.AccessToken, tokens.RefreshToken
}

// statusAndError returns the status and decoded {"error": "..."} body.
func statusAndError(t *testing.T, rec *httptest.ResponseRecorder) (int, string) {
	t.Helper()
	var body struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body.Error
}
