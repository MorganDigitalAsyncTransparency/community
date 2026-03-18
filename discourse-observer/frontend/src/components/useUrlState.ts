// Spec: specs/dashboard/url-state.md
// Tests: tests/dashboard/url-state.unit.test.ts (pure logic in urlState.ts)

import { useState, useCallback } from "react";
import type { Page } from "../types";
import type { ActivePeriod } from "./timePeriod";
import { parseUrlState, buildSearch, type UrlState } from "./urlState";

function syncUrl(state: UrlState) {
  const search = buildSearch(state);
  const url = window.location.pathname + search;
  window.history.replaceState(null, "", url);
}

export function useUrlState() {
  const [state, setStateRaw] = useState<UrlState>(() =>
    parseUrlState(window.location.search),
  );

  const setPage = useCallback((page: Page) => {
    setStateRaw((prev) => {
      const next = { ...prev, page };
      syncUrl(next);
      return next;
    });
  }, []);

  const setPeriod = useCallback((period: ActivePeriod) => {
    setStateRaw((prev) => {
      const next = { ...prev, period };
      syncUrl(next);
      return next;
    });
  }, []);

  const setTag = useCallback((tag: string | null) => {
    setStateRaw((prev) => {
      const next = { ...prev, tag };
      syncUrl(next);
      return next;
    });
  }, []);

  const setArea = useCallback((area: string | null) => {
    setStateRaw((prev) => {
      const next = { ...prev, area };
      syncUrl(next);
      return next;
    });
  }, []);

  const clearAll = useCallback(() => {
    setStateRaw((prev) => {
      const next: UrlState = {
        page: prev.page,
        period: { kind: "preset", preset: "allTime" },
        tag: null,
        area: null,
      };
      syncUrl(next);
      return next;
    });
  }, []);

  return {
    page: state.page,
    period: state.period,
    tag: state.tag,
    area: state.area,
    setPage,
    setPeriod,
    setTag,
    setArea,
    clearAll,
  };
}
