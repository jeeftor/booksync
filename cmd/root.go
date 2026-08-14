// Package cmd defines the cobra CLI for bookSync.
package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	dataDir = envOr("BOOKSYNC_DATA_DIR", "./data")
	version = "dev"
)

// SetVersion sets the application version (called from main.go via ldflags-injected build info if desired).
func SetVersion(v string) { version = v }

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func logLevel() slog.Level {
	switch strings.ToLower(os.Getenv("BOOKSYNC_LOG_LEVEL")) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel()}))
}

var rootCmd = &cobra.Command{
	Use:   "booksync",
	Short: "Sync reading progress between Kindle and Audiobookshelf",
	Long: `booksync keeps a Kindle account's reading progress and an Audiobookshelf
user's audiobook progress in sync, so you can switch between reading and
listening to the same book.

Configuration (Kindle accounts, Audiobookshelf users, profiles, book
mappings) is stored in SQLite at <data-dir>/booksync.db.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", dataDir, "directory for booksync.db")
}

// Execute runs the root cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
