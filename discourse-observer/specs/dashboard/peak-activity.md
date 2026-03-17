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

`PeakActivity` is a pure function component. It holds no state — all filtering is handled by `App` before passing props.

### Color scale

Cell background: `rgba(59, 130, 246, α)` where α = `count / maxCount`. When count is 0, no background is applied (transparent). Text color switches to white when α > 0.5 for readability.

### Data flow

`App.tsx` computes the peak activity topic list by combining unreplied and resolved topics (same as topic intake), applying both the period filter and the tag filter, then passing the result to `PeakActivity`.

### Component–requirement mapping

| Component | File | Requirements |
|-----------|------|-------------|
| `App` | App.tsx | PA-8 — combines unreplied + resolved; PA-10 — activity page; PA-11 — period filter; PA-12 — tag filter |
| `PeakActivity` | PeakActivity.tsx | PA-1 — renders heatmap; PA-2 — grid layout; PA-3 — color intensity; PA-6 — full grid; PA-7 — labels; PA-9 — legend; PA-13 — empty state |
| `peakActivityMetrics` | peakActivityMetrics.ts | PA-2 — 7×24 bucketing; PA-5 — UTC interpretation |

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

Test location: `tests/dashboard/peak-activity.unit.test.ts`

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
