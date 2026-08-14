package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/jeeftor/bookSync/internal/api"
	"github.com/jeeftor/bookSync/internal/service"
	"github.com/jeeftor/bookSync/internal/store"
	"github.com/jeeftor/bookSync/internal/webui"
)

var (
	port         string
	pollInterval time.Duration
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the booksync HTTP server (REST API, MCP endpoint, and web UI)",
	RunE:  runServe,
}

func init() {
	_ = godotenv.Load()
	serveCmd.Flags().StringVar(&port, "port", envOr("BOOKSYNC_PORT", "8686"), "HTTP port")
	serveCmd.Flags().DurationVar(&pollInterval, "poll-interval", 15*time.Minute,
		"how often to run the background sync pass across all confirmed mappings")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	log := newLogger()
	log.Info("starting booksync", "version", version, "commit", commit, "date", date)

	dbPath := filepath.Join(dataDir, "booksync.db")
	db, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	svc := service.New(db, log)

	frontendFS, err := webui.Assets()
	if err != nil {
		log.Warn("frontend assets unavailable, serving API only", "error", err)
		frontendFS = nil
	}

	e := api.New(svc, log, api.BuildInfo{Version: version, Commit: commit, Date: date}, frontendFS)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runPoller(ctx, svc, log, pollInterval)

	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Error("http server stopped", "error", err)
		}
	}()
	log.Info("listening", "port", port, "data_dir", dataDir)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	return e.Shutdown(shutdownCtx)
}

// runPoller periodically syncs every confirmed mapping in the background.
func runPoller(ctx context.Context, svc *service.Service, log *slog.Logger, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Info("running scheduled sync pass")
			svc.SyncAll(ctx)
		}
	}
}
