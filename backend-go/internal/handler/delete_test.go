package handler

import (
	"net/http"
	"testing"

	"github.com/alaya-archive/backend-go/internal/repository"
)

func TestDeleteAccount_Success(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/delete-account", map[string]string{
		"confirmation": "DELETE",
	}, access)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	// User must be gone.
	if _, err := env.userRepo.GetByEmail("alice@example.com"); err == nil {
		t.Error("user still exists after delete")
	}
}

func TestDeleteAccount_WrongConfirmation(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("alice@example.com", "alice", "supersecret")

	rec := env.do(http.MethodPost, "/api/v1/auth/delete-account", map[string]string{
		"confirmation": "delete",
	}, access)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d (should require exact 'DELETE')", rec.Code)
	}

	// User must still exist.
	if _, err := env.userRepo.GetByEmail("alice@example.com"); err != nil {
		t.Errorf("user was deleted despite wrong confirmation: %v", err)
	}
}

func TestDeleteAccount_RequiresAuth(t *testing.T) {
	env := newTestEnv(t)

	rec := env.do(http.MethodPost, "/api/v1/auth/delete-account", map[string]string{
		"confirmation": "DELETE",
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

// TestDeleteAccount_CascadesMedia verifies that FK cascades actually fire. If
// `_foreign_keys=ON` ever gets dropped from the DSN, this test will catch it.
func TestDeleteAccount_CascadesMedia(t *testing.T) {
	env := newTestEnv(t)
	access, _ := env.registerUser("alice@example.com", "alice", "supersecret")

	user, err := env.userRepo.GetByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}

	// Insert a media row owned by the user, bypassing the HTTP layer.
	mediaRepo := repository.NewMediaRepository(env.db)
	if err := mediaRepo.Create(&repository.Media{
		ID:        "media-1",
		UserID:    user.ID,
		MediaType: "manga",
		Title:     "Test Manga",
		Status:    "planned",
	}); err != nil {
		t.Fatalf("insert media: %v", err)
	}

	// Sanity: the media row exists.
	var mediaCountBefore int
	if err := env.db.QueryRow(`SELECT COUNT(*) FROM media WHERE user_id = ?`, user.ID).Scan(&mediaCountBefore); err != nil {
		t.Fatalf("count media before: %v", err)
	}
	if mediaCountBefore != 1 {
		t.Fatalf("expected 1 media row, got %d", mediaCountBefore)
	}

	// Delete the account.
	rec := env.do(http.MethodPost, "/api/v1/auth/delete-account", map[string]string{
		"confirmation": "DELETE",
	}, access)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d", rec.Code)
	}

	// Media rows must be gone via cascade.
	var mediaCountAfter int
	if err := env.db.QueryRow(`SELECT COUNT(*) FROM media WHERE user_id = ?`, user.ID).Scan(&mediaCountAfter); err != nil {
		t.Fatalf("count media after: %v", err)
	}
	if mediaCountAfter != 0 {
		t.Errorf("media not cascaded: %d rows remain", mediaCountAfter)
	}
}

// TestDeleteAccount_CascadesSocial verifies follows and friend_requests are
// cleaned up in both directions when a user is deleted.
func TestDeleteAccount_CascadesSocial(t *testing.T) {
	env := newTestEnv(t)
	aliceAccess, _ := env.registerUser("alice@example.com", "alice", "supersecret")
	_, _ = env.registerUser("bob@example.com", "bob", "supersecret")

	alice, _ := env.userRepo.GetByEmail("alice@example.com")
	bob, _ := env.userRepo.GetByEmail("bob@example.com")

	// Seed a follow (alice -> bob) and a friend request (bob -> alice).
	if _, err := env.db.Exec(`INSERT INTO follows (follower_id, following_id) VALUES (?, ?)`, alice.ID, bob.ID); err != nil {
		t.Fatalf("insert follow: %v", err)
	}
	if _, err := env.db.Exec(`INSERT INTO friend_requests (id, from_user_id, to_user_id) VALUES (?, ?, ?)`, "req-1", bob.ID, alice.ID); err != nil {
		t.Fatalf("insert friend request: %v", err)
	}

	// Delete alice.
	rec := env.do(http.MethodPost, "/api/v1/auth/delete-account", map[string]string{
		"confirmation": "DELETE",
	}, aliceAccess)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d", rec.Code)
	}

	// Follows from alice are gone.
	var followCount int
	env.db.QueryRow(`SELECT COUNT(*) FROM follows WHERE follower_id = ? OR following_id = ?`, alice.ID, alice.ID).Scan(&followCount)
	if followCount != 0 {
		t.Errorf("follows involving alice remain: %d", followCount)
	}

	// Friend requests to alice are gone.
	var reqCount int
	env.db.QueryRow(`SELECT COUNT(*) FROM friend_requests WHERE from_user_id = ? OR to_user_id = ?`, alice.ID, alice.ID).Scan(&reqCount)
	if reqCount != 0 {
		t.Errorf("friend_requests involving alice remain: %d", reqCount)
	}

	// Bob is still around.
	if _, err := env.userRepo.GetByEmail("bob@example.com"); err != nil {
		t.Errorf("bob was collateral damage: %v", err)
	}
}
