# Dashboard Components

This document specifies the behavior of the dashboard view components rendered in the frontend.

These components implement the visual layer for the requirements defined in [queue-visibility.md](queue-visibility.md), [response-metrics.md](response-metrics.md), [time-period-filter.md](time-period-filter.md), [response-time-trends.md](response-time-trends.md), [tag-distribution.md](tag-distribution.md), [slo-monitoring.md](slo-monitoring.md), [tag-area-filter.md](tag-area-filter.md), [topic-intake.md](topic-intake.md), [stalled-topics.md](stalled-topics.md), [peak-activity.md](peak-activity.md), [response-time-distribution.md](response-time-distribution.md), and [url-state.md](url-state.md). Those files define *what* the user sees and why; this file defines *how* each component behaves to fulfill those requirements.

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
| Topic  | The topic title, linked to the Discourse forum topic via `topicUrl` |
| Tags   | Tags joined by comma, or "–" if empty |
| Age    | Time since `createdAt`, formatted as `"Xd"` (days) if ≥ 24 hours, otherwise `"Xh"` (hours) |

All column headers are clickable to sort ascending or descending. Default sort: Age descending (oldest first). Sort state is managed via the shared `useTableSort` hook.

---

## UntaggedTable

Accepts `Topic[]`. Renders a table with three columns:

| Column   | Content |
|----------|---------|
| Topic    | The topic title, linked to the Discourse forum topic via `topicUrl` |
| Category | The topic category |
| Age      | Same age format as UnrepliedTable |

All column headers are clickable to sort ascending or descending. Default sort: Age descending (oldest first). Sort state is managed via the shared `useTableSort` hook.

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

## VolumeChart

Accepts `data: VolumeBucket[]` (multi-series chart data). Renders a Recharts `LineChart` inside a `ResponsiveContainer` (width 100%, height 300px).

Four `Line` series:

- "Topics created" (`created`) — colored via `--color-chart-3`.
- "Accepted answer" (`accepted`) — colored via `--color-chart-6`.
- "Topics closed" (`closed`) — colored via `--color-chart-5`.
- "Currently open" (`open`) — colored via `--color-chart-4`.

Chart features:

- `XAxis` with `dataKey="label"` showing bucket date labels.
- `YAxis` with `allowDecimals={false}` for whole numbers.
- `Tooltip` showing all series values.
- `Legend` identifying the four series.

CSS class prefix: `volume-chart-` for chart-specific elements.

---

## MedianTrendChart

Accepts three props:

| Prop | Type | Purpose |
|------|------|---------|
| `data` | `MedianBucket[]` | Per-bucket median durations |
| `color` | `string` | Line color |
| `name` | `string` | Series name for tooltip |

Renders a Recharts `LineChart` inside a `ResponsiveContainer` (width 100%, height 250px). A single `Line` series plots `medianHours`. Buckets with `undefined` values appear as gaps (`connectNulls={false}`).

Chart features:

- `XAxis` with `dataKey="label"` showing bucket date labels.
- `YAxis` with duration-formatted tick labels (hours/days via `formatDuration`).
- `Tooltip` showing formatted duration.

CSS class prefix: `median-trend-chart-` for chart-specific elements.

---

## PeriodSelector

Accepts `period: ActivePeriod`, `customDraft: CustomRange | null`, `onPresetSelect`, `onCustomOpen`, and `onCustomDraftChange` callbacks. Renders a row of period option buttons and, when `customDraft` is not null, two date inputs.

- Clicking a preset button calls `onPresetSelect` with the selected preset.
- Clicking "Custom" calls `onCustomOpen`. The custom inputs are shown when `customDraft !== null`.
- Changing a date input calls `onCustomDraftChange(from, to)` with the updated values.
- The currently active option is indicated by the `period-btn-active` CSS class.

`PeriodSelector` is a pure function component — it holds no state. All state is managed by `App` and passed as props. See [time-period-filter.md](time-period-filter.md) for the filter requirements.

---

## TagSelector

