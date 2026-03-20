# API Strategy

This document defines the overall design principles, versioning approach, and conventions for the discourse-observer HTTP API. It complements the [API contract spec](../specs/api/api-contract.md), which defines individual endpoints and their data shapes.

The design rationale — why a domain aggregate API over raw data or a query engine — is recorded in [ADR 0012](decisions/0012-api-responsibility-model.md).

---

## Purpose

The API is the system's primary data interface. It serves pre-computed domain aggregates to any consumer that needs support metrics:

- **Frontend dashboard** — renders charts, tables, and cards from API responses.
- **MCP server** — wraps endpoints as tools so AI assistants can answer questions about support activity.
- **Future consumers** — CLI tools, notification services, external integrations.

The API is the single source of computed metrics. Consumers do not reimplement calculation logic — they render or relay what the API provides.

---

## Design principles

### Domain-oriented, not view-oriented

Endpoints are organized around data concepts — queue state, response metrics, tag distribution, SLO compliance, activity patterns — not around UI components. A layout change in the dashboard does not require an API change. A new frontend page that combines existing metrics requires no additional backend endpoints.

### Machine-readable values, consumer-formatted display

The API returns values in precise, unambiguous units:

- Durations in **milliseconds** (integers).
- Percentages as **whole numbers** (0–100).
- Timestamps as **ISO 8601 UTC** strings.
- Counts as **integers**.

Consumers format these for their medium — the dashboard shows "2d 3h", an MCP tool says "51 hours", a CSV export writes raw numbers. The API never returns pre-formatted display strings.

### Field naming

Response fields use **camelCase** (JSON/JavaScript convention). Where a field corresponds to a Discourse API field, the name is aligned with Discourse's naming — converted from snake_case to camelCase (e.g., Discourse `created_at` → `createdAt`). Where a field name would be ambiguous, a qualifier is added (e.g., `categoryName` rather than `category`, since Discourse uses both `category_id` and category name).

### Consistent filtering

All data-serving endpoints share the same filter vocabulary:

| Parameter | Meaning | Default |
|-----------|---------|---------|
| `period` | Rolling window: `7d`, `30d`, `1y`, `all` | `all` |
| `from`, `to` | Custom UTC date range (overrides `period`) | — |
| `tag` | Scope to a single monitored tag | all monitored |

Endpoints that are exempt from a filter document the exemption explicitly. The filter parameters behave identically across all endpoints that accept them — no endpoint interprets `period` differently.

### Independent endpoints

Each endpoint is self-contained. A consumer can call one endpoint without calling any other. There are no sequences, no pagination tokens that chain requests, and no required preflight calls. The configuration endpoint provides the tag/area structure a consumer needs to populate filter controls, but it is not a prerequisite for data requests.

### Additive evolution

New metrics are added as new endpoints. Existing endpoints are never modified in ways that break current consumers. If a response shape needs to change incompatibly, a new endpoint is introduced and the old one is deprecated through the versioning process.

---

## Versioning

### Path-based versioning

The API is served under `/api/v1/`. The version number is part of the URL path, not a header or query parameter. This makes the version visible in logs, bookmarks, and documentation without requiring inspection of request headers.

### Version lifecycle

| Phase | Meaning |
|-------|---------|
| **Active** | Current version. All new features land here. |
| **Deprecated** | Superseded by a newer version. Still functional but no new features. Deprecation is announced with a timeline for removal. |
| **Removed** | No longer served. Requests return 410 Gone. |

### What triggers a new version

A new major version (`v2`) is introduced only when an existing endpoint's response shape must change in a way that would break current consumers. Examples:

- Renaming or removing a response field.
- Changing a field's type (e.g., string to integer).
- Changing the semantics of an existing filter parameter.

The following do **not** require a new version:

- Adding a new endpoint.
- Adding a new optional field to an existing response.
- Adding a new optional query parameter.
- Changing internal computation logic that produces the same output shape.

### Transition period

When `v2` is introduced, `v1` continues to be served for a documented period. Both versions run simultaneously. Consumers migrate at their own pace within the transition window.

---

## Consumers

### Frontend dashboard

The primary consumer. Calls endpoints on page load and filter change. Expects fast responses (under 500ms for typical data volumes). The frontend performs no metric calculations — it renders API responses directly into charts, tables, and cards.

### MCP server

Wraps API endpoints as MCP tools. Each tool maps to one or a few endpoints. The MCP server adds natural-language descriptions and parameter mapping but does not transform the data — it passes API responses through to the AI assistant.

Example tool mapping:

| MCP tool | API endpoint(s) |
|----------|-----------------|
| "Get queue status" | `/api/v1/queue/summary` |
| "List stalled topics" | `/api/v1/queue/stalled` |
| "Get response metrics" | `/api/v1/metrics/summary` |
| "Check SLO compliance" | `/api/v1/slo/compliance` |

### Adding a new consumer

A new consumer should be able to start using the API by reading the [API contract spec](../specs/api/api-contract.md) alone. No out-of-band knowledge, SDK, or client library is required. The contract spec is the complete interface documentation.

---

## Error handling

All error responses use a consistent JSON structure:

```json
{ "error": "descriptive message" }
```

| Status | Meaning |
|--------|---------|
| 400 | Invalid request — bad filter values, malformed dates, unknown parameters |
| 404 | Unknown endpoint path |
| 500 | Server error — unexpected failure during processing |

Successful responses that produce empty results (e.g., no topics match the filter) return the normal response structure with zero counts or empty arrays — never 404.

---

## Relationship to other documents

| Document | Describes |
|----------|-----------|
| [ADR 0012](decisions/0012-api-responsibility-model.md) | Why this API model was chosen over alternatives |
| [API contract](../specs/api/api-contract.md) | Every endpoint: path, parameters, response shape, requirements |
| [API contract verification](../specs/api/api-contract_verification.md) | How the contract is tested |
| [ARCHITECTURE.md](../ARCHITECTURE.md) | Where the API fits in the system layers |
| This document | Overarching principles, versioning, conventions |
