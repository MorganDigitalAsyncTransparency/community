# Architecture

This document describes the current architecture boundaries of discourse-observer and the reasoning behind them.

## Overview

discourse-observer is organized into layers that separate concerns cleanly. Each layer has a single responsibility and communicates with adjacent layers through well-defined interfaces.

### Data flow

This shows how data moves through the system at runtime:

```text
Discourse Forum (external)
        │ HTTP
        ▼
  src/discourse/          — fetches raw API data, handles auth and pagination
        │ raw API data
        ▼
  src/observer/           — normalizes, detects changes, coordinates fetch and store
        │ model types
        ▼
  src/storage/            — persists observations (SQLite initially)
```

Cross-cutting:

```text
  src/model/   — shared domain types, used by all modules above
  src/config/  — forum-specific configuration, provided at startup
```

### Dependency direction

The arrows above show *data flow*, not *import dependencies*. Imports follow dependency inversion:

- `model` has no imports. It is the innermost layer.
- `observer` imports only `model`. It defines interfaces (`FetchClient`, `StorageBackend`) for the adapters.
- `discourse` and `storage` import `model`. They implement the interfaces defined by `observer`. At runtime they are injected into the observer — the observer never imports them.
- `config` has no imports. Config values are read at startup and passed into module constructors.

## Layer responsibilities

### src/discourse/

Responsible for all communication with the Discourse API. This module knows how to authenticate, paginate, and fetch raw data from a single Discourse forum. It does **not** interpret the data — it returns it in a form close to the API response.

This isolation means that if the Discourse API changes, only this module needs to change.

### src/observer/

Responsible for change detection, normalization, and coordinating the fetch-observe-store cycle. The observer defines interfaces for its dependencies (`FetchClient`, `StorageBackend`) and works entirely in terms of `model` types.

The observer does not import `discourse` or `storage`. Those modules are injected at startup. This keeps the core logic independent of API details and persistence implementation.

### src/model/

Contains the internal normalized types and domain concepts used throughout the project. These types are independent of the Discourse API shape and represent the project's own understanding of forum activity.

The model module has no dependencies on other modules. It is a leaf dependency.

### src/config/

Holds forum-specific configuration and adaptation points. This is where a deployment specifies its forum URL, API credentials, polling intervals, and any forum-specific mappings (such as which categories to observe or which tags to track).

The config module has no imports. Config values are provided to other modules at startup — passed into constructors or init functions — rather than being imported directly by those modules.

### src/storage/

An abstraction point for persisting observed data. This module defines how observations are stored and retrieved. The initial implementation uses SQLite (decided in [ADR 0002](docs/decisions/0002-technology-choices.md)). An in-memory implementation may be added for testing. More sophisticated backends can be added later without touching the observer or model layers.

## What is intentionally not included

### Frontend / Dashboard

There is no frontend, dashboard, or reporting UI in this project yet. The observation and data layers need to be stable before building visualization on top of them. When a frontend is added, it will consume data through a backend API layer that also does not exist yet.

### Backend API

There is no HTTP API or server component yet. The project currently focuses on the data pipeline: fetching, observing, modeling, and storing. An API layer will be introduced when there is something meaningful to serve.

### Event / History model

The observation layer captures current state and changes, but a formal event sourcing or history model is not implemented yet. This is a likely future addition once the observation patterns are understood.

## Design decisions

Architecture decisions are recorded as ADRs in [docs/decisions/](docs/decisions/). Consult them for context on why the architecture is shaped this way.
