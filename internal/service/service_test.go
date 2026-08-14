package service

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jeeftor/bookSync/internal/store"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return New(db, nil)
}

func TestKeepIfBlank(t *testing.T) {
	cases := []struct {
		name, incoming, existing, want string
	}{
		{"blank keeps existing", "", "old-secret", "old-secret"},
		{"mask placeholder keeps existing", secretMask, "old-secret", "old-secret"},
		{"non-blank overrides", "new-secret", "old-secret", "new-secret"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := keepIfBlank(tc.incoming, tc.existing); got != tc.want {
				t.Errorf("keepIfBlank(%q, %q) = %q, want %q", tc.incoming, tc.existing, got, tc.want)
			}
		})
	}
}

func TestUpdateKindleAccountLeavesBlankSecretsUnchanged(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	created, err := svc.CreateKindleAccount(ctx, store.KindleAccount{
		Label:       "Test",
		UbidMain:    "ubid-1",
		AtMain:      "at-1",
		SessionID:   "sess-1",
		XMain:       "x-1",
		DeviceToken: "device-1",
		TLSProxyURL: "http://proxy:8080",
		TLSProxyKey: "key-1",
	})
	if err != nil {
		t.Fatalf("CreateKindleAccount: %v", err)
	}

	// Simulate the UI round-tripping a masked list response: only Label
	// changes, every secret field is blank.
	updated, err := svc.UpdateKindleAccount(ctx, store.KindleAccount{
		ID:          created.ID,
		Label:       "Renamed",
		TLSProxyURL: created.TLSProxyURL,
	})
	if err != nil {
		t.Fatalf("UpdateKindleAccount: %v", err)
	}

	if updated.Label != "Renamed" {
		t.Errorf("Label = %q, want %q", updated.Label, "Renamed")
	}
	if updated.UbidMain != "ubid-1" || updated.AtMain != "at-1" || updated.SessionID != "sess-1" ||
		updated.XMain != "x-1" || updated.DeviceToken != "device-1" || updated.TLSProxyKey != "key-1" {
		t.Errorf("secret fields should be unchanged, got %+v", updated)
	}

	// Now actually change one secret; the others should still be preserved.
	updated2, err := svc.UpdateKindleAccount(ctx, store.KindleAccount{
		ID:          created.ID,
		Label:       "Renamed",
		TLSProxyURL: created.TLSProxyURL,
		TLSProxyKey: "key-2",
	})
	if err != nil {
		t.Fatalf("UpdateKindleAccount (change key): %v", err)
	}
	if updated2.TLSProxyKey != "key-2" {
		t.Errorf("TLSProxyKey = %q, want %q", updated2.TLSProxyKey, "key-2")
	}
	if updated2.UbidMain != "ubid-1" {
		t.Errorf("UbidMain should still be unchanged, got %q", updated2.UbidMain)
	}
}

func TestUpdateABSUserLeavesBlankTokenUnchanged(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	created, err := svc.CreateABSUser(ctx, store.ABSUser{
		Label:    "Test",
		BaseURL:  "http://abs:80",
		APIToken: "token-1",
	})
	if err != nil {
		t.Fatalf("CreateABSUser: %v", err)
	}

	updated, err := svc.UpdateABSUser(ctx, store.ABSUser{
		ID:      created.ID,
		Label:   "Renamed",
		BaseURL: created.BaseURL,
	})
	if err != nil {
		t.Fatalf("UpdateABSUser: %v", err)
	}
	if updated.APIToken != "token-1" {
		t.Errorf("APIToken should be unchanged, got %q", updated.APIToken)
	}
	if updated.Label != "Renamed" {
		t.Errorf("Label = %q, want %q", updated.Label, "Renamed")
	}
}
