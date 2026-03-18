# Peak Activity — Dashboard View

This specification defines the requirements for UC-19: identifying when support topics typically arrive and when activity is highest. It traces to UC-19 in [use-cases.md](../use-cases.md).

This file defines *what* the user sees and why. [dashboard-components.md](dashboard-components.md) defines *how* each component behaves to fulfill these requirements.

---

## Use case traceability

| Requirement | Use case |
|-------------|----------|
| PA-1 – PA-9 | UC-19: Identify peak activity periods |
| PA-10 – PA-12 | Cross-cutting: placement, filters, empty states |

---

## Requirements

### Peak activity heatmap (UC-19)

**PA-1.** The user sees a breakdown of topic creation by day of week and hour of day, so that demand concentration is visible and staffing decisions can be informed.

**PA-2.** The breakdown is presented as a heatmap grid. Rows represent days of the week (Monday through Sunday). Columns represent hours of the day (0 through 23). Each cell shows the count of topics created in that (day, hour) slot.

**PA-3.** Cell background color intensity reflects the count relative to the maximum count in the grid. Higher counts produce darker cells. Zero-count cells have no background color. This makes peak periods visually obvious without requiring the viewer to read every number.

**PA-4.** The color scale is a single-hue gradient from white (zero) to a dark shade (maximum count). The exact color is an implementation detail, but it must provide sufficient contrast for the count text to remain readable at all intensity levels.

**PA-5.** Topics are assigned to (day, hour) slots based on their `createdAt` timestamp interpreted in UTC. This is consistent with all other time-based displays in the dashboard.

**PA-6.** The heatmap always shows the full 7×24 grid regardless of whether any topics fall in a given slot. Empty slots display 0 with no background color.

**PA-7.** Row labels show abbreviated day names: Mon, Tue, Wed, Thu, Fri, Sat, Sun. Column headers show hour numbers: 0, 1, 2, … 23.

**PA-8.** The data source is all filtered topics — unreplied and resolved combined — since the heatmap measures when topics are created regardless of their current status. This is the same data source as topic intake (TI-7).

**PA-9.** A legend below the heatmap shows the color scale from "0" (lightest) to the maximum count value (darkest), so the viewer can interpret intensity without hovering.

### Placement

**PA-10.** The peak activity heatmap appears on the existing Activity page, below the stalled topics section. Both UC-18 (stalled topics) and UC-19 (peak activity) relate to activity patterns, so grouping them on the same page is coherent.

### Filters

**PA-11.** The period filter (UC-12) applies. Only topics created within the selected time period are counted in the heatmap.

**PA-12.** The tag filter (UC-15) applies. When a tag is selected, only topics carrying that tag are counted. When no tag is selected, only topics carrying a monitored tag are counted.

### Empty state

**PA-13.** When there are no topics in the selected period and tag scope, an empty-state message is shown instead of the heatmap.

### Timezone header rows (ADR [0010](../../docs/decisions/0010-timezone-strategy.md))

**PA-14.** The heatmap displays a UTC hour header row (0–23) that is always present and cannot be removed. This row is the bottom header row and is visually labeled "UTC".

**PA-15.** The user can add up to three additional timezone header rows above the UTC row. Each row shows offset hour labels for a user-chosen IANA timezone.

**PA-16.** An "Add timezone" button appears below the heatmap (after the legend). When three timezone rows are already present, the button is disabled (greyed out) with a title attribute explaining the three-row limit.

**PA-17.** Each added timezone header row displays: the timezone short name (e.g. "CET"), the current UTC offset in parentheses (e.g. "(+1)"), and 24 offset hour labels aligned to the UTC columns. A remove button (×) at the end of the row allows the user to delete it.

**PA-18.** Offset hour labels are computed using `Intl.DateTimeFormat` with the selected IANA timezone identifier. This resolves the current UTC offset automatically, including DST transitions.

**PA-19.** Half-hour and quarter-hour offsets (e.g. IST +5:30, Nepal +5:45) display non-integer labels (e.g. "5:30", "13:30") rather than rounding.

