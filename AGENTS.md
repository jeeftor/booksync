# bookSync

Syncs reading progress between a Kindle account and an Audiobookshelf (ABS) user, so you can
switch between reading and listening to the same book. REST API, MCP endpoint, and a Svelte web UI.

## Why this exists / how it works

Amazon has no official API for Kindle reading progress. bookSync talks to the same private
endpoints `read.amazon.com`'s web reader uses (via `github.com/rodrigopero/kindle-api-go`),
authenticated with long-lived session cookies extracted manually from a browser, routed through a
local `tls-client-api` sidecar that spoofs Chrome's TLS fingerprint (Amazon blocks on fingerprint
otherwise). Audiobookshelf has a real REST API and needs no such workaround.

Kindle only exposes a coarse `percentageRead` per book (no fine-grained location we could align to
an audio timestamp), and Amazon's own Whispersync already collapses multiple devices on one
account into a single reading position. So the sync strategy is deliberately simple: **whichever
side has the higher percentage wins**, and that percentage is pushed to the other side. There is no
audio/text forced-alignment (unlike ebook-reader-to-audiobook tools such as bookbridge) — it isn't
needed here and would require downloading + transcribing audiobooks.

Kindle has no writable progress API, so sync is one-directional in practice (Kindle -> ABS); if ABS
is further ahead, that's recorded in sync history but can't be pushed back to Kindle.

## Data model

- **Kindle account**: one set of cookies (`ubid-main`, `at-main`, `session-id`, `x-main`) + device
  token + TLS proxy config. One Kindle account = one Amazon account's library/progress, regardless
  of how many physical devices are signed into it (Amazon already merges those).
