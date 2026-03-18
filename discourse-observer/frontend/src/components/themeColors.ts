// Spec: docs/design-strategy.md § Data visualization, § Branding (chart-color bridge)
// Tests: manual (visual verification)

/**
 * Reads a CSS custom property value from the document root.
 * Called once at module load — not per render.
 */
function readCssVar(property: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(property).trim();
}

/** First chart series (e.g., time to first reply) — `--color-chart-1` */
export const CHART_COLOR_1 = readCssVar("--color-chart-1");

/** Second chart series (e.g., resolution time) — `--color-chart-2` */
export const CHART_COLOR_2 = readCssVar("--color-chart-2");

/** Third chart series (e.g., intake volume) — `--color-chart-3` */
export const CHART_COLOR_3 = readCssVar("--color-chart-3");

/** RGB triplet for heatmap alpha variation — `--color-heatmap-base` */
export const HEATMAP_BASE = readCssVar("--color-heatmap-base");

/** Text color on accent/heatmap backgrounds — `--color-text-on-accent` */
export const TEXT_ON_ACCENT = readCssVar("--color-text-on-accent");
