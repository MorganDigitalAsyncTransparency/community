# Tag Distribution — Dashboard View

This specification defines the requirements for the tag distribution page: a view that helps users understand how support volume, resolution speed, and open backlog are distributed across monitored tags. It traces to UC-9, UC-10, and UC-11 in [use-cases.md](../use-cases.md).

This file defines *what* the user sees and why. [dashboard-components.md](dashboard-components.md) defines *how* each component behaves to fulfill these requirements.

---

## Use case traceability

| Requirement | Use case |
|-------------|----------|
| TD-1 – TD-5 | UC-9: Identify highest-volume tag areas |
| TD-6 – TD-11 | UC-10: Identify slowest tag areas |
| TD-12 – TD-24 | UC-11: Detect accumulating backlogs |
| TD-25 – TD-27 | Cross-cutting: navigation, period filter, duration format |

---

## Requirements

### Topics by tag (UC-9)

**TD-1.** The user sees all tags ranked by total topic count, highest first, so that demand concentration across tag areas is immediately visible.

**TD-2.** Topic count includes both open (unreplied) and resolved topics. A topic whose `tags` array contains multiple entries is counted independently toward each tag's total.

**TD-3.** The active time period filter applies by `createdAt`. Only topics created within the selected period are counted.

**TD-4.** Each row shows the tag name and its topic count.

**TD-5.** When no tagged topics exist in the selected period, the section shows an empty-state message instead of a table.

---

### Resolution time by tag (UC-10)

**TD-6.** The user sees all tags ranked by median resolution time, slowest first, so that capacity bottlenecks and expertise gaps are identifiable.

**TD-7.** Only topics that have a `resolvedAt` timestamp contribute to the median resolution time for a tag. Topics without `resolvedAt` are excluded from the median calculation.

**TD-8.** Tags where no topics have a `resolvedAt` show "–" for median resolution time and are sorted to the bottom of the table, after all tags with a numeric median.

**TD-9.** The active time period filter applies by `createdAt`.

**TD-10.** Each row shows the tag name, count of resolved topics (those with `resolvedAt`) for that tag, and median resolution time.

**TD-11.** When no resolved topics exist in the selected period, the section shows an empty-state message instead of a table.

---

### Open backlogs by tag — current snapshot (UC-11)

**TD-12.** The user sees all tags that have at least one currently open (unreplied) topic, ranked by open topic count descending, so that tag areas with the largest current backlog are visible at a glance.

**TD-13.** The active time period filter applies by `createdAt`. Only unreplied topics created within the selected period are counted.

**TD-14.** Each row shows the tag name and open topic count.

**TD-15.** When no unreplied topics exist in the selected period, the section shows an empty-state message instead of a table.

---

### Open backlogs by tag — weekly trend (UC-11)

**TD-16.** The user sees a weekly trend table showing, per calendar week, how many topics were created, how many were resolved, and how many remain open (still in the unreplied queue). This allows the user to determine whether the backlog is accumulating (still-open count is high in recent weeks) or under control (still-open count is near zero).

**TD-17.** Week boundaries follow the same Monday 00:00:00 UTC – Sunday 23:59:59.999 UTC convention as RT-2. A topic belongs to the week in which its `createdAt` falls.

**TD-18.** "Created" for a week is the count of all topics (unreplied and resolved) whose `createdAt` falls in that week.

**TD-19.** "Resolved" for a week is the count of resolved topics whose `createdAt` falls in that week.

**TD-20.** "Still open" for a week is the count of unreplied topics whose `createdAt` falls in that week.

**TD-21.** The table is ordered newest week first, so the most recent backlog state is visible without scrolling.

**TD-22.** Only weeks containing at least one topic are shown. Weeks with no topics are omitted.

**TD-23.** The weekly backlog trend spans all available history regardless of the active time period filter. The trend always reflects the full dataset so that long-running open topics remain visible even when the period selector is set to a narrow window.

**TD-24.** When there are no topics at all, the weekly trend table shows an empty-state message instead of a table.

---

### Cross-cutting

**TD-25.** A third navigation item "Distribution" is added to the header navigation, alongside "Queue" and "Response metrics". The active page is indicated visually.

**TD-26.** The Distribution page shares the same header, sync timestamp, and period selector as the other pages. The active period persists when switching between pages (per TF-11).

**TD-27.** Duration values in the resolution time column (TD-10) use the same format as RM-13: whole days (`"Xd"`) for durations of 24 hours or more, whole hours (`"Xh"`) for less than 24 hours, minimum `"1h"`.

---

## Design

### Layout

The Distribution page shares the application shell (header, sync timestamp, navigation, period selector) with the other pages. Below the period selector, it renders three sections in order:

1. **"Topics by tag"** — tag volume ranking table (TD-1 – TD-5).
2. **"Resolution time by tag"** — tag resolution time ranking table (TD-6 – TD-11).
3. **"Open backlogs by tag"** — per-tag open count table (TD-12 – TD-15), followed by the weekly backlog trend table (TD-16 – TD-24).

