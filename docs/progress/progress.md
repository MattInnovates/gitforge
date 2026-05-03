# GitForge Build Progress

## Modules

| # | Module | Status |
|---|---|---|
| 1 | Foundation | COMPLETE - Steps 1.1 through 1.6 done |
| 2 | Core / Events | COMPLETE - Event bus + tests |
| 3 | Storage | COMPLETE - Storage interfaces + sqlite stores + tests |
| 4 | Git HTTP | COMPLETE - git http-backend bridge + health endpoint |
| 5 | Git SSH | COMPLETE - SSH git upload-pack/receive-pack bridge + tests |
| 6 | API | COMPLETE - users/repos REST handlers + tests |
| 7 | Web UI | COMPLETE - HTML UI routes + forms + tests |
| 8 | Worker | COMPLETE - worker queue runtime + retries + tests |
| 9 | Auth | COMPLETE - password hashing + bearer tokens + protected routes |
| 10 | CLI Client | COMPLETE - gitforge CLI with auth and git passthrough |

---

## Module 1 - Foundation

Goal: Get the project compiling with a shared config, shared types, and database connection.

### Steps

| # | Task | Status |
|---|---|---|
| 1.1 | Initialize go.mod | DONE - github.com/MattInnovates/gitforge |
| 1.2 | Define shared config struct and loader (env vars) | DONE - core/config with env defaults |
| 1.3 | Define core domain types (User, Repo) | DONE - core/domain User + Repository models |
| 1.4 | Set up database connection + schema migrations | DONE - sqlite connection + schema_migrations + initial tables |
| 1.5 | Wire config into each cmd/main.go stub | DONE - web/git-http/git-ssh/worker use shared loader |
| 1.6 | Verify all binaries compile cleanly | DONE - go build ./... |

---

## Module 2 - Core / Events

Goal: Provide internal in-process pub/sub so services can communicate through events.

### Steps

| # | Task | Status |
|---|---|---|
| 2.1 | Define event envelope fields | DONE - name, timestamp, payload |
| 2.2 | Implement in-process event bus | DONE - subscribe/publish/unsubscribe/close |
| 2.3 | Add event bus tests | DONE - publish, queue-full, unsubscribe, close |
| 2.4 | Verify package behavior with go test | DONE - go test ./... |

---

## Module 3 - Storage

Goal: Add storage interfaces and sqlite-backed implementations for core entities.

### Steps

| # | Task | Status |
|---|---|---|
| 3.1 | Define storage interfaces for users and repositories | DONE - UserStore and RepositoryStore |
| 3.2 | Implement sqlite user store | DONE - create/get by id/get by username |
| 3.3 | Implement sqlite repository store | DONE - create/get by owner+name/list by owner |
| 3.4 | Add storage tests | DONE - create/get/list + not found path |
| 3.5 | Verify package behavior with go test | DONE - go test ./... |

---

## Module 4 - Git HTTP

Goal: Serve Git smart HTTP operations by bridging requests to git http-backend.

### Steps

| # | Task | Status |
|---|---|---|
| 4.1 | Resolve repos root and ensure directory exists | DONE - absolute path + mkdir |
| 4.2 | Bridge requests to git http-backend | DONE - net/http/cgi handler |
| 4.3 | Add basic operational endpoints/logging | DONE - /healthz + request logging |
| 4.4 | Verify package behavior with go test | DONE - go test ./... |

---

## Module 5 - Git SSH

Goal: Serve Git SSH operations by handling exec requests and running safe git pack commands.

### Steps

| # | Task | Status |
|---|---|---|
| 5.1 | Implement SSH server listener and session handling | DONE - x/crypto/ssh server with session channels |
| 5.2 | Support git exec commands | DONE - git-upload-pack and git-receive-pack |
| 5.3 | Add command and path safety checks | DONE - parser + repo root escape protection |
| 5.4 | Add module tests | DONE - parse and path resolver tests |
| 5.5 | Verify package behavior with go test | DONE - go test ./... |

