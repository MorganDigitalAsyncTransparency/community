# Getting Started

This guide explains how to configure, start, and access the discourse-observer dashboard locally.

## Quick start

```sh
make start
```

This copies `.env.example` to `.env` (if needed), builds the containers, starts them, and opens the dashboard. Edit `.env` with your Discourse credentials before the first run.

After code changes, use `make restart` to rebuild and relaunch.

The rest of this guide covers prerequisites, configuration, and how the stack works.

## Prerequisites

Install these before continuing:

1. [Docker Desktop](https://docs.docker.com/desktop/) — includes Docker Engine, Docker Compose, BuildKit, and the CLI. On Linux you can alternatively install [Docker Engine](https://docs.docker.com/engine/install/) with the [Compose plugin](https://docs.docker.com/compose/install/) and [Buildx plugin](https://docs.docker.com/build/install-buildx/) separately.
2. [Go 1.26+](https://go.dev/dl/) — only needed for local development outside Docker.
3. [Node.js 24+](https://nodejs.org/) (includes npm) — only needed for local development outside Docker.

You also need a Discourse forum API token (read-only is sufficient).

## Configure

Copy the example environment file and fill in your values:

```sh
cp .env.example .env
```

Edit `.env`:

```sh
DISCOURSE_BASE_URL=https://your-forum.example.com
DISCOURSE_API_TOKEN=your-api-token-here
DISCOURSE_API_USER=nickname
```

The `.env` file is gitignored and will not be committed.

## Build and start

Make sure you have completed the [Configure](#configure) step first — `docker compose` will fail if `.env` is missing.

```sh
make build   # build both containers
make up      # start in background
```

Or with Docker Compose directly:

```sh
docker compose build
docker compose up -d
```

## Verify

After starting, confirm the stack is healthy:

```sh
curl http://localhost:3000/api/health   # should return {"status":"ok"}
```

Open <http://localhost:3000> in a browser to see the dashboard.

## Access the dashboard

Once running, the dashboard is available at:

- **From this machine:** <http://localhost:3000>
- **From another machine on the local network:** `http://<this-machine-ip>:3000`

To find this machine's IP address:

```sh
# Linux / macOS
hostname -I

# Windows
ipconfig
```

## Stop

```sh
make down
```

## How it works

The setup runs two containers on an internal Docker network:

```text
Browser ──:3000──▸ nginx (frontend)
                    ├── /*    → serves the React dashboard
                    └── /api/ → proxies to the Go backend
                                     │
                                Go API (backend, internal only)
                                     │
                                Discourse API (external)
```

- **Frontend container** — nginx serves the React build and proxies API requests to the backend. This is the only container exposed to the host.
- **Backend container** — Go service that polls Discourse and serves the API. Not directly accessible from outside Docker.
- **Data volume** — NDJSON files and the SQLite database are stored in a Docker volume that survives container restarts.

## Rebuild after code changes

```sh
make restart
```

This stops the running containers, rebuilds changed layers, starts everything again, and opens the dashboard. Docker layer caching means only changed layers are rebuilt.
