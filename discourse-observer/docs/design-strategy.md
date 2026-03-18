# Design Strategy

This document defines the dashboard's visual structure: layout regions, CSS variable taxonomy, sidebar behavior, and responsive breakpoints. It is the implementation reference for [ADR 0011](decisions/0011-dashboard-layout-and-theme.md).

The goal is a dashboard that can be rebranded by changing variable values in one place, without modifying component code. This document defines the **names and semantics** of those variables. The actual values live in the CSS file — not here — so there is exactly one source of truth.

---

## Layout structure

The page shell is a CSS Grid with two columns: a collapsible sidebar and a main area. The main area is subdivided into three rows: filter bar, content, and footer.

```text
┌──────────┬───────────────────────────────────────┐
│          │            filter-bar                  │
│          ├───────────────────────────────────────┤
│ sidebar  │                                       │
│          │            content                     │
│          │                                       │
│          ├───────────────────────────────────────┤
│          │            footer                      │
└──────────┴───────────────────────────────────────┘
```

Grid definition:

```css
.shell {
  display: grid;
  grid-template: auto 1fr auto / auto 1fr;
  grid-template-areas:
    "sidebar  filter-bar"
    "sidebar  content"
    "sidebar  footer";
  min-height: 100vh;
}
```

---

## Regions

### Sidebar

The sidebar holds the logo, navigation links, and a collapse toggle. It spans the full viewport height. Because the content area handles its own scrolling (via `overflow-y: auto`), the sidebar stays visually fixed without needing `position: fixed`.

| State | Width | Content |
|-------|-------|---------|
| Expanded | `var(--sidebar-width-expanded)` · 200px | Icon + text label per page |
| Collapsed | `var(--sidebar-width-collapsed)` · 48px | Icon only, tooltip on hover |

The sidebar's CSS class toggles between expanded and collapsed widths. The grid column uses `auto` sizing so it follows the sidebar's actual width automatically. Transition between states is animated (`width` transition, ~200ms ease) to provide visual continuity.

If navigation items exceed the sidebar height, the sidebar scrolls independently (`overflow-y: auto`).

**Navigation items** are vertical, one per row. The active page is visually distinguished (background highlight or left border accent). Navigation can be grouped with section dividers or headings as pages grow.

**Collapse toggle** sits at the bottom of the sidebar. The user's preference is persisted in `localStorage`.

### Filter bar

The filter bar sits above the content area. It does not scroll — the grid row is `auto`-sized, so the filter bar stays at the top while the content area scrolls independently. It contains:

- Period selector (preset buttons + custom date range)
- Tag/area selector (area dropdown + tag buttons)
- Clear all filters button (conditional)

Note: `position: sticky` is not needed. The grid layout keeps the filter bar fixed — only the content row scrolls.

### Content

The scrollable area for page-specific components (tables, charts, cards). Content scrolls independently within its grid cell and constrains width to `var(--content-max-width)` (1400px), applied to the filter bar, content, and footer regions.

### Footer

Sits at the bottom of the main area. Contains version, last sync time, and a repository link.

Content example: `v0.1.0 · Last synced 2026-03-18 14:32 UTC · GitHub ↗`

---

## Responsive behavior

| Breakpoint | Sidebar | Layout |
|------------|---------|--------|
| ≥ 1024px | Expanded or collapsed (user choice) | Two-column grid |
| 768–1023px | Collapsed (icon only) | Two-column grid |
| < 768px | Hidden by default; hamburger button (fixed, top-left) opens sidebar as a full-height overlay with backdrop | Single column, filter bar below hamburger |

At the mobile breakpoint, a fixed hamburger button (`☰`) appears at the top-left corner. Tapping it opens the sidebar as a fixed overlay above a semi-transparent backdrop. Tapping the backdrop or any navigation link closes the overlay. The hamburger button and overlay styles are scoped to the `<= 767px` media query — they have no effect at wider viewports.

Mobile is not a priority. The layout should degrade gracefully rather than break.

---

## CSS variable taxonomy

All visual tokens live on `:root` in the stylesheet. A fork replaces the values to rebrand the dashboard. This document defines the variable **names and semantics** — the stylesheet is the source of truth for values.

