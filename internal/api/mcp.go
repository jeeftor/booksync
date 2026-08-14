package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jeeftor/bookSync/internal/matcher"
)

// buildMCPServer exposes bookSync's profiles/mappings/sync operations as MCP
// tools, backed by the same service.Service used by the REST handlers.
func buildMCPServer(h *Handlers, version string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "booksync", Version: version}, nil)

	type emptyInput struct{}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_profiles",
		Description: "List all Kindle<->Audiobookshelf sync profiles",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in emptyInput) (*mcp.CallToolResult, any, error) {
		profiles, err := h.svc.ListProfiles(ctx)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(profiles)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_kindle_accounts",
		Description: "List all configured Kindle accounts",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in emptyInput) (*mcp.CallToolResult, any, error) {
		accs, err := h.svc.ListKindleAccounts(ctx)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(accs)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_abs_users",
		Description: "List all configured Audiobookshelf users",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in emptyInput) (*mcp.CallToolResult, any, error) {
		users, err := h.svc.ListABSUsers(ctx)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(users)
	})

	type profileIDInput struct {
		ProfileID int64 `json:"profile_id" jsonschema:"the profile ID"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_book_mappings",
		Description: "Get all confirmed Kindle<->Audiobookshelf book mappings for a profile, with current progress on each side",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in profileIDInput) (*mcp.CallToolResult, any, error) {
		mappings, err := h.svc.ListMappings(ctx, in.ProfileID)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(mappings)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_match_suggestions",
		Description: "Fuzzy-match a profile's unmapped Kindle books against its Audiobookshelf library and return candidate pairings awaiting confirmation",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in profileIDInput) (*mcp.CallToolResult, any, error) {
		candidates, err := h.svc.Suggestions(ctx, in.ProfileID)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(candidates)
	})

	type confirmMatchInput struct {
		ProfileID   int64   `json:"profile_id" jsonschema:"the profile ID"`
		KindleASIN  string  `json:"kindle_asin" jsonschema:"the Kindle book's ASIN"`
		KindleTitle string  `json:"kindle_title" jsonschema:"the Kindle book's title"`
		ABSItemID   string  `json:"abs_item_id" jsonschema:"the Audiobookshelf library item ID"`
		ABSTitle    string  `json:"abs_title" jsonschema:"the Audiobookshelf item's title"`
		Confidence  float64 `json:"confidence,omitempty" jsonschema:"match confidence 0-1, if known"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "confirm_match",
		Description: "Confirm a Kindle<->Audiobookshelf book pairing, activating it for sync",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in confirmMatchInput) (*mcp.CallToolResult, any, error) {
		m, err := h.svc.ConfirmMatch(ctx, in.ProfileID, matcher.Candidate{
			KindleASIN:  in.KindleASIN,
			KindleTitle: in.KindleTitle,
			ABSItemID:   in.ABSItemID,
			ABSTitle:    in.ABSTitle,
			Confidence:  in.Confidence,
		})
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(m)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "reject_match",
		Description: "Reject a suggested Kindle<->Audiobookshelf pairing so it isn't proposed again for this profile; a different Audiobookshelf item may still be suggested for the same Kindle book",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in confirmMatchInput) (*mcp.CallToolResult, any, error) {
		if err := h.svc.RejectMatch(ctx, in.ProfileID, matcher.Candidate{
			KindleASIN: in.KindleASIN,
			ABSItemID:  in.ABSItemID,
		}); err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(map[string]bool{"rejected": true})
	})

	type mappingIDInput struct {
		MappingID int64 `json:"mapping_id" jsonschema:"the book mapping ID"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "sync_now",
		Description: "Immediately sync one book mapping's progress between Kindle and Audiobookshelf (furthest position wins)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in mappingIDInput) (*mcp.CallToolResult, any, error) {
		event, err := h.svc.SyncMapping(ctx, in.MappingID)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(event)
	})

	type syncHistoryInput struct {
		MappingID int64 `json:"mapping_id" jsonschema:"the book mapping ID"`
		Limit     int   `json:"limit,omitempty" jsonschema:"max number of events to return, default 50"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_sync_history",
		Description: "Get recent sync events (successes and errors) for one book mapping",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in syncHistoryInput) (*mcp.CallToolResult, any, error) {
		events, err := h.svc.SyncHistory(ctx, in.MappingID, in.Limit)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(events)
	})

	return s
}

func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errResult(err), nil, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil, nil
}

func errResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("%v", err)}}}
}
