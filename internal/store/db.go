// Package store provides SQLite persistence for Kindle accounts, Audiobookshelf
// users, sync profiles, book mappings, and sync history.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS kindle_accounts (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    label         TEXT NOT NULL,
    ubid_main     TEXT NOT NULL,
    at_main       TEXT NOT NULL,
    session_id    TEXT NOT NULL,
    x_main        TEXT NOT NULL,
    device_token  TEXT NOT NULL,
    tls_proxy_url TEXT NOT NULL DEFAULT '',
    tls_proxy_key TEXT NOT NULL DEFAULT '',
    created       DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS abs_users (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    label     TEXT NOT NULL,
    base_url  TEXT NOT NULL,
    api_token TEXT NOT NULL,
    created   DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS profiles (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    label             TEXT NOT NULL,
    kindle_account_id INTEGER NOT NULL REFERENCES kindle_accounts(id) ON DELETE CASCADE,
    abs_user_id       INTEGER NOT NULL REFERENCES abs_users(id) ON DELETE CASCADE,
    abs_library_id    TEXT NOT NULL DEFAULT '',
    poll_minutes      INTEGER NOT NULL DEFAULT 15,
    created           DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS book_mappings (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_id      INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    kindle_asin     TEXT NOT NULL,
    kindle_title    TEXT NOT NULL DEFAULT '',
    abs_item_id     TEXT NOT NULL,
    abs_title       TEXT NOT NULL DEFAULT '',
    confidence      REAL NOT NULL DEFAULT 0,
    confirmed       INTEGER NOT NULL DEFAULT 0,
    last_kindle_pct REAL NOT NULL DEFAULT 0,
    last_abs_pct    REAL NOT NULL DEFAULT 0,
    last_synced     DATETIME,
    created         DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(profile_id, kindle_asin, abs_item_id)
);

CREATE TABLE IF NOT EXISTS rejected_matches (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_id  INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    kindle_asin TEXT NOT NULL,
    abs_item_id TEXT NOT NULL,
    created     DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(profile_id, kindle_asin, abs_item_id)
);

CREATE TABLE IF NOT EXISTS sync_events (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    mapping_id INTEGER NOT NULL REFERENCES book_mappings(id) ON DELETE CASCADE,
    direction  TEXT NOT NULL,
    from_pct   REAL NOT NULL DEFAULT 0,
    to_pct     REAL NOT NULL DEFAULT 0,
    message    TEXT NOT NULL DEFAULT '',
    created    DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

// Open creates (if needed) the data directory and SQLite file at path, applies
// the schema, and returns a ready-to-use *sql.DB.
func Open(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("store: creating data dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("store: opening database: %w", err)
	}
	db.SetMaxOpenConns(1) // modernc.org/sqlite: avoid SQLITE_BUSY under concurrent writers

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: applying schema: %w", err)
	}

	return db, nil
}
