# dashboard

A personal dashboard web app written in Go. It uses:

- **chi** for routing
- **templ** for server-rendered HTML components
- **HTMX** for progressively loading widgets
- **Tailwind CSS** for styling (built via the standalone CLI)
- A small **widget framework** (`internal/widgetkit`) for caching, stale fallback, and consistent rendering

## Current Widgets

- **Weather** (`/widgets/weather`)
- **Hacker News Top 10** (`/widgets/hn`)

Widgets are loaded into the dashboard page via HTMX and rendered as HTML fragments by the server.

## Project Structure

```
cmd/dashboard/ # app entrypoint
internal/app/ # composition root (wire-up)
internal/server/ # router + middleware + routes
internal/ui/ # templ layouts/pages/components
internal/widgetkit/ # widget framework (handler, registry)
internal/widgets/ # widget implementations (weather, hn, ...)
internal/httpx/ # shared HTTP client helpers
internal/cache/ # generic TTL cache
web/ # source assets (Tailwind input)
static/ # served assets (Tailwind output, vendor js)
```


## Development

### Prereqs

This repo is intended to be used with the included **VS Code Dev Container** configuration. The devcontainer installs:

- Go
- `templ`
- `air`
- Tailwind standalone CLI
- sqlite3 CLI (for future DB work)

If you’re not using the devcontainer, you’ll need the above tools installed locally.

### Hot reload (recommended)

Start the full dev loop (templ proxy + air + tailwind watcher + static watcher):

```bash
make live
```

Then open:

Dashboard via templ proxy: http://localhost:7331

App server (direct): http://localhost:8080

### One-off build & run

```
make build
./tmp/app serve

```

### How Widgets Work

A widget has four pieces:

1. Client / API fetch (public API calls)
2. ViewModel (display-friendly data for templates)
3. templ template (HTML fragment)
4. Handler (wired through `internal/widgetkit`)

### Widget framework features

`internal/widgetkit` provides:

- TTL caching
- stale-if-error behavior (serve previous data when refresh fails)
- A standard handler shape: `Fetch`, `Render`, `Error`, and optional `MarkStale`

### Adding a new widget (recipe)

1. Create `internal/widgets/<name>/`
2. Add:
    - `client.go`
    - `models.go`
    - `viewmodel.go`
    - `template.templ`
    - `handler.go` (returns an http.Handler)
3. Register it in `internal/app/app.go` using the `widgetkit` registry

### Static Assets

- Tailwind source: `web/css/input.css`
- Built CSS output: `static/css/output.css`
- HTMX is self-hosted: `static/js/htmx.min.js`

### Notes / Next Ideas

- Add more widgets (GitHub activity, calendar, air quality, etc.)
- Add golden tests for each widget using injected fakes (no outbound network in tests)
- Add `singleflight` in `widgetkit` to avoid duplicate refresh work under load
- Consider embedding `static/` into the binary for production deployments

## TODO

- [ ] embed `static/` into the binary for production deployment