package database

import (
	"database/sql"
	"fmt"
	"log"
)

type migration struct {
	version int
	name    string
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		name:    "create_users",
		sql: `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			username TEXT UNIQUE NOT NULL,
			hashed_password TEXT NOT NULL,
			display_name TEXT DEFAULT '',
			bio TEXT DEFAULT '',
			avatar TEXT DEFAULT '',
			email_verified INTEGER DEFAULT 0,
			profile_public INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		`,
	},
	{
		version: 2,
		name:    "create_media",
		sql: `
		CREATE TABLE IF NOT EXISTS media (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			media_type TEXT NOT NULL CHECK(media_type IN ('manga', 'movie', 'anime', 'book', 'game', 'tv_show', 'music', 'other')),
			title TEXT NOT NULL,
			title_original TEXT DEFAULT '',
			description TEXT DEFAULT '',
			cover_image TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'planned' CHECK(status IN ('planned', 'in_progress', 'completed', 'dropped', 'on_hold')),
			rating INTEGER DEFAULT NULL CHECK(rating IS NULL OR (rating >= 1 AND rating <= 10)),
			notes TEXT DEFAULT '',
			year_released INTEGER DEFAULT NULL,
			creator TEXT DEFAULT '',
			genre TEXT DEFAULT '',
			volumes_total INTEGER DEFAULT NULL,
			volumes_owned INTEGER DEFAULT NULL,
			episodes_total INTEGER DEFAULT NULL,
			episodes_watched INTEGER DEFAULT NULL,
			chapters_total INTEGER DEFAULT NULL,
			chapters_read INTEGER DEFAULT NULL,
			is_public INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_media_user_id ON media(user_id);
		CREATE INDEX IF NOT EXISTS idx_media_type ON media(media_type);
		CREATE INDEX IF NOT EXISTS idx_media_status ON media(status);
		`,
	},
	{
		version: 3,
		name:    "create_tags",
		sql: `
		CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		);
		CREATE TABLE IF NOT EXISTS media_tags (
			media_id TEXT NOT NULL REFERENCES media(id) ON DELETE CASCADE,
			tag_id TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			PRIMARY KEY (media_id, tag_id)
		);
		CREATE INDEX IF NOT EXISTS idx_media_tags_media_id ON media_tags(media_id);
		`,
	},
	{
		version: 4,
		name:    "create_social",
		sql: `
		CREATE TABLE IF NOT EXISTS follows (
			follower_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			following_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (follower_id, following_id),
			CHECK (follower_id != following_id)
		);

		CREATE TABLE IF NOT EXISTS friend_requests (
			id TEXT PRIMARY KEY,
			from_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			to_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'accepted', 'rejected')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(from_user_id, to_user_id)
		);

		CREATE TABLE IF NOT EXISTS friends (
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			friend_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, friend_id),
			CHECK (user_id != friend_id)
		);
		`,
	},
	{
		version: 5,
		name:    "create_feed",
		sql: `
		CREATE TABLE IF NOT EXISTS feed_items (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			item_type TEXT NOT NULL,
			data TEXT DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_feed_items_user_id ON feed_items(user_id);
		CREATE INDEX IF NOT EXISTS idx_feed_items_created_at ON feed_items(created_at);
		`,
	},
	{
		version: 6,
		name:    "create_notifications",
		sql: `
		CREATE TABLE IF NOT EXISTS notifications (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			type TEXT NOT NULL,
			title TEXT NOT NULL,
			message TEXT DEFAULT '',
			link TEXT DEFAULT '',
			is_read INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
		`,
	},
	{
		version: 7,
		name:    "avatar_blob",
		sql: `
		ALTER TABLE users DROP COLUMN avatar;
		ALTER TABLE users ADD COLUMN avatar_data BLOB;
		ALTER TABLE users ADD COLUMN avatar_mime TEXT DEFAULT '';
		`,
	},
}

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	for _, m := range migrations {
		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}
		if exists > 0 {
			continue
		}

		log.Printf("applying migration %d: %s", m.version, m.name)
		if _, err := db.Exec(m.sql); err != nil {
			return fmt.Errorf("migration %d (%s): %w", m.version, m.name, err)
		}
		if _, err := db.Exec("INSERT INTO schema_migrations (version, name) VALUES (?, ?)", m.version, m.name); err != nil {
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
	}

	return nil
}
