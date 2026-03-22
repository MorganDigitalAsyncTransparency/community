# Tag Flows

This specification defines the tag flows metric: how topics move through the tag landscape over their lifetime, computed from revision data captured by detail sync.

---

## Context

Understanding how tags change over a topic's lifecycle reveals workflow patterns — which tags are added first, which are added later, which are removed, and which tags tend to co-occur. This data helps identify classification bottlenecks, routing inefficiencies, and tag usage patterns.

This metric uses `tag_change` events from the `topic_events` table (populated by detail sync).

---

## Requirements

Requirements use the prefix **TF** (Tag Flows).

### Transitions

**TF-1.** A transition is a change from one tag set to another on a single topic, derived from consecutive `tag_change` events.

**TF-2.** The "from" state of the first `tag_change` event is the `previous` array from its detail JSON. The "to" state is the `current` array.

**TF-3.** For topics with multiple `tag_change` events, each event produces a transition: the "from" is the previous event's `current` (or the first event's `previous` for the first transition).

**TF-4.** `from: []` means the topic was untagged before the change. `to: []` means all tags were removed.

**TF-5.** Transitions are aggregated: identical from→to pairs are grouped, with a count and median duration the topic spent in the "from" state before transitioning.

**TF-6.** Duration in "from" state is measured from the previous event's timestamp (or topic creation for the first event) to the current event's timestamp.

### Tag pairs

**TF-7.** Tag pairs are tags that end up on the same topic via revision — specifically, tags that co-exist in any `tag_change` event's `current` array where both tags were not already present together in the `previous` array.

**TF-8.** Tag pairs are reported with a count of how many topics exhibit that co-occurrence.

### Summary

**TF-9.** `topics_with_tag_changes`: number of topics that have at least one `tag_change` event.

**TF-10.** `total_topics`: total number of topics in the filtered set.

**TF-11.** `median_changes_per_topic`: median number of `tag_change` events per topic, computed only across topics that have at least one change.

**TF-12.** `most_common_first_tag`: the tag that appears most frequently as the first new tag in the earliest `tag_change` event across all topics. Null if no topics have tag changes.

**TF-13.** `most_unstable_tag`: the tag with the highest average number of add+remove operations per topic. A tag is "added" when it appears in `current` but not `previous`; "removed" when it appears in `previous` but not `current`. Null if no topics have tag changes.

### API

**TF-14.** `GET /api/v1/metrics/tag-flows` returns tag flow metrics.

Response:

```json
{
  "transitions": [
    {
      "from": [],
      "to": ["authentication"],
      "count": 87,
      "medianDurationHours": 4.2
    }
  ],
  "tagPairs": [
    { "tags": ["authentication", "sso"], "count": 23 }
  ],
  "summary": {
    "topicsWithTagChanges": 142,
    "totalTopics": 5012,
    "medianChangesPerTopic": 1.3,
    "mostCommonFirstTag": "authentication",
    "mostUnstableTag": "api"
  }
}
```

- `transitions`: array sorted by count descending (TF-5).
- `tagPairs`: array sorted by count descending, then tags alphabetically (TF-8). Tags within each pair are sorted alphabetically.
- `summary`: object with fields TF-9 through TF-13.
- `medianChangesPerTopic` is a float. Null fields use null when no data exists.

**TF-15.** The endpoint supports `period`, `from`/`to`, and `tag` query parameters with the same semantics as existing endpoints. Period and date filters apply to topic `created_at`. Tag filter scopes to topics carrying the specified tag.

**TF-16.** When no qualifying topics exist, the endpoint returns empty arrays for transitions and tagPairs, and zeroed/null summary fields. It does not return 404 (consistent with AC-5).

### Domain

**TF-17.** Tag flow computation is a set of pure functions in `backend/domain/`. They take topics and their events as input and return computed results. They do not access the database or make HTTP calls.

### Frontend

**TF-18.** The dashboard includes a **Tag Flows** page accessible from the sidebar navigation.

**TF-19.** The page has four summary cards at the top: topics with tag changes (as ratio), median changes per topic, most common first tag, most unstable tag.

**TF-20.** The page has a transitions table: From | To | Count | Median time in "from" state.

**TF-21.** The page has a tag pairs section showing top co-occurring tag pairs with counts.

---

## Verification

| Requirement | Test | What it verifies |
|-------------|------|-----------------|
| TF-1, TF-2, TF-3 | `TestTagTransitions` | Correct from→to transitions from revisions |
| TF-5, TF-6 | `TestTagTransitionMedianDuration` | Median time in "from" state |
| TF-7, TF-8 | `TestTagPairs` | Co-occurring tags via revision |
| TF-9–TF-13 | `TestTagFlowsSummary` | Summary stats computed correctly |
| TF-14, TF-15, TF-16 | `TestTagFlowsEndpoint` | API response shape and filter support |
| TF-16 | `TestTagFlowsEmptyDataset` | Empty data returns empty arrays, not errors |