**PA-20.** The data grid (7 rows × 24 columns) is unchanged — it shows topic counts bucketed by UTC day and hour. Timezone header rows are purely a reading aid; no re-bucketing occurs.

### Timezone picker

**PA-21.** Clicking "Add timezone" opens a timezone picker rendered inline below the button. The picker contains a text input for searching and a list of IANA timezones grouped by region. Selecting a timezone adds a header row and closes the picker.

**PA-22.** Duplicate timezone selections are prevented. If the user selects a timezone already displayed in a header row, no new row is added and the picker closes.

### Cookie consent

**PA-23.** When a user adds their first timezone and no cookie consent decision has been stored, a modal dialog appears explaining that the selection will be stored in a browser cookie so it persists across visits. The modal offers two choices: Accept (store the cookie, persist selections) or Deny (session-only, selections lost on reload).

**PA-24.** If the user chose Deny, the consent modal appears again the next time they add a timezone — giving them the opportunity to change their mind. If the user chose Accept, subsequent timezone additions are persisted to the cookie without re-prompting.

**PA-25.** The cookie stores only the list of selected IANA timezone identifiers (e.g. `["Europe/Berlin","Asia/Kolkata"]`). No personal data beyond timezone preference is stored.

---

## Design

### Bucketing

The `computeHeatmapData` function groups topics into a 7×24 grid:

```typescript
interface HeatmapCell {
  day: number;   // 0 = Monday, 6 = Sunday
  hour: number;  // 0–23
  count: number;
}

interface HeatmapData {
  cells: HeatmapCell[][];  // 7 rows × 24 columns
  maxCount: number;
}

function computeHeatmapData(topics: Topic[]): HeatmapData;
```

- Iterates over topics, extracts the UTC day of week and UTC hour from `createdAt`.
- JavaScript `getUTCDay()` returns 0 for Sunday — the function remaps to 0 = Monday, 6 = Sunday.
- Initializes a full 7×24 grid of zeros, then increments counts.
- Computes `maxCount` as the highest count in any cell (0 if no topics).

### Day label helper

```typescript
const DAY_LABELS: string[] = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
```

### Container component

`PeakActivity` accepts one prop:

| Prop | Type | Purpose |
|------|------|---------|
| `topics` | `Topic[]` | Filtered topics (period + tag filters already applied) |

Calls `computeHeatmapData(topics)` and renders:

- A section heading "Peak activity".
- An HTML `<table>` with column headers 0–23 and row labels Mon–Sun.
- Each cell shows the count and has a background color computed from `count / maxCount`.
- A color scale legend below the table.
- If the input is empty, renders an empty-state paragraph ("No data") instead of the table.

CSS class prefix: `peak-` for all elements specific to this component.

`PeakActivity` manages timezone header state internally via `useState`:

- `timezones: string[]` — list of selected IANA timezone identifiers (max 3).
- `pickerOpen: boolean` — whether the timezone picker is visible.
- `consentState: "pending" | "accepted" | "denied"` — cookie consent decision for the current session. Initialized from the consent cookie on mount (via `useState` initializer).

On mount, if a consent cookie with value `"accepted"` exists, the component reads the timezone list from the timezones cookie and initializes `timezones` from it. Otherwise `timezones` starts empty.

### Timezone offset computation

```typescript
function utcOffsetMinutes(timeZone: string): number;
```

Returns the UTC offset in minutes for the given IANA timezone at the current moment. Uses `Intl.DateTimeFormat` with `timeZoneName: "shortOffset"` to extract the offset string (e.g. "+5:30", "−9"), then parses it to minutes. Positive values mean ahead of UTC.

```typescript
function formatOffsetHour(utcHour: number, offsetMinutes: number): string;
```

Returns the display label for a given UTC hour adjusted by the offset. For whole-hour offsets, returns a number string (e.g. "14"). For fractional offsets, returns "H:MM" (e.g. "14:30").

