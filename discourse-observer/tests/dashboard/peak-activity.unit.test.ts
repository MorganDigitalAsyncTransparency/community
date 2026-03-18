// Spec: specs/dashboard/peak-activity.md

import { describe, expect, it } from "vitest";
import {
  computeHeatmapData,
  DAY_LABELS,
} from "../../frontend/src/components/peakActivityMetrics";
import type { Topic } from "../../frontend/src/mock/data";

function makeTopic(overrides: Partial<Topic> & { id: number }): Topic {
  return {
    title: "Test topic",
    createdAt: "2025-01-06T10:00:00Z", // Monday 10:00 UTC
    tags: ["api"],
    categoryName: "Support",
    replyCount: 0,
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// DAY_LABELS (PA-7)
// ---------------------------------------------------------------------------

describe("DAY_LABELS", () => {
  it("contains 7 labels Mon through Sun", () => {
    expect(DAY_LABELS).toEqual(["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]);
  });
});

// ---------------------------------------------------------------------------
// computeHeatmapData (PA-2, PA-3, PA-5, PA-6, PA-13)
// ---------------------------------------------------------------------------

describe("computeHeatmapData", () => {
  // PA-2, PA-5 — single topic lands in correct (day, hour) cell
  it("places a single topic in the correct day and hour cell", () => {
    // 2025-01-06 is a Monday, 10:00 UTC
    const topics = [makeTopic({ id: 1, createdAt: "2025-01-06T10:30:00Z" })];
    const result = computeHeatmapData(topics);

    expect(result.cells[0][10].count).toBe(1); // Monday, hour 10
    expect(result.maxCount).toBe(1);
  });

  // PA-2 — multiple topics in same slot increment count
  it("increments count for multiple topics in the same slot", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: "2025-01-06T10:00:00Z" }),
      makeTopic({ id: 2, createdAt: "2025-01-06T10:45:00Z" }),
      makeTopic({ id: 3, createdAt: "2025-01-06T10:59:00Z" }),
    ];
    const result = computeHeatmapData(topics);

    expect(result.cells[0][10].count).toBe(3);
    expect(result.maxCount).toBe(3);
  });

  // PA-6 — returns full 7×24 grid with zeros for empty slots
  it("returns a full 7×24 grid with zeros for empty slots", () => {
    const result = computeHeatmapData([]);

    expect(result.cells.length).toBe(7);
    for (const row of result.cells) {
      expect(row.length).toBe(24);
      for (const cell of row) {
        expect(cell.count).toBe(0);
      }
    }
  });

  // PA-3 — maxCount reflects the highest cell count
  it("maxCount reflects the highest cell count", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: "2025-01-06T10:00:00Z" }), // Mon 10
      makeTopic({ id: 2, createdAt: "2025-01-06T10:30:00Z" }), // Mon 10
      makeTopic({ id: 3, createdAt: "2025-01-07T14:00:00Z" }), // Tue 14
    ];
    const result = computeHeatmapData(topics);

    expect(result.maxCount).toBe(2);
  });

  // PA-13 — maxCount is 0 for empty input
  it("maxCount is 0 for empty input", () => {
    const result = computeHeatmapData([]);
    expect(result.maxCount).toBe(0);
  });

  // PA-5 — Sunday maps to row 6 (not row 0)
  it("maps Sunday to row 6", () => {
    // 2025-01-05 is a Sunday
    const topics = [makeTopic({ id: 1, createdAt: "2025-01-05T08:00:00Z" })];
    const result = computeHeatmapData(topics);

    expect(result.cells[6][8].count).toBe(1);
  });

  // PA-5 — Monday maps to row 0
  it("maps Monday to row 0", () => {
    // 2025-01-06 is a Monday
    const topics = [makeTopic({ id: 1, createdAt: "2025-01-06T00:00:00Z" })];
    const result = computeHeatmapData(topics);

    expect(result.cells[0][0].count).toBe(1);
  });

  // PA-5 — uses UTC day and hour (not local time)
  it("uses UTC day and hour", () => {
    // 2025-01-08 is a Wednesday — 23:00 UTC
    // In many local timezones this would be Thursday, but should stay Wed hour 23
    const topics = [makeTopic({ id: 1, createdAt: "2025-01-08T23:00:00Z" })];
    const result = computeHeatmapData(topics);

    expect(result.cells[2][23].count).toBe(1); // Wednesday, hour 23
  });

  // PA-2 — does not mutate input array
  it("does not mutate the input array", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: "2025-01-06T10:00:00Z" }),
      makeTopic({ id: 2, createdAt: "2025-01-07T14:00:00Z" }),
    ];
    const copy = [...topics];
    computeHeatmapData(topics);

    expect(topics).toEqual(copy);
  });

  // PA-6 — each cell has correct day and hour fields
  it("each cell has correct day and hour fields", () => {
    const result = computeHeatmapData([]);

    for (let day = 0; day < 7; day++) {
      for (let hour = 0; hour < 24; hour++) {
        expect(result.cells[day][hour].day).toBe(day);
        expect(result.cells[day][hour].hour).toBe(hour);
      }
    }
  });
});