### Naming convention

```text
--{category}-{target}-{variant}
```

- **category**: `color`, `font`, `spacing`, `radius`, `shadow`, `z`, `sidebar`
- **target**: what it applies to (`bg`, `text`, `border`, `size`, `weight`, `width`)
- **variant**: optional modifier (`primary`, `secondary`, `muted`, `hover`, `active`, `disabled`, `expanded`, `collapsed`)

Layout dimensions like `--sidebar-width-expanded` use the component name as category. This keeps layout tokens distinct from design tokens while following the same naming shape.

### Colors

#### Backgrounds

| Variable | Semantic meaning |
|----------|-----------------|
| `--color-bg-primary` | Page background — the base canvas |
| `--color-bg-surface` | Card/table background — raised above the canvas |
| `--color-bg-raised` | Slightly elevated surface (picker panels, consent dialogs) |
| `--color-bg-sidebar` | Sidebar background — typically darker than the page |
| `--color-bg-sidebar-hover` | Sidebar interactive element on hover (semi-transparent overlay) |
| `--color-bg-sidebar-active` | Sidebar active/selected item background (semi-transparent overlay) |
| `--color-bg-hover` | Interactive element on hover |
| `--color-bg-active` | Interactive element while pressed or selected |

#### Text

| Variable | Semantic meaning |
|----------|-----------------|
| `--color-text-primary` | Default body text — highest contrast against bg-primary |
| `--color-text-secondary` | De-emphasized text (labels, captions, table headers) |
| `--color-text-muted` | Lowest-emphasis text (empty states, hints, timestamps) |
| `--color-text-disabled` | Disabled controls — must still be legible but clearly inactive |
| `--color-text-sidebar` | Sidebar text on the sidebar background |
| `--color-text-sidebar-active` | Active sidebar item — highest contrast on sidebar background |
| `--color-text-on-accent` | Text on accent-colored backgrounds (buttons, badges) |

#### Borders

| Variable | Semantic meaning |
|----------|-----------------|
| `--color-border-default` | Standard border (cards, inputs, tables) |
| `--color-border-subtle` | Light separator (table rows, section dividers) |
| `--color-border-strong` | Emphasized separator (section boundaries, UTC header row) |
| `--color-border-active` | Border for active/selected state (active filter buttons) |
| `--color-border-disabled` | Border for disabled controls |
| `--color-border-sidebar` | Sidebar internal border (collapse toggle separator) |

#### Accent and semantic

| Variable | Semantic meaning |
|----------|-----------------|
| `--color-accent-primary` | Primary action color (links, active indicators) |
| `--color-accent-primary-hover` | Hover state of primary accent |
| `--color-status-warning` | Warning indicators (default SLO, fallback states) |
| `--color-status-error` | Error and destructive actions (remove buttons, violations) |

#### Data visualization

| Variable | Semantic meaning |
|----------|-----------------|
| `--color-heatmap-base` | RGB triplet (no `#`) for heatmap alpha variation: `rgba(var(--color-heatmap-base) / α)` |
| `--color-chart-1` | First chart series (e.g., time to first reply) |
| `--color-chart-2` | Second chart series (e.g., resolution time) |
| `--color-chart-3` | Third chart series (e.g., intake volume) |

#### Table

| Variable | Semantic meaning |
|----------|-----------------|
| `--color-table-header-bg` | Table header row background |
| `--color-table-row-alt` | Alternating (zebra) row background |
| `--color-table-row-hover` | Row hover highlight |

### Contrast pairs

When rebranding, these pairs must maintain sufficient contrast. Changing one without checking the other will produce unreadable text or invisible borders.

