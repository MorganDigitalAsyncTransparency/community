# Mock

Hardcoded topic fixtures for development and testing.

## Responsibility

Provides a realistic set of topics covering all API endpoint scenarios: unreplied, resolved (solved and self-closed), replied-open, untagged, stalled, and multi-tag topics. Topics span a range of creation dates to exercise period filtering and time bucketing.

## Does not

- Implement any calculation logic
- Access external services or files
- Represent production data

## Dependencies

- `backend/model/` — Topic type

## Current usage

API handlers now query SQLite directly. This package remains as a test fixture provider: it is used by the mock Discourse server (`discourse/mockserver/`) for pipeline integration tests and by API contract tests (seeded into a temporary SQLite database). It is not imported at runtime.
