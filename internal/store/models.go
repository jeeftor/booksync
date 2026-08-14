package store

import "time"

// KindleAccount holds the manually-extracted Amazon session cookies and
// device token needed to talk to Kindle's private API, plus the local
// tls-client-api proxy used to spoof Amazon's expected TLS fingerprint.
type KindleAccount struct {
	ID          int64     `json:"id"`
	Label       string    `json:"label"`
	UbidMain    string    `json:"ubidMain"`
	AtMain      string    `json:"atMain"`
	SessionID   string    `json:"sessionId"`
	XMain       string    `json:"xMain"`
	DeviceToken string    `json:"deviceToken"`
	TLSProxyURL string    `json:"tlsProxyUrl"`
	TLSProxyKey string    `json:"tlsProxyKey"`
	Created     time.Time `json:"created"`
}

// ABSUser holds credentials for one Audiobookshelf user.
type ABSUser struct {
	ID       int64     `json:"id"`
	Label    string    `json:"label"`
	BaseURL  string    `json:"baseUrl"`
	APIToken string    `json:"apiToken"`
	Created  time.Time `json:"created"`
}

// Profile pairs one Kindle account with one Audiobookshelf user and owns a
// set of confirmed book mappings between the two libraries.
type Profile struct {
	ID              int64     `json:"id"`
	Label           string    `json:"label"`
	KindleAccountID int64     `json:"kindleAccountId"`
	ABSUserID       int64     `json:"absUserId"`
	ABSLibraryID    string    `json:"absLibraryId"`
	PollMinutes     int       `json:"pollMinutes"`
	Created         time.Time `json:"created"`
}

// BookMapping links a Kindle ASIN to an Audiobookshelf library item within a
// profile, tracking last-known progress on each side.
type BookMapping struct {
	ID            int64      `json:"id"`
	ProfileID     int64      `json:"profileId"`
	KindleASIN    string     `json:"kindleAsin"`
	KindleTitle   string     `json:"kindleTitle"`
	ABSItemID     string     `json:"absItemId"`
	ABSTitle      string     `json:"absTitle"`
	Confidence    float64    `json:"confidence"`
	Confirmed     bool       `json:"confirmed"`
	LastKindlePct float64    `json:"lastKindlePct"`
	LastABSPct    float64    `json:"lastAbsPct"`
	LastSynced    *time.Time `json:"lastSynced,omitempty"`
	Created       time.Time  `json:"created"`
}

// SyncEvent is one row of sync history/activity log for a mapping.
type SyncEvent struct {
	ID        int64     `json:"id"`
	MappingID int64     `json:"mappingId"`
	Direction string    `json:"direction"` // kindle_to_abs | abs_to_kindle | noop | error
	FromPct   float64   `json:"fromPct"`
	ToPct     float64   `json:"toPct"`
	Message   string    `json:"message"`
	Created   time.Time `json:"created"`
}
