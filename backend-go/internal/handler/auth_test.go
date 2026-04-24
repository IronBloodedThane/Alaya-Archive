package handler

import (
	"net/http"
	"testing"
	"time"

	"github.com/alaya-archive/backend-go/internal/auth"
)

// -----------------------------------------------------------------------------
// Register
// -----------------------------------------------------------------------------

func TestRegister_Success(t *testing.T) {
	env := newTestEnv(t)

	rec := env.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":        "alice@example.com",
		"username":     "alice",
		"password":     "supersecret",
		"display_name": "Alice",
	}, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	decodeJSON(t, rec, &body)
	if body.AccessToken == "" || body.RefreshToken == "" {
		t.Fatalf("missing tokens in response: %+v", body)
	}

	claims, err := auth.ValidateToken(body.AccessToken, auth.AccessToken, env.cfg.SecretKey)
	if err != nil {
		t.Fatalf("access token invalid: %v", err)
	}
	if claims.UserID == "" {
		t.Fatal("access token has empty user id")
	}

	if len(env.mailer.verifyCalls) != 1 {
		t.Fatalf("expected 1 verification email, got %d", len(env.mailer.verifyCalls))
	}
	if env.mailer.verifyCalls[0].To != "alice@example.com" {
		t.Fatalf("verification sent to wrong address: %s", env.mailer.verifyCalls[0].To)
	}

	user, err := env.userRepo.GetByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("user not in db: %v", err)
	}
	if user.EmailVerified {
		t.Error("user should not be email_verified yet")
	}
}

func TestRegister_NormalizesCase(t *testing.T) {
	env := newTestEnv(t)

	_ = env.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    "Alice@Example.COM",
		"username": "ALICE",
		"password": "supersecret",
	}, "")

	user, err := env.userRepo.GetByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("expected lowercase email lookup to succeed: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("username not lowercased: %q", user.Username)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    "alice@example.com",
		"username": "alice2",
		"password": "supersecret",
	}, "")

	code, msg := statusAndError(t, rec)
	if code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", code, rec.Body.String())
	}
	if msg == "" {
		t.Error("expected error message on conflict")
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    "alice2@example.com",
		"username": "alice",
		"password": "supersecret",
	}, "")

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	env := newTestEnv(t)

	rec := env.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    "alice@example.com",
		"username": "alice",
		"password": "short",
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRegister_ShortUsername(t *testing.T) {
	env := newTestEnv(t)

	rec := env.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    "alice@example.com",
		"username": "ab",
		"password": "supersecret",
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRegister_MissingFields(t *testing.T) {
	env := newTestEnv(t)

	rec := env.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    "alice@example.com",
		"username": "alice",
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

// -----------------------------------------------------------------------------
// Login
// -----------------------------------------------------------------------------

func TestLogin_WithEmail(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"login":    "alice@example.com",
		"password": "supersecret",
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestLogin_WithUsername(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"login":    "alice",
		"password": "supersecret",
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestLogin_CaseInsensitive(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"login":    "ALICE@EXAMPLE.COM",
		"password": "supersecret",
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"login":    "alice",
		"password": "nope",
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	env := newTestEnv(t)

	rec := env.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"login":    "ghost@example.com",
		"password": "whatever",
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

// -----------------------------------------------------------------------------
// Refresh
// -----------------------------------------------------------------------------

func TestRefresh_Success(t *testing.T) {
	env := newTestEnv(t)
	_, refresh := env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": refresh,
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	decodeJSON(t, rec, &body)
	if body.AccessToken == "" || body.RefreshToken == "" {
		t.Fatal("missing tokens on refresh")
	}
}

func TestRefresh_RejectsAccessTokenAsRefresh(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": access,
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	env := newTestEnv(t)

	rec := env.do(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": "not-a-token",
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

// -----------------------------------------------------------------------------
// Verify Email
// -----------------------------------------------------------------------------

func TestVerifyEmail_Success(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "supersecret")

	if len(env.mailer.verifyCalls) != 1 {
		t.Fatalf("expected 1 verification mail, got %d", len(env.mailer.verifyCalls))
	}
	token := tokenFromURL(t, env.mailer.verifyCalls[0].URL)

	rec := env.do(http.MethodPost, "/api/v1/auth/verify-email", map[string]string{
		"token": token,
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	user, err := env.userRepo.GetByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if !user.EmailVerified {
		t.Error("user.EmailVerified should be true after verify")
	}
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	env := newTestEnv(t)

	rec := env.do(http.MethodPost, "/api/v1/auth/verify-email", map[string]string{
		"token": "garbage",
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestVerifyEmail_WrongTokenType(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("alice@example.com", "alice", "supersecret")

	// Access token shouldn't work for verification.
	rec := env.do(http.MethodPost, "/api/v1/auth/verify-email", map[string]string{
		"token": access,
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

// -----------------------------------------------------------------------------
// Resend verification
// -----------------------------------------------------------------------------

func TestResendVerification_Success(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("alice@example.com", "alice", "supersecret")

	initial := len(env.mailer.verifyCalls)

	rec := env.do(http.MethodPost, "/api/v1/auth/resend-verification", nil, access)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(env.mailer.verifyCalls) != initial+1 {
		t.Fatalf("expected a new verify email, got %d total", len(env.mailer.verifyCalls))
	}
}

func TestResendVerification_AlreadyVerified(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("alice@example.com", "alice", "supersecret")

	// Manually verify the user.
	user, _ := env.userRepo.GetByEmail("alice@example.com")
	if err := env.userRepo.VerifyEmail(user.ID); err != nil {
		t.Fatalf("verify: %v", err)
	}

	rec := env.do(http.MethodPost, "/api/v1/auth/resend-verification", nil, access)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestResendVerification_RequiresAuth(t *testing.T) {
	env := newTestEnv(t)

	rec := env.do(http.MethodPost, "/api/v1/auth/resend-verification", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

// Regression cover: verification URL uses FRONTEND_URL + /verify-email?token=...
func TestVerifyEmail_URLFormat(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "supersecret")

	got := env.mailer.verifyCalls[0].URL
	want := env.cfg.FrontendURL + "/verify-email?token="
	if len(got) <= len(want) || got[:len(want)] != want {
		t.Fatalf("verify URL %q does not start with %q", got, want)
	}
	_ = time.Second // keep time import if tests add time-based checks later
}
