# Response Time Distribution — Dashboard View

This specification defines the requirements for UC-20: understanding how response times are distributed to identify whether the median hides a long tail of slow responses. It traces to UC-20 in [use-cases.md](../use-cases.md).

This file defines *what* the user sees and why. [dashboard-components.md](dashboard-components.md) defines *how* each component behaves to fulfill these requirements.

---

## Use case traceability

| Requirement | Use case |
|-------------|----------|
| RD-1 – RD-10 | UC-20: Understand response time spread |
| RD-11 – RD-14 | Cross-cutting: placement, filters, empty states, configuration |

---

## Requirements

### Distribution histograms (UC-20)

**RD-1.** The user sees two histograms — one for time-to-first-reply and one for time-to-resolution — so that the spread of response times is visible beyond the median.

**RD-2.** Each histogram groups topics into time brackets (buckets). The x-axis shows bucket labels. The y-axis shows the count of topics in each bucket. This makes it immediately clear how many topics fall into each speed range.

**RD-3.** Bucket boundaries are configurable via `config/distributionBuckets.json`. The file contains an array of upper ceilings in hours. Each ceiling defines the upper bound of a bucket. The final bucket captures all values above the last ceiling. This allows the team to adjust the granularity without code changes.

**RD-4.** Bucket labels are derived from the ceilings. The first bucket is labeled "< Xh" (or "< Xd" when X is a multiple of 24) where X is the first ceiling. Middle buckets are labeled "X–Y" using the previous ceiling as the lower bound and the current ceiling as the upper bound. The final bucket is labeled "> X" using the last ceiling. Hours are displayed as "Xh" and multiples of 24 hours as "Xd".

**RD-5.** For the first-reply histogram, the duration is `firstReplyAt − createdAt`. Topics without `firstReplyAt` are excluded. This matches the existing median first reply metric.

**RD-6.** For the resolution histogram, the duration is `resolvedAt − createdAt`. Topics without `resolvedAt` are excluded. This matches the existing median resolution metric.

**RD-7.** The histograms are rendered as Recharts bar charts, consistent with the topic intake chart style. Each histogram uses a distinct bar color to visually distinguish the two series.

**RD-8.** Each histogram includes a tooltip showing the bucket label and exact count on hover.

**RD-9.** The first-reply histogram uses the color `#8884d8` and the resolution histogram uses the color `#82ca9d`. These match the colors used in the response time trend chart for visual consistency.

**RD-10.** Each histogram has a heading — "First reply distribution" and "Resolution time distribution" — above the chart.

### Placement

**RD-11.** The distribution histograms appear on the Response Metrics page, below the existing response metrics cards and trend chart. UC-20 adds depth to the same metrics already shown on this page.

### Filters

**RD-12.** The period filter (UC-12) applies. Only resolved topics created within the selected time period are counted in the histograms.

**RD-13.** The tag filter (UC-15) applies. When a tag is selected, only topics carrying that tag are counted. When no tag is selected, only topics carrying a monitored tag are counted.

### Empty state

**RD-14.** When there are no qualifying topics for a histogram (no topics with `firstReplyAt` for the first histogram, or no topics with `resolvedAt` for the second), an empty-state message is shown instead of that chart.

---

## Design

### Configuration

`config/distributionBuckets.json`:

```json
{
  "bucketCeilingsHours": [1, 4, 12, 24, 48, 96, 168]
}
```

This produces 8 buckets: < 1h, 1–4h, 4–12h, 12h–1d, 1–2d, 2–4d, 4–7d, > 7d.

The configuration is imported in `App.tsx` and passed as a prop to the distribution component, following the same pattern as `sloThresholds.json`.

### Bucket label formatting

```typescript
function formatBucketCeiling(hours: number): string;
```

Returns `"Xd"` when hours is a multiple of 24 (e.g., 24 → "1d", 168 → "7d"), otherwise `"Xh"` (e.g., 4 → "4h").

### Bucketing function

```typescript
interface DistributionBucket {
  label: string;
  count: number;
}

function bucketDurations(
  durationsMs: number[],
  ceilingsHours: number[],
): DistributionBucket[];
```

- Converts each ceiling to milliseconds.
- Creates N+1 buckets (where N is the number of ceilings).
- For each duration, finds the first ceiling it falls below and increments that bucket.
- Returns all buckets including those with count 0.

### Duration extraction helpers

```typescript
function firstReplyDurations(topics: Topic[]): number[];
function resolutionDurations(topics: Topic[]): number[];
```

These extract the raw duration arrays from filtered topics, excluding topics that lack the relevant timestamp.

### Container component

`ResponseTimeDistribution` accepts two props:

| Prop | Type | Purpose |
|------|------|---------|
| `topics` | `Topic[]` | Filtered resolved topics (period + tag filters already applied) |
| `ceilingsHours` | `number[]` | Bucket ceilings from configuration |

Calls `firstReplyDurations` and `resolutionDurations`, then `bucketDurations` for each, and renders:

- A section for each histogram with its heading.
- A `DistributionChart` for each non-empty result.
- An empty-state paragraph when the duration array is empty.

CSS class prefix: `rd-` for all elements specific to this component.

`ResponseTimeDistribution` is a pure function component. It holds no state — all filtering is handled by `App` before passing props.

### Chart component

`DistributionChart` accepts three props:

| Prop | Type | Purpose |
|------|------|---------|
| `data` | `DistributionBucket[]` | Bucketed data to display |
| `color` | `string` | Bar fill color |
| `name` | `string` | Series name for tooltip |

Renders a Recharts `BarChart` inside a `ResponsiveContainer` (width 100%, height 300px).

- `XAxis` with `dataKey="label"` showing bucket labels.
- `YAxis` with `allowDecimals={false}` for whole numbers.
- `Tooltip` showing bucket label and count.
- No legend (single series per chart).

CSS class prefix: `rd-chart-` for chart-specific elements.

### Data flow

`App.tsx` passes `filteredData.resolvedTopics` (already period- and tag-filtered) and the bucket ceilings to `ResponseTimeDistribution`. The component handles duration extraction and bucketing internally.

### Component–requirement mapping

| Component | File | Requirements |
|-----------|------|-------------|
| `App` | App.tsx | RD-11 — response metrics page; RD-12 — period filter; RD-13 — tag filter |
| `ResponseTimeDistribution` | ResponseTimeDistribution.tsx | RD-1 — two histograms; RD-5 — first reply durations; RD-6 — resolution durations; RD-10 — headings; RD-14 — empty state |
| `DistributionChart` | DistributionChart.tsx | RD-7 — bar chart; RD-8 — tooltip; RD-9 — colors |
| `distributionMetrics` | distributionMetrics.ts | RD-2 — bucketing; RD-3 — configurable ceilings; RD-4 — labels; RD-5 — first reply extraction; RD-6 — resolution extraction |

---

## Validation

### Automated tests

| What | Requirements | Rationale |
|------|-------------|-----------|
| `formatBucketCeiling` — hours < 24 returns "Xh" | RD-4 | Label formatting for hours. |
| `formatBucketCeiling` — multiples of 24 returns "Xd" | RD-4 | Label formatting for days. |
| `bucketDurations` — empty input returns all-zero buckets | RD-2, RD-14 | Empty state produces full bucket structure. |
| `bucketDurations` — durations land in correct buckets | RD-2, RD-3 | Core bucketing correctness. |
| `bucketDurations` — duration exceeding all ceilings lands in last bucket | RD-2 | Overflow bucket. |
| `bucketDurations` — duration exactly equal to a ceiling lands in the next bucket (not the one below the ceiling) | RD-2 | Boundary precision — buckets use strict less-than for upper bounds. |
| `bucketDurations` — bucket labels match expected format | RD-4 | Label generation correctness. |
| `bucketDurations` — single ceiling produces two buckets | RD-2 | Minimum configuration. |
| `firstReplyDurations` — excludes topics without firstReplyAt | RD-5 | Exclusion correctness. |
| `firstReplyDurations` — computes correct durations in ms | RD-5 | Duration computation. |
| `resolutionDurations` — excludes topics without resolvedAt | RD-6 | Exclusion correctness. |
| `resolutionDurations` — computes correct durations in ms | RD-6 | Duration computation. |
| `bucketDurations` — does not mutate input array | RD-2 | Pure function contract. |

Test location: `tests/dashboard/response-time-distribution.unit.test.ts`

### Manual verification

| What | Requirements | Rationale |
|------|-------------|-----------|
| Two histograms render on the Response Metrics page below trends | RD-1, RD-11 | Layout concern. |
| First reply histogram uses color #8884d8 | RD-9 | Visual rendering concern. |
| Resolution histogram uses color #82ca9d | RD-9 | Visual rendering concern. |
| Tooltip shows bucket label and count on hover | RD-8 | Interactive rendering concern. |
| Switching period filter updates histograms | RD-12 | Cross-component interaction. |
| Selecting a tag scopes histograms to that tag | RD-13 | Filter interaction. |
| Empty-state message shown when no qualifying topics | RD-14 | Requires visual confirmation. |
| Changing distributionBuckets.json ceilings changes histogram buckets | RD-3 | Configuration concern. |
