# Tag and Area Filter — Dashboard

This specification defines the requirements for UC-15 (filter dashboard by tag) and UC-16 (navigate tags by area). It applies to all dashboard pages.

This file defines *what* the user can do and how filtering behaves. Component details are in [dashboard-components.md](dashboard-components.md). Traceability is in [traceability.md](traceability.md).

---

## Use case traceability

| Requirement | Use case |
|-------------|----------|
| TA-1 – TA-8 | UC-15: Filter dashboard by tag |
| TA-9 – TA-14 | UC-16: Navigate tags by area |
| TA-15 – TA-21 | Cross-cutting: configuration, placement, defaults, empty states |

---

## Configuration

**TA-15.** Tags and areas are defined in a JSON configuration file (`config/tagConfig.json`). The file contains an array of area objects. Each area has a `name`, a `primaryTag`, and a `tags` array listing all tags in the area (including the primary tag).

**TA-16.** A committed example file (`config/tagConfig.example.json`) documents the expected schema. The runtime file is gitignored and created from the example during setup, following the same pattern as `config/sloThresholds.json`.

**TA-17.** The union of all tags across all areas defines the set of monitored tags. Only topics carrying at least one monitored tag are counted in metrics when no tag is selected.

---

## Requirements

### Tag filtering (UC-15)

**TA-1.** The dashboard provides a tag selector that lets the user choose a single monitored tag. Selecting a tag filters all metrics, lists, and charts on the active page to show only topics carrying that tag.

**TA-2.** When no tag is selected, data covers all monitored tags aggregated together. This is the default state.

**TA-3.** The tag filter applies to the same topic collections as the time period filter: unreplied topics, untagged topics, and resolved topics.

**TA-4.** The tag filter composes with the time period filter. Both filters are applied — a topic must satisfy both to appear.

**TA-5.** A topic matches a tag filter when its `tags` array includes the selected tag.

**TA-6.** When a tag is selected, the untagged topics list is empty — untagged topics have no tags and therefore cannot match any tag filter.

**TA-7.** Response time trends (RT-8) and weekly backlog trends (TD-23) are intentionally unfiltered by time period. However, the tag filter applies to these trend views — when a tag is selected, trends show only data for that tag. This asymmetry exists because tag selection is a scope decision (the user wants to examine one area), while time period is a window decision (trends need full history to show trajectory).

**TA-8.** The selected tag persists when the user navigates between pages. It does not reset on page switch.

### Area navigation (UC-16)

**TA-9.** The dashboard provides an area selector that groups the tag list by named areas defined in the configuration.

**TA-10.** When no area is selected, all monitored tags are visible in the tag selector.

**TA-11.** Selecting an area narrows the visible tag list to tags belonging to that area. It does not select a tag — the current tag selection (or lack of one) is preserved.

**TA-12.** Within an area, the primary tag appears first. Remaining tags are sorted alphabetically.

**TA-13.** When no area is selected, tags are sorted alphabetically across all areas.

**TA-14.** An "All areas" option is always available to return to the full tag list.

### Placement and defaults (TA-18 – TA-21)

**TA-18.** The tag and area selectors are rendered in the toolbar area alongside the period selector, visible on all pages.

**TA-19.** The default state on application load is: no area selected, no tag selected (showing all monitored tags aggregated).

**TA-20.** When the active tag or area filter yields no topics for a given list or metric, each list and metric shows its existing empty state. No additional empty-filter message is required.

**TA-21.** The tag selector indicates which tag is currently active. The area selector indicates which area is currently active.

---

## Design

### Configuration type

The tag configuration is loaded from `config/tagConfig.json`:

```typescript
interface AreaConfig {
  name: string;
  primaryTag: string;
  tags: string[];
}

type TagConfig = AreaConfig[];
```

### Filter functions

Defined in `tagFilter.ts`:

- `monitoredTags(config)` — returns the deduplicated set of all tags across all areas.
- `filterByTag(topics, tag)` — returns topics whose `tags` array includes the given tag. When `tag` is `null`, returns all topics unchanged.
- `filterByMonitoredTags(topics, monitored)` — returns topics that carry at least one monitored tag.
- `tagsForArea(config, area)` — returns the tag list for a given area with the primary tag first and the rest sorted alphabetically. When `area` is `null`, returns all monitored tags sorted alphabetically.

### Component

