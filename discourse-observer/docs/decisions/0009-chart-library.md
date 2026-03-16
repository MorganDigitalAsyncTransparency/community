# 9. Chart Library

**Status:** Proposed
**Date:** 2026-03-16

## Context

The dashboard frontend needs to visualize support trends across a hierarchical team structure. The organization is divided into areas (~5 currently, expected to grow), each can contain about ~10 teams, sometimes no team, only area. The dashboard operates at three levels: top (all areas aggregated), area (teams within one area), and individual team. Graphs appear at each level, meaning the typical view shows ~10 series (teams within an area) — not all 30+ teams simultaneously.

The existing implementation (UC-8) renders weekly trend data as HTML tables, which is sufficient for raw numbers but does not communicate trends, comparisons, or distributions effectively. The goal is to show how each team's support performance evolves week by week: how many topics are opened and closed, how fast issues are resolved, and how long things stay open.

The visualization needs fall into two categories:

**Current requirements:**

- Line charts showing weekly trends per team (~10 series): topic volume (opened/closed), resolution time, time-to-first-reply, open duration
- Pie charts showing proportional distribution within a level (e.g., share of resolved, missed, or incoming topics across teams in an area, or distribution between areas)
- Hover tooltips for exact values
- Legend interaction to toggle individual series on/off
- Time period filtering (already implemented via UC-12)

**Anticipated requirements:**

- Stacked area charts for volume distribution over time (seeing total volume and per-team breakdown simultaneously)
- Bar charts for side-by-side team/area comparison at a point in time
- Dual y-axes (volume on left, duration on right) in a single chart
- Export charts as images for reports and presentations
- Responsive layout for varying screen sizes
- Annotations to mark events in the timeline (process changes, holidays) that explain anomalies

The project uses React 19.2.4, Vite 8, and TypeScript 5.9. There are no existing charting dependencies. A future design specification will define visual styling for the full site, so the library should not impose a rigid visual identity but should be possible to theme.

## Alternatives Considered

### Recharts

A declarative, React-component-based charting library built on D3 submodules. Charts are composed from JSX components (`<LineChart>`, `<Line>`, `<XAxis>`, `<Tooltip>`, `<Legend>`).

**Strengths:**

- Idiomatic React API — charts are component trees, not configuration objects
- Covers line, bar, area, pie, and composed charts out of the box
- Built-in responsive container component
- Legend click-to-toggle works without custom code
- Most actively maintained option — latest release 3.8.0 on 2026-03-06
- ~24k GitHub stars
- MIT license

**Weaknesses:**

- Moderate bundle size
- No built-in annotation support — requires custom reference lines or shapes
- No built-in image export — requires supplementary library (e.g., html-to-image)
- Dual y-axes supported but documentation is sparse

### Chart.js (via react-chartjs-2)

A canvas-based charting library with a React wrapper. Configuration is passed as a JavaScript options object; the wrapper translates React props into imperative Chart.js calls.

**Strengths:**

- Broad chart type coverage including line, bar, area, pie, and mixed charts
- Built-in annotation support via the chartjs-plugin-annotation plugin
- Canvas rendering performs well with many data points
- Large community (~65k GitHub stars for Chart.js)
- Chart.js 4.5.1 (2025-10-13), react-chartjs-2 5.3.1 — React 19 supported from 5.3.0
- MIT license

**Weaknesses:**

- Two dependencies to maintain (Chart.js + react-chartjs-2) with separate release cycles
- Imperative API under a React wrapper — configuration is an options object, not a component tree
- Canvas rendering gives React no control over individual chart elements
- Styling is configured through the Chart.js options API, not CSS
- Responsive by default, but resizing behavior can be surprising with canvas

### Nivo

A rich visualization library built on D3, designed specifically for React. Provides high-level components with extensive theming and animation support.

**Strengths:**

- Strong React integration — built for React from the ground up
- Rich theming system with a centralized theme object
- Supports line, bar, area, pie, and many specialized chart types
- Interactive features (tooltips, legends) included
- Server-side rendering support
- MIT license

**Weaknesses:**

- ~80–120 kB gzip depending on which chart packages are imported — largest option
- Pulls in large portions of D3 as transitive dependencies
- Configuration-heavy — many options to learn before getting a chart right
- Annotation support is limited
- Latest release 0.99.0 on 2025-05-23 (~10 months ago), ~14k GitHub stars
- Image export not built-in

