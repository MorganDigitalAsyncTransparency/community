# observer/

This module is responsible for turning raw Discourse data into structured observations.

## Responsibility

The observer sits between the raw API layer and the rest of the system. It:

- Receives raw data from `discourse/`
- Normalizes it into internal types defined in `model/`
- Detects meaningful changes by comparing current data with previous observations
- Produces observation records that capture what changed, when, and in what context
- Coordinates fetch operations (scheduled sync, on-demand, backfill)

## Boundaries

This module does **not**:

- Call the Discourse API directly (it receives data from `discourse/`)
- Define its own data types (it uses types from `model/`)
- Persist data directly (it passes observations to `storage/`)
- Perform analytics or derive higher-level events

## Current implementation

`Observer` in `observer.go` coordinates sync cycles using two interfaces: `FetchClient` (implemented by `discourse/`) and `StorageBackend` (implemented by `storage/`). It supports two sync modes: initial sync (full crawl with page-by-page resume) and delta sync (incremental from a stored watermark). `Run()` auto-detects which mode to use based on whether a watermark exists. `Normalize` maps `model.RawTopic` fields to domain types, deriving outcome (solved/self-closed/open) from Discourse flags and constructing topic URLs.

## Design expectations

- Observation logic should be composed of pure transformation functions where possible
- Change detection should be deterministic and testable with known inputs
- The module should not assume which categories, tags, or workflows matter — that comes from `config/`
- Observation operations (sync, fetch, backfill) should be composable and independently testable
