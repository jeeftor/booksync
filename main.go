// Command booksync syncs reading progress between a Kindle account and an
// Audiobookshelf user, with a REST API, MCP endpoint, and web UI for
// managing accounts and confirming book matches.
package main

import "github.com/jeeftor/bookSync/cmd"

// Version is set via -ldflags "-X main.Version=..." at build time.
var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	cmd.Execute()
}
