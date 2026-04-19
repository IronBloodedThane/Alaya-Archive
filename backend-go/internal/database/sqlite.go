package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	journalMode := os.Getenv("DB_JOURNAL_MODE")
	if journalMode == "" {
		journalMode = "WAL"
	}

	db, err := sql.Open("sqlite", path+"?_journal_mode="+journalMode+"&_busy_timeout=5000&_foreign_keys=ON")
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