```typescript
function timezoneShortName(timeZone: string): string;
```

Returns the short display name for a timezone (e.g. "CET", "IST") using `Intl.DateTimeFormat` with `timeZoneName: "short"`.

```typescript
function formatUtcOffset(offsetMinutes: number): string;
```

Returns a human-readable offset string like "+1", "−5", "+5:30".

### Cookie logic

```typescript
function readTimezoneCookie(): string[];
function writeTimezoneCookie(timezones: string[]): void;
function readConsentCookie(): "accepted" | null;
function writeConsentCookie(): void;
```

- The timezones cookie name is `peak_tz`. It stores a JSON-encoded array of IANA timezone strings.
- The consent cookie name is `peak_tz_consent`. It stores the string `"accepted"`.
- Both cookies use `SameSite=Lax`, no expiry (session default for timezones if consent denied; 365-day expiry if accepted).
- `readTimezoneCookie` returns `[]` if the cookie is absent or unparseable.

### Timezone picker component

`TimezonePicker` is a controlled component that accepts:

| Prop | Type | Purpose |
|------|------|---------|
| `onSelect` | `(tz: string) => void` | Called when the user selects a timezone |
| `onClose` | `() => void` | Called when the picker should close |
| `excludeTimezones` | `string[]` | Timezones already selected (hidden from the list) |

It renders a search input and a filtered list of IANA timezones grouped by region. The timezone list is a curated set of ~80 commonly used timezones (not the full ~400+ IANA database). Clicking a timezone calls `onSelect`.

CSS class prefix: `peak-tz-picker-` for all picker elements.

### Cookie consent modal

`CookieConsentModal` accepts:

| Prop | Type | Purpose |
|------|------|---------|
| `onAccept` | `() => void` | Called when the user accepts cookie storage |
| `onDeny` | `() => void` | Called when the user denies cookie storage |

It renders a modal overlay with an explanation of what will be stored and two buttons: "Accept" and "Deny".

CSS class prefix: `peak-consent-` for all modal elements.

### Color scale

Cell background: `rgba(59, 130, 246, α)` where α = `count / maxCount`. When count is 0, no background is applied (transparent). Text color switches to white when α > 0.5 for readability.

### Data flow

`App.tsx` computes the peak activity topic list by combining unreplied and resolved topics (same as topic intake), applying both the period filter and the tag filter, then passing the result to `PeakActivity`.

### Component–requirement mapping

| Component | File | Requirements |
|-----------|------|-------------|
| `App` | App.tsx | PA-8 — combines unreplied + resolved; PA-10 — activity page; PA-11 — period filter; PA-12 — tag filter |
| `PeakActivity` | PeakActivity.tsx | PA-1 — renders heatmap; PA-2 — grid layout; PA-3 — color intensity; PA-6 — full grid; PA-7 — labels; PA-9 — legend; PA-13 — empty state; PA-14 — UTC header row; PA-15 — timezone header rows; PA-16 — add button disabled at limit; PA-17 — row display; PA-20 — grid unchanged; PA-22 — duplicate prevention |
| `peakActivityMetrics` | peakActivityMetrics.ts | PA-2 — 7×24 bucketing; PA-5 — UTC interpretation |
| `timezoneUtils` | timezoneUtils.ts | PA-18 — offset computation via Intl; PA-19 — fractional offset labels |
| `TimezonePicker` | TimezonePicker.tsx | PA-21 — searchable grouped picker; PA-22 — duplicate prevention (via excludeTimezones) |
| `CookieConsentModal` | CookieConsentModal.tsx | PA-23 — consent modal; PA-24 — re-prompt on deny |
| `timezoneCookies` | timezoneCookies.ts | PA-25 — cookie storage |

---

## Validation

### Automated tests

