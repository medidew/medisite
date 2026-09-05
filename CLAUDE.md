# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`medisite` is a small Go web server for a personal blog/portfolio/link-hub site (`medidew.name`), deployed behind an existing nginx reverse proxy. It renders server-side HTML via `html/template` — no frontend framework, no JS build step, no client-side JS beyond what's in the static CSS.

## Commands

```bash
go build ./...              # build everything
go run .                    # run with defaults (config.yaml, or built-in defaults if absent)
go run . -config path.yaml  # run with an explicit config file
go vet ./...                # static checks
gofmt -l -w .                # format (run before considering any change done)
```

There is no test suite yet. There's also no Makefile/task runner — the commands above are the whole workflow.

To manually verify a change, run the server and hit the routes:
```bash
go build -o /tmp/medisite-build . && /tmp/medisite-build &
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8080/
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8080/blog
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8080/blog/example-post
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8080/portfolio
```
Kill the background process afterward and remove any `/tmp` binary and `medisite.log` produced by the run — don't leave test artifacts in the repo.

## Configuration

All runtime settings live in a YAML config file (`config.yaml` by default, overridable with `-config`), loaded by `internal/config`. The file is optional — `config.Default()` supplies every field, and `config.Load` merges the file on top of those defaults, so an absent file is not an error but a malformed one is a fail-fast startup error. Fields: `addr`, `content_dir`, `templates_dir`, `static_dir`, `log_file`.

## Architecture

**Startup sequence (`main.go`):** parse `-config` flag → `config.Load` → build the zap logger from `cfg.LogFile` (`internal/logging`) → `content.Load` (reads and parses everything under `content_dir`) → `web.New` (parses templates, builds routes) → `http.ListenAndServe`. Every fatal error after the logger exists goes through `logger.Fatal`; the two errors that can happen *before* the logger exists (bad `-config` path, unwritable log file) fall back to stdlib `log.Fatalf` since there's nothing to log to yet.

**Content model (`internal/content`)** is the site's data layer, entirely decoupled from HTTP:
- Posts are Markdown files in `content/posts/*.md` with a `---`-delimited YAML frontmatter block (`title`, `date` as `2006-01-02`, `tags`, `summary`, optional `slug` override — see `posts.go`'s `frontmatter` struct). The Markdown body is rendered to `template.HTML` via `goldmark` at load time, not per-request.
- Portfolio projects come from `content/portfolio.yaml` (a `projects:` list matching the `Project` struct).
- `content.Load` composes both into a single immutable `*Site` (sorted posts, slug→post map, project list) built once at startup and shared read-only across request goroutines — no locking anywhere in the request path.
- Load errors are split deliberately: a single malformed post is logged (`logger.Warn`) and skipped so one bad file doesn't take the site down; a missing `content/` or `posts/` dir, an unparseable `portfolio.yaml`, or two posts resolving to the same slug are fail-fast errors from `content.Load`.
- **Content and templates are read from disk at process startup, not embedded (`go:embed`).** Publishing a new post or editing a template/CSS file means restarting the process, not rebuilding it — this is the deliberate deploy model here, don't "fix" it by switching to embedding without being asked.

**HTTP layer (`internal/web`)**:
- Routing (`routes.go`) uses the Go 1.22+ `http.ServeMux` pattern syntax. `GET /{$}` is used for the home page specifically because a bare `"/"` is a *subtree* match in this mux and would swallow the catch-all; the trailing `mux.HandleFunc("/", s.handleNotFound)` is the standard stdlib idiom for a 404 handler since `ServeMux` has no built-in `NotFound` hook. Keep both patterns' distinction in mind when adding routes.
- Templates (`server.go`) are built as one `*template.Template` per page, each `Clone()`d from `templates/base.html.tmpl` before parsing that page's own file. This is required, not incidental: `html/template` has no real inheritance, and parsing every page's `{{define "content"}}` block into one shared template would let the last-parsed page silently win for all of them. Add new pages by extending `pageNames` in `server.go` and adding a template file that defines a `content` block; don't try to collapse this into a single shared template tree.
- `handlers.go` renders through one shared `render()` helper (page name, HTTP status, data struct) so status-code handling and template-execution error logging live in one place. A missing `/blog/{slug}` calls the same `handleNotFound` as the router's catch-all, so all 404s look identical regardless of cause.
- The `Server` holds the loaded `*content.Site`, the pre-parsed page templates, and the `*zap.Logger` — all read-only after construction.
- Search (`GET /search?q=`) is a plain server-rendered round trip, no JS: `content.Site.Search` (in `internal/content/search.go`) does a case-insensitive substring match over each post's title/summary/tags/`Body` (the raw Markdown source, kept on `Post` specifically so search doesn't match rendered-HTML tag noise) and returns matches in newest-first order. The `search_form` template (defined once in `base.html.tmpl`, alongside — not inside — the `base` define) is a shared partial included via `{{template "search_form" .}}` from both `blog_index.html.tmpl` and `search.html.tmpl`; any data struct passed to a page including it needs a `Query` field for the input's prefill value, even if it's always empty (see `blogIndexData`).

**Logging (`internal/logging`)**: one `*zap.Logger` built once in `main.go` and passed down into `content.Load` and `web.New` (both packages take a `*zap.Logger` parameter rather than importing a global logger). It writes console-encoded (human-readable, not JSON) lines to both stdout and the file at `cfg.LogFile` via `zapcore.NewTee`, creating the log file's parent directory if needed. The file sink is a `lumberjack.Logger` (`gopkg.in/natefinch/lumberjack.v2`) with `MaxAge: 30` — note lumberjack only prunes *rotated* backup files older than that; rotation itself is still triggered by size (default 100MB, unconfigured here), not purely by elapsed time, so a long-running process with light log volume won't rotate on a monthly cadence by itself.

## Design constraints worth preserving

These came from explicit product requirements, not accidents of implementation — don't undo them incidentally while making unrelated changes:
- Visual design is intentionally plain: no colour gradients, no drop shadows, no card UI, no client-side JS, one hand-written stylesheet (`static/style.css`) with a serif body font and sans-serif chrome (nav/metadata), Wikipedia-ish palette, and a single centered content column with fluid mobile-friendly padding.
- The nav only ever has three items (Home / Blog / Portfolio) and relies on flexbox wrap rather than a hamburger menu/JS on narrow screens — don't add JS-driven responsive behavior for this.
- Third-party dependencies are kept deliberately minimal: `github.com/yuin/goldmark` (Markdown), `gopkg.in/yaml.v3` (frontmatter + config + portfolio parsing), `go.uber.org/zap` (logging), and `gopkg.in/natefinch/lumberjack.v2` (log file rotation). Avoid pulling in a router, ORM, or template-engine replacement — stdlib `net/http`/`html/template` are sufficient for this site's scope.
- Explicitly out of scope unless asked for: RSS/Atom feed, draft workflow, pagination, tag-filter pages, comments, auth.
