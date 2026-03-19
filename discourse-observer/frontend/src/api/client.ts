// Spec: specs/api/api-contract.md (AC-1, AC-6, AC-7, AC-8, AC-9, AC-10)
// Tests: tests/dashboard/api-integration.unit.test.ts

import type { ActivePeriod } from "../components/timePeriod";

const BASE = "/api/v1";

const PRESET_TO_PARAM: Record<string, string> = {
  last7: "7d",
  last30: "30d",
  lastYear: "1y",
  allTime: "all",
};

export interface FilterParams {
  period: ActivePeriod;
  tag: string | null;
}

function buildQuery(filters?: FilterParams): string {
  if (!filters) return "";

  const params = new URLSearchParams();

  if (filters.period.kind === "custom") {
    params.set("from", filters.period.range.from);
    params.set("to", filters.period.range.to);
  } else {
    params.set("period", PRESET_TO_PARAM[filters.period.preset] ?? "all");
  }

  if (filters.tag) {
    params.set("tag", filters.tag);
  }

  const qs = params.toString();
  return qs ? `?${qs}` : "";
}

export class ApiClientError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiClientError";
  }
}

export async function apiFetch<T>(path: string, filters?: FilterParams): Promise<T> {
  const url = `${BASE}${path}${buildQuery(filters)}`;
  const res = await fetch(url);

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiClientError(res.status, body.error ?? res.statusText);
  }

  return res.json() as Promise<T>;
}
