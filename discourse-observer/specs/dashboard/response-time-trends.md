# Response Time Trends — Dashboard View

This specification defines the requirements for the weekly response time trend table: a view that shows median first reply time and median resolution time for each calendar week, allowing users to see whether response quality is improving or worsening over time. It traces to UC-8 in [use-cases.md](../use-cases.md).

This file defines *what* the user sees and why. [dashboard-components.md](dashboard-components.md) defines *how* each component behaves to fulfill these requirements.

---

## Use case traceability

| Requirement | Use case |
|-------------|----------|
| RT-1 – RT-9, RT-11 – RT-18 | UC-8: Track response time trends |
| RT-10 | Cross-cutting: duration display format (shared with RM-13) |

---

## Requirements

### Weekly trend table (UC-8)

**RT-1.** The user sees median first reply time and median resolution time for each calendar week that contains at least one resolved topic, across all historical data, so that response quality over time is visible.

**RT-2.** Each week bucket covers Monday 00:00:00 UTC through Sunday 23:59:59.999 UTC. A resolved topic belongs to the week in which its `createdAt` date falls.

**RT-3.** The table is ordered newest week first, so the most recent performance is visible without scrolling.

**RT-4.** Each row displays: the Monday date identifying the week, the count of resolved topics in that week, the median first reply time, and the median resolution time.

**RT-5.** Only weeks containing at least one resolved topic are shown. Consecutive weeks with no resolved topics are omitted.

**RT-6.** Median first reply time for a week excludes topics without `firstReplyAt`. When no topics in the week have `firstReplyAt`, the cell displays "–".

**RT-7.** Median resolution time for a week excludes topics without `resolvedAt`. When no topics in the week have `resolvedAt`, the cell displays "–".

**RT-8.** The trend table shows all resolved topics regardless of the active time period filter (UC-12). The trend always spans all available history.

**RT-9.** When there are no resolved topics at all, the trend table is replaced by an empty-state message.

### Duration formatting

**RT-10.** Duration values use the same format as RM-13: whole days (`"Xd"`) for durations of 24 hours or more, whole hours (`"Xh"`) for less than 24 hours, minimum `"1h"`.

### Trend charts (UC-8)

**RT-12.** A line chart displays median first reply time and median resolution time as two separate lines across all calendar weeks that contain at least one resolved topic, so that the user can see the trend visually without reading individual table rows.

**RT-13.** The chart X-axis shows weeks in chronological order (oldest left, newest right), labelled by the Monday date of each week. When there are more weeks than fit legibly, the axis may thin labels to avoid overlap.

**RT-14.** The chart Y-axis shows duration in hours. Values of 24 hours or more are displayed as whole days in axis labels and tooltips (e.g. "2d"), values below 24 hours as whole hours (e.g. "12h"), consistent with the RT-10 duration format.

**RT-15.** Hovering over a data point shows a tooltip with the week label, the formatted duration value, and the series name.

**RT-16.** The chart includes a legend identifying the two series ("Median first reply" and "Median resolution"). Clicking a legend entry toggles that series on or off.

**RT-17.** Weeks where a metric value is "–" (no qualifying topics) are represented as gaps in the line — not as zero values — so that missing data is visually distinct from fast response times.

**RT-18.** The chart appears between the section heading and the trend table, so that the visual overview comes before the detailed numbers.

### Placement

**RT-11.** The trend table appears on the response metrics page, below the summary cards (`ResponseMetricsCards`), so that aggregate metrics and weekly detail are co-located.

---

## Design

### Layout

The trend table appears on the response metrics page, below the summary cards (`ResponseMetricsCards`). It is introduced by a section heading "Weekly trends".

### Week label

Each week is identified by the ISO date (YYYY-MM-DD) of the Monday that begins it. The component formats this for display as a locale-aware short date.

### Topic count column

The topic count per week is shown so that rows with "–" metrics can be understood in context (e.g. a week with one topic but no resolved timestamps).

### Relationship to per-tag views (UC-9–11)

The core computation function (`computeWeeklyTrends`) accepts a `Topic[]` parameter. A caller wanting per-tag trends passes a pre-filtered list. No tag-specific logic lives inside the function. UC-9–11 are implemented in [tag-distribution.md](tag-distribution.md), which uses `mondayOf` (exported from `trendMetrics.ts`) for the weekly backlog trend table.

### Component–requirement mapping

| Component | File | Requirements |
|-----------|------|-------------|
| `App` | App.tsx | RT-8 — passes unfiltered `resolvedTopics` to trend component; RT-11 — renders trend section below summary cards |
| `ResponseTimeTrends` | ResponseTimeTrends.tsx | RT-3, RT-4, RT-5, RT-9 — renders the trend table |
| `trendMetrics` | trendMetrics.ts | RT-1, RT-2, RT-5, RT-6, RT-7, RT-10 — computes weekly buckets and metrics |
| `topicFormatting` | topicFormatting.ts | Week label formatting — `formatWeekLabel` (shared with TagDistribution) |

### Data flow

`App` passes `MOCK_DATA.resolvedTopics` (unfiltered) directly to `ResponseTimeTrends`. The component calls `computeWeeklyTrends` and renders the result. `computeWeeklyTrends` calls the existing `medianFirstReplyTime` and `medianResolutionTime` from `responseMetrics.ts`, reusing their "–" empty-state behaviour (RT-6, RT-7).

---

## Validation

### Automated tests

| What | Requirements | Rationale |
|------|-------------|-----------|
| `computeWeeklyTrends` — empty input returns empty array | RT-1, RT-5 | No topics → no rows. |
| `computeWeeklyTrends` — topics in the same week produce one row | RT-2, RT-4 | Week bucketing must not split within a week. |
| `computeWeeklyTrends` — topics in different weeks produce one row per week | RT-1, RT-4 | Each distinct week must appear. |
| `computeWeeklyTrends` — rows are ordered newest first | RT-3 | Ordering determines what the user sees first. |
| `computeWeeklyTrends` — `topicCount` matches the number of topics in the week | RT-4 | Wrong count would mislead the user about activity levels. |
| `computeWeeklyTrends` — week with no `firstReplyAt` shows "–" for first reply | RT-6 | Absence of reply data must not produce a spurious value. |
| `computeWeeklyTrends` — week with no `resolvedAt` shows "–" for resolution | RT-7 | Absence of resolution data must not produce a spurious value. |
| `computeWeeklyTrends` — Monday and Sunday of the same week land in the same bucket | RT-2 | Boundary correctness for the Monday–Sunday definition. |
| `computeWeeklyTrends` — Sunday and Monday on the week boundary land in different buckets | RT-2 | Ensures the Sunday/Monday split is correct. |
| `computeWeeklyTrends` — does not mutate the input array | RT-1 | Pure function contract. |

Test location: `tests/dashboard/response-time-trends.unit.test.ts`

### Manual verification

| What | Requirements | Rationale |
|------|-------------|-----------|
| Trend table appears below summary cards on the response metrics page | RT-11 | Layout concern. |
| Rows display readable week dates and non-overlapping columns | RT-4 | Formatting and CSS concern. |
| Empty-state message shown when no resolved topics exist | RT-9 | Requires visual confirmation. |
| Trend table remains unchanged when the time period filter is changed | RT-8 | Interaction concern not coverable by unit tests. |
