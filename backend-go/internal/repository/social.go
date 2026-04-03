package repository

import (
	"database/sql"
	"time"
)

type Follow struct {
	FollowerID  string    `json:"follower_id"`
	FollowingID string    `json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type FriendRequest struct {
	ID         string    `json:"id"`
	FromUserID string    `json:"from_user_id"`
	ToUserID   string    `json:"to_user_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	// Populated from join
	FromUsername    string `json:"from_username,omitempty"`
	FromDisplayName string `json:"from_display_name,omitempty"`
	ToUsername      string `json:"to_username,omitempty"`
	ToDisplayName   string `json:"to_display_name,omitempty"`
}

type FeedItem struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ItemType  string    `json:"item_type"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	// Populated from join
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

type SocialRepository struct {
	db *sql.DB
}

func NewSocialRepository(db *sql.DB) *SocialRepository {
	return &SocialRepository{db: db}
}

// Follows

func (r *SocialRepository) Follow(followerID, followingID string) error {
	_, err := r.db.Exec(`INSERT OR IGNORE INTO follows (follower_id, following_id) VALUES (?, ?)`, followerID, followingID)
	return err
}

func (r *SocialRepository) Unfollow(followerID, followingID string) error {
	_, err := r.db.Exec(`DELETE FROM follows WHERE follower_id = ? AND following_id = ?`, followerID, followingID)
	return err
}

func (r *SocialRepository) GetFollowers(userID string) ([]*User, error) {
	rows, err := r.db.Query(
		`SELECT u.id, u.email, u.username, '', u.display_name, u.bio, u.avatar, u.email_verified, u.profile_public, u.created_at, u.updated_at
		 FROM users u JOIN follows f ON u.id = f.follower_id WHERE f.following_id = ?`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUsers(rows)
}

func (r *SocialRepository) GetFollowing(userID string) ([]*User, error) {
	rows, err := r.db.Query(
		`SELECT u.id, u.email, u.username, '', u.display_name, u.bio, u.avatar, u.email_verified, u.profile_public, u.created_at, u.updated_at
		 FROM users u JOIN follows f ON u.id = f.following_id WHERE f.follower_id = ?`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUsers(rows)
}

// Friend Requests

func (r *SocialRepository) CreateFriendRequest(id, fromUserID, toUserID string) error {
	_, err := r.db.Exec(
		`INSERT INTO friend_requests (id, from_user_id, to_user_id) VALUES (?, ?, ?)`,
		id, fromUserID, toUserID,
	)
	return err
}

func (r *SocialRepository) GetFriendRequest(id string) (*FriendRequest, error) {
	req := &FriendRequest{}
	err := r.db.QueryRow(
		`SELECT fr.id, fr.from_user_id, fr.to_user_id, fr.status, fr.created_at, fr.updated_at
		 FROM friend_requests fr WHERE fr.id = ?`, id,
	).Scan(&req.ID, &req.FromUserID, &req.ToUserID, &req.Status, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (r *SocialRepository) GetPendingRequest(fromUserID, toUserID string) (*FriendRequest, error) {
	req := &FriendRequest{}
	err := r.db.QueryRow(
		`SELECT id, from_user_id, to_user_id, status, created_at, updated_at
		 FROM friend_requests WHERE from_user_id = ? AND to_user_id = ? AND status = 'pending'`,
		fromUserID, toUserID,
	).Scan(&req.ID, &req.FromUserID, &req.ToUserID, &req.Status, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (r *SocialRepository) AcceptFriendRequest(requestID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var fromID, toID string
	err = tx.QueryRow(`SELECT from_user_id, to_user_id FROM friend_requests WHERE id = ? AND status = 'pending'`, requestID).Scan(&fromID, &toID)
	if err != nil {
		return err
	}

	tx.Exec(`UPDATE friend_requests SET status = 'accepted', updated_at = CURRENT_TIMESTAMP WHERE id = ?`, requestID)
	tx.Exec(`INSERT OR IGNORE INTO friends (user_id, friend_id) VALUES (?, ?)`, fromID, toID)
	tx.Exec(`INSERT OR IGNORE INTO friends (user_id, friend_id) VALUES (?, ?)`, toID, fromID)

	return tx.Commit()
}

func (r *SocialRepository) RejectFriendRequest(requestID string) error {
	_, err := r.db.Exec(`UPDATE friend_requests SET status = 'rejected', updated_at = CURRENT_TIMESTAMP WHERE id = ?`, requestID)
	return err
}

func (r *SocialRepository) GetFriends(userID string) ([]*User, error) {
	rows, err := r.db.Query(
		`SELECT u.id, u.email, u.username, '', u.display_name, u.bio, u.avatar, u.email_verified, u.profile_public, u.created_at, u.updated_at
		 FROM users u JOIN friends f ON u.id = f.friend_id WHERE f.user_id = ?`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUsers(rows)
}

func (r *SocialRepository) GetFriendRequests(userID string) ([]*FriendRequest, error) {
	rows, err := r.db.Query(
		`SELECT fr.id, fr.from_user_id, fr.to_user_id, fr.status, fr.created_at, fr.updated_at,
		        fu.username, fu.display_name, tu.username, tu.display_name
		 FROM friend_requests fr
		 JOIN users fu ON fr.from_user_id = fu.id
		 JOIN users tu ON fr.to_user_id = tu.id
		 WHERE (fr.to_user_id = ? OR fr.from_user_id = ?) AND fr.status = 'pending'
		 ORDER BY fr.created_at DESC`, userID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*FriendRequest
	for rows.Next() {
		req := &FriendRequest{}
		if err := rows.Scan(&req.ID, &req.FromUserID, &req.ToUserID, &req.Status, &req.CreatedAt, &req.UpdatedAt, &req.FromUsername, &req.FromDisplayName, &req.ToUsername, &req.ToDisplayName); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *SocialRepository) RemoveFriend(userID, friendID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Exec(`DELETE FROM friends WHERE user_id = ? AND friend_id = ?`, userID, friendID)
	tx.Exec(`DELETE FROM friends WHERE user_id = ? AND friend_id = ?`, friendID, userID)

	return tx.Commit()
}

func (r *SocialRepository) AreFriends(userID, otherID string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM friends WHERE user_id = ? AND friend_id = ?`, userID, otherID).Scan(&count)
	return count > 0, err
}