Accepts five props:

| Prop | Type | Purpose |
|------|------|---------|
| `config` | `TagConfig` | Unified tag/area/SLO configuration loaded from `config/tagConfig.json` |
| `activeTag` | `string \| null` | Currently selected tag, or `null` for all |
| `activeArea` | `string \| null` | Currently selected area, or `null` for all |
| `onTagSelect` | `(tag: string \| null) => void` | Called when a tag button is clicked |
| `onAreaSelect` | `(area: string \| null) => void` | Called when an area is selected |

Renders two controls in a single row:

1. **Area selector** — a `<select>` dropdown with "All areas" as the default option, followed by one option per area from the configuration.
2. **Tag buttons** — an "All" button (representing no tag selected) followed by one button per visible tag. Visible tags are determined by `tagsForArea(config, activeArea)`.

- Selecting an area calls `onAreaSelect`. The tag selection is preserved.
- Clicking a tag button calls `onTagSelect` with the tag name, or `null` for the "All" button.
- The currently active tag button is indicated by the `tag-btn-active` CSS class.
- Primary tags (defined by `areas[].primaryTag` in the configuration) are marked with an asterisk suffix (e.g. `api*`) in all views.

`TagSelector` is a pure function component — it holds no state. All state is managed by `App` and passed as props. See [tag-area-filter.md](tag-area-filter.md) for the filter requirements.

CSS class prefix: `tag-` for all elements specific to this component.

---

## TagDistribution

Accepts five props:

| Prop | Type | Purpose |
|------|------|---------|
| `allTopics` | `Topic[]` | Filtered unreplied + resolved topics combined — used for UC-9 volume ranking |
| `resolvedTopics` | `Topic[]` | Filtered resolved topics — used for UC-10 resolution time ranking |
| `openTopics` | `Topic[]` | Filtered unreplied topics — used for UC-11 per-tag snapshot |
| `allTopicsHistory` | `Topic[]` | Unfiltered unreplied + resolved combined — used for UC-11 weekly trend |
| `openTopicsHistory` | `Topic[]` | Unfiltered unreplied topics — used for UC-11 weekly trend |

Renders three sections in order:

**1. "Topics by tag" (UC-9):** Calls `tagVolumeRanking(allTopics)` and renders a table with columns Tag and Topics, sorted highest count first. Shows an empty-state paragraph ("No data") when the result is empty.

**2. "Resolution time by tag" (UC-10):** Calls `tagResolutionRanking(resolvedTopics)` and renders a table with columns Tag, Resolved, and Median resolution. Tags with "–" median sort to the bottom. Shows an empty-state paragraph when the result is empty.

**3. "Open backlogs by tag" (UC-11):** Two parts:

- Calls `tagBacklogRanking(openTopics)` and renders a table with columns Tag and Open topics, sorted highest count first. Shows an empty-state paragraph when empty.
- Calls `computeWeeklyBacklog(allTopicsHistory, openTopicsHistory)` and renders a weekly trend table with columns Week, Created, Resolved, and Still open, ordered newest first. Shows an empty-state paragraph when empty. Week labels are formatted as UTC short dates via `toLocaleDateString`.

`TagDistribution` is a pure function component. It holds no state — all filtering and history scoping is handled by `App` before passing props. See [tag-distribution.md](tag-distribution.md) for the requirements.

CSS class prefix: `dist-` for all elements specific to this component.

---

## SloMonitor

Accepts four props:

| Prop | Type | Purpose |
|------|------|---------|
| `unrepliedTopics` | `Topic[]` | Filtered unreplied topics — used for first reply and inactivity violation checks and compliance |
| `resolvedTopics` | `Topic[]` | Filtered resolved topics — used for all three violation checks and compliance |
| `sloConfig` | `SloConfig` | Tag-to-threshold mapping scoped to visible tags (filtered by active area/tag selection) |
| `defaultSloTags` | `Set<string>` | Tags using default SLO thresholds — shown with "(default thresholds)" indicator |

Renders two sections in order:

