# Stalled Topics — Dashboard View

This specification defines the requirements for UC-18: detecting open topics that have received at least one reply but have gone quiet without resolution. It traces to UC-18 in [use-cases.md](../use-cases.md).

This file defines *what* the user sees and why. [dashboard-components.md](dashboard-components.md) defines *how* each component behaves to fulfill these requirements.

---

## Use case traceability

| Requirement | Use case |
|-------------|----------|
| ST-1 – ST-9 | UC-18: Detect stalled topics |
| ST-10 – ST-13 | Cross-cutting: configuration, placement, empty states |

---

## Requirements

### Stalled topic detection (UC-18)

**ST-1.** The user sees a list of open topics that have received at least one reply but have had no activity for longer than a configured number of days, so that conversations at risk of being abandoned can be followed up.

**ST-2.** A topic is considered stalled when all of the following are true:

- It has received at least one reply (`replyCount >= 1`).
- It has no accepted answer (`resolvedAt` is absent).
- It does not carry the configured closed tag.
- The time since its last activity exceeds the configured stalled-days threshold.

**ST-3.** Last activity is represented by the `lastActivityAt` field on the topic. This is the timestamp of the most recent post or activity on the topic, supplied by the backend (Discourse API `last_posted_at` or `bumped_at`).

**ST-4.** The stalled-days threshold is defined per tag in the configuration file (`config/tagConfig.json`), under each tag's optional `stalledDays` field. Tags without an explicit value inherit from `defaults.stalledDays`. When a topic has multiple configured tags, the strictest (lowest) `stalledDays` determines whether it is stalled.

**ST-5.** The closed tag is defined per tag in the configuration file, under each tag's optional `closedTag` field. There is no default — absence means the tag does not participate in closed-tag exclusion. When a topic has multiple configured tags, the closed tags from all of them are checked. A topic carrying any of its tags' configured closed tags is considered closed (not stalled).

**ST-6.** The default sort order is by days since last activity, descending (highest first). The user can click any column header to toggle sort direction for that column. Only one column is sorted at a time; clicking a different column switches to that column in its default direction (Topic ascending, Tag ascending, Threshold ascending, Days inactive descending).

**ST-7.** Each row displays: topic title (linked to the Discourse forum topic), tag (first monitored tag on the topic, or "–" if none), threshold (the per-topic stalled-days value with a "(default)" indicator when inherited from `defaults.stalledDays`), and days since last activity (whole days, truncated).

**ST-7a.** The section heading reads "Stalled topics" without a specific threshold number, because the threshold varies per topic.

**ST-8.** The period filter (UC-12) applies. Only topics created within the selected time period are shown.

**ST-9.** The tag filter (UC-15) applies. When a tag is selected, only topics carrying that tag are shown. When no tag is selected, only topics carrying a monitored tag are shown.

### Configuration (ST-4, ST-5 detail)

**ST-10.** The tag configuration file uses the unified `config/tagConfig.json` structure. Each tag entry may include `closedTag` (string, optional, no default) and `stalledDays` (number, optional, falls back to `defaults.stalledDays`). See [tag-area-filter.md](tag-area-filter.md) for the full schema.

**ST-11.** The example configuration file (`config/tagConfig.example.json`) documents the unified structure including `closedTag` and `stalledDays` fields.

### Placement

**ST-12.** The stalled topics list appears on the Queue page, after summary cards and before the awaiting reply and untagged topics sections.

### Empty state

**ST-13.** When no topics meet the stalled criteria for the selected period and tag, an empty-state message is shown instead of the table.

---

## Design

### Tag configuration

The `TagConfig` type is the unified configuration structure defined in [tag-area-filter.md](tag-area-filter.md). `closedTag` and `stalledDays` are per-tag optional fields. The `resolveAllTags` function merges each tag entry with defaults and tracks provenance (`stalledDaysIsDefault`). The stalled detection logic uses resolved tags to determine per-topic thresholds.

### Data model

The `Topic` interface gains an optional field:

```typescript
lastActivityAt?: string; // ISO 8601 UTC — most recent activity on the topic
```

For topics with no replies, `lastActivityAt` may be absent or equal to `createdAt`. For stalled detection, only topics with `replyCount >= 1` are candidates, so `lastActivityAt` is expected to be present on those topics.

### Data source

`DashboardData` gains a new list:

```typescript
repliedOpenTopics: Topic[]; // open topics with replyCount >= 1
```

These are topics that have replies but no resolution and no closed tag. The stalled filter is applied on top of this list at display time.

### Stalled filter

The `filterStalledTopics` function identifies stalled topics from a pre-filtered list of replied open topics:

```typescript
function filterStalledTopics(
  topics: Topic[],
  resolved: Record<string, ResolvedTag>,
  now?: Date,
): Topic[];
```

- For each topic, collects `closedTag` values from its configured tags and excludes the topic if it carries any of them.
- Uses the strictest (lowest) `stalledDays` across the topic's configured tags.
- Excludes topics with no configured tags (threshold cannot be determined).
- Excludes topics whose `lastActivityAt` is within the threshold.
- Returns the remaining topics sorted by `lastActivityAt` ascending (oldest first).
- `now` defaults to `new Date()` and is injectable for testing.

### Days-since-last-activity calculation

```typescript
function daysSinceLastActivity(topic: Topic, now?: Date): number;
```

Returns the number of whole days (truncated) between `lastActivityAt` and `now`. If `lastActivityAt` is absent, falls back to `createdAt`.

### Per-topic threshold lookup