// Feed

func (r *SocialRepository) CreateFeedItem(item *FeedItem) error {
	_, err := r.db.Exec(
		`INSERT INTO feed_items (id, user_id, item_type, data) VALUES (?, ?, ?, ?)`,
		item.ID, item.UserID, item.ItemType, item.Data,
	)
	return err
}

func (r *SocialRepository) GetFeed(userID string, limit, offset int) ([]*FeedItem, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.Query(
		`SELECT fi.id, fi.user_id, fi.item_type, fi.data, fi.created_at, u.username, u.display_name
		 FROM feed_items fi
		 JOIN users u ON fi.user_id = u.id
		 WHERE fi.user_id IN (SELECT following_id FROM follows WHERE follower_id = ?)
		    OR fi.user_id IN (SELECT friend_id FROM friends WHERE user_id = ?)
		 ORDER BY fi.created_at DESC LIMIT ? OFFSET ?`,
		userID, userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*FeedItem
	for rows.Next() {
		item := &FeedItem{}
		if err := rows.Scan(&item.ID, &item.UserID, &item.ItemType, &item.Data, &item.CreatedAt, &item.Username, &item.DisplayName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func scanUsers(rows *sql.Rows) ([]*User, error) {
	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.HashedPassword, &u.DisplayName, &u.Bio, &u.Avatar, &u.EmailVerified, &u.ProfilePublic, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.HashedPassword = ""
		users = append(users, u)
	}
	return users, nil
}
