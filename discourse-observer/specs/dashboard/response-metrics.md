# Response Metrics — Dashboard View

This specification defines the requirements for the dashboard's response metrics page: a view that helps users understand response times and outcomes for resolved support topics. It traces to the response time and outcome use cases defined in [use-cases.md](../use-cases.md).

This file defines *what* the user sees and why. [dashboard-components.md](dashboard-components.md) defines *how* each component behaves to fulfill these requirements.

---

## Use case traceability

| Requirement | Use case |
|-------------|----------|
| RM-1 – RM-2 | UC-4: Measure time to first reply |
| RM-3 – RM-4 | UC-5: Measure time to resolution |
| RM-5 – RM-7 | UC-6: Compare solved versus self-closed topics |
| RM-8 – RM-9 | UC-7: Measure answer rate |
| RM-13 | Cross-cutting: duration display format |
| RM-14 | Cross-cutting: median calculation method |
| RM-10 – RM-12 | Cross-cutting: navigation, empty states, data freshness |

---

## Requirements

### Time to first reply (UC-4)

**RM-1.** The user sees the median time from topic creation to first reply, calculated across all resolved topics that have a first reply, so that team responsiveness is visible at a glance.

**RM-2.** Topics without a first reply (self-closed without any reply) are excluded from the median first reply calculation, because they do not represent a response event.

### Time to resolution (UC-5)

**RM-3.** The user sees the median time from topic creation to resolution, calculated across all resolved topics, so that end-to-end handling speed is visible.

**RM-4.** Both solved and self-closed topics are included in the resolution time calculation, because both outcomes represent a topic leaving the open queue. Topics without a `resolvedAt` timestamp are excluded, as no resolution duration can be calculated.

### Solved versus self-closed (UC-6)

**RM-5.** The user sees the count of solved topics and the count of self-closed topics, so that the balance between real answers and unanswered closures is visible.

**RM-6.** The user sees the ratio of solved to self-closed topics expressed as a textual summary (e.g. "12 solved / 5 self-closed"), so that the comparison is immediate without mental arithmetic.

**RM-7.** When there are no resolved topics, the solved/self-closed display shows "0 solved / 0 self-closed".

### Answer rate (UC-7)

**RM-8.** The user sees the percentage of resolved topics that were solved (not self-closed), so that overall support quality is quantified.

**RM-9.** When there are no resolved topics, the answer rate displays "–" instead of a percentage, because a percentage of zero is misleading.

### Duration formatting

**RM-13.** Time durations (median first reply, median resolution) are displayed as whole days (`"Xd"`) when the duration is 24 hours or more, and as whole hours (`"Xh"`) when less than 24 hours. The minimum displayed value is `"1h"`. This matches the age format used in the queue visibility view (QV-3).

### Median calculation

**RM-14.** When computing a median across an odd number of values, the result is the middle value. When computing across an even number of values, the result is the average of the two middle values, truncated to a whole number of milliseconds.

### Navigation, empty states, and data freshness

**RM-10.** The dashboard provides navigation between the queue visibility page and the response metrics page, so that users can switch views without reloading.

**RM-11.** The response metrics page shares the same header and sync timestamp as the queue visibility page (reusing QV-17, QV-18 from [queue-visibility.md](queue-visibility.md)).

**RM-12.** When there are no resolved topics with a first reply, the median first reply metric displays "–". When there are no resolved topics at all, the median resolution time displays "–".

---

## Design

### Layout

The response metrics page shares the application shell (header, sync timestamp, navigation) with the queue visibility page. Below the navigation, it renders four summary cards:

1. **"Median first reply"** — median time to first reply (RM-1, RM-2, RM-12).
2. **"Median resolution"** — median time to resolution (RM-3, RM-4, RM-12).
3. **"Outcomes"** — solved and self-closed counts (RM-5, RM-6, RM-7).
4. **"Answer rate"** — percentage solved (RM-8, RM-9).

### Navigation

