package store

import (
	"database/sql"
	"fmt"
)

// ListKindleAccounts returns all configured Kindle accounts.
func ListKindleAccounts(db *sql.DB) ([]KindleAccount, error) {
	rows, err := db.Query(`SELECT id, label, ubid_main, at_main, session_id, x_main,
		device_token, tls_proxy_url, tls_proxy_key, created FROM kindle_accounts ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("store: listing kindle accounts: %w", err)
	}
	defer rows.Close()

	var out []KindleAccount
	for rows.Next() {
		var a KindleAccount
		if err := rows.Scan(&a.ID, &a.Label, &a.UbidMain, &a.AtMain, &a.SessionID, &a.XMain,
			&a.DeviceToken, &a.TLSProxyURL, &a.TLSProxyKey, &a.Created); err != nil {
			return nil, fmt.Errorf("store: scanning kindle account: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// GetKindleAccount fetches a single Kindle account by ID.
func GetKindleAccount(db *sql.DB, id int64) (*KindleAccount, error) {
	var a KindleAccount
	err := db.QueryRow(`SELECT id, label, ubid_main, at_main, session_id, x_main,
		device_token, tls_proxy_url, tls_proxy_key, created FROM kindle_accounts WHERE id = ?`, id).
		Scan(&a.ID, &a.Label, &a.UbidMain, &a.AtMain, &a.SessionID, &a.XMain,
			&a.DeviceToken, &a.TLSProxyURL, &a.TLSProxyKey, &a.Created)
	if err != nil {
		return nil, fmt.Errorf("store: getting kindle account %d: %w", id, err)
	}
	return &a, nil
}

// CreateKindleAccount inserts a new Kindle account and returns it with its ID set.
func CreateKindleAccount(db *sql.DB, a KindleAccount) (*KindleAccount, error) {
	res, err := db.Exec(`INSERT INTO kindle_accounts
		(label, ubid_main, at_main, session_id, x_main, device_token, tls_proxy_url, tls_proxy_key)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.Label, a.UbidMain, a.AtMain, a.SessionID, a.XMain, a.DeviceToken, a.TLSProxyURL, a.TLSProxyKey)
	if err != nil {
		return nil, fmt.Errorf("store: creating kindle account: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("store: reading new kindle account id: %w", err)
	}
	return GetKindleAccount(db, id)
}

// UpdateKindleAccount overwrites an existing Kindle account's fields.
func UpdateKindleAccount(db *sql.DB, a KindleAccount) (*KindleAccount, error) {
	_, err := db.Exec(`UPDATE kindle_accounts SET label = ?, ubid_main = ?, at_main = ?,
		session_id = ?, x_main = ?, device_token = ?, tls_proxy_url = ?, tls_proxy_key = ?
		WHERE id = ?`,
		a.Label, a.UbidMain, a.AtMain, a.SessionID, a.XMain, a.DeviceToken,
		a.TLSProxyURL, a.TLSProxyKey, a.ID)
	if err != nil {
		return nil, fmt.Errorf("store: updating kindle account %d: %w", a.ID, err)
	}
	return GetKindleAccount(db, a.ID)
}

// DeleteKindleAccount removes a Kindle account (cascades to dependent profiles).
func DeleteKindleAccount(db *sql.DB, id int64) error {
	if _, err := db.Exec(`DELETE FROM kindle_accounts WHERE id = ?`, id); err != nil {
		return fmt.Errorf("store: deleting kindle account %d: %w", id, err)
	}
	return nil
}
