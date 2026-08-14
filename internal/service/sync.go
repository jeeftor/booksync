package service

import (
	"context"
	"fmt"
	"math"

	"github.com/jeeftor/bookSync/internal/store"
)

// progressEpsilon is the minimum percentage-point difference worth writing
// back to a service. Amazon/ABS progress values jitter by fractions of a
// percent just from rounding, so this avoids sync "flapping".
const progressEpsilon = 0.5

// SyncMapping compares Kindle and Audiobookshelf progress for one mapping and
// pushes whichever side is further along to the other. Amazon's own
// Whispersync already collapses multiple devices on one account into a
// single reading position, so bookSync mirrors that same "furthest position
// wins" rule across services rather than attempting fine-grained location
// alignment (which Kindle's API doesn't expose anyway).
func (s *Service) SyncMapping(ctx context.Context, mappingID int64) (*store.SyncEvent, error) {
	m, err := store.GetMapping(s.db, mappingID)
	if err != nil {
		return nil, err
	}
	if !m.Confirmed {
		return nil, fmt.Errorf("service: mapping %d is not confirmed", mappingID)
	}

	kc, ac, _, err := s.profileClients(ctx, m.ProfileID)
	if err != nil {
		return s.recordFailure(m.ID, err)
	}

	kBook, err := kc.GetBook(ctx, m.KindleASIN)
	if err != nil {
		return s.recordFailure(m.ID, fmt.Errorf("fetching kindle progress: %w", err))
	}

	item, err := ac.GetItem(ctx, m.ABSItemID)
	if err != nil {
		return s.recordFailure(m.ID, fmt.Errorf("fetching abs item: %w", err))
	}
	progress, err := ac.MediaProgressByItem(ctx)
	if err != nil {
		return s.recordFailure(m.ID, fmt.Errorf("fetching abs progress: %w", err))
	}
	absPct := 0.0
	if p, ok := progress[m.ABSItemID]; ok {
		absPct = p.Progress * 100
	}
	kindlePct := kBook.PercentageRead

	decision := decideSyncDirection(kindlePct, absPct)
	event := store.SyncEvent{
		MappingID: m.ID,
		FromPct:   math.Max(kindlePct, absPct),
		Direction: decision.Direction,
		ToPct:     decision.ToPct,
		Message:   decision.Message,
	}

	if decision.Direction == "kindle_to_abs" {
		if err := ac.SetProgress(ctx, m.ABSItemID, item.Duration, kindlePct/100); err != nil {
			return s.recordFailure(m.ID, fmt.Errorf("updating abs progress: %w", err))
		}
		absPct = kindlePct
	}

	if err := store.UpdateMappingProgress(s.db, m.ID, kindlePct, absPct); err != nil {
		return nil, err
	}
	if err := store.RecordSyncEvent(s.db, event); err != nil {
		return nil, err
	}
	return &event, nil
}

// syncDecision is the outcome of comparing Kindle and Audiobookshelf
// progress for a mapping, before any network write happens.
type syncDecision struct {
	Direction string // kindle_to_abs | abs_ahead_no_kindle_write | noop
	ToPct     float64
	Message   string
}

// decideSyncDirection implements the "furthest position wins" rule in
// isolation from any network I/O, so it can be unit tested directly. Kindle
// has no writable progress API, so when Audiobookshelf is ahead we can only
// record the gap, not push it back.
func decideSyncDirection(kindlePct, absPct float64) syncDecision {
	switch {
	case kindlePct-absPct > progressEpsilon:
		return syncDecision{Direction: "kindle_to_abs", ToPct: kindlePct}
	case absPct-kindlePct > progressEpsilon:
		return syncDecision{
			Direction: "abs_ahead_no_kindle_write",
			ToPct:     absPct,
			Message:   "Audiobookshelf is ahead of Kindle; Kindle has no writable progress API",
		}
	default:
		return syncDecision{Direction: "noop", ToPct: math.Max(kindlePct, absPct)}
	}
}

func (s *Service) recordFailure(mappingID int64, cause error) (*store.SyncEvent, error) {
	event := store.SyncEvent{MappingID: mappingID, Direction: "error", Message: cause.Error()}
	if err := store.RecordSyncEvent(s.db, event); err != nil {
		s.log.Error("recording sync failure", "mapping_id", mappingID, "error", err)
	}
	return nil, cause
}

// SyncProfile syncs every confirmed mapping belonging to one profile.
func (s *Service) SyncProfile(ctx context.Context, profileID int64) error {
	mappings, err := store.ListMappings(s.db, profileID)
	if err != nil {
		return err
	}
	var firstErr error
	for _, m := range mappings {
		if !m.Confirmed {
			continue
		}
		if _, err := s.SyncMapping(ctx, m.ID); err != nil {
			s.log.Warn("sync failed", "mapping_id", m.ID, "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// SyncAll syncs every confirmed mapping across all profiles. Used by the
// background poller.
func (s *Service) SyncAll(ctx context.Context) {
	mappings, err := store.ListAllMappings(s.db)
	if err != nil {
		s.log.Error("listing mappings for sync", "error", err)
		return
	}
	for _, m := range mappings {
		if _, err := s.SyncMapping(ctx, m.ID); err != nil {
			s.log.Warn("sync failed", "mapping_id", m.ID, "error", err)
		}
	}
}
