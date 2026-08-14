// Package api provides the REST HTTP server and MCP endpoint for bookSync.
package api

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jeeftor/bookSync/internal/service"
)

// Handlers holds shared dependencies for REST and MCP handlers.
type Handlers struct {
	svc            *service.Service
	log            *slog.Logger
	kindleDefaults service.KindleAccountDefaults
}

// BuildInfo describes the running binary's build, injected via -ldflags at
// build time (see Dockerfile and .github/workflows/release.yml). Reported by
// /api/health and shown in the web UI so a running instance is identifiable.
type BuildInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

// New builds the Echo server: REST API under /api, the MCP endpoint at /mcp,
// and (if frontendFS is non-nil) the built Svelte SPA served at /.
func New(svc *service.Service, log *slog.Logger, build BuildInfo, kindleDefaults service.KindleAccountDefaults, frontendFS fs.FS) *echo.Echo {
	if log == nil {
		log = slog.Default()
	}
	h := &Handlers{svc: svc, log: log, kindleDefaults: kindleDefaults}

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	api := e.Group("/api")
	api.GET("/health", h.Health(build))

	api.GET("/kindle-accounts", h.ListKindleAccounts)
	api.GET("/kindle-accounts/defaults", h.KindleAccountDefaults)
	api.POST("/kindle-accounts", h.CreateKindleAccount)
	api.PUT("/kindle-accounts/:id", h.UpdateKindleAccount)
	api.DELETE("/kindle-accounts/:id", h.DeleteKindleAccount)
	api.POST("/kindle-accounts/:id/test", h.TestKindleAccount)
	api.POST("/kindle-accounts/test", h.TestKindleAccountDraft)

	api.GET("/abs-users", h.ListABSUsers)
	api.POST("/abs-users", h.CreateABSUser)
	api.PUT("/abs-users/:id", h.UpdateABSUser)
	api.DELETE("/abs-users/:id", h.DeleteABSUser)
	api.POST("/abs-users/:id/test", h.TestABSUser)

	api.GET("/profiles", h.ListProfiles)
	api.POST("/profiles", h.CreateProfile)
	api.PUT("/profiles/:id", h.UpdateProfile)
	api.DELETE("/profiles/:id", h.DeleteProfile)
	api.GET("/profiles/:id/suggestions", h.Suggestions)
	api.GET("/profiles/:id/mappings", h.ListMappings)
	api.POST("/profiles/:id/mappings", h.ConfirmMatch)
	api.POST("/profiles/:id/suggestions/reject", h.RejectMatch)
	api.POST("/profiles/:id/sync", h.SyncProfile)

	api.DELETE("/mappings/:id", h.DeleteMapping)
	api.POST("/mappings/:id/sync", h.SyncMapping)
	api.GET("/mappings/:id/history", h.MappingHistory)

	api.GET("/activity", h.RecentActivity)

	mcpServer := buildMCPServer(h, build.Version)
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return mcpServer }, nil)
	e.Any("/mcp", echo.WrapHandler(mcpHandler))

	if frontendFS != nil {
		fileServer := http.FileServer(http.FS(frontendFS))
		e.GET("/*", echo.WrapHandler(spaFallback(frontendFS, fileServer)))
	}

	return e
}

// spaFallback serves static files when they exist and falls back to
// index.html otherwise, so client-side routing works on refresh/deep links.
func spaFallback(files fs.FS, fileServer http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		if _, err := fs.Stat(files, path[1:]); err != nil {
			r.URL.Path = "/index.html"
		}
		fileServer.ServeHTTP(w, r)
	})
}
