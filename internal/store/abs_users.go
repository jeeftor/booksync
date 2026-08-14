package store

import (
	"database/sql"
	"fmt"
)

// ListABSUsers returns all configured Audiobookshelf users.
func ListABSUsers(db *sql.DB) ([]ABSUser, error) {
	rows, err := db.Query(`SELECT id, label, base_url, api_token, created FROM abs_users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("store: listing abs users: %w", err)
	}
	defer rows.Close()

	var out []ABSUser
	for rows.Next() {
		var u ABSUser
		if err := rows.Scan(&u.ID, &u.Label, &u.BaseURL, &u.APIToken, &u.Created); err != nil {
			return nil, fmt.Errorf("store: scanning abs user: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// GetABSUser fetches a single Audiobookshelf user by ID.
func GetABSUser(db *sql.DB, id int64) (*ABSUser, error) {
	var u ABSUser
	err := db.QueryRow(`SELECT id, label, base_url, api_token, created FROM abs_users WHERE id = ?`, id).
		Scan(&u.ID, &u.Label, &u.BaseURL, &u.APIToken, &u.Created)
	if err != nil {
		return nil, fmt.Errorf("store: getting abs user %d: %w", id, err)
	}
	return &u, nil
}

// CreateABSUser inserts a new Audiobookshelf user and returns it with its ID set.
func CreateABSUser(db *sql.DB, u ABSUser) (*ABSUser, error) {
	res, err := db.Exec(`INSERT INTO abs_users (label, base_url, api_token) VALUES (?, ?, ?)`,
		u.Label, u.BaseURL, u.APIToken)
	if err != nil {
		return nil, fmt.Errorf("store: creating abs user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("store: reading new abs user id: %w", err)
	}
	return GetABSUser(db, id)
}

// UpdateABSUser overwrites an existing Audiobookshelf user's fields.
func UpdateABSUser(db *sql.DB, u ABSUser) (*ABSUser, error) {
	_, err := db.Exec(`UPDATE abs_users SET label = ?, base_url = ?, api_token = ? WHERE id = ?`,
		u.Label, u.BaseURL, u.APIToken, u.ID)
	if err != nil {
		return nil, fmt.Errorf("store: updating abs user %d: %w", u.ID, err)
	}
	return GetABSUser(db, u.ID)
}

// DeleteABSUser removes an Audiobookshelf user (cascades to dependent profiles).
func DeleteABSUser(db *sql.DB, id int64) error {
	if _, err := db.Exec(`DELETE FROM abs_users WHERE id = ?`, id); err != nil {
		return fmt.Errorf("store: deleting abs user %d: %w", id, err)
	}
	return nil
}
