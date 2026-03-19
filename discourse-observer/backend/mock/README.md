# Mock

Hardcoded topic fixtures for development and testing.

## Responsibility

Provides a realistic set of 44 topics covering all API endpoint scenarios: unreplied, resolved (solved and self-closed), replied-open, untagged, stalled, and multi-tag topics. Topics span a range of creation dates to exercise period filtering and time bucketing.

## Does not

- Implement any calculation logic
- Access external services or files
- Represent production data

## Dependencies

- `backend/model/` — Topic type

## Future

This package will be removed when the data pipeline and SQLite analytical store ([ADR 0006](../../docs/decisions/0006-analytical-storage.md)) are implemented. At that point, API handlers will query the store instead of reading mock data.
