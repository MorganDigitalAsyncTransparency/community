# Architecture

This document describes the current architecture boundaries of discourse-observer and the reasoning behind them.

## Overview

discourse-observer is organized into layers that separate concerns cleanly. Each layer has a single responsibility and communicates with adjacent layers through well-defined interfaces.

```text
┌─────────────────────────────────┐
│  Discourse Forum (external)     │
└──────────────┬──────────────────┘
               │ API calls
┌──────────────▼──────────────────┐
│  src/discourse/                 │
│  Raw API integration            │
│  Fetches data from Discourse    │
└──────────────┬──────────────────┘
               │ Raw API responses
┌──────────────▼──────────────────┐
│  src/observer/                  │
│  Observation logic              │
│  Detects changes, normalizes    │
└──────────────┬──────────────────┘
               │ Normalized observations
┌──────────────▼──────────────────┐
│  src/model/                     │
│  Internal domain types          │
│  Normalized, API-independent    │
└──────────────┬──────────────────┘
               │ Domain objects
┌──────────────▼──────────────────┐
│  src/storage/                   │
│  Persistence abstraction        │
│  Stores observations for later  │
└─────────────────────────────────┘
```

Cross-cutting:

```text
┌─────────────────────────────────┐
│  src/config/                    │
│  Forum-specific configuration   │
│  Adaptation points              │
└─────────────────────────────────┘
```

## Layer responsibilities

### src/discourse/

Responsible for all communication with the Discourse API. This module knows how to authenticate, paginate, and fetch raw data from a single Discourse forum. It does **not** interpret the data — it returns it in a form close to the API response.

This isolation means that if the Discourse API changes, only this module needs to change.

### src/observer/

Responsible for turning raw Discourse data into meaningful observations. This is where change detection, filtering, and normalization happen. The observer takes raw API data from the discourse module and produces structured observations using types from the model module.

The observer does not call the Discourse API directly — it receives data from the discourse layer.

### src/model/

Contains the internal normalized types and domain concepts used throughout the project. These types are independent of the Discourse API shape and represent the project's own understanding of forum activity.

The model module has no dependencies on other modules. It is a leaf dependency.

### src/config/

Holds forum-specific configuration and adaptation points. This is where a deployment specifies its forum URL, API credentials, polling intervals, and any forum-specific mappings (such as which categories to observe or which tags to track).

The config module is read by other modules but does not depend on them.

### src/storage/

An abstraction point for persisting observed data. This module defines how observations are stored and retrieved. The initial implementation may be as simple as local file storage or an in-memory store. More sophisticated backends (databases, cloud storage) can be added later behind this abstraction.

## What is intentionally not included

### Frontend / Dashboard

There is no frontend, dashboard, or reporting UI in this project yet. The observation and data layers need to be stable before building visualization on top of them. When a frontend is added, it will consume data through a backend API layer that also does not exist yet.

### Backend API

There is no HTTP API or server component yet. The project currently focuses on the data pipeline: fetching, observing, modeling, and storing. An API layer will be introduced when there is something meaningful to serve.

### Event / History model

The observation layer captures current state and changes, but a formal event sourcing or history model is not implemented yet. This is a likely future addition once the observation patterns are understood.

## Design decisions

Architecture decisions are recorded as ADRs in [docs/decisions/](docs/decisions/). Consult them for context on why the architecture is shaped this way.
