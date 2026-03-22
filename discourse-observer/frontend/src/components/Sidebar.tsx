// Spec: specs/dashboard/dashboard-components.md, specs/dashboard/sidebar-about-section.md
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
  { page: "activity", label: "Activity", icon: "\u26A1" },
  { page: "tag-flows", label: "Tag flows", icon: "\u21C4" },
];

const STORAGE_KEY = "sidebar-collapsed";

const GITHUB_URL = "https://github.com/MorganDigitalAsyncTransparency/community";
const DOCS_URL = "https://morgandigitalasynctransparency.github.io/community/";

interface SidebarProps {
  activePage: Page;
  onNavigate: (page: Page) => void;
  mobileOpen?: boolean;
  onMobileClose?: () => void;
  version: string;
  lastSyncedAt: string | null;
  onSyncLogClick?: () => void;
}

function formatSyncTime(isoString: string): string {
  return new Date(isoString).toLocaleString(undefined, {
    dateStyle: "short",
    timeStyle: "short",
  });
}

export function Sidebar({
  activePage, onNavigate, mobileOpen, onMobileClose,
  version, lastSyncedAt, onSyncLogClick,
}: SidebarProps) {
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

  const mobileClass = mobileOpen ? " sidebar-mobile-open" : "";

  function handleNavClick(page: Page) {
    onNavigate(page);
    onMobileClose?.();
  }

  function handleSyncLogClick() {
    onSyncLogClick?.();
    onMobileClose?.();
  }

  return (
    <>
      {mobileOpen && (
        <div className="sidebar-backdrop" onClick={onMobileClose} />
      )}
      <aside className={`sidebar${collapsed ? " sidebar-collapsed" : ""}${mobileClass}`}>
        <div className="sidebar-logo">
          {collapsed ? "d-o" : "discourse-observer"}
        </div>

        <nav className="sidebar-nav">
          {NAV_ITEMS.map(({ page, label, icon }) => (
            <button
              key={page}
              className={`sidebar-link${page === activePage ? " sidebar-link-active" : ""}`}
              onClick={() => handleNavClick(page)}
              title={collapsed ? label : undefined}
            >
              <span className="sidebar-icon">{icon}</span>
              <span className="sidebar-label">{label}</span>
            </button>
          ))}
        </nav>

        <div className="sidebar-about" title={collapsed ? "About" : undefined}>
          <div className="sidebar-about-heading">
            <span className="sidebar-icon">{"\u2139"}</span>
            <span className="sidebar-label">About</span>
          </div>
          <div className="sidebar-about-items">
            <span className="sidebar-about-item">{version}</span>
            <span className="sidebar-about-item">
              {lastSyncedAt
                ? `Last synced ${formatSyncTime(lastSyncedAt)}`
                : "Not yet synced"}
            </span>
            <button className="sidebar-about-link" onClick={handleSyncLogClick}>
              Sync log
            </button>
            <a
              className="sidebar-about-link"
              href={GITHUB_URL}
              target="_blank"
              rel="noopener noreferrer"
            >
              GitHub &#8599;
            </a>
            <a
              className="sidebar-about-link"
              href={DOCS_URL}
              target="_blank"
              rel="noopener noreferrer"
            >
              Read more &#8599;
            </a>
          </div>
        </div>

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
    </>
  );
}
