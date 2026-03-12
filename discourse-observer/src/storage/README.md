# storage/

This module provides an abstraction for persisting observed data.

## Responsibility

The storage module defines how observations are saved and retrieved. It acts as a boundary between the observation logic and whatever persistence mechanism is used.

## Current status

This module is not yet fully implemented. The abstraction point exists so that the observer can be built against a storage interface without committing to a specific backend.

## Planned approach

The initial implementation will likely be simple — local file storage, SQLite, or an in-memory store for development. The abstraction allows swapping to a more capable backend (PostgreSQL, cloud storage, a time-series database) later without changing the observer or model layers.

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
