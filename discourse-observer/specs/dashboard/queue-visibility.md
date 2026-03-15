# Queue Visibility — Dashboard View

This specification defines the requirements for the dashboard's first iteration: a queue visibility view that helps users understand the current state of unreplied and untagged topics. It traces to the queue visibility use cases defined in [use-cases.md](../use-cases.md).

This file defines *what* the user sees and why. [dashboard-components.md](dashboard-components.md) defines *how* each component behaves to fulfill these requirements.

---

## Use case traceability

| Requirement | Use case |
|-------------|----------|
| QV-1 – QV-4 | UC-1: Identify topics waiting longest for a reply |
| QV-5 – QV-9 | UC-2: See all unreplied support topics |
| QV-10 – QV-14 | UC-3: Detect untagged topics |
| QV-15 – QV-18 | Cross-cutting: edge cases and data freshness |

---

## Requirements

### Unreplied topics — identifying the most neglected (UC-1)

**QV-1.** The user sees a list of support topics that have received no reply, sorted oldest first, so that the most neglected topics appear at the top.

**QV-2.** Each entry in the unreplied list shows the topic title, the topic's age, and the associated tag(s).

**QV-3.** Topic age is displayed as a relative duration: whole days (e.g. "14d") when the topic is 24 hours or older, whole hours (e.g. "8h") when younger than 24 hours. The minimum displayed value is 1h.

**QV-4.** When a topic has multiple tags, all tags are shown separated by commas. When a topic has no tags, a dash ("–") is displayed instead.

### Unreplied topics — queue size overview (UC-2)

**QV-5.** The user sees a count of all unreplied support topics, providing an at-a-glance understanding of the current queue size.

**QV-6.** The user sees the age of the oldest unreplied topic, displayed as whole days (e.g. "14d"), so that the worst-case wait time is immediately visible.

**QV-7.** When there are no unreplied topics, the oldest-age indicator shows a dash ("–") instead of a number.

**QV-8.** The unreplied count and oldest-age indicator are visible without scrolling, separate from the detailed list.

**QV-9.** The detailed unreplied list shows every unreplied support topic — not a truncated subset.

### Untagged topics (UC-3)

**QV-10.** The user sees a count of topics that have no tag at all, so that the scale of untagged topics is immediately visible.

**QV-11.** The user sees a list of individual untagged topics, sorted oldest first, so that long-standing tagging gaps are surfaced first.

**QV-12.** Each entry in the untagged list shows the topic title, the topic's age, and the topic's category.

**QV-13.** Topic age in the untagged list uses the same relative duration format as the unreplied list (QV-3).

**QV-14.** The untagged count is visible without scrolling, separate from the detailed list.

### Edge cases and data freshness

**QV-15.** When the unreplied list is empty, the table area shows no rows — there is no error state or placeholder message required beyond the count showing zero and the oldest-age indicator showing a dash.

**QV-16.** When the untagged list is empty, the table area shows no rows — there is no error state or placeholder message required beyond the count showing zero.

**QV-17.** The dashboard displays the timestamp of the most recent data sync, so that the user knows how current the information is.

**QV-18.** The sync timestamp is formatted as a localized date and time (date and time of day, not a relative duration).

---

## Design

This section describes how the requirements above are fulfilled by the current component structure. Component behavior is specified in detail in [dashboard-components.md](dashboard-components.md).

### Layout

The dashboard view ([App.tsx](../../frontend/src/App.tsx)) composes three areas:

1. **Header** — application title and sync timestamp (QV-17, QV-18).
2. **Summary cards** — at-a-glance counts and oldest-age indicator, always visible at the top of the content area (QV-5, QV-6, QV-7, QV-8, QV-10, QV-14).
3. **Detail tables** — full lists of unreplied and untagged topics (QV-1, QV-2, QV-9, QV-11, QV-12).

### Component–requirement mapping

