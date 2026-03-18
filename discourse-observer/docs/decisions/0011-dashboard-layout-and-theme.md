# 11. Dashboard Layout and Theme

**Status:** Proposed
**Date:** 2026-03-18

## Context

The discourse-observer dashboard is a six-page application displaying support metrics through tables, charts, summary cards, and a heatmap. The current frontend ([App.tsx](../../frontend/src/App.tsx), [App.css](../../frontend/src/App.css)) was built feature-by-feature without a deliberate layout or theming strategy. Every color, spacing value, and font size is hardcoded directly in CSS rules — approximately 1000 lines of ad-hoc styling.

This creates two problems:

1. **Rebranding requires editing hundreds of lines.** Changing the background color means finding and replacing every `#f5f6f8`, `#fff`, `#f0f1f3`, and related values across the file. There is no single place to define "what does this dashboard look like."

2. **Layout structure is implicit.** The header, navigation, filters, content area, and (missing) footer are positioned through scattered margin/padding rules rather than an explicit layout grid. Adding a new region (e.g., a footer, a sidebar, or a status bar) means reverse-engineering how the current regions relate to each other.

This ADR evaluates alternatives for two related decisions:

- **Layout structure:** How should the dashboard's visual regions — header, navigation, filters, content, footer — be arranged?
- **Theme architecture:** How should visual properties (colors, typography, spacing, radii) be organized so that rebranding requires changing variable values, not component code?

### Design reference

The layout evaluation draws inspiration from Grafana dashboards: dense but readable layouts, clear section boundaries, compact filter bars, and dark/light theme support. The goal is not to copy Grafana's visual identity but to achieve the same information density and clarity.

### Current layout

The existing structure, as rendered from [App.tsx](../../frontend/src/App.tsx):

```
┌──────────────────────────────────────────────────┐
│  Title            Nav tabs          Sync status   │  header (flex row)
├──────────────────────────────────────────────────┤
│  Period: [7d] [30d] [90d] [All] [Custom]         │  period selector
│  Area: [▾]  Tags: [tag1] [tag2] [All]            │  tag selector
│  [Clear all filters]                              │  conditional
├──────────────────────────────────────────────────┤
│                                                  │
│              Page content                         │  main (flex column)
│              (tables, charts, cards)              │
│                                                  │
└──────────────────────────────────────────────────┘
```

Properties: single-column, `max-width: 960px`, centered. No footer. Navigation is inline with the title. Filters are not sticky — they scroll out of view.

### Constraints

- Six pages today, potentially more as the system grows.
- Two global filters (period, tag/area) apply across all pages.
- Primary audience uses desktop screens (often large monitors), but the layout should degrade gracefully on smaller screens.
- The dashboard will be branded differently per deployment — colors, fonts, and logos change, but structure does not. The branding model is fork-based: a deployment forks the repository and replaces visual tokens. There is no runtime theme switcher in the application itself.
- The project uses plain CSS (no preprocessor, no CSS-in-JS). ADR 0002 favors minimal tooling.

## Alternatives Considered

### Layout

#### A. Topbar with sticky filter row

The header occupies the full width at the top, containing logo, navigation tabs, theme toggle, and sync status in one or two rows. Below it, a sticky filter bar holds period and tag selectors. Content fills the remaining viewport height.

```
┌──────────────────────────────────────────────────┐
│  Logo              Nav tabs        [☀/☾]   Sync  │  topbar
├──────────────────────────────────────────────────┤
│  Period: [7d] [30d] [All]   Tags: [...]  [Clear] │  sticky filter bar
├──────────────────────────────────────────────────┤
│                                                  │
│              Page content (scrollable)            │
│                                                  │
├──────────────────────────────────────────────────┤
│  v0.1.0 · Last synced 14:32 UTC · GitHub ↗      │  footer
└──────────────────────────────────────────────────┘
```

**Strengths:**

- Familiar pattern from Grafana and most analytics dashboards — users know where to look.
- Filter bar is always visible during scroll, reducing interaction cost.
- Simple CSS: stacked flex or grid rows. No sidebar breakpoint logic.
- Content area gets full width, which benefits wide tables and charts.
- Low implementation complexity — closest to the current structure.

**Weaknesses:**

- Navigation competes for horizontal space with logo and status. With 6+ tabs this works; with 10+ it could overflow.
- No persistent indication of "where am I" — the active tab highlight is the only cue, and it shares the row with other elements.
- Adding grouped navigation (e.g., "Operations" / "Analytics" sections) requires either dropdowns or a second row, complicating the header.

#### B. Sidebar navigation with header and filter bar

A fixed-width sidebar on the left holds the logo, navigation links, and theme toggle. The main area has a header strip with filters and a content region below. The sidebar provides persistent context about the current page.

```
┌──────────┬───────────────────────────────────────┐
│  Logo    │  Period: [7d] [30d]  Tags: [..]  Clear│  filter bar
│          ├───────────────────────────────────────┤
│  • Queue │                                       │
│    Resp. │          Page content (scrollable)     │
│    Dist  │                                       │
│    SLO   │                                       │
│    Vol.  │                                       │
│    Activ.│                                       │
│          ├───────────────────────────────────────┤
│  [☀/☾]  │  v0.1.0 · Synced 14:32 UTC · GitHub ↗│  footer
└──────────┴───────────────────────────────────────┘
```

**Strengths:**

- Persistent "where am I" — the active page is always visible in the sidebar, even while scrolled deep into content.
- Natural grouping for navigation — sections, dividers, or collapsible groups are trivial in a vertical list.
- Scales to more pages without layout pressure — a vertical list accommodates 20+ items.
- Familiar from Grafana, GitHub, and most admin/monitoring tools.
- Sidebar width is fixed, so content-area width is predictable for table column sizing.

