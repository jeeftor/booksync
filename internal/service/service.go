// Package service contains bookSync's business logic: managing accounts and
// profiles, proposing/confirming book matches, and running the
// percentage-based progress sync between Kindle and Audiobookshelf. Both the
// REST API (internal/api) and the MCP server (internal/mcpserver) call into
// this package so there is exactly one implementation of the sync logic.
package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jeeftor/bookSync/internal/absclient"
	"github.com/jeeftor/bookSync/internal/kindleclient"
	"github.com/jeeftor/bookSync/internal/matcher"
	"github.com/jeeftor/bookSync/internal/store"
)

// Service is the shared application core.
type Service struct {
	db  *sql.DB
	log *slog.Logger
}

// New builds a Service backed by db.
func New(db *sql.DB, log *slog.Logger) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{db: db, log: log}
}

// --- Kindle accounts ---------------------------------------------------

// secretMask replaces a non-empty secret with a fixed-width placeholder so
// list/read endpoints (REST and MCP) never echo back raw cookies/tokens.
// The real values are only ever read internally via store.Get*, never
// returned to callers.
const secretMask = "••••••••"

func maskSecret(v string) string {
	if v == "" {
		return ""
	}
	return secretMask
}

func (s *Service) ListKindleAccounts(ctx context.Context) ([]store.KindleAccount, error) {
	accs, err := store.ListKindleAccounts(s.db)
	if err != nil {
		return nil, err
	}
	for i := range accs {
		accs[i].UbidMain = maskSecret(accs[i].UbidMain)
		accs[i].AtMain = maskSecret(accs[i].AtMain)
		accs[i].SessionID = maskSecret(accs[i].SessionID)
		accs[i].XMain = maskSecret(accs[i].XMain)
		accs[i].DeviceToken = maskSecret(accs[i].DeviceToken)
		accs[i].TLSProxyKey = maskSecret(accs[i].TLSProxyKey)
	}
	return accs, nil
}

func (s *Service) CreateKindleAccount(ctx context.Context, a store.KindleAccount) (*store.KindleAccount, error) {
	return store.CreateKindleAccount(s.db, a)
}

// UpdateKindleAccount updates an existing account. Since ListKindleAccounts
// masks secret fields, a caller round-tripping list data back through this
// method would otherwise overwrite real cookies/tokens with the mask
// placeholder; any blank or masked-looking field is left unchanged instead,
// so the UI only needs to send the fields the user actually edited.
func (s *Service) UpdateKindleAccount(ctx context.Context, a store.KindleAccount) (*store.KindleAccount, error) {
	existing, err := store.GetKindleAccount(s.db, a.ID)
	if err != nil {
		return nil, err
	}
	a.UbidMain = keepIfBlank(a.UbidMain, existing.UbidMain)
	a.AtMain = keepIfBlank(a.AtMain, existing.AtMain)
	a.SessionID = keepIfBlank(a.SessionID, existing.SessionID)
	a.XMain = keepIfBlank(a.XMain, existing.XMain)
	a.DeviceToken = keepIfBlank(a.DeviceToken, existing.DeviceToken)
	a.TLSProxyKey = keepIfBlank(a.TLSProxyKey, existing.TLSProxyKey)
	return store.UpdateKindleAccount(s.db, a)
}

// keepIfBlank returns existing when incoming is empty or still the mask
// placeholder (i.e. the caller didn't intend to change it), otherwise incoming.
func keepIfBlank(incoming, existing string) string {
	if incoming == "" || incoming == secretMask {
		return existing
	}
	return incoming
}

func (s *Service) DeleteKindleAccount(ctx context.Context, id int64) error {
	return store.DeleteKindleAccount(s.db, id)
}

// TestKindleAccount verifies the stored credentials work by initializing a
// session and returns how many books were found in the library.
func (s *Service) TestKindleAccount(ctx context.Context, id int64) (int, error) {
	acc, err := store.GetKindleAccount(s.db, id)
	if err != nil {
		return 0, err
	}
	c, err := kindleclient.New(ctx, *acc)
	if err != nil {
		return 0, err
	}
	return len(c.Library()), nil
}

// --- Audiobookshelf users -----------------------------------------------

func (s *Service) ListABSUsers(ctx context.Context) ([]store.ABSUser, error) {
	users, err := store.ListABSUsers(s.db)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].APIToken = maskSecret(users[i].APIToken)
	}
	return users, nil
}

func (s *Service) CreateABSUser(ctx context.Context, u store.ABSUser) (*store.ABSUser, error) {
	return store.CreateABSUser(s.db, u)
}

// UpdateABSUser updates an existing user; see UpdateKindleAccount for why a
// blank/masked apiToken is treated as "leave unchanged".
func (s *Service) UpdateABSUser(ctx context.Context, u store.ABSUser) (*store.ABSUser, error) {
	existing, err := store.GetABSUser(s.db, u.ID)
	if err != nil {
		return nil, err
	}
	u.APIToken = keepIfBlank(u.APIToken, existing.APIToken)
	return store.UpdateABSUser(s.db, u)
}

func (s *Service) DeleteABSUser(ctx context.Context, id int64) error {
	return store.DeleteABSUser(s.db, id)
}

