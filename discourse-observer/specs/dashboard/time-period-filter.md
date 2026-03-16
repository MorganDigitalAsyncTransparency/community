# Time Period Filter — Dashboard

This specification defines the requirements for UC-12: filtering dashboard metrics and lists by a selected time window. It applies to both the Queue Visibility page and the Response Metrics page.

This file defines *what* the user can do and how filtering behaves. Component details are in [dashboard-components.md](dashboard-components.md). Traceability is in [traceability.md](traceability.md).

---

## Use case traceability

| Requirement | Use case |
|-------------|----------|
| TF-1 – TF-14 | UC-12: Filter by time period |

---

## Requirements

### Period options (TF-1 – TF-2)

**TF-1.** The dashboard provides a time period selector visible on both the Queue and Response Metrics pages. It controls which topics appear across all metrics and lists on the active page.

**TF-2.** The available period options are: Last 7 days, Last 30 days, Last year, All time, and Custom range.

### Filter semantics (TF-3 – TF-9)

**TF-3.** The time period filters topics by `createdAt` — a topic is included when its creation timestamp falls within the selected window.

**TF-4.** "Last 7 days" includes topics with `createdAt >= now − 7 × 24 h`. The window is a rolling interval relative to the current clock time, not a calendar boundary.

**TF-5.** "Last 30 days" includes topics with `createdAt >= now − 30 × 24 h`.

**TF-6.** "Last year" includes topics with `createdAt >= now − 365 × 24 h`.

**TF-7.** "All time" includes all topics regardless of creation date.

**TF-8.** "Custom range" lets the user specify a start date and an end date. Topics with `createdAt` on or after the start date (from 00:00:00 UTC) and on or before the end date (through 23:59:59.999 UTC) are included. UTC boundaries are used because topic timestamps are stored in UTC, making the filter consistent and predictable regardless of the user's timezone.

**TF-9.** For a custom range, both a start date and an end date are required before the filter takes effect. While either is absent, the dashboard shows all topics — equivalent to "All time".

### Scope and persistence (TF-10 – TF-11)

**TF-10.** The filter applies to all three topic collections: unreplied topics, untagged topics, and resolved topics.

**TF-11.** The selected period persists when the user navigates between the Queue and Response Metrics pages. It does not reset on page switch.

### Defaults and empty states (TF-12 – TF-14)

**TF-12.** The default period on application load is "All time".

**TF-13.** When the active period yields no topics for a given list or metric, each list and metric shows its existing empty state: zero counts, dash for age or median time, no rows in tables. No additional empty-filter message is required.

**TF-14.** The period selector indicates which option is currently active, so the user always knows what window they are viewing.

---

## Design

### Types

Two types are defined in `timePeriod.ts`:

- `PeriodPreset` — union of the four named options: `"last7" | "last30" | "lastYear" | "allTime"`
- `ActivePeriod` — discriminated union distinguishing preset from custom:
  - `{ kind: "preset"; preset: PeriodPreset }` — one of the four named options
  - `{ kind: "custom"; range: { from: string; to: string } }` — ISO date strings (`YYYY-MM-DD`)

### Filter function

`filterByPeriod(topics, period)` is a pure function that takes a `Topic[]` and an `ActivePeriod` and returns the subset matching the period. It reads `Date.now()` for preset boundary calculations.

### Component

`PeriodSelector` renders the selector row. It is a pure function component — it holds no state. All state is managed in `App` and passed as props:

- `period: ActivePeriod` — the currently applied filter (used to highlight the active button).
- `customDraft: CustomRange | null` — the in-progress date input values; `null` means the custom inputs are not shown.
- `onPresetSelect(preset)` — called when the user clicks a preset button.
- `onCustomOpen()` — called when the user clicks the Custom button; `App` initialises `customDraft`.
- `onCustomDraftChange(from, to)` — called on every date input change; `App` applies the filter once both dates are non-empty.

### Placement

The `PeriodSelector` is rendered in `App.tsx` below the header and above the page content area. It is shared across both pages so the selected period persists without duplication.

### Data flow

`App.tsx` holds `activePeriod` state and applies `filterByPeriod` to each of the three topic lists from `DashboardData` before passing them to child components. Child component interfaces are unchanged — they continue to receive already-filtered `Topic[]` arrays.

### Scope

The filter applies to queue visibility and the response metrics summary cards. The weekly response time trend table (`ResponseTimeTrends`, UC-8) is explicitly excluded — it always receives the full unfiltered topic history regardless of the active period. This is by design: trend analysis requires consistent historical windows to be meaningful. See RT-8 in [response-time-trends.md](response-time-trends.md).

### Component–requirement mapping

| Component | File | Requirements |
|-----------|------|-------------|
| `PeriodSelector` | [PeriodSelector.tsx](../../frontend/src/components/PeriodSelector.tsx) | TF-2 — period options; TF-9 — incomplete custom; TF-14 — active indicator |
| `filterByPeriod` | [timePeriod.ts](../../frontend/src/components/timePeriod.ts) | TF-3 – TF-9 — filter semantics |
| `App` | [App.tsx](../../frontend/src/App.tsx) | TF-1 — shared placement; TF-10 — all collections filtered; TF-11 — persistence across pages; TF-12 — default; TF-13 — empty states (delegated to child components) |

---

## Validation

### Automated tests

Pure logic with well-defined boundary conditions. A defect in `filterByPeriod` would silently show wrong topics across every view.

| What | Requirements | Rationale |
|------|-------------|-----------|
| `filterByPeriod` with `last7` — includes topics within window, excludes topics outside, boundary topic included | TF-4 | Boundary correctness. Off-by-one would drop or include the wrong set of topics. |
| `filterByPeriod` with `last30` — excludes topics older than 30 days | TF-5 | Parallel boundary check. |
| `filterByPeriod` with `lastYear` — excludes topics older than 365 days | TF-6 | Parallel boundary check. |
| `filterByPeriod` with `allTime` — returns all topics unmodified | TF-7 | Confirms no accidental filtering in the default case. |
| `filterByPeriod` with custom range — includes topics on or after from date, on or before to date, excludes topics outside | TF-8 | Both bounds checked independently. |
| `filterByPeriod` with incomplete custom — not applicable; TF-9 is handled in component state | TF-9 | Component never calls `filterByPeriod` with a partial range. |
| `filterByPeriod` does not mutate input array | TF-3 | Safety property. |
| `filterByPeriod` returns empty array for empty input | TF-13 | Confirms empty-state delegation works correctly. |

Test location: `tests/dashboard/time-period-filter.unit.test.ts`

### Manual verification

| What | Requirements | Rationale |
|------|-------------|-----------|
| Period selector is visible on both the Queue and Response Metrics pages | TF-1 | Placement is a layout concern. |
| Period selector renders all five options | TF-2 | Presence and labels depend on rendered output. |
| Changing the period updates all three lists (unreplied, untagged, resolved) | TF-10 | Cross-list effect, best confirmed by observing counts change together. |
| Active period button is visually distinguished | TF-14 | Visual styling concern. |
| Custom date inputs appear when "Custom" is selected | TF-9 | Conditional rendering, best confirmed visually. |
| Custom filter does not apply until both dates are set | TF-9 | Stateful interaction, not unit-testable without a browser. |
| Switching pages preserves the selected period | TF-11 | Interaction concern, depends on component state lifecycle. |
| Default period on load is "All time" | TF-12 | Initial render state. |
