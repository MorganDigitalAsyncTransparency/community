# Getting Started

This guide explains how to configure, start, and access the discourse-observer dashboard locally.

## Quick start

```sh
make start
```

This single command handles the full onboarding flow: installs dependencies, creates config files, runs verification, builds containers, and opens the dashboard.

No Discourse forum is needed to get started. The default `.env` has no API token, so `make start` automatically seeds mock topics into a local SQLite database. The dashboard opens fully populated with realistic test data.

To connect to a real forum later, edit `.env` with your Discourse credentials (see [Configure for a real forum](#configure-for-a-real-forum)). The next `make start` detects the token and skips seeding.

After code changes, use `make restart` to rebuild and relaunch.

If you only want to install dependencies without launching, run `make setup` separately.

The rest of this guide covers prerequisites, configuration, and how the stack works.

## Prerequisites

Run the setup script to check what you have and what's missing:

```sh
sh scripts/setup.sh
```

If you already have `make` installed, you can use `make check` instead.

The following tools are needed. Go and Node.js are only required for local development outside Docker.

| Tool | Windows | macOS | Linux |
|------|---------|-------|-------|
| **[GNU Make](https://www.gnu.org/software/make/)** | `choco install make` | included with Xcode CLT (`xcode-select --install`) | `sudo apt install make` (Debian) / `sudo dnf install make` (Fedora) |
| **[Docker Desktop](https://docs.docker.com/desktop/)** | `choco install docker-desktop` | `brew install --cask docker` | [Docker Engine](https://docs.docker.com/engine/install/) + [Compose plugin](https://docs.docker.com/compose/install/) |
| **[Go 1.26+](https://go.dev/dl/)** | `choco install golang` | `brew install go` | [go.dev/dl](https://go.dev/dl/) (distro packages are often outdated) |
| **[Node.js 24+](https://nodejs.org/)** | `choco install nodejs-lts` | `brew install node` | `sudo apt install nodejs npm` (Debian) / `sudo dnf install nodejs npm` (Fedora) |

To connect to a real Discourse forum, you also need an API token (read-only is sufficient). This is not required for local development with mock data.

### VS Code extensions

The repository includes a `.vscode/extensions.json` with recommended extensions. VS Code will prompt you to install them when you open the project. The recommendations include:

- **Go** — Go language support, debugging, and linting
- **ESLint** — JavaScript/TypeScript linting (frontend)
- **Stylelint** — CSS linting (frontend)
- **markdownlint** — Markdown linting (documentation)
- **Docker** — Dockerfile and Compose support

### VS Code terminal setup

This project uses shell scripts and `make`, which require a Unix-compatible shell. If your VS Code terminal defaults to something else (e.g. PowerShell on Windows), switch it to a Unix shell:

1. Open the command palette: `Ctrl+Shift+P`
2. Search for **Terminal: Select Default Profile**
3. Choose a Unix-compatible shell (e.g. **Bash**, or **Git Bash** on Windows via [Git for Windows](https://git-scm.com/downloads/win))

New terminal panels will then run commands like `make start` and `sh scripts/setup.sh` directly.

## Configure for a real forum

`make start` works out of the box with mock data. This section is only needed when you want to connect to a real Discourse forum.

### Discourse credentials

Edit `.env` (created automatically by `make start` from `.env.example`):

```sh
DISCOURSE_BASE_URL=https://your-forum.example.com
DISCOURSE_API_TOKEN=your-api-token-here
DISCOURSE_API_USER=nickname
```

When `DISCOURSE_API_TOKEN` has a value, `make start` skips mock seeding and the backend fetches from your forum instead.

The `.env` file is gitignored and will not be committed.

### Tag configuration

Edit `config/tagConfig.json` (created automatically from the example file) to define your monitored tags, area groupings, SLO thresholds, and stalled-topic settings. The file has four sections:

- **`defaults`** — fallback values for `stalledDays`, `area`, and `slo`. Applied to tags that don't override them. Tags using defaults are marked in the UI so viewers know the values are not explicitly agreed.
- **`areas`** — named groups with a `primaryTag` for display ordering.
- **`tags`** — one entry per monitored tag. Each entry can set `area`, `closedTag`, `stalledDays`, and `slo`. All fields are optional — absent values fall back to defaults (except `closedTag`, which has no default). Writing a value explicitly means it has been agreed upon; changing defaults later won't affect tags with explicit values.

The file is gitignored.

## Build and start

If connecting to a real forum, complete the [Configure for a real forum](#configure-for-a-real-forum) step first. For mock data, no configuration is needed — `make start` handles everything.

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
curl http://localhost:3000/api/v1/status   # should return {"lastSyncedAt":...,"version":"..."}
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
- **Data directory** — the SQLite database (`data/analytics.db`) is bind-mounted from the project's `data/` directory, shared between host and container.

## Rebuild after code changes

```sh
make restart
```

This stops the running containers, rebuilds changed layers, starts everything again, and opens the dashboard. Docker layer caching means only changed layers are rebuilt.

## Make targets

| Command | What it does |
|---|---|
| `make start` | One-command onboarding: setup, verify, configure, build, launch, open browser (auto-seeds mock data if no API token) |
| `make seed` | Populate SQLite with mock topics for development |
| `make restart` | Verify, rebuild, and relaunch after code changes |
| `make verify` | Run all linters and tests |
| `make lint` | Run all linters (Go + markdown + frontend) |
| `make test` | Run all tests (Go + frontend) |
| `make build` | Build Docker containers |
| `make up` | Start containers in background |
| `make down` | Stop containers |
| `make check` | Check that prerequisites are installed |
| `make setup` | Install dependencies and configure git hooks |
| `make docs` | Build and serve documentation locally with live reload |
| `make open` | Open dashboard in browser |