| Component | File | Requirements |
|-----------|------|-------------|
| `App` | [App.tsx](../../frontend/src/App.tsx) | QV-17, QV-18 — renders sync timestamp in the header |
| `SummaryCards` | [SummaryCards.tsx](../../frontend/src/components/SummaryCards.tsx) | QV-5 — unreplied count; QV-6, QV-7 — oldest unreplied age; QV-8 — above-the-fold placement; QV-10, QV-14 — untagged count |
| `UnrepliedTable` | [UnrepliedTable.tsx](../../frontend/src/components/UnrepliedTable.tsx) | QV-1 — oldest-first sort; QV-2 — title, age, tags per row; QV-3 — relative age format; QV-4 — tag display; QV-9 — complete list; QV-15 — empty state |
| `UntaggedTable` | [UntaggedTable.tsx](../../frontend/src/components/UntaggedTable.tsx) | QV-11 — oldest-first sort; QV-12 — title, age, category per row; QV-13 — shared age format; QV-16 — empty state |
| `topicFormatting` | [topicFormatting.ts](../../frontend/src/components/topicFormatting.ts) | QV-3, QV-13 — age formatting; QV-4 — tag display; QV-6, QV-7 — oldest unreplied age |

### Data flow

All components receive data from a single `DashboardData` object. In the current prototype this object is populated from mock data ([mock/data.ts](../../frontend/src/mock/data.ts)). When the backend API is available, only the data source changes — the component interfaces remain the same.

### What is not covered

This specification covers the first iteration of queue visibility only. The following are explicitly out of scope:

- Time period filtering (UC-2 mentions "filterable by time period" — deferred to a future iteration that addresses UC-12).
- Untagged share as a percentage of all topics (UC-3 mentions "the share they represent" — deferred until total topic count is available from the backend).
- Linking topic titles to the Discourse forum.
- Sorting controls or column reordering.

---

## Validation

This section defines how the requirements above are verified, and why each item is tested automatically or manually.

### Automated tests

Pure logic that produces deterministic, observable output — these are the highest-value automated tests because a regression would silently corrupt what users see.

| What | Requirements | Rationale |
|------|-------------|-----------|
| `formatAge` — returns `"Xd"` for ≥ 24 h, `"Xh"` for < 24 h, minimum 1 h | QV-3, QV-13 | Pure function with well-defined boundary conditions. A formatting error would mislead users about topic urgency. |
| `sortedByOldest` — returns topics in ascending `createdAt` order | QV-1, QV-11 | Pure function. Incorrect sort order would hide the most neglected topics. |
| `oldestUnrepliedDays` — returns `"Xd"` for non-empty lists, `"–"` for empty lists | QV-6, QV-7 | Pure function with an edge case (empty list). A wrong value in the summary card would give a false sense of queue health. |
| `formatTags` — joins tags with comma, returns `"–"` for empty array | QV-4 | Pure function. Incorrect output would hide tag information or display confusing placeholder text. |

Test location: `tests/dashboard/queue-visibility.unit.test.ts`

### Manual verification

Visual and layout concerns that depend on CSS rendering and browser behavior — these cannot be meaningfully asserted in unit tests.

| What | Requirements | Rationale |
|------|-------------|-----------|
| Summary cards are visible without scrolling | QV-8, QV-14 | Depends on viewport size, CSS layout, and surrounding content height. No unit-testable assertion captures "above the fold". |
| Table columns show correct headers and alignment | QV-2, QV-12 | Column order and header text are in JSX, but visual alignment and readability depend on CSS. |
| Sync timestamp appears in the header area | QV-17, QV-18 | Placement is a layout concern. The timestamp formatting itself uses `toLocaleString`, which varies by browser locale. |
| Empty table shows no rows and no error state | QV-15, QV-16 | The absence of visual artifacts (no placeholder, no broken layout) is best confirmed visually. |

Manual verification is performed by loading the dashboard with mock data and inspecting the rendered page in a browser.
