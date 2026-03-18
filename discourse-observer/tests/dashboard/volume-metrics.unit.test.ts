// Spec: specs/dashboard/response-metrics.md

import { describe, expect, it } from "vitest";
import { computeVolumeBuckets } from "../../frontend/src/components/volumeMetrics";
import type { Topic } from "../../frontend/src/mock/data";
import type { TimeRange } from "../../frontend/src/components/intakeMetrics";

function makeTopic(overrides: Partial<Topic> & { createdAt: string }): Topic {
  return {
    id: 1,
    title: "Test topic",
    tags: [],
    categoryName: "Support",
    replyCount: 0,
    ...overrides,
  };
}

function utcNoon(year: number, month: number, day: number): string {
  return new Date(Date.UTC(year, month - 1, day, 12, 0, 0)).toISOString();
}

describe("computeVolumeBuckets", () => {
  it("returns empty when range is null", () => {
    const result = computeVolumeBuckets(
      { allTopics: [], solvedTopics: [], selfClosedTopics: [], openTopics: [] },
      "daily",
      null,
    );
    expect(result).toEqual([]);
  });

  it("counts topics into correct series for a single day", () => {
    const created = utcNoon(2025, 3, 10);
    const all = [
      makeTopic({ id: 1, createdAt: created, outcome: "solved" }),
      makeTopic({ id: 2, createdAt: created, outcome: "self-closed" }),
      makeTopic({ id: 3, createdAt: created }),
    ];
    const solved = [all[0]];
    const selfClosed = [all[1]];
    const open = [all[2]];

    const range: TimeRange = { first: "2025-03-10", last: "2025-03-10" };
    const result = computeVolumeBuckets(
      { allTopics: all, solvedTopics: solved, selfClosedTopics: selfClosed, openTopics: open },
      "daily",
      range,
    );

    expect(result).toHaveLength(1);
    expect(result[0].created).toBe(3);
    expect(result[0].accepted).toBe(1);
    expect(result[0].closed).toBe(1);
    expect(result[0].open).toBe(1);
  });

  it("fills gaps with zeros for days without topics", () => {
    const topic = makeTopic({ createdAt: utcNoon(2025, 3, 10) });
    const range: TimeRange = { first: "2025-03-10", last: "2025-03-12" };

    const result = computeVolumeBuckets(
      { allTopics: [topic], solvedTopics: [], selfClosedTopics: [], openTopics: [] },
      "daily",
      range,
    );

    expect(result).toHaveLength(3);
    expect(result[1].created).toBe(0);
    expect(result[1].accepted).toBe(0);
    expect(result[1].closed).toBe(0);
    expect(result[1].open).toBe(0);
  });

  it("produces weekly buckets keyed by Monday", () => {
    // 2025-03-12 is Wednesday → Monday is 2025-03-10
    const topic = makeTopic({ createdAt: utcNoon(2025, 3, 12) });
    const range: TimeRange = { first: "2025-03-10", last: "2025-03-10" };

    const result = computeVolumeBuckets(
      { allTopics: [topic], solvedTopics: [], selfClosedTopics: [], openTopics: [] },
      "weekly",
      range,
    );

    expect(result).toHaveLength(1);
    expect(result[0].bucketKey).toBe("2025-03-10");
    expect(result[0].created).toBe(1);
  });

  it("does not mutate input arrays", () => {
    const topics = [makeTopic({ createdAt: utcNoon(2025, 3, 10) })];
    const copy = [...topics];
    const range: TimeRange = { first: "2025-03-10", last: "2025-03-10" };

    computeVolumeBuckets(
      { allTopics: topics, solvedTopics: [], selfClosedTopics: [], openTopics: [] },
      "daily",
      range,
    );

    expect(topics).toEqual(copy);
  });
});
