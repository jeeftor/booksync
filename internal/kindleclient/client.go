// Package kindleclient wraps github.com/rodrigopero/kindle-api-go with the
// small surface bookSync actually needs: list the library and fetch each
// book's current reading percentage. There is no official Kindle API; this
// talks to the same private endpoints read.amazon.com's web reader uses,
// authenticated with long-lived session cookies extracted manually from a
// browser, and routed through a local tls-client-api proxy to match Amazon's
// expected TLS fingerprint.
package kindleclient

import (
	"context"
	"fmt"

	kindle "github.com/rodrigopero/kindle-api-go"

	"github.com/jeeftor/bookSync/internal/store"
)

// Book is the subset of Kindle library/progress data bookSync cares about.
type Book struct {
	ASIN           string
	Title          string
	Authors        []string
	PercentageRead float64
}

// Client fetches library and progress data for one Kindle account.
type Client struct {
	inner *kindle.Client
}

// New builds a Client from a stored Kindle account and loads its library.
// It must succeed once before GetBook can be called.
func New(ctx context.Context, acc store.KindleAccount) (*Client, error) {
	cookies := kindle.Cookies{
		UbidMain:  acc.UbidMain,
		AtMain:    acc.AtMain,
		SessionID: acc.SessionID,
		XMain:     acc.XMain,
	}

	c, err := kindle.NewClient(cookies, acc.DeviceToken, acc.TLSProxyURL, acc.TLSProxyKey)
	if err != nil {
		return nil, fmt.Errorf("kindleclient: building client for %q: %w", acc.Label, err)
	}

	if err := c.Init(ctx); err != nil {
		return nil, fmt.Errorf("kindleclient: initializing session for %q: %w", acc.Label, err)
	}

	return &Client{inner: c}, nil
}

// Library returns the basic metadata (title/author/ASIN) for every book in
// the account's library. It does not include reading progress; call GetBook
// per ASIN for that.
func (c *Client) Library() []Book {
	books := make([]Book, 0, len(c.inner.Books))
	for _, b := range c.inner.Books {
		books = append(books, Book{ASIN: b.ASIN, Title: b.Title, Authors: b.Authors})
	}
	return books
}

// GetBook fetches full details for one ASIN, including PercentageRead.
func (c *Client) GetBook(ctx context.Context, asin string) (*Book, error) {
	details, err := c.inner.GetBookDetails(ctx, asin)
	if err != nil {
		return nil, fmt.Errorf("kindleclient: fetching details for %s: %w", asin, err)
	}
	return &Book{
		ASIN:           details.ASIN,
		Title:          details.Title,
		Authors:        details.Authors,
		PercentageRead: details.PercentageRead,
	}, nil
}
