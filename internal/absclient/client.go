// Package absclient is a small, purpose-built REST client for Audiobookshelf
// (https://api.audiobookshelf.org). It only implements the handful of
// endpoints bookSync needs: listing libraries/items and reading/writing a
// user's media progress.
package absclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to one Audiobookshelf server as one user.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// New builds a Client for the given server URL and user API token.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Library is a minimal Audiobookshelf library.
type Library struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LibraryItem is the subset of an Audiobookshelf library item bookSync needs
// for matching and progress sync.
type LibraryItem struct {
	ID       string
	Title    string
	Authors  []string
	Duration float64 // seconds
}

// MediaProgress is one entry of a user's per-item progress.
type MediaProgress struct {
	LibraryItemID string  `json:"libraryItemId"`
	Duration      float64 `json:"duration"`
	Progress      float64 `json:"progress"` // 0..1
	CurrentTime   float64 `json:"currentTime"`
	IsFinished    bool    `json:"isFinished"`
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("absclient: encoding request body: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("absclient: building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("absclient: request to %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("absclient: %s %s returned %d: %s", method, path, resp.StatusCode, string(data))
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("absclient: decoding response from %s: %w", path, err)
	}
	return nil
}

// Libraries returns all libraries visible to this user.
func (c *Client) Libraries(ctx context.Context) ([]Library, error) {
	var out struct {
		Libraries []Library `json:"libraries"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/libraries", nil, &out); err != nil {
		return nil, err
	}
	return out.Libraries, nil
}

// LibraryItems returns every item in a library, with author/title metadata.
func (c *Client) LibraryItems(ctx context.Context, libraryID string) ([]LibraryItem, error) {
	var out struct {
		Results []struct {
			ID    string `json:"id"`
			Media struct {
				Duration float64 `json:"duration"`
				Metadata struct {
					Title       string   `json:"title"`
					AuthorName  string   `json:"authorName"`
					AuthorNames []string `json:"authorNames"`
				} `json:"metadata"`
			} `json:"media"`
		} `json:"results"`
	}

	path := fmt.Sprintf("/api/libraries/%s/items?limit=0", libraryID)
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}

	items := make([]LibraryItem, 0, len(out.Results))
	for _, r := range out.Results {
		authors := r.Media.Metadata.AuthorNames
		if len(authors) == 0 && r.Media.Metadata.AuthorName != "" {
			authors = []string{r.Media.Metadata.AuthorName}
		}
		items = append(items, LibraryItem{
			ID:       r.ID,
			Title:    r.Media.Metadata.Title,
			Authors:  authors,
			Duration: r.Media.Duration,
		})
	}
	return items, nil
}

// GetItem fetches one library item's metadata and duration.
func (c *Client) GetItem(ctx context.Context, itemID string) (*LibraryItem, error) {
	var out struct {
		ID    string `json:"id"`
		Media struct {
			Duration float64 `json:"duration"`
			Metadata struct {
				Title       string   `json:"title"`
				AuthorName  string   `json:"authorName"`
				AuthorNames []string `json:"authorNames"`
			} `json:"metadata"`
		} `json:"media"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/items/"+itemID, nil, &out); err != nil {
		return nil, err
	}

	authors := out.Media.Metadata.AuthorNames
	if len(authors) == 0 && out.Media.Metadata.AuthorName != "" {
		authors = []string{out.Media.Metadata.AuthorName}
	}
	return &LibraryItem{
		ID:       out.ID,
		Title:    out.Media.Metadata.Title,
		Authors:  authors,
		Duration: out.Media.Duration,
	}, nil
}

// MediaProgressByItem returns this user's current progress for every library
// item they've started, keyed by library item ID.
func (c *Client) MediaProgressByItem(ctx context.Context) (map[string]MediaProgress, error) {
	var out struct {
		MediaProgress []MediaProgress `json:"mediaProgress"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/me", nil, &out); err != nil {
		return nil, err
	}

	byItem := make(map[string]MediaProgress, len(out.MediaProgress))
	for _, p := range out.MediaProgress {
		byItem[p.LibraryItemID] = p
	}
	return byItem, nil
}

// SetProgress sets this user's completion percentage (0..1) for a library
// item. Audiobookshelf derives currentTime from progress*duration when
// duration is supplied.
func (c *Client) SetProgress(ctx context.Context, itemID string, duration float64, progress float64) error {
	body := map[string]any{
		"duration":    duration,
		"progress":    progress,
		"currentTime": duration * progress,
	}
	return c.do(ctx, http.MethodPatch, "/api/me/progress/"+itemID, body, nil)
}
