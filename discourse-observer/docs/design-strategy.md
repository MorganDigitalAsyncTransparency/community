# Design Strategy

This document defines the dashboard's visual structure: layout regions, CSS variable taxonomy, sidebar behavior, and responsive breakpoints. It is the implementation reference for [ADR 0011](decisions/0011-dashboard-layout-and-theme.md).

The goal is a dashboard that can be rebranded by changing variable values in one place, without modifying component code.

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
  grid-template-columns: var(--sidebar-width) 1fr;
  grid-template-rows: auto 1fr auto;
  grid-template-areas:
    "sidebar  filter-bar"
    "sidebar  content"
    "sidebar  footer";
  min-height: 100vh;
}
```

The content area centers its children with `max-width: 1400px` and horizontal auto-margins.

---

## Regions

### Sidebar

The sidebar holds the logo, navigation links, and a collapse toggle. It spans the full viewport height and stays fixed during scroll.

| State | Width | Content |
|-------|-------|---------|
| Expanded | `var(--sidebar-width-expanded)` · 200px | Icon + text label per page |
| Collapsed | `var(--sidebar-width-collapsed)` · 48px | Icon only, tooltip on hover |

The `--sidebar-width` variable is toggled between expanded and collapsed values. The grid column follows automatically.

**Navigation items** are vertical, one per row. The active page is visually distinguished (e.g., background highlight, left border accent). Navigation can be grouped with section dividers or headings as pages grow.

**Collapse toggle** sits at the bottom of the sidebar. It switches the sidebar between expanded and collapsed states. The user's preference can be persisted in `localStorage`.

### Filter bar

The filter bar sits above the content area and is sticky — it remains visible when the content scrolls. It contains:

- Period selector (preset buttons + custom date range)
- Tag/area selector (area dropdown + tag buttons)
- Clear all filters button (conditional)

```css
.filter-bar {
  grid-area: filter-bar;
  position: sticky;
  top: 0;
  z-index: var(--z-filter-bar);
  background: var(--color-bg-primary);
  border-bottom: 1px solid var(--color-border-subtle);
  padding: var(--spacing-sm) var(--spacing-md);
}
```

### Content

The scrollable area for page-specific components (tables, charts, cards). Content is constrained to `max-width: 1400px` and centered within the grid cell.

```css
.content {
  grid-area: content;
  overflow-y: auto;
  padding: var(--spacing-lg) var(--spacing-md);
}

.content-inner {
  max-width: 1400px;
  margin: 0 auto;
}
```

### Footer

The footer sits at the bottom of the main area. It contains version, last sync time, and a repository link.

```css
.footer {
  grid-area: footer;
  border-top: 1px solid var(--color-border-subtle);
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}
```

Content example: `v0.1.0 · Last synced 2026-03-18 14:32 UTC · GitHub ↗`

---

## Responsive behavior

| Breakpoint | Sidebar | Layout |
|------------|---------|--------|
| ≥ 1024px | Expanded or collapsed (user choice) | Two-column grid |
| 768–1023px | Collapsed (icon only) | Two-column grid |
| < 768px | Hidden, hamburger toggle | Single column, filter bar below hamburger |

Mobile is not a priority. The layout should degrade gracefully rather than break.

---

## CSS variable taxonomy

All visual tokens live on `:root`. A fork replaces this block to rebrand the dashboard.

### Naming convention

```text
--{category}-{target}-{variant}
```

- **category**: `color`, `font`, `spacing`, `radius`, `shadow`, `z`
- **target**: what it applies to (`bg`, `text`, `border`, `sidebar`, `filter-bar`)
- **variant**: optional modifier (`primary`, `secondary`, `subtle`, `hover`, `active`)

### Colors

```css
:root {
  /* Backgrounds */
  --color-bg-primary: #f5f6f8;
  --color-bg-surface: #ffffff;
  --color-bg-sidebar: #1a1c22;
  --color-bg-hover: #eef0f4;
  --color-bg-active: #e2e5ea;

  /* Text */
  --color-text-primary: #1a1a1a;
  --color-text-secondary: #666666;
  --color-text-muted: #888888;
  --color-text-sidebar: #d8d9da;
  --color-text-sidebar-active: #ffffff;

  /* Borders */
  --color-border-default: #dddddd;
  --color-border-subtle: #eeeeee;

  /* Accent */
  --color-accent-primary: #3b82f6;
  --color-accent-primary-hover: #2563eb;

  /* Semantic */
  --color-status-warning: #b08000;
  --color-status-error: #e55555;

  /* Data visualization */
  --color-heatmap-base: 59, 130, 246;  /* RGB triplet for alpha variation */
  --color-chart-1: #8884d8;
  --color-chart-2: #82ca9d;
  --color-chart-3: #5b8ff9;

  /* Table */
  --color-table-header-bg: #f0f1f3;
  --color-table-row-alt: #fafbfc;
  --color-table-row-hover: #f0f4ff;
}
```

### Typography

```css
:root {
  --font-family-base: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
    Helvetica, Arial, sans-serif;

  --font-size-xs: 0.7rem;
  --font-size-sm: 0.85rem;
  --font-size-base: 0.9rem;
  --font-size-md: 1rem;
  --font-size-lg: 1.15rem;
  --font-size-xl: 1.5rem;
  --font-size-display: 2rem;

  --font-weight-normal: 400;
  --font-weight-medium: 600;
  --font-weight-bold: 700;

  --line-height-base: 1.5;
  --line-height-tight: 1.2;
}
```

### Spacing

An 8px-based scale. Each step doubles or halves the base.

```css
:root {
  --spacing-xs: 0.25rem;   /*  4px */
  --spacing-sm: 0.5rem;    /*  8px */
  --spacing-md: 1rem;      /* 16px */
  --spacing-lg: 1.5rem;    /* 24px */
  --spacing-xl: 2rem;      /* 32px */
}
```

### Borders and radii

```css
:root {
  --radius-sm: 4px;
  --radius-md: 6px;
  --radius-lg: 8px;
}
```

### Shadows

```css
:root {
  --shadow-dialog: 0 4px 24px rgb(0 0 0 / 15%);
}
```

### Z-index

```css
:root {
  --z-sidebar: 100;
  --z-filter-bar: 50;
  --z-dialog-backdrop: 1000;
}
```

### Sidebar dimensions

```css
:root {
  --sidebar-width-expanded: 200px;
  --sidebar-width-collapsed: 48px;
  --sidebar-width: var(--sidebar-width-expanded);
}
```

---

## Branding

To rebrand a fork:

1. Copy the `:root` variable block into a new file (e.g., `theme-acme.css`).
2. Replace color, font, and spacing values.
3. Import the theme file after the base stylesheet, or replace the `:root` block directly.

No component code changes. No class name changes. The variable names are the contract between structure and style.
