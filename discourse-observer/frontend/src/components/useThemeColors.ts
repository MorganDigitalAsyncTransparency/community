// Spec: docs/design-strategy.md § Data visualization, § Branding (chart-color bridge)
// Tests: manual (visual verification)

/**
 * Reads CSS custom property values from the document root.
 * Used as a bridge between CSS theme variables and JavaScript chart libraries
 * (Recharts) that require color strings as props.
 */
export function getThemeColor(property: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(property).trim();
}