### Victory

A declarative, component-based charting library built by Nearform (formerly Formidable). Designed for React and React Native.

**Strengths:**

- Declarative React component API similar to Recharts
- Supports line, bar, area, pie charts and compositions
- Built-in animation system
- React Native support (not needed here, but indicates deep React integration)
- MIT license

**Weaknesses:**

- Latest release 37.3.6 on 2025-01-14 (~14 months ago) — maintenance has slowed significantly
- ~11k GitHub stars
- No built-in annotation or image export support
- React Native abstractions add weight without benefit for a web-only project
- Dual y-axis support requires manual axis configuration

### visx (Airbnb)

Low-level visualization primitives that combine D3 utilities with React components. Not a charting library — a toolkit for building charts.

**Strengths:**

- Maximum flexibility — full control over every visual element
- Small per-package bundle size (import only what you use)
- Truly React-native — every element is a React component rendering SVG
- ~19k GitHub stars, though latest release 3.12.0 was 2024-11-07 (~16 months ago)
- Perfect for custom, bespoke visualizations
- MIT license

**Weaknesses:**

- No ready-made charts — line chart, tooltips, legends, and interactions must be built from primitives
- Significant development effort for every chart type
- No built-in legend toggle, annotations, or image export
- Requires D3 knowledge for scales, axes, and data transforms
- The flexibility becomes a maintenance burden when the need is standard chart types

### Apache ECharts (via echarts-for-react)

A comprehensive, canvas-based visualization library originally developed by Baidu. One of the most feature-complete options available.

**Strengths:**

- Extremely broad chart type support — covers every type in our requirements list
- Built-in annotation (markLine, markArea), dual y-axes, and image export (saveAsImage)
- Built-in legend toggle, rich tooltip formatting, responsive resize
- Large global community (~63k GitHub stars), latest release 6.0.0 on 2025-07-30
- Strong performance with large datasets via canvas rendering
- Tree-shakeable — selective imports can significantly reduce bundle size
- Apache 2.0 license

**Weaknesses:**

- ~100 kB gzip for a typical build (tree-shakeable but core is large)
- Configuration-driven API — charts are defined as option objects, not React component trees
- React wrapper (echarts-for-react) is a thin layer, not an idiomatic React integration
- Learning curve for the configuration schema, which is broad and deep
- Documentation is extensive but originally Chinese-first; English docs have improved but can be uneven

## Decision

*To be decided.*

## Consequences

*To be determined after decision.*

## Appendix: Requirements Coverage

| Requirement | Recharts | Chart.js | Nivo | Victory | visx | ECharts |
|---|---|---|---|---|---|---|
| Line chart (~10 series) | Yes | Yes | Yes | Yes | Build | Yes |
| Pie chart | Yes | Yes | Yes | Yes | Build | Yes |
| Hover tooltips | Yes | Yes | Yes | Yes | Build | Yes |
| Legend toggle | Yes | Plugin | Yes | Yes | Build | Yes |
| Stacked area chart | Yes | Yes | Yes | Yes | Build | Yes |
| Bar chart | Yes | Yes | Yes | Yes | Build | Yes |
| Dual y-axes | Partial | Yes | Partial | Manual | Build | Yes |
| Image export | No | Plugin | No | No | No | Yes |
| Responsive | Yes | Yes | Yes | Yes | Manual | Yes |
| Annotations | No | Plugin | Limited | No | Build | Yes |
| React-idiomatic API | High | Low | High | High | High | Low |
| License | MIT | MIT | MIT | MIT | MIT | Apache 2.0 |

### Package Versions and Activity

Data retrieved from npm registry on 2026-03-16.

| Library | Version | Latest release | React 19 |
|---|---|---|---|
| Recharts | 3.8.0 | 2026-03-06 | Yes |
| Chart.js + react-chartjs-2 | 4.5.1 + 5.3.1 | 2025-10-13 | Yes (from 5.3.0) |
| Nivo (@nivo/line, /pie, /core) | 0.99.0 | 2025-05-23 | Yes (from 0.98.0) |
| Victory | 37.3.6 | 2025-01-14 | Unverified |
| visx (@visx/visx) | 3.12.0 | 2024-11-07 | Unverified |
| ECharts + echarts-for-react | 6.0.0 + 3.0.6 | 2025-07-30 | Yes |
