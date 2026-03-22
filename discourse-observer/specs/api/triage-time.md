# Triage Time

This specification defines the triage time metric: how long from topic creation until the first tag is set, computed from revision data captured by detail sync.

---

## Context

When a topic is created without tags, someone must triage it — read the topic and assign an appropriate tag. The time between creation and the first tag addition is a measurable indicator of triage responsiveness.

This metric uses `tag_change` events from the `topic_events` table (populated by detail sync). Topics created with tags already set have no triage event and are excluded.

---

## Requirements

Requirements use the prefix **TT** (Triage Time).

### Data

**TT-1.** A topic's triage time is the duration from `created_at` to the `happened_at` of its earliest `tag_change` event.

**TT-2.** Topics with no `tag_change` events are excluded from triage time computation. There is no triage event to measure.

**TT-3.** The "first tag" for a topic is the first tag that appears in the `current` field of its earliest `tag_change` event's detail (a JSON-encoded `RevisionTagChange` with `previous` and `current` string arrays).

### Computation

**TT-4.** The overall median is computed across all qualifying topics' triage durations.

**TT-5.** The by-tag breakdown groups topics by their first tag (TT-3) and computes the median triage duration within each group.

**TT-6.** The count for each group (overall and per-tag) is the number of qualifying topics in that group.

**TT-7.** Triage durations are reported in hours as floating-point values, not milliseconds. This matches the natural scale of the metric (hours to days, not sub-second precision).

### API

**TT-8.** `GET /api/v1/metrics/triage-time` returns triage time metrics.

Response:

```json
{
  "overall": {
    "medianHours": 4.2,
    "count": 142
  },
  "byTag": [
    { "tag": "authentication", "medianHours": 3.1, "count": 34 },
    { "tag": "installation", "medianHours": 6.8, "count": 22 }
  ]
}
```

- `overall.medianHours`: median triage duration in hours (TT-4, TT-7). Null when count is 0.
- `overall.count`: number of qualifying topics (TT-6).
- `byTag`: array of per-tag breakdowns (TT-5), sorted by count descending, then tag ascending.
- Each entry: `tag` (string), `medianHours` (float, nullable), `count` (integer).

**TT-9.** The endpoint supports `period`, `from`/`to`, and `tag` query parameters with the same semantics as existing endpoints (AC-8, AC-9, AC-10). Period and date filters apply to topic `created_at`. Tag filter scopes to topics whose first tag matches the filter value.

**TT-10.** When no qualifying topics exist for the given filters, the endpoint returns `overall.medianHours` as null and `overall.count` as 0, with an empty `byTag` array. It does not return 404 (consistent with AC-5).

### Domain

**TT-11.** Triage time computation is a pure function in `backend/domain/`. It takes topics and their events as input and returns computed results. It does not access the database or make HTTP calls.

**TT-12.** The domain function parses the `Detail` field of `tag_change` events as JSON (`RevisionTagChange`) to extract the first tag added.

### Frontend

**TT-13.** The dashboard displays overall median triage time as a summary card on the Response Metrics page.

---

## Verification

| Requirement | Test | What it verifies |
|-------------|------|-----------------|
| TT-1, TT-4, TT-6 | `TestTriageTimeMedian` | Correct median computation from revision timestamps |
| TT-3, TT-5 | `TestTriageTimeByTag` | Breakdown by first tag added |
| TT-2 | `TestTriageTimeNoEvents` | Topics without tag_change events are excluded |
| TT-8, TT-9, TT-10 | `TestTriageTimeEndpoint` | API response shape and filter support |
| TT-12 | `TestTriageTimeDetailParsing` | Detail JSON is parsed correctly to extract first tag |
