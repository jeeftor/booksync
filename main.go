// Command booksync syncs reading progress between a Kindle account and an
// Audiobookshelf user, with a REST API, MCP endpoint, and web UI for
// managing accounts and confirming book matches.
package main

import "github.com/jeeftor/bookSync/cmd"

// Version, Commit, and Date are set via -ldflags "-X main.X=..." at build
// time (see Dockerfile and .github/workflows/release.yml).
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	cmd.SetBuildInfo(Version, Commit, Date)
	cmd.Execute()
}