**1. "Threshold violations" (UC-13):** Three subsections, one per threshold type — "First reply violations", "Resolution violations", "Inactivity violations". Each calls the corresponding violation function and renders a table with columns: Topic, Tag, Threshold, Actual, Excess. The Topic column shows the topic title as a clickable link to the Discourse forum topic. All column headers are clickable to sort ascending or descending; default sort is by Excess descending. Each subsection shows an empty-state paragraph ("No violations") when the list is empty.

**2. "SLO compliance" (UC-14):** Calls the compliance function and renders a table with columns: Tag, First reply, Resolution, Inactivity. Each cell shows a percentage or "–" when no topics are eligible. All column headers are clickable to sort; default sort is by Tag ascending. Shows an empty-state paragraph ("No data") when no monitored tags have eligible topics.

When the SLO configuration is empty (no tags configured), the entire component renders a single empty-state message ("No SLO thresholds configured") instead of the sections above.

`SloMonitor` manages sort state for each sub-table via the shared `useTableSort` hook. All topic filtering is handled by `App` before passing props. See [slo-monitoring.md](slo-monitoring.md) for the requirements.

CSS class prefix: `slo-` for all elements specific to this component.

---

## StalledTopics

Accepts three props:

| Prop | Type | Purpose |
|------|------|---------|
| `topics` | `Topic[]` | Filtered replied open topics (period + tag filters already applied) |
| `resolvedTags` | `Record<string, ResolvedTag>` | Resolved tag configuration with per-tag stalledDays and closedTag |
| `monitoredTags` | `string[]` | All monitored tags — used to display the first monitored tag per topic |

Calls `filterStalledTopics(topics, resolvedTags)` and renders:

- A section heading: "Stalled topics" (no threshold number — thresholds vary per topic).
- A table with four columns:

| Column | Content |
|--------|---------|
| Topic | The topic title, linked to the Discourse forum topic via `topicUrl` |
| Tag | First monitored tag on the topic, or "–" if none |
| Threshold | Per-topic stalled-days value via `stalledThresholdForTopic`, with "(default)" indicator when inherited from defaults |
| Days inactive | Whole days since `lastActivityAt` (truncated), falling back to `createdAt` |

- All column headers are clickable to sort ascending or descending. Default sort: Days inactive descending (highest first). Clicking a column header toggles direction; clicking a different column switches to that column's default direction (Topic: ascending, Tag: ascending, Threshold: ascending, Days inactive: descending).
- The active sort column header shows an arrow indicator (▲ ascending, ▼ descending).
- If no topics pass the stalled filter, renders an empty-state paragraph ("No stalled topics") instead of the table.

`StalledTopics` manages sort state via the shared `useTableSort` hook. All topic filtering is handled by `App` before passing props. See [stalled-topics.md](stalled-topics.md) for the requirements.

CSS class prefix: `stalled-` for all elements specific to this component.

---

## PeakActivity

Accepts one prop:

| Prop | Type | Purpose |
|------|------|---------|
| `topics` | `Topic[]` | Filtered unreplied + resolved topics combined (period + tag filters already applied) |

Calls `computeHeatmapData(topics)` and renders:

- A section heading "Peak activity".
- An HTML `<table>` heatmap with timezone header rows (see below), a UTC hour header row, 7 data rows (Mon–Sun), and 24 columns (hours 0–23).
- Each data cell displays the topic count and has a background color whose intensity reflects the count relative to the grid maximum (`count / maxCount`).
- Cell background uses `rgba(var(--color-heatmap-base), α)`. Text color switches to white when α > 0.5. The heatmap base color is read from the CSS custom property via the `HEATMAP_BASE` constant exported by `themeColors.ts`.
- Zero-count cells show "0" with no background color (transparent).
- A color scale legend below the table showing the range from 0 to the maximum count.
- An "Add timezone" button above the table (after the heading). Disabled when 3 timezone rows are present.
- If the input is empty, renders an empty-state paragraph ("No data") instead of the table.

