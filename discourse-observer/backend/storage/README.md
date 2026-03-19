# storage/

This module provides an abstraction for persisting observed data.

## Responsibility

The storage module defines how observations are saved and retrieved. It acts as a boundary between the observation logic and whatever persistence mechanism is used.

## Current implementation

`SQLiteStore` in `sqlite.go` persists normalized topics to a SQLite database using `modernc.org/sqlite` (pure Go, no CGO). It implements the `StorageBackend` interface defined by `observer/`. Topics are upserted by ID, making the pipeline idempotent. Schema migrations run on startup.

The abstraction allows swapping to a different backend (PostgreSQL, in-memory for testing) without changing the observer or model layers.

## Boundaries

This module:

- Depends on types from `model/` for what it stores and returns
- Does not depend on `discourse/` or `observer/`
- Does not interpret or analyze stored data
- Provides a consistent interface regardless of the underlying backend

## Design expectations

- The storage interface should be defined in terms of `model/` types, not raw data shapes
- Implementations should be swappable behind the interface
- The initial implementation should prioritize simplicity and local development over scalability
- Storage configuration (paths, connection strings) comes from `config/`
