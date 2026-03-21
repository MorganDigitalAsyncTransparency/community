# Sidebar About Section

This specification defines the "About" section added to the sidebar, replacing the standalone footer component.

Parent spec: [dashboard-components.md](dashboard-components.md)

---

## Context

The footer displayed version, last sync time, sync log link, and GitHub link in a horizontal bar at the bottom of the main content area. This change moves that information into the sidebar as a structured "About" section, creating a cleaner layout and making operational information accessible from the persistent navigation.

---

## Requirements

### DS-15: About section in sidebar

The sidebar displays an "About" section below the navigation links and above the collapse toggle. The section contains:

1. **Heading** — An info icon (ℹ) followed by the text "About". Uses the same styling pattern as navigation items but is not clickable.

2. **Version** — The application version string (e.g. "0.1.0"), displayed on its own row.

3. **Last synced** — The formatted timestamp of the last successful sync, or "Not yet synced" if no sync has occurred. Displayed on its own row. Time format matches the previous footer: locale-aware short date and short time.

4. **Sync log** — A clickable link that navigates to the sync log page. Displayed on its own row.

5. **GitHub** — An external link to the GitHub repository, opening in a new tab. Displayed on its own row.

6. **Read more** — An external link to the GitHub Pages documentation site, opening in a new tab. Displayed on its own row.

### DS-16: Collapsed sidebar behavior

When the sidebar is collapsed, the About section follows the same pattern as navigation items:

- The info icon (ℹ) remains visible.
- Text labels are hidden (opacity 0, width 0) — same CSS mechanism as nav labels.
- A tooltip shows "About" on hover.

### DS-17: Footer removal

The standalone footer component and its grid area are removed. The shell layout changes from a three-row main area (filter-bar, content, footer) to a two-row main area (filter-bar, content).

### DS-18: Sync log page access

The sync log page remains accessible via the "Sync log" link in the About section. It is not added to the main navigation items. This preserves the current information hierarchy where sync log is an operational tool, not a primary dashboard page.

---

## Verification

- **DS-15:** Manual — sidebar displays About section with all five items, each on its own row, between nav and collapse toggle.
- **DS-16:** Manual — collapsed sidebar shows ℹ icon only; hovering shows "About" tooltip.
- **DS-17:** Manual — no footer bar visible; content area extends to the bottom of the viewport.
- **DS-18:** Manual — clicking "Sync log" in the About section navigates to the sync log page.
