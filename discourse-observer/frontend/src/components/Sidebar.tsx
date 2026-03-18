// Spec: specs/dashboard/dashboard-components.md
// ADR: docs/decisions/0011-dashboard-layout-and-theme.md
// Tests: manual (sidebar interaction)

import { useState } from "react";
import type { Page } from "../types";

interface NavItem {
  page: Page;
  label: string;
  icon: string;
}

const NAV_ITEMS: NavItem[] = [
  { page: "queue", label: "Queue", icon: "\u25A6" },
  { page: "response-metrics", label: "Response metrics", icon: "\u25F7" },
  { page: "distribution", label: "Distribution", icon: "\u25D4" },
  { page: "slo", label: "SLO", icon: "\u25C9" },
  { page: "volume", label: "Volume", icon: "\u25A4" },
  { page: "activity", label: "Activity", icon: "\u26A1" },
];

const STORAGE_KEY = "sidebar-collapsed";

interface SidebarProps {
  activePage: Page;
  onNavigate: (page: Page) => void;
}

export function Sidebar({ activePage, onNavigate }: SidebarProps) {
  const [collapsed, setCollapsed] = useState(() => {
    try {
      return localStorage.getItem(STORAGE_KEY) === "true";
    } catch {
      return false;
    }
  });

  function toggleCollapsed() {
    setCollapsed((prev) => {
      const next = !prev;
      try {
        localStorage.setItem(STORAGE_KEY, String(next));
      } catch {
        // localStorage unavailable — ignore
      }
      return next;
    });
  }

  return (
    <aside className={`sidebar${collapsed ? " sidebar-collapsed" : ""}`}>
      <div className="sidebar-logo">
        {collapsed ? "d-o" : "discourse-observer"}
      </div>

      <nav className="sidebar-nav">
        {NAV_ITEMS.map(({ page, label, icon }) => (
          <button
            key={page}
            className={`sidebar-link${page === activePage ? " sidebar-link-active" : ""}`}
            onClick={() => onNavigate(page)}
            title={collapsed ? label : undefined}
          >
            <span className="sidebar-icon">{icon}</span>
            <span className="sidebar-label">{label}</span>
          </button>
        ))}
      </nav>

      <button
        className="sidebar-toggle"
        onClick={toggleCollapsed}
        title={collapsed ? "Expand sidebar" : "Collapse sidebar"}
      >
        <span className="sidebar-toggle-icon">{collapsed ? "\u00BB" : "\u00AB"}</span>
        <span className="sidebar-toggle-label">
          {collapsed ? "Expand" : "Collapse"}
        </span>
      </button>
    </aside>
  );
}
