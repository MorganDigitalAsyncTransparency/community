# Dashboard Components

This document specifies the behavior of the dashboard view components rendered in the frontend.

These components implement the visual layer for the requirements defined in [queue-visibility.md](queue-visibility.md) and [response-metrics.md](response-metrics.md). Those files define *what* the user sees and why; this file defines *how* each component behaves to fulfill those requirements.

---

## Data types

Components consume the `DashboardData` and `Topic` interfaces defined in the mock data layer. These types will later be provided by the backend API; until then, mock data is used.

---

## SummaryCards

Accepts `DashboardData`. Renders three summary cards:

1. **"Awaiting reply"** — displays the count of `unrepliedTopics`.
2. **"Untagged"** — displays the count of `untaggedTopics`.
3. **"Oldest unreplied"** — displays the number of whole days since the oldest `createdAt` in `unrepliedTopics`. If the list is empty, displays "–".

The day calculation uses the difference between now and the oldest `createdAt`, truncated to whole days.

---

## UnrepliedTable

Accepts `Topic[]`. Renders a table with three columns:

| Column | Content |
|--------|---------|
| Age    | Time since `createdAt`, formatted as `"Xd"` (days) if ≥ 24 hours, otherwise `"Xh"` (hours) |
| Title  | The topic title |
| Tags   | Tags joined by comma, or "–" if empty |

Rows are sorted oldest first (ascending by `createdAt`).

---

## UntaggedTable

Accepts `Topic[]`. Renders a table with three columns:

| Column   | Content |
|----------|---------|
| Age      | Same age format as UnrepliedTable |
| Title    | The topic title |
| Category | The topic category |

Rows are sorted oldest first (ascending by `createdAt`).

---

## ResponseMetricsCards

Accepts `Topic[]` (the resolved topics). Renders four summary cards:

1. **"Median first reply"** — displays the median time from `createdAt` to `firstReplyAt` across topics that have a `firstReplyAt`. Topics without `firstReplyAt` are excluded. If no topics qualify, displays "–". Time is formatted using the shared duration format (see Duration formatting below).
2. **"Median resolution"** — displays the median time from `createdAt` to `resolvedAt` across all resolved topics. If the list is empty, displays "–". Time is formatted using the shared duration format.
3. **"Outcomes"** — displays solved and self-closed counts as `"X solved / Y self-closed"`. If the list is empty, displays `"0 solved / 0 self-closed"`.
4. **"Answer rate"** — displays the percentage of topics with `outcome === "solved"`, rounded to the nearest whole number, followed by "%". If the list is empty, displays "–".

### Median calculation

Given a sorted array of durations, the median is:

- If odd count: the middle value.
- If even count: the average of the two middle values, truncated to a whole number of milliseconds.

---

## Navigation

The `App` component renders two navigation links in the header: "Queue" and "Response metrics". Clicking a link switches the visible page content. The active link is visually distinguished using a CSS class (`nav-link-active`).

Navigation uses component state — no client-side router.

---

## Shared behavior

### Age formatting

Both table components format age identically:

- Compute hours elapsed since `createdAt`.
- If ≥ 24 hours: display as `"Xd"` where X is whole days (truncated).
- If < 24 hours: display as `"Xh"` where X is whole hours (truncated, minimum 1).

### Duration formatting

Metrics that display a time duration (median first reply, median resolution) use the same format as age:

- If ≥ 24 hours: display as `"Xd"` where X is whole days (truncated).
- If < 24 hours: display as `"Xh"` where X is whole hours (truncated, minimum 1).

This reuses the `formatAge` logic but accepts a duration in milliseconds rather than an ISO date string.

### Styling

- No inline styles. All styling uses CSS classes.
- Class name prefixes: `summary-` for SummaryCards, `unreplied-` for UnrepliedTable, `untagged-` for UntaggedTable, `response-` for ResponseMetricsCards, `nav-` for navigation.

### Implementation constraints

- Pure function components. No React hooks.
- Each component file stays under 200 lines.
- Types are imported from the mock data module.
