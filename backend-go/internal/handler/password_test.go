package handler

import (
	"net/http"
	"testing"
)

// -----------------------------------------------------------------------------
// Forgot Password
// -----------------------------------------------------------------------------

func TestForgotPassword_ExistingUser_SendsEmail(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
		"email": "alice@example.com",
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if len(env.mailer.resetCalls) != 1 {
		t.Fatalf("expected 1 reset email, got %d", len(env.mailer.resetCalls))
	}
	if env.mailer.resetCalls[0].To != "alice@example.com" {
		t.Errorf("reset sent to %q, expected alice@example.com", env.mailer.resetCalls[0].To)
	}
}

func TestForgotPassword_NonexistentUser_SilentSuccess(t *testing.T) {
	env := newTestEnv(t)

	rec := env.do(http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
		"email": "ghost@example.com",
	}, "")

	// Must return 200 regardless to avoid email enumeration.
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d (must be 200 for anti-enumeration)", rec.Code)
	}
	if len(env.mailer.resetCalls) != 0 {
		t.Errorf("reset email sent for non-user: %+v", env.mailer.resetCalls)
	}
}

func TestForgotPassword_CaseInsensitive(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "supersecret")

	_ = env.do(http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
		"email": "ALICE@EXAMPLE.COM",
	}, "")

	if len(env.mailer.resetCalls) != 1 {
		t.Fatalf("expected 1 reset email on case-mismatched lookup, got %d", len(env.mailer.resetCalls))
	}
}

// -----------------------------------------------------------------------------
// Reset Password
// -----------------------------------------------------------------------------

func TestResetPassword_EndToEnd(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "oldpassword")

	// Trigger reset email.
	_ = env.do(http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
		"email": "alice@example.com",
	}, "")
	if len(env.mailer.resetCalls) != 1 {
		t.Fatalf("expected reset email, got %d", len(env.mailer.resetCalls))
	}
	token := tokenFromURL(t, env.mailer.resetCalls[0].URL)

	// Consume token to set new password.
	rec := env.do(http.MethodPost, "/api/v1/auth/reset-password", map[string]string{
		"token":        token,
		"new_password": "newpassword1",
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("reset status = %d, body = %s", rec.Code, rec.Body.String())
	}

	// New password must work.
	rec = env.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"login":    "alice",
		"password": "newpassword1",
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login with new password: status = %d", rec.Code)
	}

	// Old password must be rejected.
	rec = env.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"login":    "alice",
		"password": "oldpassword",
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("old password should be rejected, got status = %d", rec.Code)
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/reset-password", map[string]string{
		"token":        "garbage",
		"new_password": "newpassword1",
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestResetPassword_WeakPassword(t *testing.T) {
	env := newTestEnv(t)
	env.registerUser("alice@example.com", "alice", "supersecret")

	_ = env.do(http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
		"email": "alice@example.com",
	}, "")
	token := tokenFromURL(t, env.mailer.resetCalls[0].URL)

	rec := env.do(http.MethodPost, "/api/v1/auth/reset-password", map[string]string{
		"token":        token,
		"new_password": "short",
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestResetPassword_WrongTokenType(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/reset-password", map[string]string{
		"token":        access,
		"new_password": "newpassword1",
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("access token accepted as reset token, status = %d", rec.Code)
	}
}

// -----------------------------------------------------------------------------
// Change Password
// -----------------------------------------------------------------------------

func TestChangePassword_Success(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("alice@example.com", "alice", "oldpassword")

	rec := env.do(http.MethodPost, "/api/v1/auth/change-password", map[string]string{
		"current_password": "oldpassword",
		"new_password":     "newpassword1",
	}, access)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	// New password must work.
	rec = env.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"login":    "alice",
		"password": "newpassword1",
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login with new password: status = %d", rec.Code)
	}

	// Old password must not.
	rec = env.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"login":    "alice",
		"password": "oldpassword",
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("old password still works after change, status = %d", rec.Code)
	}
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("alice@example.com", "alice", "oldpassword")

	rec := env.do(http.MethodPost, "/api/v1/auth/change-password", map[string]string{
		"current_password": "wrong",
		"new_password":     "newpassword1",
	}, access)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d (should reject wrong current)", rec.Code)
	}
}

func TestChangePassword_WeakNew(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("alice@example.com", "alice", "oldpassword")

	rec := env.do(http.MethodPost, "/api/v1/auth/change-password", map[string]string{
		"current_password": "oldpassword",
		"new_password":     "short",
	}, access)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestChangePassword_RequiresAuth(t *testing.T) {
	env := newTestEnv(t)

	rec := env.do(http.MethodPost, "/api/v1/auth/change-password", map[string]string{
		"current_password": "oldpassword",
		"new_password":     "newpassword1",
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}
