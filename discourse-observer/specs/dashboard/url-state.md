# URL State Synchronization — Dashboard

This specification defines the requirements for persisting filter and page state in the browser URL so that dashboard views can be bookmarked and shared. It applies to all dashboard pages.

This file defines *what* is persisted and how URL state maps to application state. Component details are in [dashboard-components.md](dashboard-components.md). Traceability is in [traceability.md](traceability.md).

---

## Use case traceability

| Requirement | Use case |
|-------------|----------|
| US-1 – US-12 | UC-24: Persist filter state in URL |

---

## Requirements

### URL parameters (US-1 – US-4)

**US-1.** The dashboard persists the following state in URL query parameters: active page, period selection, active tag, and active area.

**US-2.** The URL parameter names are:

| State | Parameter | Valid values | Default (omitted) |
|-------|-----------|-------------|-------------------|
| Page | `page` | `queue`, `response-metrics`, `distribution`, `slo`, `activity`, `tag-flows`, `sync-log` | `queue` |
| Period preset | `period` | `last7`, `last30`, `lastYear`, `allTime` | `allTime` |
| Custom range start | `from` | ISO date `YYYY-MM-DD` | *(none)* |
| Custom range end | `to` | ISO date `YYYY-MM-DD` | *(none)* |
| Active tag | `tag` | Any monitored tag name | *(none — all tags)* |
| Active area | `area` | Any area name from configuration | *(none — all areas)* |

**US-3.** When the period is `custom`, the `from` and `to` parameters are included. The `period` parameter is omitted when the range is custom (the presence of both `from` and `to` implies custom mode). When the period is a preset, `from` and `to` are omitted.

**US-4.** Parameters at their default value are omitted from the URL to keep it clean. An empty query string represents the default state.

### Reading state from URL (US-5 – US-8)

**US-5.** On application load, the dashboard reads query parameters from the current URL and uses them as the initial state for page, period, tag, and area.

**US-6.** Unrecognized or missing parameters fall back to defaults: page `queue`, period `allTime`, no tag, no area.

**US-7.** Invalid parameter values are treated as missing — the default is used. An unrecognized `page` value falls back to `queue`. An unrecognized `period` value falls back to `allTime`. A `from` or `to` that is not a valid `YYYY-MM-DD` date is ignored.

**US-8.** When `from` is present without `to` (or vice versa), the incomplete range is ignored and the period falls back to `allTime`.

### Writing state to URL (US-9 – US-10)

**US-9.** When any filter or page state changes, the URL is updated to reflect the new state. The update uses `history.replaceState` — it does not create browser history entries and does not trigger a page reload.

**US-10.** The URL is updated synchronously with the state change. There is no debounce or delay.

### Scope (US-11 – US-12)

**US-11.** URL state synchronization does not change the behavior of any existing filter. The period filter, tag filter, area selector, and page navigation continue to work as defined in their respective specifications. URL synchronization is an additional persistence layer.

**US-12.** The custom draft state (in-progress date inputs before both dates are set) is not persisted in the URL. Only the applied filter state is reflected.

---

## Design

### Pure functions

Defined in `urlState.ts`:

- `parseUrlState(search: string): UrlState` — parses a URL search string into a state object. Validates each parameter and falls back to defaults for invalid values.
- `buildSearch(state: UrlState): string` — serializes a state object into a URL search string. Omits parameters at their default value.

The `UrlState` type mirrors the filter state managed in `App`:

```typescript
interface UrlState {
  page: Page;
  period: ActivePeriod;
  tag: string | null;
  area: string | null;
}
```

### Hook

`useUrlState` is a custom hook that:

1. Calls `parseUrlState(window.location.search)` once on mount to produce the initial state.
2. Returns the current state and setter functions that both update React state and call `history.replaceState` with the new URL.

### Data flow

`App.tsx` replaces its individual `useState` calls for `page`, `activePeriod`, `activeTag`, and `activeArea` with the `useUrlState` hook. The hook returns the same state shape and setter interface, so child components are unaffected.

### Component–requirement mapping

| Component | File | Requirements |
|-----------|------|-------------|
| `parseUrlState` | [urlState.ts](../../frontend/src/components/urlState.ts) | US-5 – US-8 — parse and validate |
| `buildSearch` | [urlState.ts](../../frontend/src/components/urlState.ts) | US-2 – US-4 — serialize and omit defaults |
| `useUrlState` | [useUrlState.ts](../../frontend/src/components/useUrlState.ts) | US-9 – US-10 — sync state to URL |
| `App` | [App.tsx](../../frontend/src/App.tsx) | US-1 — integration; US-11 — no behavior change; US-12 — draft not persisted |

---

## Validation

### Automated tests

Pure parsing and serialization logic with well-defined inputs and outputs.

| What | Requirements | Rationale |
|------|-------------|-----------|
| `parseUrlState` with valid page — returns parsed page | US-5 | Core parse path. |
| `parseUrlState` with invalid page — falls back to `queue` | US-7 | Invalid value handling. |
| `parseUrlState` with valid preset period — returns preset | US-5 | Preset parsing. |
| `parseUrlState` with invalid period — falls back to `allTime` | US-7 | Invalid value handling. |
| `parseUrlState` with `from` and `to` — returns custom period | US-5, US-3 | Custom range parsing. |
| `parseUrlState` with `from` only — falls back to `allTime` | US-8 | Incomplete range. |
| `parseUrlState` with `to` only — falls back to `allTime` | US-8 | Incomplete range. |
| `parseUrlState` with invalid date format — falls back to `allTime` | US-7 | Malformed date handling. |
| `parseUrlState` with tag parameter — returns tag | US-5 | Tag parsing. |
| `parseUrlState` with area parameter — returns area | US-5 | Area parsing. |
| `parseUrlState` with empty string — returns all defaults | US-6 | Empty/missing parameters. |
| `buildSearch` with all defaults — returns empty string | US-4 | Default omission. |
| `buildSearch` with non-default page — includes page param | US-2 | Page serialization. |
| `buildSearch` with preset period — includes period param | US-2 | Preset serialization. |
| `buildSearch` with custom period — includes from and to, omits period | US-3 | Custom serialization. |
| `buildSearch` with tag — includes tag param | US-2 | Tag serialization. |
| `buildSearch` with area — includes area param | US-2 | Area serialization. |
| `buildSearch` round-trip — `parseUrlState(buildSearch(state))` equals input | US-5 | Round-trip consistency. |

Test location: `tests/dashboard/url-state.unit.test.ts`

### Manual verification

| What | Requirements | Rationale |
|------|-------------|-----------|
| Bookmarking a filtered view and reopening restores filters | US-1, US-5 | End-to-end persistence. |
| Changing a filter updates the URL bar without page reload | US-9 | Visual confirmation of URL sync. |
| Pasting a URL with filters applies them on load | US-5 | Shareability. |
| Invalid URL parameters show default state without errors | US-6, US-7 | Robustness. |
| Clear all filters resets URL to clean state | US-4, US-11 | Integration with existing clear behavior. |