`PeakActivity` manages timezone-related state internally via `useState`: the list of selected IANA timezones (max 3), whether the timezone picker is open, and the cookie consent state. On mount it reads persisted state from cookies if consent was previously accepted. See [peak-activity.md](peak-activity.md) for the requirements.

CSS class prefix: `peak-` for all elements specific to this component.

### Timezone header rows

The `<thead>` contains:

1. Zero to three user-added timezone header rows, each showing: a label cell with the timezone short name and UTC offset, 24 offset hour labels, and a remove (×) button cell. CSS class: `peak-tz-row`.
2. A UTC header row (always present, always last in `<thead>`). Label cell shows "UTC". CSS class: `peak-header-utc`.

A visual separator (heavier border) distinguishes the header rows from the data rows.

---

## TimezonePicker

A controlled component that renders a searchable timezone list. Accepts:

| Prop | Type | Purpose |
|------|------|---------|
| `onSelect` | `(tz: string) => void` | Called when a timezone is selected |
| `onClose` | `() => void` | Called to close the picker |
| `excludeTimezones` | `string[]` | Timezones to hide (already selected) |

Renders a text input for filtering and a scrollable list of IANA timezones sorted by UTC offset. Each entry shows the offset, short code (e.g. "IST", "CET"), and representative cities. The list is a curated set of ~24 entries covering all major offsets. Search matches code, city name, offset string, and IANA identifier. Clicking a timezone calls `onSelect` and then `onClose`.

CSS class prefix: `peak-tz-picker-` for all picker elements.

---

## CookieConsentModal

A modal dialog for timezone cookie consent. Accepts:

| Prop | Type | Purpose |
|------|------|---------|
| `onAccept` | `() => void` | User accepts cookie persistence |
| `onDeny` | `() => void` | User denies (session-only) |

Renders a backdrop overlay and a centered dialog explaining that timezone selections will be stored in a browser cookie. Two buttons: "Accept" and "Deny".

CSS class prefix: `peak-consent-` for all modal elements.

---

## ResponseTimeDistribution

Accepts two props:

| Prop | Type | Purpose |
|------|------|---------|
| `topics` | `Topic[]` | Filtered resolved topics (period + tag filters already applied) |
| `ceilingsHours` | `number[]` | Bucket ceilings from `config/distributionBuckets.json` |

Calls `firstReplyDurations` and `resolutionDurations` to extract duration arrays, then `bucketDurations` for each, and renders:

- Two histogram sections, each with a heading ("First reply distribution", "Resolution time distribution").
- A `DistributionChart` for each non-empty result.
- An empty-state paragraph ("No data") when the duration array is empty.

`ResponseTimeDistribution` is a pure function component. It holds no state — all filtering is handled by `App` before passing props. See [response-time-distribution.md](response-time-distribution.md) for the requirements.

CSS class prefix: `rd-` for all elements specific to this component.

---

## DistributionChart

Accepts three props:

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

`DistributionChart` uses Recharts' `ResponsiveContainer`, which requires a parent with defined dimensions. The chart wrapper div provides this via CSS.

---

## Sidebar

Accepts four props:

| Prop | Type | Purpose |
|------|------|---------|
| `activePage` | `Page` | Currently visible page |
| `onNavigate` | `(page: Page) => void` | Called when a navigation link is clicked |
| `mobileOpen` | `boolean \| undefined` | When `true`, sidebar renders as a fixed overlay (mobile breakpoint) |
| `onMobileClose` | `(() => void) \| undefined` | Called to close the mobile overlay — triggered by backdrop click or navigation |

Renders a vertical sidebar spanning the full viewport height with:

- A logo section showing "discourse-observer" (expanded) or "d-o" (collapsed).
- Five navigation links — Queue, Response metrics, Distribution, SLO, Activity — each with an icon and text label. The active page is visually distinguished (`sidebar-link-active`).
- A collapse/expand toggle at the bottom.

**States:**

- **Expanded** (200px): icon + text label per page.
- **Collapsed** (48px): icon only, with `title` tooltip showing the page name.

