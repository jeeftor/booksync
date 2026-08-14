// Package webui embeds the built Svelte frontend (frontend/dist, copied here
// by `make frontend`) into the booksync binary.
package webui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// Assets returns the embedded frontend filesystem rooted at dist/, ready to
// be served directly (e.g. via http.FileServer).
func Assets() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
