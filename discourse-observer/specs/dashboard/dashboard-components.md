# Dashboard Components

This document specifies the behavior of the dashboard view components rendered in the frontend.

These components implement the visual layer for the requirements defined in [queue-visibility.md](queue-visibility.md). That file defines *what* the user sees and why; this file defines *how* each component behaves to fulfill those requirements.

---

## Data types

Components consume the `DashboardData` and `Topic` interfaces defined in the mock data layer. These types will later be provided by the backend API; until then, mock data is used.

---

## SummaryCards

Accepts `DashboardData`. Renders three summary cards:

1. **"Väntar på svar"** — displays the count of `unrepliedTopics`.
2. **"Otaggade"** — displays the count of `untaggedTopics`.
3. **"Äldsta utan svar"** — displays the number of whole days since the oldest `createdAt` in `unrepliedTopics`. If the list is empty, displays "–".

The day calculation uses the difference between now and the oldest `createdAt`, truncated to whole days.

---

## UnrepliedTable

Accepts `Topic[]`. Renders a table with three columns:

| Column | Content |
|--------|---------|
| Ålder  | Time since `createdAt`, formatted as `"Xd"` (days) if ≥ 24 hours, otherwise `"Xh"` (hours) |
| Titel  | The topic title |
| Tagg   | Tags joined by comma, or "–" if empty |

Rows are sorted oldest first (ascending by `createdAt`).

---

## UntaggedTable

Accepts `Topic[]`. Renders a table with three columns:

| Column   | Content |
|----------|---------|
| Ålder    | Same age format as UnrepliedTable |
| Titel    | The topic title |
| Kategori | The topic category |

Rows are sorted oldest first (ascending by `createdAt`).

---

## Shared behavior

### Age formatting

Both table components format age identically:

- Compute hours elapsed since `createdAt`.
- If ≥ 24 hours: display as `"Xd"` where X is whole days (truncated).
- If < 24 hours: display as `"Xh"` where X is whole hours (truncated, minimum 1).

### Styling

- No inline styles. All styling uses CSS classes.
- Class name prefixes: `summary-` for SummaryCards, `unreplied-` for UnrepliedTable, `untagged-` for UntaggedTable.

### Implementation constraints

- Pure function components. No React hooks.
- Each component file stays under 200 lines.
- Types are imported from the mock data module.