// TestABSUser verifies the stored credentials work and returns the visible libraries.
func (s *Service) TestABSUser(ctx context.Context, id int64) ([]absclient.Library, error) {
	u, err := store.GetABSUser(s.db, id)
	if err != nil {
		return nil, err
	}
	return absclient.New(u.BaseURL, u.APIToken).Libraries(ctx)
}

// --- Profiles -------------------------------------------------------------

func (s *Service) ListProfiles(ctx context.Context) ([]store.Profile, error) {
	return store.ListProfiles(s.db)
}

func (s *Service) CreateProfile(ctx context.Context, p store.Profile) (*store.Profile, error) {
	return store.CreateProfile(s.db, p)
}

func (s *Service) UpdateProfile(ctx context.Context, p store.Profile) (*store.Profile, error) {
	return store.UpdateProfile(s.db, p)
}

func (s *Service) DeleteProfile(ctx context.Context, id int64) error {
	return store.DeleteProfile(s.db, id)
}

// --- Matching ---------------------------------------------------------

// profileClients resolves a profile's Kindle and Audiobookshelf clients.
func (s *Service) profileClients(ctx context.Context, profileID int64) (*kindleclient.Client, *absclient.Client, *store.Profile, error) {
	profile, err := store.GetProfile(s.db, profileID)
	if err != nil {
		return nil, nil, nil, err
	}
	kAcc, err := store.GetKindleAccount(s.db, profile.KindleAccountID)
	if err != nil {
		return nil, nil, nil, err
	}
	aUser, err := store.GetABSUser(s.db, profile.ABSUserID)
	if err != nil {
		return nil, nil, nil, err
	}

	kc, err := kindleclient.New(ctx, *kAcc)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("service: connecting to kindle account %q: %w", kAcc.Label, err)
	}
	ac := absclient.New(aUser.BaseURL, aUser.APIToken)

	return kc, ac, profile, nil
}

func (s *Service) resolveABSLibraryID(ctx context.Context, ac *absclient.Client, profile *store.Profile) (string, error) {
	if profile.ABSLibraryID != "" {
		return profile.ABSLibraryID, nil
	}
	libs, err := ac.Libraries(ctx)
	if err != nil {
		return "", fmt.Errorf("service: listing abs libraries: %w", err)
	}
	if len(libs) == 0 {
		return "", fmt.Errorf("service: no audiobookshelf libraries visible to this user")
	}
	return libs[0].ID, nil
}

// Suggestions fetches both libraries for a profile and proposes fuzzy match
// candidates the caller can confirm via ConfirmMatch. Existing confirmed
// mappings are excluded.
func (s *Service) Suggestions(ctx context.Context, profileID int64) ([]matcher.Candidate, error) {
	kc, ac, profile, err := s.profileClients(ctx, profileID)
	if err != nil {
		return nil, err
	}

	libraryID, err := s.resolveABSLibraryID(ctx, ac, profile)
	if err != nil {
		return nil, err
	}
	items, err := ac.LibraryItems(ctx, libraryID)
	if err != nil {
		return nil, fmt.Errorf("service: listing abs library items: %w", err)
	}

	existing, err := store.ListMappings(s.db, profileID)
	if err != nil {
		return nil, err
	}
	mapped := make(map[string]bool, len(existing))
	for _, m := range existing {
		mapped[m.KindleASIN] = true
	}

	var unmatched []kindleclient.Book
	for _, b := range kc.Library() {
		if !mapped[b.ASIN] {
			unmatched = append(unmatched, b)
		}
	}

	rejected, err := store.RejectedPairs(s.db, profileID)
	if err != nil {
		return nil, err
	}
	isRejected := func(asin, absItemID string) bool { return rejected[asin+"|"+absItemID] }

	return matcher.Suggest(unmatched, items, isRejected), nil
}

// RejectMatch records that a suggested pairing was reviewed and declined, so
// it won't be proposed again for this profile; a different Audiobookshelf
// item can still be suggested for the same Kindle book.
func (s *Service) RejectMatch(ctx context.Context, profileID int64, c matcher.Candidate) error {
	return store.RejectMatch(s.db, profileID, c.KindleASIN, c.ABSItemID)
}

// ConfirmMatch persists a Kindle<->ABS pairing for a profile as confirmed,
// activating it for sync.
func (s *Service) ConfirmMatch(ctx context.Context, profileID int64, c matcher.Candidate) (*store.BookMapping, error) {
	m, err := store.CreateMapping(s.db, store.BookMapping{
		ProfileID:   profileID,
		KindleASIN:  c.KindleASIN,
		KindleTitle: c.KindleTitle,
		ABSItemID:   c.ABSItemID,
		ABSTitle:    c.ABSTitle,
		Confidence:  c.Confidence,
		Confirmed:   true,
	})
	if err != nil {
		return nil, err
	}
	return store.ConfirmMapping(s.db, m.ID)
}

func (s *Service) ListMappings(ctx context.Context, profileID int64) ([]store.BookMapping, error) {
	return store.ListMappings(s.db, profileID)
}

func (s *Service) DeleteMapping(ctx context.Context, id int64) error {
	return store.DeleteMapping(s.db, id)
}

func (s *Service) SyncHistory(ctx context.Context, mappingID int64, limit int) ([]store.SyncEvent, error) {
	return store.ListSyncEvents(s.db, mappingID, limit)
}

func (s *Service) RecentActivity(ctx context.Context, limit int) ([]store.SyncEvent, error) {
	return store.ListRecentEvents(s.db, limit)
}
