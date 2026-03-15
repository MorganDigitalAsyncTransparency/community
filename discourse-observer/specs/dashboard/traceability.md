# Dashboard — Traceability Matrix

This matrix shows how use cases decompose into specifications, requirements, and verification artifacts for the dashboard module.

---

## Queue visibility

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-1: Identify topics waiting longest for a reply | [queue-visibility.md](queue-visibility.md) | QV-1 – QV-4 | — |
| UC-2: See all unreplied support topics | [queue-visibility.md](queue-visibility.md) | QV-5 – QV-9 | — |
| UC-3: Detect untagged topics | [queue-visibility.md](queue-visibility.md) | QV-10 – QV-14 | — |
| (cross-cutting) | [queue-visibility.md](queue-visibility.md) | QV-15 – QV-18 | — |

### Component behavior

| Spec | Requirements | Verification |
|------|-------------|--------------|
| [dashboard-components.md](dashboard-components.md) | All | — |

---

## Gaps

- No verification artifacts exist yet for any dashboard spec. This is expected — the specs were written retroactively for an existing prototype. Verification will be added in a subsequent PR.
- UC-2 partial: time period filtering (deferred to UC-12 implementation).
- UC-3 partial: untagged share as percentage of all topics (deferred until total topic count is available from backend).
