# Architecture Overview

## Services

GitForge is split into four independent binaries:

| Binary | Purpose |
|--------|---------|
| `cmd/web` | Serves the web UI and REST API |
| `cmd/git-http` | Handles Git smart HTTP protocol (`git push`, `git clone` over HTTP) |
| `cmd/git-ssh` | Handles Git over SSH |
| `cmd/worker` | Runs background jobs (webhooks, indexing, cleanup) |

## Package Layout

```
api/          - HTTP route definitions
cmd/          - Entry points for each binary
  git-http/
  git-ssh/
  web/
  worker/
core/         - Shared business logic
  events/     - Internal event bus
docs/         - Documentation
```

## Request Flow

```
Browser / Git Client
        │
        ▼
   cmd/web  ──── api/routes.go ──── core/
        │
        └── publishes events ──── cmd/worker
```

## Data Flow

- Git operations go through `cmd/git-http` or `cmd/git-ssh`
- Both backends delegate to shared `core/` logic
- Side effects (webhooks, notifications) are dispatched via the event bus in `core/events`

## Further Reading

- [API Reference](../api/reference.md)
- [Configuration](../configuration/configuration.md)