```typescript
function stalledThresholdForTopic(
  topic: Topic,
  resolved: Record<string, ResolvedTag>,
): { days: number; isDefault: boolean } | null;
```

Returns the strictest (lowest) `stalledDays` across the topic's configured tags and whether that value was inherited from defaults (`stalledDaysIsDefault`). Returns `null` when the topic has no configured tags.

### Container component

`StalledTopics` accepts three props:

| Prop | Type | Purpose |
|------|------|---------|
| `topics` | `Topic[]` | Filtered replied open topics (period + tag filters already applied) |
| `resolvedTags` | `Record<string, ResolvedTag>` | Resolved tag configuration with provenance tracking |
| `monitoredTags` | `string[]` | All monitored tags — used to display the first monitored tag per topic |

Calls `filterStalledTopics(topics, resolvedTags)` and renders:

- A section heading: "Stalled topics" (no threshold number, since thresholds vary per topic).
- A table with columns: Topic, Tag, Threshold, Days inactive.
- The Topic column shows the topic title as a clickable link to the Discourse forum topic.
- The Threshold column shows the per-topic stalled-days value via `stalledThresholdForTopic`. When the value is inherited from defaults, a "(default)" indicator is appended.
- All column headers are clickable to sort. The active sort column shows an arrow indicator (▲ ascending, ▼ descending). Clicking the same column toggles direction; clicking a different column switches to that column's default direction (Topic: ascending, Tag: ascending, Threshold: ascending, Days inactive: descending).
- Default sort: Days inactive descending (highest first).
- If `filterStalledTopics` returns an empty array, renders an empty-state paragraph ("No stalled topics") instead of the table.

CSS class prefix: `stalled-` for all elements specific to this component.

`StalledTopics` manages sort state via the shared `useTableSort` hook. All topic filtering is handled by `App` before passing props.

### Data flow

`App.tsx` computes the stalled topic source by taking `repliedOpenTopics` from `DashboardData`, applying both the period filter and the tag filter, then passing the result to `StalledTopics` along with resolved tag configuration. The `StalledTopics` component is rendered on the Queue page, below the existing sections.

### Component–requirement mapping

| Component | File | Requirements |
|-----------|------|-------------|
| `App` | App.tsx | ST-8 — period filter applied; ST-9 — tag filter applied; ST-12 — queue page placement |
| `StalledTopics` | StalledTopics.tsx | ST-1 — renders stalled list; ST-6 — sort order and interactive sorting on all columns; ST-7 — columns including threshold and topic links; ST-13 — empty state |
| `useTableSort` | useTableSort.ts | ST-6 — shared sort state management (column, direction, toggle, arrow indicator) |
| `stalledMetrics` | stalledMetrics.ts | ST-2 — stalled criteria; ST-3 — lastActivityAt usage; ST-4 — stalledDays threshold; ST-5 — closedTag exclusion |
| `tagFilter` | tagFilter.ts | ST-10 — TagConfig structure; resolveAllTags for per-tag resolution |
| Config file | tagConfig.json (tagConfig.example.json) | ST-10 — unified structure; ST-11 — example updated |

---

## Validation

### Automated tests

| What | Requirements | Rationale |
|------|-------------|-----------|
| `filterStalledTopics` — excludes topics with closed tag | ST-5 | Closed-tag exclusion correctness. |
| `filterStalledTopics` — excludes topics with recent activity within threshold | ST-2, ST-4 | Threshold boundary correctness. |
| `filterStalledTopics` — includes topics with activity older than threshold | ST-2, ST-4 | Threshold boundary correctness. |
| `filterStalledTopics` — boundary: exactly stalledDays ago is not stalled | ST-4 | Boundary precision. |
| `filterStalledTopics` — boundary: stalledDays + 1 day is stalled | ST-4 | Boundary precision. |
| `filterStalledTopics` — sorts by lastActivityAt ascending (oldest first) | ST-6 | Default data order from filter. |
| `filterStalledTopics` — empty input returns empty array | ST-13 | Empty state. |
| `filterStalledTopics` — does not mutate input array | ST-2 | Pure function contract. |
| `filterStalledTopics` — falls back to createdAt when lastActivityAt is absent | ST-3 | Fallback behavior. |
| `stalledThresholdForTopic` — returns strictest threshold and isDefault flag | ST-4, ST-7 | Per-topic threshold display. |
| `stalledThresholdForTopic` — returns null for topic with no configured tags | ST-4 | Edge case. |
| `daysSinceLastActivity` — returns whole days truncated | ST-7 | Display accuracy. |
| `daysSinceLastActivity` — uses lastActivityAt when present | ST-3 | Field priority. |
| `daysSinceLastActivity` — falls back to createdAt when lastActivityAt absent | ST-3 | Fallback behavior. |
| `formatStalledTag` — returns first monitored tag or "–" | ST-7 | Tag display. |

Test location: `tests/dashboard/stalled-topics.unit.test.ts`

### Manual verification

| What | Requirements | Rationale |
|------|-------------|-----------|
| Stalled topics section visible on Queue page | ST-12 | Layout concern. |
| Table renders with correct columns (Topic, Tag, Threshold, Days inactive) and topic links | ST-7 | Visual rendering concern. |
| Clicking any column header sorts by that column with arrow indicator | ST-6 | Interactive sort on all four columns. |
| Switching period filter updates the list | ST-8 | Cross-component interaction. |
| Selecting a tag scopes the list to that tag | ST-9 | Filter interaction. |
| Empty-state message shown when no stalled topics | ST-13 | Requires visual confirmation. |
| Closed-tag topics do not appear in the list | ST-5 | Requires mock data with closed tag. |
