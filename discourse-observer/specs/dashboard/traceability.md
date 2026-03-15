# Dashboard — Traceability Matrix

This matrix shows how use cases decompose into specifications, requirements, and verification artifacts for the dashboard module.

---

## Queue visibility

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-1: Identify topics waiting longest for a reply | [queue-visibility.md](queue-visibility.md) | QV-1 – QV-4 | [queue-visibility.unit.test.ts](../../tests/dashboard/queue-visibility.unit.test.ts) (QV-1, QV-3, QV-4); manual (QV-2) |
| UC-2: See all unreplied support topics | [queue-visibility.md](queue-visibility.md) | QV-5 – QV-9 | [queue-visibility.unit.test.ts](../../tests/dashboard/queue-visibility.unit.test.ts) (QV-6, QV-7); manual (QV-5, QV-8, QV-9) |
| UC-3: Detect untagged topics | [queue-visibility.md](queue-visibility.md) | QV-10 – QV-14 | [queue-visibility.unit.test.ts](../../tests/dashboard/queue-visibility.unit.test.ts) (QV-11, QV-13); manual (QV-10, QV-12, QV-14) |
| (cross-cutting) | [queue-visibility.md](queue-visibility.md) | QV-15 – QV-18 | manual (QV-15 – QV-18) |

### Component behavior

| Spec | Requirements | Verification |
|------|-------------|--------------|
| [dashboard-components.md](dashboard-components.md) | All | [queue-visibility.unit.test.ts](../../tests/dashboard/queue-visibility.unit.test.ts) (shared logic); manual (rendering) |

---

## Response metrics

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-4: Measure time to first reply | [response-metrics.md](response-metrics.md) | RM-1, RM-2 | [response-metrics.unit.test.ts](../../tests/dashboard/response-metrics.unit.test.ts) (RM-1, RM-2, RM-12) |
| UC-5: Measure time to resolution | [response-metrics.md](response-metrics.md) | RM-3, RM-4 | [response-metrics.unit.test.ts](../../tests/dashboard/response-metrics.unit.test.ts) (RM-3, RM-4, RM-12) |
| UC-6: Compare solved versus self-closed | [response-metrics.md](response-metrics.md) | RM-5 – RM-7 | [response-metrics.unit.test.ts](../../tests/dashboard/response-metrics.unit.test.ts) (RM-5, RM-6, RM-7) |
| UC-7: Measure answer rate | [response-metrics.md](response-metrics.md) | RM-8, RM-9 | [response-metrics.unit.test.ts](../../tests/dashboard/response-metrics.unit.test.ts) (RM-8, RM-9) |
| (cross-cutting: format) | [response-metrics.md](response-metrics.md) | RM-13 | [response-metrics.unit.test.ts](../../tests/dashboard/response-metrics.unit.test.ts) (RM-13) |
| (cross-cutting: nav/empty) | [response-metrics.md](response-metrics.md) | RM-10 – RM-12 | [response-metrics.unit.test.ts](../../tests/dashboard/response-metrics.unit.test.ts) (RM-12); manual (RM-10, RM-11) |

---

## Gaps

- UC-2 partial: time period filtering (deferred to UC-12 implementation).
- UC-3 partial: untagged share as percentage of all topics (deferred until total topic count is available from backend).
- Visual design: no spec exists for the frontend's visual language (colors, typography, spacing, layout system). CSS implementation exists without a corresponding specification. Tracked as a separate work package.
