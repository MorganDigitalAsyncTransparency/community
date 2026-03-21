// Spec: specs/dashboard/url-state.md
// Tests: tests/dashboard/url-state.unit.test.ts

import type { Page } from "../types";
import type { ActivePeriod, PeriodPreset } from "./timePeriod";

export interface UrlState {
  page: Page;
  period: ActivePeriod;
  tag: string | null;
  area: string | null;
}

const VALID_PAGES: ReadonlySet<string> = new Set<Page>([
  "queue",
  "response-metrics",
  "distribution",
  "slo",
  "activity",
  "sync-log",
]);

const VALID_PRESETS: ReadonlySet<string> = new Set<PeriodPreset>([
  "last7",
  "last30",
  "lastYear",
  "allTime",
]);

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

function isValidDate(value: string): boolean {
  if (!DATE_PATTERN.test(value)) return false;
  const ms = Date.parse(value + "T00:00:00Z");
  return !Number.isNaN(ms);
}

const DEFAULT_STATE: UrlState = {
  page: "queue",
  period: { kind: "preset", preset: "allTime" },
  tag: null,
  area: null,
};

export function parseUrlState(search: string): UrlState {
  const params = new URLSearchParams(search);

  const rawPage = params.get("page");
  const page: Page = rawPage !== null && VALID_PAGES.has(rawPage)
    ? (rawPage as Page)
    : DEFAULT_STATE.page;

  const rawFrom = params.get("from");
  const rawTo = params.get("to");
  const rawPreset = params.get("period");

  let period: ActivePeriod;
  if (rawFrom !== null && rawTo !== null && isValidDate(rawFrom) && isValidDate(rawTo)) {
    period = { kind: "custom", range: { from: rawFrom, to: rawTo } };
  } else if (rawPreset !== null && VALID_PRESETS.has(rawPreset)) {
    period = { kind: "preset", preset: rawPreset as PeriodPreset };
  } else {
    period = DEFAULT_STATE.period;
  }

  const rawTag = params.get("tag");
  const tag = rawTag && rawTag.length > 0 ? rawTag : null;

  const rawArea = params.get("area");
  const area = rawArea && rawArea.length > 0 ? rawArea : null;

  return { page, period, tag, area };
}

export function buildSearch(state: UrlState): string {
  const params = new URLSearchParams();

  if (state.page !== DEFAULT_STATE.page) {
    params.set("page", state.page);
  }

  if (state.period.kind === "custom") {
    params.set("from", state.period.range.from);
    params.set("to", state.period.range.to);
  } else if (state.period.preset !== "allTime") {
    params.set("period", state.period.preset);
  }

  if (state.tag !== null) {
    params.set("tag", state.tag);
  }

  if (state.area !== null) {
    params.set("area", state.area);
  }

  const result = params.toString();
  return result.length > 0 ? "?" + result : "";
}
