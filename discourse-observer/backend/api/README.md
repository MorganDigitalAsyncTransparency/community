# API

HTTP handlers for all `/api/v1/` endpoints defined in the [API contract](../../specs/api/api-contract.md).

## Responsibility

- Route registration for all 17 endpoints
- Query parameter parsing and validation (`period`, `from`/`to`, `tag`)
- Filter application (delegated to `domain/` functions)
- JSON response encoding with correct field names and types
- Error responses with consistent `{"error": "message"}` structure

## Does not

- Contain business logic or domain calculations — those live in `backend/domain/`
- Access the Discourse API, storage, or file system
- Define data types — uses `backend/model/` types

## Dependencies

- `backend/model/` — Topic and config types
- `backend/domain/` — Calculation functions
