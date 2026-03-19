# Topic Intake — Dashboard View

This specification defines the requirements for UC-17: tracking how many new support topics are created per time period. It traces to UC-17 in [use-cases.md](../use-cases.md).

This file defines *what* the user sees and why. [dashboard-components.md](dashboard-components.md) defines *how* each component behaves to fulfill these requirements.

---

## Use case traceability

| Requirement | Use case |
|-------------|----------|
| TI-1 – TI-11 | UC-17: Track topic intake over time |
| TI-12 – TI-14 | Cross-cutting: granularity, placement, empty states |

---

## Requirements

### Intake volume (UC-17)

**TI-1.** The user sees the number of new support topics created per time bucket for the selected time period (UC-12), so that demand volume and patterns are visible.

**TI-2.** A time bucket is either one day or one week. The granularity is chosen automatically based on the active period: daily for periods spanning fewer than 90 days, weekly for 90 days or more.

**TI-3.** Daily buckets cover a single UTC calendar day (00:00:00 UTC through 23:59:59.999 UTC). A topic belongs to the day in which its `createdAt` falls.

**TI-4.** Weekly buckets cover Monday 00:00:00 UTC through Sunday 23:59:59.999 UTC — the same definition used by response time trends (RT-2). A topic belongs to the week in which its `createdAt` falls.

**TI-5.** The chart shows only time buckets that fall within the selected period. Unlike response time trends (RT-8), intake volume respects the time period filter because the user wants to see demand for the chosen window, not all history.

**TI-6.** When a tag is selected (UC-15), the chart shows intake for that tag only. When no tag is selected, the chart shows aggregate intake across all monitored tags.

**TI-7.** The data source for intake is all support topics regardless of outcome status (unreplied, resolved, and replied-open), since intake measures when topics are created, not their current state.

**TI-8.** The chart is a line chart. A continuous line communicates volume trends — rises and falls over time — more clearly than discrete bars, especially when zero-count periods are included.

**TI-9.** The chart X-axis shows time buckets in chronological order (oldest left, newest right), labelled by date. When there are more buckets than fit legibly, the axis may thin labels to avoid overlap.

**TI-10.** The chart Y-axis shows topic count as whole numbers starting from zero.

**TI-11.** Hovering over a data point shows a tooltip with the bucket label (date or week) and the topic count.

**TI-8a.** All time periods between the earliest and latest bucket are included in the chart, even if no topics were created in that period. Empty periods show a count of zero. This ensures the x-axis is continuous and accurately represents time gaps.

### Granularity (TI-2 detail)

**TI-12.** The 90-day threshold is evaluated as follows:

- Preset "Last 7 days" → daily.
- Preset "Last 30 days" → daily.
- Preset "Last year" → weekly.
- Preset "All time" → weekly.
- Custom range → daily if the span is fewer than 90 days, weekly otherwise.

### Placement

**TI-13.** The intake chart appears on its own dashboard page titled "Volume", accessible via a navigation link alongside the existing pages (Queue, Response metrics, Distribution, SLO).

### Empty state

**TI-14.** When there are no topics in the selected period, an empty-state message is shown instead of the chart.

---

## Design

### Layout

The Volume page contains the intake line chart as a full-width section. A section heading "Topic intake" introduces the chart.

### Granularity selection

The `intakeGranularity` function determines whether the active period produces daily or weekly buckets:

```typescript
type IntakeGranularity = "daily" | "weekly";

function intakeGranularity(period: ActivePeriod): IntakeGranularity;
```

For presets, the mapping is hardcoded (TI-12). For custom ranges, the function computes the span in days between `from` and `to` and compares against the 90-day threshold.

### Bucketing

The `computeIntakeBuckets` function groups topics into time buckets and returns chart-ready data:

```typescript
interface IntakeBucket {
  label: string;     // formatted date label for X-axis
  count: number;     // topic count in this bucket
  bucketKey: string; // YYYY-MM-DD (day) or Monday YYYY-MM-DD (week) — for sorting
}

interface TimeRange {
  first: string; // YYYY-MM-DD bucket key
  last: string;  // YYYY-MM-DD bucket key
}

function computeTimeRange(
  topics: Topic[],
  granularity: IntakeGranularity,
): TimeRange | null;

function computeIntakeBuckets(
  topics: Topic[],
  granularity: IntakeGranularity,
  range: TimeRange | null,
): IntakeBucket[];
```

`computeTimeRange` finds the earliest and latest bucket keys across the given topics. It is called with all monitored-tag topics (before tag filtering) so the x-axis stays consistent when switching between individual tags.

`computeIntakeBuckets`:

- Groups topics by `createdAt` into the appropriate bucket (daily or weekly) using a shared `bucketKey` helper.
- Returns buckets in chronological order (oldest first).
- All periods between `range.first` and `range.last` are included (TI-8a). Missing periods are filled with count zero. Returns empty when `range` is `null`.
- For weekly buckets, reuses `mondayOf` from `trendMetrics.ts`.
- For daily buckets, a `dayOf` function extracts the UTC date as `YYYY-MM-DD`.
- Labels are formatted using `toLocaleDateString` with `{ year: "numeric", month: "short", day: "numeric", timeZone: "UTC" }` — consistent with week labels in `ResponseTimeTrends`.