**Weaknesses:**

- Sidebar consumes ~200px of horizontal space on every screen size. On a 1280px laptop, content gets ~1080px — still generous, but less than full-width.
- Requires a responsive breakpoint: below ~768px the sidebar should collapse to a hamburger menu or a top bar. This adds a media query and a toggle interaction.
- Slightly more CSS than the topbar alternative: grid template, sidebar styling, collapse behavior. Not complex, but more than zero.

#### C. Compact single-row header, full-width content

Everything — logo, navigation, filters, status — compressed into one or two narrow rows at the top. No sidebar. Content gets maximum vertical and horizontal space.

```
┌──────────────────────────────────────────────────┐
│ Logo Queue│Resp│Dist│SLO│Vol│Act [☀/☾] Sync:14:32│  single header row
├──────────────────────────────────────────────────┤
│ Period: [7d] [30d]  Tags: [support] [billing]    │  filter row
├──────────────────────────────────────────────────┤
│                                                  │
│              Page content                         │
│                                                  │
├──────────────────────────────────────────────────┤
│  v0.1.0 · GitHub ↗                              │  minimal footer
└──────────────────────────────────────────────────┘
```

**Strengths:**

- Maximum content area — no sidebar, minimal header height.
- Simplest CSS of all alternatives.
- Nearly identical to the current structure, making migration trivial.

**Weaknesses:**

- Dense header becomes cramped with 6+ navigation items plus filters plus status in limited horizontal space.
- No persistent "where am I" — identical weakness to Alternative A, but worse because the tab labels are shorter and harder to scan.
- No room for navigation grouping or future page additions without overflow.
- Essentially a tighter version of the current layout — solves the theming problem but not the navigation clarity problem.

### Theme architecture

#### I. CSS custom properties on `:root`

Define all visual tokens as `--variable-name` on `:root`. Theme switching overrides these variables on a `[data-theme="dark"]` selector. Components reference only variables, never literal values.

```css
:root {
  --color-bg-primary: #f5f6f8;
  --color-text-primary: #1a1a1a;
  --spacing-md: 1rem;
  /* ... */
}

[data-theme="dark"] {
  --color-bg-primary: #1a1c22;
  --color-text-primary: #d8d9da;
}
```

**Strengths:**

- Native CSS — no build step, no preprocessor, no runtime cost.
- Rebranding is a single variable block change.
- Dark/light switching is instant (attribute toggle, no class list manipulation needed).
- Consistent with the project's minimal-tooling philosophy (ADR 0002).
- Variables cascade naturally — a component can override a variable locally without affecting siblings.

**Weaknesses:**

- No compile-time validation — a typo in a variable name silently falls back to the property's initial value.
- IDE support for variable autocomplete varies across editors.
- Cannot express computed values (e.g., "half of `--spacing-md`") without `calc()`.

#### II. CSS preprocessor variables (Sass/Less)

Use Sass `$variables` or Less `@variables` with separate theme files compiled into distinct CSS bundles.

**Strengths:**

- Compile-time resolution — typos are caught at build time.
- Rich expression support (math, functions, mixins).
- Familiar to teams with Sass/Less experience.

**Weaknesses:**

- Introduces a build dependency the project does not currently have (ADR 0002 chose plain CSS, Vite, and minimal tooling).
- Theme switching requires loading a different CSS file or toggling class-scoped overrides — no runtime variable cascade.
- Cannot switch themes without either a page reload or maintaining two compiled sheets in the DOM.

#### III. CSS-in-JS theme provider (styled-components, Emotion)

Define theme tokens as a JavaScript object. A React context provider distributes them. Components access tokens via template literals or props.

**Strengths:**

- Full type safety when used with TypeScript.
- Dynamic theming with runtime variable access.
- Co-locates styles with components.

**Weaknesses:**

- Adds a runtime dependency and a build-time dependency.
- Contradicts the project's existing approach: all current styling is in plain CSS files with BEM-like class naming.
- Migration from 1000 lines of plain CSS to CSS-in-JS is a rewrite, not a refactor.
- Runtime style injection has a performance cost on initial render.

### Content width strategy

#### Fixed max-width

Content area uses `max-width` (e.g., 1400px) with `margin: 0 auto` centering. Prevents content from becoming unreadably wide on ultra-wide monitors.

**Strengths:** Readable line lengths. Predictable table column widths. Consistent appearance across screen sizes.

**Weaknesses:** Wastes screen space on very large monitors. May feel restrictive if future components need full viewport width.

#### Fluid full-width

Content area expands to fill available viewport width with only padding constraints.

**Strengths:** Maximum use of available space. Tables and charts can spread out.

**Weaknesses:** Text and sparse content become hard to scan on wide screens. Table rows stretch to uncomfortable widths on ultrawide monitors (45"+).

#### Adaptive max-width

Content area uses `max-width` that increases at defined breakpoints, or individual components can opt into full width while text sections remain constrained.

**Strengths:** Balances readability with space utilization. Tables get full width while text stays constrained.

**Weaknesses:** More CSS to maintain. Breakpoint definitions require testing across screen sizes.

## Decision

**Theme architecture — CSS custom properties (Alternative I).** The project already uses plain CSS without a preprocessor (ADR 0002). CSS custom properties solve the branding problem natively — a fork replaces one block of `--variable` values on `:root` and the entire dashboard follows. No build dependency, no runtime cost, no component code changes.

Sass/Less (Alternative II) would add a build-time dependency to solve the same problem. CSS-in-JS (Alternative III) would require rewriting 1000 lines of plain CSS into a different paradigm and add a runtime dependency. Neither is justified when the branding model is fork-based with no runtime theme switching.

**Layout and content width** — pending. These decisions have not been made.

## Consequences

*To be determined after the decision is made.*
