package store

import (
	"database/sql"
	"fmt"
)

// ListProfiles returns all sync profiles.
func ListProfiles(db *sql.DB) ([]Profile, error) {
	rows, err := db.Query(`SELECT id, label, kindle_account_id, abs_user_id, abs_library_id, poll_minutes, created
		FROM profiles ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("store: listing profiles: %w", err)
	}
	defer rows.Close()

	var out []Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.Label, &p.KindleAccountID, &p.ABSUserID, &p.ABSLibraryID, &p.PollMinutes, &p.Created); err != nil {
			return nil, fmt.Errorf("store: scanning profile: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetProfile fetches a single profile by ID.
func GetProfile(db *sql.DB, id int64) (*Profile, error) {
	var p Profile
	err := db.QueryRow(`SELECT id, label, kindle_account_id, abs_user_id, abs_library_id, poll_minutes, created
		FROM profiles WHERE id = ?`, id).
		Scan(&p.ID, &p.Label, &p.KindleAccountID, &p.ABSUserID, &p.ABSLibraryID, &p.PollMinutes, &p.Created)
	if err != nil {
		return nil, fmt.Errorf("store: getting profile %d: %w", id, err)
	}
	return &p, nil
}

// CreateProfile inserts a new profile and returns it with its ID set.
func CreateProfile(db *sql.DB, p Profile) (*Profile, error) {
	if p.PollMinutes <= 0 {
		p.PollMinutes = 15
	}
	res, err := db.Exec(`INSERT INTO profiles (label, kindle_account_id, abs_user_id, abs_library_id, poll_minutes)
		VALUES (?, ?, ?, ?, ?)`, p.Label, p.KindleAccountID, p.ABSUserID, p.ABSLibraryID, p.PollMinutes)
	if err != nil {
		return nil, fmt.Errorf("store: creating profile: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("store: reading new profile id: %w", err)
	}
	return GetProfile(db, id)
}

// UpdateProfile overwrites an existing profile's fields.
func UpdateProfile(db *sql.DB, p Profile) (*Profile, error) {
	_, err := db.Exec(`UPDATE profiles SET label = ?, kindle_account_id = ?, abs_user_id = ?,
		abs_library_id = ?, poll_minutes = ? WHERE id = ?`,
		p.Label, p.KindleAccountID, p.ABSUserID, p.ABSLibraryID, p.PollMinutes, p.ID)
	if err != nil {
		return nil, fmt.Errorf("store: updating profile %d: %w", p.ID, err)
	}
	return GetProfile(db, p.ID)
}

// DeleteProfile removes a profile (cascades to its book mappings).
func DeleteProfile(db *sql.DB, id int64) error {
	if _, err := db.Exec(`DELETE FROM profiles WHERE id = ?`, id); err != nil {
		return fmt.Errorf("store: deleting profile %d: %w", id, err)
	}
	return nil
}