### Chart component

`IntakeChart` renders a Recharts `LineChart` inside a `ResponsiveContainer` (width 100%, height 300px).

- A single `Line` series for topic count, colored `#5b8ff9`, with `type="monotone"` and small dots (radius 3) so individual data points remain visible.
- `XAxis` with `dataKey="label"` showing bucket date labels.
- `YAxis` with `allowDecimals={false}` to show whole numbers only.
- `Tooltip` showing the bucket label and count.
- No legend needed (single series).

CSS class prefix: `intake-chart-` for chart-specific elements.

### Container component

`TopicIntake` accepts `topics: Topic[]`, `granularity: IntakeGranularity`, and `timeRange: TimeRange | null`. It calls `computeIntakeBuckets(topics, granularity, timeRange)` and renders:

- A section heading "Topic intake".
- The `IntakeChart` with the bucket data.
- If `computeIntakeBuckets` returns an empty array, renders an empty-state paragraph ("No data") instead of the chart.

CSS class prefix: `intake-` for section-level elements.

### Data flow

`App.tsx` computes the intake topic list by combining unreplied and resolved topics, applying both the period filter and the tag filter. It computes `intakeGranularity(activePeriod)` and `computeTimeRange` from all monitored-tag period-filtered topics (before tag filtering), then passes the filtered topics, granularity, and time range to `TopicIntake`.

The Volume page is added as a new navigation option in `App`. The `Page` type is extended with `"volume"`.

### Component–requirement mapping

| Component | File | Requirements |
|-----------|------|-------------|
| `App` | App.tsx | TI-5 — period filter applied; TI-6 — tag filter applied; TI-7 — combines unreplied and resolved; TI-13 — volume page navigation |
| `TopicIntake` | TopicIntake.tsx | TI-1 — renders intake view; TI-14 — empty state |
| `IntakeChart` | IntakeChart.tsx | TI-8 — line chart; TI-9 — X-axis chronological; TI-10 — Y-axis whole numbers; TI-11 — tooltip |
| `intakeMetrics` | intakeMetrics.ts | TI-2 — granularity selection; TI-3 — daily bucketing; TI-4 — weekly bucketing; TI-8a — gap filling; TI-12 — threshold rules |

---

## Validation

### Automated tests

| What | Requirements | Rationale |
|------|-------------|-----------|
| `intakeGranularity` — "last7" returns "daily" | TI-12 | Granularity mapping correctness. |
| `intakeGranularity` — "last30" returns "daily" | TI-12 | Granularity mapping correctness. |
| `intakeGranularity` — "lastYear" returns "weekly" | TI-12 | Granularity mapping correctness. |
| `intakeGranularity` — "allTime" returns "weekly" | TI-12 | Granularity mapping correctness. |
| `intakeGranularity` — custom range under 90 days returns "daily" | TI-12 | Threshold boundary. |
| `intakeGranularity` — custom range of exactly 90 days returns "weekly" | TI-12 | Boundary correctness. |
| `intakeGranularity` — custom range over 90 days returns "weekly" | TI-12 | Threshold boundary. |
| `computeIntakeBuckets` — daily: topics on same day produce one bucket | TI-3 | Daily bucketing correctness. |
| `computeIntakeBuckets` — daily: topics on different days produce one bucket per day | TI-3 | Multi-bucket correctness. |
| `computeIntakeBuckets` — daily: bucket count matches topic count per day | TI-1 | Count accuracy. |
| `computeIntakeBuckets` — daily: buckets are in chronological order | TI-9 | Chart X-axis ordering. |
| `computeIntakeBuckets` — weekly: topics in same week produce one bucket | TI-4 | Weekly bucketing correctness. |
| `computeIntakeBuckets` — weekly: Monday and Sunday of same week land in same bucket | TI-4 | Boundary correctness (same as RT-2). |
| `computeIntakeBuckets` — weekly: Sunday and Monday on week boundary land in different buckets | TI-4 | Week-split correctness. |
| `computeIntakeBuckets` — weekly: buckets are in chronological order | TI-9 | Chart X-axis ordering. |
| `computeIntakeBuckets` — daily: fills gaps between days with zero-count buckets | TI-8a | Continuous x-axis. |
| `computeIntakeBuckets` — weekly: fills gaps between weeks with zero-count buckets | TI-8a | Continuous x-axis. |
| `computeIntakeBuckets` — empty input returns empty array | TI-14 | Empty state. |
| `computeIntakeBuckets` — does not mutate input array | TI-1 | Pure function contract. |

Test location: `backend/api/contract_test.go`, `backend/domain/*_test.go`

### Manual verification

| What | Requirements | Rationale |
|------|-------------|-----------|
| Volume page is accessible via navigation link | TI-13 | Layout concern. |
| Line chart renders with correct shape | TI-8, TI-10 | Visual rendering concern. |
| Hovering a data point shows tooltip with date and count | TI-11 | Interaction concern. |
| Switching period filter updates the chart | TI-5 | Cross-component interaction. |
| Selecting a tag scopes the chart to that tag | TI-6 | Filter interaction. |
| Granularity switches from daily to weekly for "Last year" | TI-2, TI-12 | Visual granularity change. |
| Empty-state message shown when no topics in period | TI-14 | Requires visual confirmation. |
