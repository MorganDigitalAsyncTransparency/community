# Architecture

This document describes the current architecture boundaries of discourse-observer and the reasoning behind them.

## Overview

discourse-observer is organized into layers that separate concerns cleanly. Each layer has a single responsibility and communicates with adjacent layers through well-defined interfaces.

### Data flow

Data flow at runtime:

```text
Discourse Forum (external)
        │ HTTP
        ▼
  backend/discourse/      — fetches raw API data, handles auth and pagination
        │ raw API data
        ▼
  backend/observer/       — normalizes, detects changes, coordinates fetch and store
        │ model types
        ▼
  backend/storage/        — persists raw observations (NDJSON files)
```

API serving:

```text
  backend/api/           — HTTP handlers, routing, filter parsing, JSON responses
        │ domain types
        ▼
  backend/domain/        — pure calculation functions (medians, bucketing, SLO, heatmap)
        │ model types
        ▼
  backend/model/         — shared domain types
```

Cross-cutting:

```text
  backend/model/   — shared domain types, used by all modules above
  backend/config/  — forum-specific configuration, provided at startup
  backend/mock/    — hardcoded topic fixtures for development and testing
```

### Dependency direction

The arrows above show *data flow*, not *import dependencies*. Imports follow dependency inversion:

- `model` has no imports. It is the innermost layer.
- `observer` imports only `model`. It defines interfaces (`FetchClient`, `StorageBackend`) for the adapters.
- `discourse` and `storage` import `model`. They implement the interfaces defined by `observer`. At runtime they are injected into the observer — the observer never imports them.
- `domain` imports only `model`. It contains pure calculation functions with no framework or I/O dependencies.
- `api` imports `domain` and `model`. It handles HTTP concerns (routing, JSON, filter parsing) and delegates computation to `domain`.
- `mock` imports only `model`. It provides hardcoded topic fixtures for development and testing.
- `config` has no imports. Config values are read at startup and passed into module constructors.

## Layer responsibilities

### backend/discourse/

Responsible for all communication with the Discourse API. This module knows how to authenticate, paginate, and fetch raw data from a single Discourse forum. It does **not** interpret the data — it returns it in a form close to the API response.

This isolation means that if the Discourse API changes, only this module needs to change.

### backend/observer/

Responsible for change detection, normalization, and coordinating the fetch-observe-store cycle. The observer defines interfaces for its dependencies (`FetchClient`, `StorageBackend`) and works entirely in terms of `model` types.

The observer does not import `discourse` or `storage`. Those modules are injected at startup. This keeps the core logic independent of API details and persistence implementation.

### backend/model/

Contains the internal normalized types and domain concepts used throughout the project. These types are independent of the Discourse API shape and represent the project's own understanding of forum activity.

The model module has no dependencies on other modules.

### backend/config/

Holds forum-specific configuration and adaptation points. This is where a deployment specifies its forum URL, API credentials, polling intervals, and any forum-specific mappings (such as which categories to observe or which tags to track).

The config module has no imports. Config values are provided to other modules at startup — passed into constructors or init functions — rather than being imported directly by those modules.

### backend/storage/

Persists normalized topics. The current implementation uses SQLite (decided in [ADR 0006](docs/decisions/0006-analytical-storage.md)) to store the pipeline output. Raw append-only NDJSON files (decided in [ADR 0005](docs/decisions/0005-storage-format.md)) are planned for a future layer. The storage interface (`StorageBackend`) is defined by `observer/` and implemented here.

### backend/api/

HTTP handlers for all `/api/v1/` endpoints defined in the [API contract](specs/api/api-contract.md). Responsible for routing, query parameter parsing and validation, filter application, and JSON response encoding. Delegates all computation to `backend/domain/`. Does not contain business logic.

### backend/domain/

Pure calculation functions implementing domain aggregates: medians, time bucketing, histogram distribution, tag rankings, SLO violation detection, compliance computation, and heatmap generation. These functions receive pre-filtered topic slices and return computed results. No HTTP, I/O, or framework dependencies.

### backend/mock/

Hardcoded topic fixtures used as the data source during development. Provides a realistic set of topics covering all endpoint scenarios (unreplied, resolved, stalled, untagged, multi-tag). Also used by the mock Discourse server (`discourse/mockserver/`) for pipeline integration tests. The API layer still reads from `mock/`; wiring it to SQLite is the next step.

## Terminology

These terms have specific meanings in this project. Other documentation uses them consistently with these definitions.

| Term | Meaning |
|------|---------|
| **Fetch** | Retrieve raw data from the Discourse API. Happens in `backend/discourse/`. |
| **Normalize** | Transform raw API data into internal `model` types. Happens in `backend/observer/`. |
| **Observation** | A structured record of what the observer saw — a snapshot of one or more topics at a point in time, including what changed since the previous observation. |
| **Sync** | A complete fetch-normalize-store cycle: poll the API for updates, produce observations, persist them. |
| **Poll** | Check the API for new or changed data. Polling is the mechanism; sync is the full cycle. |

## What is intentionally not included

### Frontend / Dashboard

A React/TypeScript frontend exists in `frontend/` and renders a multi-page dashboard (Queue, Response metrics, Distribution, SLO, Activity) using mock data. Individual feature specs live in `specs/dashboard/`: queue-visibility, response-metrics, time-period-filter, tag-distribution, slo-monitoring, tag-area-filter, topic-intake (bucketing logic reused by volume charts), stalled-topics (on the Queue page), and peak-activity (on the Activity page). Cross-cutting component behavior is in `specs/dashboard/dashboard-components.md`. When the backend API is available, the frontend will consume domain aggregate endpoints ([specs/api/api-contract.md](specs/api/api-contract.md)) and its calculation logic (medians, bucketing, rankings) will be removed — the backend takes over that responsibility ([ADR 0012](docs/decisions/0012-api-responsibility-model.md)).

### Backend API

The API contract is specified in [specs/api/api-contract.md](specs/api/api-contract.md). It defines domain aggregate endpoints that serve pre-computed data to the frontend and future consumers (MCP servers, CLI tools). The responsibility model — why the backend computes aggregates rather than serving raw data — is recorded in [ADR 0012](docs/decisions/0012-api-responsibility-model.md).

The API is implemented in `backend/api/` with domain calculations in `backend/domain/`. It currently serves mock data from `backend/mock/`. The data pipeline (fetch → observe → store) is implemented and tested, but the API layer has not yet been wired to read from SQLite. That integration is the next step.

### Event / History model

The observation layer captures current state and changes, but a formal event sourcing or history model is not implemented yet. This is a likely future addition once the observation patterns are understood.

## Design decisions

Architecture decisions are recorded as ADRs in [docs/decisions/](docs/decisions/README.md). Consult them for context on why the architecture is shaped this way.