| What | Requirements | Rationale |
|------|-------------|-----------|
| `computeHeatmapData` — single topic lands in correct (day, hour) cell | PA-2, PA-5 | Bucketing correctness. |
| `computeHeatmapData` — multiple topics in same slot increment count | PA-2 | Accumulation correctness. |
| `computeHeatmapData` — returns full 7×24 grid with zeros for empty slots | PA-6 | Grid completeness. |
| `computeHeatmapData` — maxCount reflects the highest cell count | PA-3 | Color scale basis. |
| `computeHeatmapData` — maxCount is 0 for empty input | PA-13 | Empty state. |
| `computeHeatmapData` — Sunday maps to row 6 (not row 0) | PA-5 | Day remapping correctness. |
| `computeHeatmapData` — Monday maps to row 0 | PA-5 | Day remapping correctness. |
| `computeHeatmapData` — uses UTC day and hour (not local time) | PA-5 | UTC correctness. |
| `computeHeatmapData` — does not mutate input array | PA-2 | Pure function contract. |
| `DAY_LABELS` — contains 7 labels Mon through Sun | PA-7 | Label correctness. |
| `utcOffsetMinutes` — returns correct offset for whole-hour timezone | PA-18 | Offset computation. |
| `utcOffsetMinutes` — returns correct offset for half-hour timezone (e.g. IST) | PA-18, PA-19 | Fractional offset. |
| `formatOffsetHour` — whole-hour offset produces integer label | PA-17 | Display correctness. |
| `formatOffsetHour` — fractional offset produces "H:MM" label | PA-19 | Half-hour display. |
| `formatOffsetHour` — wraps around 24 correctly (e.g. UTC hour 23 + offset 3 → 2) | PA-17 | Wraparound arithmetic. |
| `formatUtcOffset` — positive, negative, zero, and fractional offsets | PA-17 | Offset string formatting. |
| `readTimezoneCookie` — returns empty array when no cookie | PA-25 | Cookie absent. |
| `readTimezoneCookie` — returns parsed array from valid cookie | PA-25 | Cookie present. |
| `readTimezoneCookie` — returns empty array for malformed cookie | PA-25 | Robustness. |
| `writeTimezoneCookie` — writes JSON-encoded array | PA-25 | Cookie write. |
| `readConsentCookie` / `writeConsentCookie` — round-trip | PA-23, PA-24 | Consent persistence. |
| `TIMEZONE_LIST` — no duplicate entries | PA-22 | Data integrity. |
| `TIMEZONE_LIST` — every entry is a valid IANA identifier | PA-21 | Data integrity. |

Test locations: `tests/dashboard/peak-activity.unit.test.ts`, `tests/dashboard/timezone-utils.unit.test.ts`

### Manual verification

| What | Requirements | Rationale |
|------|-------------|-----------|
| Heatmap renders on the Activity page below stalled topics | PA-10 | Layout concern. |
| Cell colors intensify with higher counts | PA-3, PA-4 | Visual rendering concern. |
| Color scale legend is visible below the heatmap | PA-9 | Visual rendering concern. |
| Switching period filter updates the heatmap | PA-11 | Cross-component interaction. |
| Selecting a tag scopes the heatmap to that tag | PA-12 | Filter interaction. |
| Empty-state message shown when no topics in period | PA-13 | Requires visual confirmation. |
| Count text is readable at all intensity levels | PA-4 | Accessibility concern. |
| UTC header row is always visible and cannot be removed | PA-14 | Visual confirmation. |
| Added timezone header rows show correct offset hours | PA-15, PA-17, PA-18 | Offset display. |
| "Add timezone" button is disabled when 3 rows present | PA-16 | Interaction limit. |
| Timezone picker opens, allows search, and closes on selection | PA-21 | Picker interaction. |
| Selecting a duplicate timezone does not add a second row | PA-22 | Duplicate prevention. |
| Cookie consent modal appears on first timezone addition | PA-23 | Consent flow. |
| Accepting consent persists timezones across page reload | PA-24 | Persistence. |
| Denying consent loses timezones on page reload, re-prompts | PA-24 | Session-only behavior. |
| Half-hour offset timezones display correctly (e.g. IST +5:30) | PA-19 | Fractional offset rendering. |