`TagSelector` renders the area dropdown and tag buttons. It is a pure function component — it holds no state. All state is managed in `App` and passed as props:

- `config: TagConfig` — the area/tag configuration.
- `activeTag: string | null` — the currently selected tag.
- `activeArea: string | null` — the currently selected area.
- `onTagSelect(tag: string | null)` — called when the user clicks a tag or clears the selection.
- `onAreaSelect(area: string | null)` — called when the user selects an area or "All areas".

### Placement

The `TagSelector` is rendered in `App.tsx` in the toolbar area, alongside `PeriodSelector`, below the header and above the page content.

### Data flow

`App.tsx` holds `activeTag` and `activeArea` state. It applies `filterByTag` after `filterByPeriod` to each topic list. When no tag is selected, `filterByMonitoredTags` is applied so that only topics with monitored tags appear in metrics.

Child component interfaces are unchanged — they continue to receive already-filtered `Topic[]` arrays.

### Component–requirement mapping

| Component | File | Requirements |
|-----------|------|-------------|
| `TagSelector` | [TagSelector.tsx](../../frontend/src/components/TagSelector.tsx) | TA-1 — tag selection; TA-9 – TA-14 — area navigation; TA-18 — placement; TA-21 — active indicators |
| `filterByTag` | [tagFilter.ts](../../frontend/src/components/tagFilter.ts) | TA-5 — match semantics; TA-6 — untagged excluded |
| `filterByMonitoredTags` | [tagFilter.ts](../../frontend/src/components/tagFilter.ts) | TA-17 — monitored tag scope |
| `monitoredTags` | [tagFilter.ts](../../frontend/src/components/tagFilter.ts) | TA-17 — extract monitored set |
| `tagsForArea` | [tagFilter.ts](../../frontend/src/components/tagFilter.ts) | TA-12, TA-13 — tag ordering |
| `App` | [App.tsx](../../frontend/src/App.tsx) | TA-2 — default aggregation; TA-3, TA-4 — filter composition; TA-7 — trend scoping; TA-8 — persistence; TA-19 — defaults |

---

## Validation

### Automated tests

Pure logic with well-defined inputs and outputs.

| What | Requirements | Rationale |
|------|-------------|-----------|
| `filterByTag` with a matching tag — includes topics carrying the tag | TA-5 | Core filter semantics. |
| `filterByTag` with a non-matching tag — excludes topics without the tag | TA-5 | Inverse correctness. |
| `filterByTag` with `null` — returns all topics unchanged | TA-2 | Default behavior. |
| `filterByTag` with multi-tag topics — includes when any tag matches | TA-5 | Multi-tag handling. |
| `filterByTag` does not mutate input array | TA-5 | Safety property. |
| `filterByMonitoredTags` — includes only topics with at least one monitored tag | TA-17 | Monitored scope. |
| `filterByMonitoredTags` — excludes untagged topics | TA-6, TA-17 | Untagged exclusion. |
| `filterByMonitoredTags` — excludes topics with only non-monitored tags | TA-17 | Non-monitored exclusion. |
| `monitoredTags` — returns deduplicated union of all area tags | TA-17 | Config extraction. |
| `monitoredTags` — returns empty set for empty config | TA-17 | Edge case. |
| `tagsForArea` with area — primary tag first, rest alphabetical | TA-12 | Area ordering. |
| `tagsForArea` with `null` — all tags alphabetical | TA-13 | Global ordering. |
| Composition: `filterByTag` after `filterByPeriod` — both constraints apply | TA-4 | Filter composition. |

Test location: `tests/dashboard/tag-area-filter.unit.test.ts`

### Manual verification

| What | Requirements | Rationale |
|------|-------------|-----------|
| Tag selector is visible on all pages | TA-18 | Placement is a layout concern. |
| Area selector shows all areas plus "All areas" | TA-9, TA-14 | Rendering and labels. |
| Selecting an area narrows the tag list without selecting a tag | TA-11 | Interaction behavior. |
| Selecting a tag updates all metrics and lists | TA-1, TA-3 | Cross-component effect. |
| Active tag and area are visually distinguished | TA-21 | Visual styling. |
| Switching pages preserves tag and area selection | TA-8 | State lifecycle. |
| Default on load is no tag, no area selected | TA-19 | Initial render state. |
| Tag filter composes with period filter visually | TA-4 | Cross-filter interaction. |