- **Audiobookshelf user**: one server URL + user API token.
- **Profile**: pairs one Kindle account with one ABS user; owns its own confirmed book mappings.
  Use one profile per person (e.g. a shared family Kindle account paired with your ABS user, and a
  second profile for your spouse's separate Amazon account + their ABS user).
- **Book mapping**: a confirmed Kindle ASIN <-> ABS library item ID link within a profile, with
  last-known progress on each side.

Matching Kindle books to ABS items is fuzzy (title/author similarity, see `internal/matcher`)
because filenames/metadata rarely line up exactly — suggestions must be confirmed via the UI/API
before they become active.

## Build & Run

```bash
make help        # list all targets
make build       # build frontend + Go binary (embeds frontend/dist into internal/webui/dist)
make run         # run the server locally (go run . serve) — API only unless frontend was built
make frontend    # build the Svelte frontend only (bun) and copy it into internal/webui/dist
make test        # go test ./...
make docker      # build the Docker image
make clean       # remove build artifacts
```

**Prerequisite**: `make build` requires the frontend to be built first, since `internal/webui`
embeds `internal/webui/dist/` via `embed.FS`. The `build` target depends on `frontend`
automatically. Without a real build, that directory contains a placeholder `index.html` explaining
how to build it, so `go build ./...` still compiles on a fresh checkout.

## Versioning & Releases

Version/commit/build-date are injected via `-ldflags -X main.Version=... -X main.Commit=...
-X main.Date=...` (see `Makefile` and `Dockerfile`), surfaced at `GET /api/health`, and shown as a
badge in the web UI header (hover for commit + build date). `internal/api.BuildInfo` is the shared
type; thread any new consumer through `cmd/serve.go`'s `api.New(...)` call rather than adding a
separate ad-hoc version string.

**Docker images are only built/published for a `v*` git tag push**, never a plain `master` push
(matches `ics-mcp`'s convention) — `.github/workflows/release.yml`'s `test` job runs on every push
for fast feedback, but the `meta`/`docker-amd64`/`docker-arm64`/`manifest` jobs are gated on
`startsWith(github.ref, 'refs/tags/v')`. This means "push a version tag" is *always* the trigger for
a new pullable image (`ghcr.io/jeeftor/booksync:vX.Y.Z` + `:latest` + `:<sha>`, multi-arch
amd64+arm64) — there's no ambiguity about whether `latest` actually contains a given change.

**Practical rule for whoever/whatever is making changes to this repo: after any change meant to be
deployed, bump and push a version tag** (`git tag -a vX.Y.Z -m "..." && git push origin vX.Y.Z`)
before telling the user it's ready to pull. Don't leave a change sitting on `master` unreleased if
the intent was to get it running.

Follow SemVer deliberately, matching `ics-mcp`'s convention: patch releases for scoped bug fixes,
minor releases for a coherent group of user-facing additions (don't tag a string of patches for a
stream of unrelated features — batch them into the next minor). In practice for this project that
usually means: patch bump per work session/fix, minor bump when a session adds a distinct new
capability.

## Configuration

Env vars (`BOOKSYNC_*`) + CLI flags; see `.env.example`. Kindle accounts, ABS users, profiles, and
book mappings are all configured through the web UI or REST API and stored in SQLite at
`<data-dir>/booksync.db` — no YAML config files.

| Variable              | Description                          | Default   |
| ---------------------- | ------------------------------------ | --------- |
| `BOOKSYNC_DATA_DIR`    | Data directory (SQLite DB)            | `./data`  |
| `BOOKSYNC_PORT`        | HTTP server port                      | `8686`    |
| `BOOKSYNC_LOG_LEVEL`   | debug, info, warn, error              | `info`    |

### Getting Kindle cookies/device token

1. Log into <https://read.amazon.com> in a browser.
2. Open DevTools -> Network tab, reload.
3. Copy the `Cookie` header from any request; you need `ubid-main`, `at-main`, `session-id`,
   `x-main`.
4. Find the request to `getDeviceToken?serialNumber=...&deviceType=...` — both params are your
   device token.
5. These are entered once per Kindle account in the web UI (Kindle Accounts tab) and are valid for
   about a year.

### TLS proxy sidecar

Required regardless of how you run bookSync — see `docker-compose.yml` and
`tls-client-config.yml.example`. Copy the example to `tls-client-config.yml`, set your own
`api_auth_keys`, and point a Kindle account's "TLS proxy URL"/"TLS proxy API key" at it.

## Architecture

- `cmd/` — Cobra CLI (`serve`)
- `internal/api/` — Echo HTTP server, REST handlers, MCP endpoint (`internal/api/mcp.go`)
- `internal/service/` — business logic shared by REST and MCP: accounts/profiles CRUD, matching,
  and the percentage-based sync (`internal/service/sync.go`)
- `internal/kindleclient/` — thin wrapper over `github.com/rodrigopero/kindle-api-go`
- `internal/absclient/` — hand-rolled Audiobookshelf REST client (libraries, items, media progress)
- `internal/matcher/` — fuzzy title/author matching (Levenshtein-based) between the two libraries
- `internal/store/` — SQLite schema + CRUD (`database/sql` + `modernc.org/sqlite`, pure Go)
- `internal/webui/` — embeds the built Svelte frontend (`embed.FS`)
- `frontend/` — Svelte 5 + Vite + Bun SPA

### REST API

- `GET/POST/PUT/DELETE /api/kindle-accounts[/:id]`, `POST /api/kindle-accounts/:id/test`
- `GET/POST/PUT/DELETE /api/abs-users[/:id]`, `POST /api/abs-users/:id/test`
- `GET/POST/PUT/DELETE /api/profiles[/:id]`
- `GET /api/profiles/:id/suggestions` — fuzzy match candidates
- `GET/POST /api/profiles/:id/mappings` — list / confirm a match
- `POST /api/profiles/:id/sync` — sync every mapping in a profile
- `DELETE /api/mappings/:id`, `POST /api/mappings/:id/sync`, `GET /api/mappings/:id/history`
- `GET /api/activity` — recent sync events across all profiles
- `ANY /mcp` — MCP endpoint

### MCP tools

`list_profiles`, `list_kindle_accounts`, `list_abs_users`, `get_book_mappings`,
`get_match_suggestions`, `confirm_match`, `sync_now`, `get_sync_history` — all thin wrappers over
`internal/service`, so REST and MCP never diverge in behavior.

## Tech stack

- Go, Cobra, Echo, `modernc.org/sqlite` (pure Go, no CGO), `github.com/modelcontextprotocol/go-sdk`
- Svelte 5 + Vite + Bun + Tailwind CSS v4

## Known limitations / follow-ups

- No auth on the web UI/API — intended for LAN/VPN-only exposure. Add auth before exposing publicly.
- Kindle account/ABS user secrets are stored in plaintext in SQLite and returned as-is by the list
  endpoints; fine for a single-operator self-hosted tool, but worth masking/encrypting before wider use.
- Background poller runs a single global interval (`--poll-interval`, default 15m) rather than
  honoring each profile's stored `pollMinutes` individually.
- No highlights/notes sync (out of scope per current requirements — progress only).
