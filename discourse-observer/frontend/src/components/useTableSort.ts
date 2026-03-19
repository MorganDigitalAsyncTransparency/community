// Spec: specs/dashboard/queue-visibility.md, specs/dashboard/stalled-topics.md,
//       specs/dashboard/slo-monitoring.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts, backend/api/contract_test.go

import { useState } from "react";

export type SortDirection = "asc" | "desc";

interface UseTableSortResult<C extends string> {
  sortColumn: C;
  sortDirection: SortDirection;
  handleSort: (column: C) => void;
  arrow: (column: C) => string;
}

export function useTableSort<C extends string>(
  defaultColumn: C,
  defaultDirections: Record<C, SortDirection>,
): UseTableSortResult<C> {
  const [sortColumn, setSortColumn] = useState<C>(defaultColumn);
  const [sortDirection, setSortDirection] = useState<SortDirection>(
    defaultDirections[defaultColumn],
  );

  function handleSort(column: C) {
    if (column === sortColumn) {
      setSortDirection((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortColumn(column);
      setSortDirection(defaultDirections[column]);
    }
  }

  function arrow(column: C): string {
    if (column !== sortColumn) return "";
    return sortDirection === "asc" ? " ▲" : " ▼";
  }

  return { sortColumn, sortDirection, handleSort, arrow };
}
