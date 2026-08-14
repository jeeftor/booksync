package store

import (
	"database/sql"
	"fmt"
)

// ListMappings returns every book mapping for a profile.
func ListMappings(db *sql.DB, profileID int64) ([]BookMapping, error) {
	rows, err := db.Query(`SELECT id, profile_id, kindle_asin, kindle_title, abs_item_id, abs_title,
		confidence, confirmed, last_kindle_pct, last_abs_pct, last_synced, created
		FROM book_mappings WHERE profile_id = ? ORDER BY id`, profileID)
	if err != nil {
		return nil, fmt.Errorf("store: listing mappings for profile %d: %w", profileID, err)
	}
	defer rows.Close()
	return scanMappings(rows)
}

// ListAllMappings returns every confirmed mapping across all profiles, used by the sync poller.
func ListAllMappings(db *sql.DB) ([]BookMapping, error) {
	rows, err := db.Query(`SELECT id, profile_id, kindle_asin, kindle_title, abs_item_id, abs_title,
		confidence, confirmed, last_kindle_pct, last_abs_pct, last_synced, created
		FROM book_mappings WHERE confirmed = 1 ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("store: listing all mappings: %w", err)
	}
	defer rows.Close()
	return scanMappings(rows)
}

func scanMappings(rows *sql.Rows) ([]BookMapping, error) {
	var out []BookMapping
	for rows.Next() {
		var m BookMapping
		var confirmed int
		if err := rows.Scan(&m.ID, &m.ProfileID, &m.KindleASIN, &m.KindleTitle, &m.ABSItemID, &m.ABSTitle,
			&m.Confidence, &confirmed, &m.LastKindlePct, &m.LastABSPct, &m.LastSynced, &m.Created); err != nil {
			return nil, fmt.Errorf("store: scanning mapping: %w", err)
		}
		m.Confirmed = confirmed != 0
		out = append(out, m)
	}
	return out, rows.Err()
}

// GetMapping fetches a single mapping by ID.
func GetMapping(db *sql.DB, id int64) (*BookMapping, error) {
	row := db.QueryRow(`SELECT id, profile_id, kindle_asin, kindle_title, abs_item_id, abs_title,
		confidence, confirmed, last_kindle_pct, last_abs_pct, last_synced, created
		FROM book_mappings WHERE id = ?`, id)
	var m BookMapping
	var confirmed int
	if err := row.Scan(&m.ID, &m.ProfileID, &m.KindleASIN, &m.KindleTitle, &m.ABSItemID, &m.ABSTitle,
		&m.Confidence, &confirmed, &m.LastKindlePct, &m.LastABSPct, &m.LastSynced, &m.Created); err != nil {
		return nil, fmt.Errorf("store: getting mapping %d: %w", id, err)
	}
	m.Confirmed = confirmed != 0
	return &m, nil
}

// CreateMapping inserts a new (unconfirmed by default) book mapping.
func CreateMapping(db *sql.DB, m BookMapping) (*BookMapping, error) {
	confirmed := 0
	if m.Confirmed {
		confirmed = 1
	}
	res, err := db.Exec(`INSERT INTO book_mappings
		(profile_id, kindle_asin, kindle_title, abs_item_id, abs_title, confidence, confirmed)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(profile_id, kindle_asin, abs_item_id) DO UPDATE SET
		  kindle_title = excluded.kindle_title, abs_title = excluded.abs_title,
		  confidence = excluded.confidence, confirmed = excluded.confirmed`,
		m.ProfileID, m.KindleASIN, m.KindleTitle, m.ABSItemID, m.ABSTitle, m.Confidence, confirmed)
	if err != nil {
		return nil, fmt.Errorf("store: creating mapping: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil || id == 0 {
		// ON CONFLICT UPDATE path: look the row back up by its unique key.
		row := db.QueryRow(`SELECT id FROM book_mappings WHERE profile_id = ? AND kindle_asin = ? AND abs_item_id = ?`,
			m.ProfileID, m.KindleASIN, m.ABSItemID)
		if scanErr := row.Scan(&id); scanErr != nil {
			return nil, fmt.Errorf("store: resolving upserted mapping id: %w", scanErr)
		}
	}
	return GetMapping(db, id)
}

// ConfirmMapping marks a mapping as confirmed, activating it for sync.
func ConfirmMapping(db *sql.DB, id int64) (*BookMapping, error) {
	if _, err := db.Exec(`UPDATE book_mappings SET confirmed = 1 WHERE id = ?`, id); err != nil {
		return nil, fmt.Errorf("store: confirming mapping %d: %w", id, err)
	}
	return GetMapping(db, id)
}

// UpdateMappingProgress records the last-known progress percentages after a sync pass.
func UpdateMappingProgress(db *sql.DB, id int64, kindlePct, absPct float64) error {
	_, err := db.Exec(`UPDATE book_mappings SET last_kindle_pct = ?, last_abs_pct = ?,
		last_synced = CURRENT_TIMESTAMP WHERE id = ?`, kindlePct, absPct, id)
	if err != nil {
		return fmt.Errorf("store: updating mapping progress %d: %w", id, err)
	}
	return nil
}

// DeleteMapping removes a mapping (and its sync history).
func DeleteMapping(db *sql.DB, id int64) error {
	if _, err := db.Exec(`DELETE FROM book_mappings WHERE id = ?`, id); err != nil {
		return fmt.Errorf("store: deleting mapping %d: %w", id, err)
	}
	return nil
}

// RejectMatch records that a proposed Kindle<->ABS pairing was reviewed and
// rejected, so it won't be suggested again for this profile.
func RejectMatch(db *sql.DB, profileID int64, kindleASIN, absItemID string) error {
	_, err := db.Exec(`INSERT INTO rejected_matches (profile_id, kindle_asin, abs_item_id)
		VALUES (?, ?, ?) ON CONFLICT(profile_id, kindle_asin, abs_item_id) DO NOTHING`,
		profileID, kindleASIN, absItemID)
	if err != nil {
		return fmt.Errorf("store: rejecting match: %w", err)
	}
	return nil
}

// RejectedPairs returns the set of previously-rejected "kindleASIN|absItemID"
// pairs for a profile, so suggestion logic can skip them.
func RejectedPairs(db *sql.DB, profileID int64) (map[string]bool, error) {
	rows, err := db.Query(`SELECT kindle_asin, abs_item_id FROM rejected_matches WHERE profile_id = ?`, profileID)
	if err != nil {
		return nil, fmt.Errorf("store: listing rejected matches for profile %d: %w", profileID, err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var asin, itemID string
		if err := rows.Scan(&asin, &itemID); err != nil {
			return nil, fmt.Errorf("store: scanning rejected match: %w", err)
		}
		out[asin+"|"+itemID] = true
	}
	return out, rows.Err()
}
