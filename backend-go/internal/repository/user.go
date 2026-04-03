package repository

import (
	"database/sql"
	"time"
)

type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	HashedPassword string    `json:"-"`
	DisplayName    string    `json:"display_name"`
	Bio            string    `json:"bio"`
	Avatar         string    `json:"avatar,omitempty"`
	EmailVerified  bool      `json:"email_verified"`
	ProfilePublic  bool      `json:"profile_public"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *User) error {
	_, err := r.db.Exec(
		`INSERT INTO users (id, email, username, hashed_password, display_name) VALUES (?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.Username, user.HashedPassword, user.DisplayName,
	)
	return err
}

func (r *UserRepository) GetByID(id string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT id, email, username, hashed_password, display_name, bio, avatar, email_verified, profile_public, created_at, updated_at FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Email, &user.Username, &user.HashedPassword, &user.DisplayName, &user.Bio, &user.Avatar, &user.EmailVerified, &user.ProfilePublic, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(email string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT id, email, username, hashed_password, display_name, bio, avatar, email_verified, profile_public, created_at, updated_at FROM users WHERE email = ?`,
		email,
	).Scan(&user.ID, &user.Email, &user.Username, &user.HashedPassword, &user.DisplayName, &user.Bio, &user.Avatar, &user.EmailVerified, &user.ProfilePublic, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(username string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT id, email, username, hashed_password, display_name, bio, avatar, email_verified, profile_public, created_at, updated_at FROM users WHERE username = ?`,
		username,
	).Scan(&user.ID, &user.Email, &user.Username, &user.HashedPassword, &user.DisplayName, &user.Bio, &user.Avatar, &user.EmailVerified, &user.ProfilePublic, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmailOrUsername(login string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT id, email, username, hashed_password, display_name, bio, avatar, email_verified, profile_public, created_at, updated_at FROM users WHERE email = ? OR username = ?`,
		login, login,
	).Scan(&user.ID, &user.Email, &user.Username, &user.HashedPassword, &user.DisplayName, &user.Bio, &user.Avatar, &user.EmailVerified, &user.ProfilePublic, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Update(user *User) error {
	_, err := r.db.Exec(
		`UPDATE users SET display_name = ?, bio = ?, avatar = ?, profile_public = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		user.DisplayName, user.Bio, user.Avatar, user.ProfilePublic, user.ID,
	)
	return err
}

func (r *UserRepository) UpdatePassword(id, hashedPassword string) error {
	_, err := r.db.Exec(`UPDATE users SET hashed_password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, hashedPassword, id)
	return err
}

func (r *UserRepository) VerifyEmail(id string) error {
	_, err := r.db.Exec(`UPDATE users SET email_verified = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}

func (r *UserRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}
