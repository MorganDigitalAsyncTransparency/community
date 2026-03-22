# Escalation Patterns

This specification defines the escalation patterns metric: topics where tags changed after the first reply, indicating the initial classification was wrong or the topic needed re-routing.

---

## Context

When a topic's tags change after someone has already replied, it suggests the initial triage was incorrect — the topic needed re-classification or escalation to a different area. Tracking these patterns reveals classification accuracy and routing effectiveness.

This metric uses `tag_change` events from the `topic_events` table and `first_reply_at` from the `topics` table.

---

## Requirements

Requirements use the prefix **EP** (Escalation Patterns).

### Detection

**EP-1.** A topic is an "escalation" if it has at least one `tag_change` event with `happened_at` after the topic's `first_reply_at`.

**EP-2.** Topics without a `first_reply_at` (no replies) are excluded — there is no reply to classify against.

**EP-3.** Topics with tag changes only before or at the time of the first reply are not escalations.

### Computation

**EP-4.** `total`: number of topics matching EP-1.

**EP-5.** `rate`: fraction of all replied topics (topics with `first_reply_at`) that are escalations. Float between 0 and 1.

**EP-6.** `by_period`: weekly trend of escalation counts. Each entry has a period label (ISO week, e.g., `"2026-W10"`) and a count of escalations where the tag change `happened_at` falls in that week.

**EP-7.** `common_patterns`: the most frequent before/after tag patterns among escalations. "Before" is the tag set from `previous` in the tag change detail; "after" is the tags added (present in `current` but not in `previous`). Sorted by count descending.

### API

**EP-8.** `GET /api/v1/metrics/escalations` returns escalation pattern metrics.

Response:

```json
{
  "total": 48,
  "rate": 0.12,
  "byPeriod": [
    { "period": "2026-W10", "count": 8 },
    { "period": "2026-W11", "count": 12 }
  ],
  "commonPatterns": [
    {
      "originalTags": ["api"],
      "addedAfterReply": ["authentication"],
      "count": 7
    }
  ]
}
```

- `total`: integer (EP-4).
- `rate`: float, null when no replied topics exist (EP-5).
- `byPeriod`: array sorted chronologically (EP-6). Empty when no escalations.
- `commonPatterns`: array sorted by count descending (EP-7). Empty when no escalations.

**EP-9.** The endpoint supports `period`, `from`/`to`, and `tag` query parameters with the same semantics as existing endpoints.

**EP-10.** When no qualifying topics exist, the endpoint returns `total: 0`, `rate: null`, and empty arrays. It does not return 404 (consistent with AC-5).

### Domain

**EP-11.** Escalation computation is a pure function in `backend/domain/`. It takes topics and their events as input and returns computed results. No database or HTTP access.

### Frontend

**EP-12.** The dashboard displays the escalation rate as a summary card on the Response Metrics page, showing the rate as a percentage and the total count.

---

## Verification

| Requirement | Test | What it verifies |
|-------------|------|-----------------|
| EP-1, EP-2, EP-3 | `TestEscalationDetection` | Tag change after first_reply_at flagged |
| EP-4, EP-5 | `TestEscalationRate` | Correct fraction of replied topics |
| EP-6 | `TestEscalationByPeriod` | Weekly trend bucketing |
| EP-7 | `TestEscalationPatterns` | Common before/after patterns |
| EP-8, EP-9, EP-10 | `TestEscalationEndpoint` | API response shape and filter support |
| EP-10 | `TestNoEscalationsWhenNoRevisions` | Zero escalations with no data |