| Foreground | Background | Context |
|------------|------------|---------|
| `--color-text-primary` | `--color-bg-primary` | Body text on page |
| `--color-text-primary` | `--color-bg-surface` | Body text on cards/tables |
| `--color-text-secondary` | `--color-bg-primary` | Labels and captions on page |
| `--color-text-secondary` | `--color-bg-surface` | Table headers, labels in cards |
| `--color-text-muted` | `--color-bg-primary` | Timestamps and hints on page |
| `--color-text-muted` | `--color-bg-surface` | Empty state messages in cards |
| `--color-text-disabled` | `--color-bg-surface` | Disabled controls in cards |
| `--color-text-sidebar` | `--color-bg-sidebar` | Sidebar navigation text |
| `--color-text-sidebar` | `--color-bg-sidebar-hover` | Sidebar text on hovered item |
| `--color-text-sidebar-active` | `--color-bg-sidebar-active` | Active sidebar item |
| `--color-text-on-accent` | `--color-accent-primary` | Text on accent buttons |
| `--color-status-warning` | `--color-bg-surface` | SLO warning indicators in tables |
| `--color-status-error` | `--color-bg-surface` | Error markers and remove buttons |

### Typography

| Variable | Semantic meaning |
|----------|-----------------|
| `--font-family-base` | System font stack for all text |
| `--font-size-xs` | Smallest text (heatmap cells, timezone labels) |
| `--font-size-sm` | Small text (filter labels, table headers, footer) |
| `--font-size-md` | Default body text and table cells |
| `--font-size-lg` | Section headings |
| `--font-size-xl` | Page title |
| `--font-size-display` | Hero numbers (summary card values, metric cards) |
| `--font-weight-normal` | Default weight |
| `--font-weight-medium` | Emphasis (table headers, labels, active nav) |
| `--font-weight-bold` | Strong emphasis (card values, page title) |
| `--line-height-base` | Default line height |
| `--line-height-tight` | Compact line height (card values, display numbers) |

### Spacing

A 4px-based scale: 4, 8, 16, 24, 32.

| Variable | Semantic meaning |
|----------|-----------------|
| `--spacing-xs` | Tight padding (heatmap cells, compact controls) |
| `--spacing-sm` | Small padding (filter bar, footer, button padding) |
| `--spacing-md` | Standard padding (horizontal content padding, gaps) |
| `--spacing-lg` | Section spacing (vertical content padding) |
| `--spacing-xl` | Large separation (between major sections) |

### Borders and radii

| Variable | Semantic meaning |
|----------|-----------------|
| `--radius-sm` | Buttons, inputs, small controls |
| `--radius-md` | Cards, tables, dropdowns |
| `--radius-lg` | Dialogs, large panels |

### Shadows

| Variable | Semantic meaning |
|----------|-----------------|
| `--shadow-dialog` | Modal/dialog elevation |

### Z-index

Stacking order from back to front:

| Layer | Variable | Value | Reason |
|-------|----------|-------|--------|
| Filter bar | `--z-filter-bar` | 50 | Above content scroll area |
| Sidebar | `--z-sidebar` | 100 | Overlays filter bar and content when expanded |
| Dropdowns | `--z-dropdown` | 200 | Timezone picker and similar overlays |
| Dialog backdrop | `--z-dialog-backdrop` | 1000 | Modal overlays — above everything |

Z-index values are defined in the document (not just the stylesheet) because they express a stacking contract that CSS alone does not make obvious.

### Layout dimensions

Layout dimension values are defined in the document (not just the stylesheet) because they are structural constraints that affect the grid layout, responsive breakpoints, and content sizing. Changing them has cascading consequences beyond visual appearance.

| Variable | Semantic meaning | Value |
|----------|-----------------|-------|
| `--sidebar-width-expanded` | Sidebar width when showing icon + text | 200px |
| `--sidebar-width-collapsed` | Sidebar width when showing icon only | 48px |
| `--content-max-width` | Maximum width of the filter bar, content area, and footer | 1400px |

---

## Branding

To rebrand a fork:

1. Open the stylesheet and locate the `:root` block.
2. Replace color, font, and spacing values.
3. Verify contrast pairs (see table above) — changing a background without adjusting its text color will break readability.
4. Test chart colors for accessibility — data visualization colors must be distinguishable for color-blind users.

No class name changes. The variable names are the contract between structure and style.

**Note on chart colors:** Chart libraries (Recharts) receive colors as JavaScript props, not via CSS inheritance. Changing `--color-chart-*` or `--color-heatmap-base` in the stylesheet requires a JavaScript bridge that reads the computed variable values and passes them to chart components. This bridge is an implementation concern — but be aware that chart rebranding is not pure CSS.
