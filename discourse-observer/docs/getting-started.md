# Getting Started

This guide explains how to configure, start, and access the discourse-observer dashboard locally.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- A Discourse forum API token (read-only is sufficient)

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

```sh
make build   # build both containers
make up      # start in background
```

Or with Docker Compose directly:

```sh
docker compose build
docker compose up -d
```

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
make build   # rebuild changed containers
make down && make up   # restart
```

Docker layer caching means only changed layers are rebuilt.
