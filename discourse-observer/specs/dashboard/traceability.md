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
| UC-4: Measure time to first reply | [response-metrics.md](response-metrics.md) | RM-1, RM-2 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (RM-1, RM-2, RM-12) |
| UC-5: Measure time to resolution | [response-metrics.md](response-metrics.md) | RM-3, RM-4 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (RM-3, RM-4, RM-12) |
| UC-6: Compare solved versus self-closed | [response-metrics.md](response-metrics.md) | RM-5 – RM-7 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (RM-5, RM-6, RM-7) |
| UC-7: Measure answer rate | [response-metrics.md](response-metrics.md) | RM-8, RM-9 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (RM-8, RM-9) |
| (cross-cutting: format) | [response-metrics.md](response-metrics.md) | RM-13 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (RM-13) |
| (cross-cutting: median) | [response-metrics.md](response-metrics.md) | RM-14 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (RM-14) |
| (cross-cutting: nav/empty) | [response-metrics.md](response-metrics.md) | RM-10 – RM-12 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (RM-12); manual (RM-10, RM-11) |

---

## Response time trends

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-8: Track response time trends | [response-time-trends.md](response-time-trends.md) | RT-1 – RT-18 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (RT-1 – RT-7, RT-10, RT-14, RT-17); manual (RT-8, RT-9, RT-11 – RT-13, RT-15, RT-16, RT-18) |

---

## Time period filter

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-12: Filter by time period | [time-period-filter.md](time-period-filter.md) | TF-1 – TF-14 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (TF-3 – TF-8, TF-13); manual (TF-1, TF-2, TF-9, TF-10, TF-11, TF-12, TF-14) |

---