### Period filter scope

The period selector applies to the volume ranking, the resolution time ranking, and the per-tag open count snapshot. The weekly backlog trend table is exempt (TD-23) — it always spans all history, consistent with how the response time trends table (RT-8) behaves.

### Gaps

- **UC-11 per-tag trend:** The "growing over time" dimension of UC-11 is addressed at the aggregate level by the weekly backlog trend table. Per-tag weekly trend breakdowns are not included; they would require a tag selector and a more complex table structure, and are deferred until there is evidence of need.

---

## Component–requirement mapping

| Component | File | Requirements |
|-----------|------|-------------|
| `App` | App.tsx | TD-25 — nav item; TD-26 — shared header and period selector |
| `TagDistribution` | TagDistribution.tsx | TD-1, TD-4, TD-5 — volume table; TD-6, TD-8, TD-10, TD-11 — resolution table; TD-12, TD-14, TD-15 — backlog snapshot table; TD-16, TD-21, TD-22, TD-24 — weekly backlog trend table |
| `tagMetrics` | tagMetrics.ts | TD-1, TD-2, TD-3 — tagVolumeRanking; TD-6, TD-7, TD-8, TD-9 — tagResolutionRanking; TD-12, TD-13 — tagBacklogRanking; TD-17 – TD-23 — computeWeeklyBacklog |
| `topicFormatting` | topicFormatting.ts | TD-27 — formatDuration (shared with RM-13) |

---

## Validation

### Automated tests

| What | Requirements | Rationale |
|------|-------------|-----------|
| `tagVolumeRanking` — empty input returns empty array | TD-1, TD-5 | No topics → no rows. |
| `tagVolumeRanking` — each tag in a multi-tag topic is counted independently | TD-2 | Incorrect multi-tag handling would misrepresent demand per tag area. |
| `tagVolumeRanking` — topics without tags do not appear | TD-2 | Untagged topics have no tag area to rank. |
| `tagVolumeRanking` — rows are ordered by topic count descending | TD-1 | Wrong order would bury the highest-demand tags. |
| `tagVolumeRanking` — does not mutate input | TD-1 | Pure function contract. |
| `tagResolutionRanking` — topics without `resolvedAt` excluded from median | TD-7 | Including unresolved topics would distort resolution time rankings. |
| `tagResolutionRanking` — tag with no `resolvedAt` across its topics shows "–" | TD-8 | Absence of data must not produce a spurious value. |
| `tagResolutionRanking` — tags with "–" sort after tags with a numeric median | TD-8 | Tags with incomplete data must not displace genuinely slow tags. |
| `tagResolutionRanking` — rows with a numeric median are ordered slowest first | TD-6 | Wrong order would hide the most problematic tag areas. |
| `tagResolutionRanking` — resolvedCount reflects only topics with `resolvedAt` | TD-10 | Wrong count would mislead users about how much data backs the median. |
| `tagBacklogRanking` — empty input returns empty array | TD-12, TD-15 | No open topics → no rows. |
| `tagBacklogRanking` — rows are ordered by open count descending | TD-12 | Wrong order would hide the largest backlogs. |
| `computeWeeklyBacklog` — empty input returns empty array | TD-16, TD-24 | No topics → no rows. |
| `computeWeeklyBacklog` — topics in the same week produce one row | TD-17, TD-18 | Week bucketing must not split within a week. |
| `computeWeeklyBacklog` — created = unreplied + resolved in that week | TD-18 | Wrong count would misrepresent weekly activity. |
| `computeWeeklyBacklog` — stillOpen = unreplied topics in that week | TD-20 | Wrong still-open count is the core metric; incorrect value breaks the UC-11 goal. |
| `computeWeeklyBacklog` — resolved = resolved topics in that week | TD-19 | Wrong resolved count would distort the backlog picture. |
| `computeWeeklyBacklog` — rows are ordered newest first | TD-21 | Wrong order would bury recent backlog state. |
| `computeWeeklyBacklog` — only weeks with at least one topic are shown | TD-22 | Empty weeks must not produce phantom rows. |
| `computeWeeklyBacklog` — does not mutate either input array | TD-17 | Pure function contract. |

Test location: `tests/dashboard/tag-distribution.unit.test.ts`

### Manual verification

| What | Requirements | Rationale |
|------|-------------|-----------|
| "Distribution" nav item is visible and switches to the distribution page | TD-25 | Interaction and layout concern. |
| Period selector affects all three tables but not the weekly backlog trend | TD-26, TD-23 | Interaction concern not coverable by unit tests. |
| Empty-state messages appear when selections produce no data | TD-5, TD-11, TD-15, TD-24 | Requires visual confirmation. |
| Three sections are visible in order on the distribution page | (layout) | Layout concern. |
