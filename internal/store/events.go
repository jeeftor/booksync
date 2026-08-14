package store

import (
	"database/sql"
	"fmt"
)

// RecordSyncEvent appends an entry to a mapping's sync history/activity log.
func RecordSyncEvent(db *sql.DB, e SyncEvent) error {
	_, err := db.Exec(`INSERT INTO sync_events (mapping_id, direction, from_pct, to_pct, message)
		VALUES (?, ?, ?, ?, ?)`, e.MappingID, e.Direction, e.FromPct, e.ToPct, e.Message)
	if err != nil {
		return fmt.Errorf("store: recording sync event: %w", err)
	}
	return nil
}

// ListSyncEvents returns the most recent sync events for a mapping, newest first.
func ListSyncEvents(db *sql.DB, mappingID int64, limit int) ([]SyncEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.Query(`SELECT id, mapping_id, direction, from_pct, to_pct, message, created
		FROM sync_events WHERE mapping_id = ? ORDER BY id DESC LIMIT ?`, mappingID, limit)
	if err != nil {
		return nil, fmt.Errorf("store: listing sync events for mapping %d: %w", mappingID, err)
	}
	defer rows.Close()

	var out []SyncEvent
	for rows.Next() {
		var e SyncEvent
		if err := rows.Scan(&e.ID, &e.MappingID, &e.Direction, &e.FromPct, &e.ToPct, &e.Message, &e.Created); err != nil {
			return nil, fmt.Errorf("store: scanning sync event: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListRecentEvents returns the most recent sync events across all mappings, newest first.
func ListRecentEvents(db *sql.DB, limit int) ([]SyncEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := db.Query(`SELECT id, mapping_id, direction, from_pct, to_pct, message, created
		FROM sync_events ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("store: listing recent sync events: %w", err)
	}
	defer rows.Close()

	var out []SyncEvent
	for rows.Next() {
		var e SyncEvent
		if err := rows.Scan(&e.ID, &e.MappingID, &e.Direction, &e.FromPct, &e.ToPct, &e.Message, &e.Created); err != nil {
			return nil, fmt.Errorf("store: scanning sync event: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