## Distribution and bottlenecks

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-9: Identify highest-volume tag areas | [tag-distribution.md](tag-distribution.md) | TD-1 – TD-5 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (TD-1 – TD-3); manual (TD-4, TD-5) |
| UC-10: Identify slowest tag areas | [tag-distribution.md](tag-distribution.md) | TD-6 – TD-11 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (TD-6 – TD-10); manual (TD-11) |
| UC-11: Detect accumulating backlogs | [tag-distribution.md](tag-distribution.md) | TD-12 – TD-24 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (TD-12 – TD-14, TD-16 – TD-23); manual (TD-15, TD-24) |
| (cross-cutting) | [tag-distribution.md](tag-distribution.md) | TD-25 – TD-27 | manual (TD-25, TD-26); [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (TD-27 via RM-13) |

---

## SLO monitoring

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-13: Flag topics exceeding SLO thresholds | [slo-monitoring.md](slo-monitoring.md) | SL-1 – SL-12 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (SL-2 – SL-4, SL-6, SL-9, SL-11); manual (SL-1, SL-5, SL-7, SL-8, SL-10, SL-12) |
| UC-14: Evaluate SLO compliance | [slo-monitoring.md](slo-monitoring.md) | SL-13 – SL-20, SL-18a | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (SL-14, SL-15, SL-17, SL-19); [tag-area-filter.unit.test.ts](../../tests/dashboard/tag-area-filter.unit.test.ts) (SL-18a via scopeSloConfig); manual (SL-13, SL-16, SL-18, SL-20) |
| (cross-cutting: config, nav, empty) | [slo-monitoring.md](slo-monitoring.md) | SL-21 – SL-25 | manual (SL-21 – SL-25) |

---

## Tag and area filter

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-15: Filter dashboard by tag | [tag-area-filter.md](tag-area-filter.md) | TA-1 – TA-8 | [tag-area-filter.unit.test.ts](../../tests/dashboard/tag-area-filter.unit.test.ts) (TA-2, TA-4, TA-5, TA-6, TA-17); manual (TA-1, TA-3, TA-7, TA-8) |
| UC-16: Navigate tags by area | [tag-area-filter.md](tag-area-filter.md) | TA-9 – TA-14 | [tag-area-filter.unit.test.ts](../../tests/dashboard/tag-area-filter.unit.test.ts) (TA-12, TA-13); manual (TA-9, TA-10, TA-11, TA-14) |
| (cross-cutting: config, placement, defaults) | [tag-area-filter.md](tag-area-filter.md) | TA-15 – TA-22 | [tag-area-filter.unit.test.ts](../../tests/dashboard/tag-area-filter.unit.test.ts) (TA-17); manual (TA-15, TA-16, TA-18, TA-19, TA-20, TA-21, TA-22) |

---

## Topic intake

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-17: Track topic intake over time | [topic-intake.md](topic-intake.md) | TI-1 – TI-11, TI-8a | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (TI-1, TI-3, TI-4, TI-8a, TI-9); manual (TI-2, TI-5, TI-6, TI-7, TI-8, TI-10, TI-11) |
| (cross-cutting: granularity, placement, empty) | [topic-intake.md](topic-intake.md) | TI-12 – TI-14 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (TI-12, TI-14); manual (TI-13) |

---

## Stalled topics

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-18: Detect stalled topics | [stalled-topics.md](stalled-topics.md) | ST-1 – ST-9 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (ST-2 – ST-7, ST-13); manual (ST-1, ST-8, ST-9) |
| (cross-cutting: config, placement, empty) | [stalled-topics.md](stalled-topics.md) | ST-10 – ST-13 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (ST-3, ST-13); manual (ST-10, ST-11, ST-12) |

---

## Peak activity

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-19: Identify peak activity periods | [peak-activity.md](peak-activity.md) | PA-1 – PA-9 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (PA-2, PA-3, PA-5, PA-6, PA-7, PA-13); manual (PA-1, PA-4, PA-9) |
| (cross-cutting: placement, filters, empty) | [peak-activity.md](peak-activity.md) | PA-10 – PA-13 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (PA-13); manual (PA-10, PA-11, PA-12) |
| (timezone headers: ADR 0010) | [peak-activity.md](peak-activity.md) | PA-14 – PA-20 | [timezone-utils.unit.test.ts](../../tests/dashboard/timezone-utils.unit.test.ts) (PA-17, PA-18, PA-19); manual (PA-14, PA-15, PA-16, PA-20) |
| (timezone picker) | [peak-activity.md](peak-activity.md) | PA-21 – PA-22 | [timezone-utils.unit.test.ts](../../tests/dashboard/timezone-utils.unit.test.ts) (PA-22 via data integrity); manual (PA-21, PA-22) |
| (cookie consent) | [peak-activity.md](peak-activity.md) | PA-23 – PA-25 | [timezone-utils.unit.test.ts](../../tests/dashboard/timezone-utils.unit.test.ts) (PA-25); manual (PA-23, PA-24) |

---

## Response time distribution

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-20: Understand response time spread | [response-time-distribution.md](response-time-distribution.md) | RD-1 – RD-10 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (RD-2 – RD-6); manual (RD-1, RD-7, RD-8, RD-9, RD-10) |
| (cross-cutting: placement, filters, empty, config) | [response-time-distribution.md](response-time-distribution.md) | RD-11 – RD-14 | [contract_test.go, domain/*_test.go](../../backend/api/contract_test.go) (RD-14); manual (RD-11, RD-12, RD-13) |

---

## URL state synchronization

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-24: Persist filter state in URL | [url-state.md](url-state.md) | US-1 – US-12 | [url-state.unit.test.ts](../../tests/dashboard/url-state.unit.test.ts) (US-2 – US-8); manual (US-1, US-9, US-10, US-11, US-12) |

---

## Gaps

- UC-3 partial: untagged share as percentage of all topics (deferred until total topic count is available from backend).
- UC-11 partial: per-tag weekly backlog trend deferred — current implementation shows aggregate weekly trend only. Per-tag breakdown requires a tag selector and is deferred until need is demonstrated.
