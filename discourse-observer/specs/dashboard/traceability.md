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
| (cross-cutting: median) | [response-metrics.md](response-metrics.md) | RM-14 | [response-metrics.unit.test.ts](../../tests/dashboard/response-metrics.unit.test.ts) (RM-14) |
| (cross-cutting: nav/empty) | [response-metrics.md](response-metrics.md) | RM-10 – RM-12 | [response-metrics.unit.test.ts](../../tests/dashboard/response-metrics.unit.test.ts) (RM-12); manual (RM-10, RM-11) |

---

## Response time trends

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-8: Track response time trends | [response-time-trends.md](response-time-trends.md) | RT-1 – RT-18 | [response-time-trends.unit.test.ts](../../tests/dashboard/response-time-trends.unit.test.ts) (RT-1 – RT-7, RT-10, RT-14, RT-17); manual (RT-8, RT-9, RT-11 – RT-13, RT-15, RT-16, RT-18) |

---

## Time period filter

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-12: Filter by time period | [time-period-filter.md](time-period-filter.md) | TF-1 – TF-14 | [time-period-filter.unit.test.ts](../../tests/dashboard/time-period-filter.unit.test.ts) (TF-3 – TF-8, TF-13); manual (TF-1, TF-2, TF-9, TF-10, TF-11, TF-12, TF-14) |

---

## Distribution and bottlenecks

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-9: Identify highest-volume tag areas | [tag-distribution.md](tag-distribution.md) | TD-1 – TD-5 | [tag-distribution.unit.test.ts](../../tests/dashboard/tag-distribution.unit.test.ts) (TD-1 – TD-3); manual (TD-4, TD-5) |
| UC-10: Identify slowest tag areas | [tag-distribution.md](tag-distribution.md) | TD-6 – TD-11 | [tag-distribution.unit.test.ts](../../tests/dashboard/tag-distribution.unit.test.ts) (TD-6 – TD-10); manual (TD-11) |
| UC-11: Detect accumulating backlogs | [tag-distribution.md](tag-distribution.md) | TD-12 – TD-24 | [tag-distribution.unit.test.ts](../../tests/dashboard/tag-distribution.unit.test.ts) (TD-12 – TD-14, TD-16 – TD-23); manual (TD-15, TD-24) |
| (cross-cutting) | [tag-distribution.md](tag-distribution.md) | TD-25 – TD-27 | manual (TD-25, TD-26); [response-metrics.unit.test.ts](../../tests/dashboard/response-metrics.unit.test.ts) (TD-27 via RM-13) |

---

## Gaps

- UC-3 partial: untagged share as percentage of all topics (deferred until total topic count is available from backend).
- UC-11 partial: per-tag weekly backlog trend deferred — current implementation shows aggregate weekly trend only. Per-tag breakdown requires a tag selector and is deferred until need is demonstrated.
- Visual design: no spec exists for the frontend's visual language (colors, typography, spacing, layout system). CSS implementation exists without a corresponding specification. Tracked as a separate work package.