The toggle persists the user's preference in `localStorage` under `sidebar-collapsed`. On mount, the component reads this value to restore the previous state. The CSS Grid uses `auto 1fr` columns so it follows the sidebar's actual width — no JavaScript-to-CSS bridge is needed.

Width transition is ~200ms ease for visual continuity.

`Sidebar` uses `useState` for collapsed state. localStorage persistence is synchronous in the toggle handler — no `useEffect`. Navigation state is managed by `App` via `useUrlState` — page changes are reflected in the URL (see [url-state.md](url-state.md)).

**Mobile overlay (< 768px):** When `mobileOpen` is `true`, the sidebar renders as a fixed overlay (`sidebar-mobile-open`) above a semi-transparent backdrop (`sidebar-backdrop`). Clicking the backdrop or any navigation link calls `onMobileClose`. The backdrop and mobile-open class are only styled in the `<= 767px` media query — at wider viewports they have no visual effect.

CSS class prefix: `sidebar-` for all elements specific to this component.

---

## Footer

Accepts two props:

| Prop | Type | Purpose |
|------|------|---------|
| `version` | `string` | Application version string (e.g. "v0.1.0") |
| `lastSyncedAt` | `string` | ISO 8601 timestamp of the last data sync |

Renders a footer bar at the bottom of the main area containing version, formatted last sync time, and a GitHub repository link.

`Footer` is a pure function component — no hooks, no state.

CSS class prefix: `footer-` for all elements specific to this component.

---

## Shared behavior

### Age formatting

Both table components format age identically:

- Compute hours elapsed since `createdAt`.
- If ≥ 24 hours: display as `"Xd"` where X is whole days (truncated).
- If < 24 hours: display as `"Xh"` where X is whole hours (truncated, minimum 1).

### Duration formatting

All time displays — both topic age and response time metrics — use a single formatting function (`formatDuration`) that accepts a duration in milliseconds:

- If ≥ 24 hours: display as `"Xd"` where X is whole days (truncated).
- If < 24 hours: display as `"Xh"` where X is whole hours (truncated, minimum 1).

`formatAge` delegates to `formatDuration` after computing the elapsed time from an ISO date string.

### Styling

- No inline styles. All styling uses CSS classes.
- Class name prefixes: `summary-` for SummaryCards, `unreplied-` for UnrepliedTable, `untagged-` for UntaggedTable, `response-` for ResponseMetricsCards, `sidebar-` for Sidebar (including `sidebar-backdrop` and `sidebar-mobile-open`), `hamburger` for the mobile menu button, `footer-` for Footer, `period-` for PeriodSelector, `tag-` for TagSelector, `volume-chart-` for VolumeChart, `median-trend-chart-` for MedianTrendChart, `slo-` for SloMonitor, `stalled-` for StalledTopics, `peak-` for PeakActivity, `peak-tz-picker-` for TimezonePicker, `peak-consent-` for CookieConsentModal, `rd-` for ResponseTimeDistribution, `rd-chart-` for DistributionChart.

### Implementation constraints

- Pure function components. No React hooks. Exceptions: `App` uses `useUrlState` for page, active period, active tag, and active area (persisted in URL query parameters — see [url-state.md](url-state.md)) and `useState` for custom range draft and mobile sidebar visibility state, as it is the application shell — not a display component. `Sidebar` uses `useState` for collapsed state — this state is local to the sidebar and does not affect other components. `UnrepliedTable`, `UntaggedTable`, `StalledTopics`, and `SloMonitor` (via its `ViolationTable` and `ComplianceTable` sub-components) use the shared `useTableSort` hook for sort column and direction — this state is local to each table and does not affect other components. `PeakActivity` uses `useState` for timezone selections, picker visibility, and cookie consent state — this state is local to the heatmap and does not affect other components. `TimezonePicker` uses `useState` for the search input — this is local input state that does not affect other components.
- Each component file stays under 200 lines.
- Types are imported from the mock data module.
