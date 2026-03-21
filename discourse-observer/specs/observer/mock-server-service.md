# Mock Server Service

The mock Discourse server runs as a docker-compose service so the full sync pipeline works in dev mode without real Discourse credentials.

Related: [scheduler.md](scheduler.md) (SC-5), [sync-strategy.md](../../docs/sync-strategy.md), [getting-started.md](../../docs/getting-started.md)

---

## Requirements

MS-1 — **Exported handler.** The mock server package exports a `Handler()` function returning `http.Handler`. This allows both `httptest.Server` usage (tests) and standalone HTTP server usage (docker service). Existing `New()` and `NewWithPageSize()` delegate to the handler.

MS-2 — **Standalone entrypoint.** A `cmd/mockserver/main.go` binary listens on `:9920` using the exported handler. It logs its listen address on startup and shuts down gracefully on SIGINT. The `MOCK_PAGE_SIZE` env var overrides the default page size (30); docker-compose sets it to 5 so pagination is visible in sync logs.

MS-3 — **Docker service.** A `mockserver` service in `docker-compose.yml` builds from `docker/mockserver.Dockerfile` and exposes port 9920 on the internal Docker network. The backend service depends on mockserver being healthy.

MS-4 — **Dockerfile.** `docker/mockserver.Dockerfile` follows the same multi-stage pattern as `backend.Dockerfile`: build in `golang:alpine`, run in `alpine`. The binary is built from `./backend/cmd/mockserver`.

MS-5 — **Dev-mode detection.** The backend starts the sync scheduler when `DISCOURSE_BASE_URL` is set, regardless of whether `DISCOURSE_API_TOKEN` is empty. Dev mode (sync disabled) is when `DISCOURSE_BASE_URL` is empty. This replaces the current check on `DISCOURSE_API_TOKEN`.

MS-6 — **Environment defaults.** `.env.example` sets `DISCOURSE_BASE_URL=http://mockserver:9920` so the backend container reaches the mock server by Docker service name. A comment documents `http://localhost:9920` for local non-Docker development.

MS-7 — **No seeding required.** The mock server service replaces the old `maybe-seed` approach. The scheduler performs a full initial sync from the mock server on first launch — no pre-seeding needed. `make seed` remains available for manual use.

---

## Verification

| Req | Method |
|-----|--------|
| MS-1 | Existing mock server tests pass unchanged (they use `New()` which delegates to `Handler()`). |
| MS-2 | `go build ./backend/cmd/mockserver` compiles. Manual: run binary, curl `localhost:9920/latest.json`. |
| MS-3 | `docker compose config` validates. `docker compose up` starts all three services. |
| MS-4 | `docker compose build mockserver` succeeds. |
| MS-5 | Backend logs "sync scheduler started" when `DISCOURSE_BASE_URL` is set and `DISCOURSE_API_TOKEN` is empty. |
| MS-6 | `.env.example` contains the mock server URL. |
| MS-7 | `make start` launches mock server, scheduler runs initial sync automatically. |