The application header includes two navigation links: "Queue" and "Response metrics". The active page is indicated visually. Navigation switches the page content without a full page reload (RM-10). Page state is persisted in the URL via query parameters (see [url-state.md](url-state.md)), so views can be bookmarked and shared.

### Component–requirement mapping

| Component | File | Requirements |
|-----------|------|-------------|
| `App` | App.tsx | RM-10 — navigation; RM-11 — shared header |
| `ResponseMetricsCards` | ResponseMetricsCards.tsx | RM-1, RM-2, RM-12 — median first reply; RM-3, RM-4, RM-12 — median resolution; RM-5, RM-6, RM-7 — outcomes; RM-8, RM-9 — answer rate |
| `responseMetrics` | responseMetrics.ts | RM-1, RM-2 — medianFirstReplyTime; RM-3, RM-4 — medianResolutionTime; RM-5 — outcomeCounts; RM-6, RM-7 — formatOutcomes; RM-8, RM-9 — answerRate; RM-14 — median |
| `topicFormatting` | topicFormatting.ts | RM-13 — formatDuration (shared with QV-3) |

### Data flow

All components receive data from the same `DashboardData` object, using the `resolvedTopics` array. The computation functions operate on `Topic[]` and return formatted strings suitable for display.

### Page composition

The Response Metrics page is built from multiple sections that each cover a distinct perspective on the same data:

1. **This file** — summary cards: median first reply, median resolution, outcomes, answer rate (UC-4, UC-5, UC-6, UC-7).
2. **Topic volume** — four-line chart showing topics created, accepted answer, topics closed, and currently open per time bucket (UC-17, merged from the former Volume page).
3. **Median first reply** — line chart showing median first reply time per time bucket.
4. **Median first resolution** — line chart showing median resolution time per time bucket.
5. **[response-time-distribution.md](response-time-distribution.md)** — histogram showing how response times are spread across time brackets (UC-20).

Volume and median trend charts use the same daily/weekly bucketing as the former Topic Intake (TI-2) and respect the period filter (UC-12) and tag filter (UC-15).

Per-tag breakdown of metrics (UC-9, UC-10) is on the Distribution page — see [tag-distribution.md](tag-distribution.md).

---

## Validation

### Automated tests

| What | Requirements | Rationale |
|------|-------------|-----------|
| `medianFirstReplyTime` — returns formatted median for topics with `firstReplyAt`, excludes topics without | RM-1, RM-2, RM-12 | Pure function. Incorrect median would misrepresent team responsiveness. |
| `medianResolutionTime` — returns formatted median for all resolved topics | RM-3, RM-4, RM-12 | Pure function. Incorrect median would misrepresent handling speed. |
| `outcomeCounts` — returns correct solved and self-closed counts | RM-5 | Pure function. Wrong counts would distort the solved/self-closed comparison. |
| `formatOutcomes` — formats counts as `"X solved / Y self-closed"`, including zero case | RM-6, RM-7 | Pure function. Wrong format would misrepresent the outcome summary. |
| `answerRate` — returns percentage of solved topics, dash for empty input | RM-8, RM-9 | Pure function. Wrong percentage would misrepresent support quality. |
| `formatDuration` — returns `"Xd"` for ≥ 24 h, `"Xh"` for < 24 h, minimum `"1h"` | RM-13 | Pure function with boundary conditions. Wrong format would misrepresent time durations. |
| `median` — returns middle value for odd count, truncated average of two middle values for even count | RM-14 | Pure function. Wrong median would corrupt all time-based metrics. |

Test location: `backend/api/contract_test.go`, `backend/domain/*_test.go`

### Manual verification

| What | Requirements | Rationale |
|------|-------------|-----------|
| Navigation switches between pages without reload | RM-10 | Interaction and layout concern. |
| Response metrics cards are visible without scrolling | RM-11 | Depends on viewport and CSS. |
| Sync timestamp appears on the response metrics page | RM-11 | Layout concern, shared with QV-17/QV-18. |
| Empty-state displays show dash where specified | RM-7, RM-9 | Absence of visual artifacts best confirmed visually. |