---

## Module 6 - API

Goal: Provide REST endpoints for core entities and wire them into the web server.

### Steps

| # | Task | Status |
|---|---|---|
| 6.1 | Implement users API endpoints | DONE - POST /api/users and GET /api/users/{username} |
| 6.2 | Implement repositories API endpoints | DONE - POST/GET /api/repos and GET /api/repos/{owner}/{name} |
| 6.3 | Add API health endpoint | DONE - GET /api/healthz |
| 6.4 | Wire routes into web service | DONE - cmd/web uses api.SetupRoutes |
| 6.5 | Add API tests | DONE - create/get/list/not-found flows |
| 6.6 | Verify package behavior with go test | DONE - go test ./... |

---

## Module 7 - Web UI

Goal: Provide a simple browser UI for core create/list actions while keeping API routes available.

### Steps

| # | Task | Status |
|---|---|---|
| 7.1 | Add HTML-rendered home page | DONE - dashboard with forms and repo list panel |
| 7.2 | Add UI create user flow | DONE - POST /ui/users with redirect |
| 7.3 | Add UI create repository flow | DONE - POST /ui/repos with redirect |
| 7.4 | Compose API and UI in web server | DONE - /api routed to API, / to UI |
| 7.5 | Add Web UI tests | DONE - render and form redirect flows |
| 7.6 | Verify package behavior with go test | DONE - go test ./... |
| 7.7 | Add owner profile and repository pages | DONE - /u/{owner} and /u/{owner}/{repo} views wired |
| 7.8 | Add profile/repo route coverage tests | DONE - route render + not-found tests |

---

## Module 8 - Worker

Goal: Provide a background worker runtime with queueing and retry semantics.

### Steps

| # | Task | Status |
|---|---|---|
| 8.1 | Implement reusable worker runtime | DONE - core/worker Runner + Job model |
| 8.2 | Add queue, concurrency, and retry logic | DONE - buffered queue, worker pool, retry attempts |
| 8.3 | Wire runtime into worker command | DONE - cmd/worker boots runner and event-driven enqueue |
| 8.4 | Add worker module tests | DONE - processing, retry, queue-full coverage |
| 8.5 | Verify package behavior with go test | DONE - go test ./... |

---

## Module 9 - Auth

Goal: Add secure authentication primitives and protected API flows.

### Steps

| # | Task | Status |
|---|---|---|
| 9.1 | Add password hashing and token service | DONE - core/auth service (bcrypt + HMAC token claims) |
| 9.2 | Add auth configuration | DONE - token secret and token TTL env settings |
| 9.3 | Add auth API endpoints | DONE - POST /api/auth/login and GET /api/auth/me |
| 9.4 | Protect repo creation with bearer auth | DONE - /api/repos POST requires valid token |
| 9.5 | Wire auth into web/API/UI constructors | DONE - cmd/web passes auth service to handlers |
| 9.6 | Add auth tests | DONE - core/auth tests + API login/me coverage |
| 9.7 | Verify package behavior with go test | DONE - go test ./... |

---

## Module 10 - CLI Client

Goal: Provide a practical CLI for basic Git operations plus GitForge auth/API workflows.

### Steps

| # | Task | Status |
|---|---|---|
| 10.1 | Create CLI binary entrypoint | DONE - cmd/gitforge main command router |
| 10.2 | Add Git command passthrough | DONE - init/add/commit/status/log/clone/push/pull/branch/checkout |
| 10.3 | Add auth login and identity commands | DONE - login and me commands with bearer token handling |
| 10.4 | Add API helper commands | DONE - user-create and repo-create commands |
| 10.5 | Add local session storage | DONE - config-dir session.json save/load |
| 10.6 | Add CLI tests | DONE - session storage round-trip test |
| 10.7 | Verify package behavior with go test | DONE - go test ./... |
