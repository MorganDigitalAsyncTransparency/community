// Spec: specs/dashboard/tag-distribution.md

import { describe, expect, it } from "vitest";
import {
  tagVolumeRanking,
  tagResolutionRanking,
  tagBacklogRanking,
  computeWeeklyBacklog,
} from "../../frontend/src/components/tagMetrics";
import type { Topic } from "../../frontend/src/mock/data";

const DAY_MS = 86_400_000;

function makeTopic(overrides: Partial<Topic> & { createdAt: string }): Topic {
  return {
    id: 1,
    title: "Test topic",
    tags: [],
    category: "Support",
    replyCount: 0,
    ...overrides,
  };
}

function utcNoon(year: number, month: number, day: number): string {
  return new Date(Date.UTC(year, month - 1, day, 12, 0, 0)).toISOString();
}

// ---------------------------------------------------------------------------
// tagVolumeRanking (TD-1 – TD-3)
// ---------------------------------------------------------------------------

describe("tagVolumeRanking", () => {
  // TD-1, TD-5
  it("returns an empty array when given no topics", () => {
    expect(tagVolumeRanking([])).toEqual([]);
  });

  // TD-1 — ordered by count descending
  it("ranks tags by topic count, highest first", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3), tags: ["api"] }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 4), tags: ["api"] }),
      makeTopic({ id: 3, createdAt: utcNoon(2025, 3, 5), tags: ["webhooks"] }),
    ];

    const result = tagVolumeRanking(topics);

    expect(result[0]).toEqual({ tag: "api", topicCount: 2 });
    expect(result[1]).toEqual({ tag: "webhooks", topicCount: 1 });
  });

  // TD-2 — multi-tag topics credit each tag independently
  it("counts each tag independently for multi-tag topics", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3), tags: ["api", "authentication"] }),
    ];

    const result = tagVolumeRanking(topics);

    expect(result).toHaveLength(2);
    expect(result.find((r) => r.tag === "api")?.topicCount).toBe(1);
    expect(result.find((r) => r.tag === "authentication")?.topicCount).toBe(1);
  });

  // TD-2 — untagged topics produce no entries
  it("does not include entries for topics with no tags", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3), tags: [] }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 4), tags: ["api"] }),
    ];

    const result = tagVolumeRanking(topics);

    expect(result).toHaveLength(1);
    expect(result[0].tag).toBe("api");
  });

  // TD-1 — does not mutate input
  it("does not mutate the input array", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3), tags: ["api"] }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 4), tags: ["webhooks"] }),
    ];
    const original = [...topics];

    tagVolumeRanking(topics);

    expect(topics).toEqual(original);
  });

  // tie-breaking: equal counts sorted alphabetically
  it("breaks ties alphabetically by tag name", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3), tags: ["webhooks"] }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 4), tags: ["api"] }),
    ];

    const result = tagVolumeRanking(topics);

    expect(result[0].tag).toBe("api");
    expect(result[1].tag).toBe("webhooks");
  });
});

// ---------------------------------------------------------------------------
// tagResolutionRanking (TD-6 – TD-10)
// ---------------------------------------------------------------------------

describe("tagResolutionRanking", () => {
  // TD-11
  it("returns an empty array when given no topics", () => {
    expect(tagResolutionRanking([])).toEqual([]);
  });

  // TD-7 — topics without resolvedAt are excluded from the median
  it("excludes topics without resolvedAt from the median calculation", () => {
    const created = utcNoon(2025, 3, 3);
    const topics = [
      makeTopic({
        id: 1,
        createdAt: created,
        tags: ["api"],
        resolvedAt: new Date(new Date(created).getTime() + 2 * DAY_MS).toISOString(),
        outcome: "solved",
      }),
      makeTopic({ id: 2, createdAt: created, tags: ["api"] }), // no resolvedAt
    ];

    const result = tagResolutionRanking(topics);

    expect(result).toHaveLength(1);
    expect(result[0].resolvedCount).toBe(1); // only the one with resolvedAt counts
    expect(result[0].medianResolution).toBe("2d");
  });

  // TD-8 — tag with no resolvedAt across its topics shows "–"
  it("shows '–' for a tag where no topics have resolvedAt", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3), tags: ["api"] }),
    ];

    const result = tagResolutionRanking(topics);

    expect(result[0].medianResolution).toBe("–");
    expect(result[0].resolvedCount).toBe(0);
  });

  // TD-8 — tags with "–" sort after tags with a numeric median
  it("sorts tags with '–' median after tags with a numeric median", () => {
    const created = utcNoon(2025, 3, 3);
    const resolvedAt = new Date(new Date(created).getTime() + 3 * DAY_MS).toISOString();
    const topics = [
      makeTopic({ id: 1, createdAt: created, tags: ["missing-data"] }), // no resolvedAt
      makeTopic({ id: 2, createdAt: created, tags: ["api"], resolvedAt, outcome: "solved" }),
    ];

    const result = tagResolutionRanking(topics);

    expect(result[0].tag).toBe("api");
    expect(result[1].tag).toBe("missing-data");
  });

  // TD-6 — rows with a numeric median are ordered slowest first
  it("ranks tags by median resolution time, slowest first", () => {
    const created = utcNoon(2025, 3, 3);
    const topics = [
      makeTopic({
        id: 1,
        createdAt: created,
        tags: ["fast"],
        resolvedAt: new Date(new Date(created).getTime() + 1 * DAY_MS).toISOString(),
        outcome: "solved",
      }),
      makeTopic({
        id: 2,
        createdAt: created,
        tags: ["slow"],
        resolvedAt: new Date(new Date(created).getTime() + 7 * DAY_MS).toISOString(),
        outcome: "solved",
      }),
    ];

    const result = tagResolutionRanking(topics);

    expect(result[0].tag).toBe("slow");
    expect(result[1].tag).toBe("fast");
  });

  // TD-10 — resolvedCount reflects only topics with resolvedAt
  it("sets resolvedCount to only topics with resolvedAt", () => {
    const created = utcNoon(2025, 3, 3);
    const topics = [
      makeTopic({
        id: 1,
        createdAt: created,
        tags: ["api"],
        resolvedAt: new Date(new Date(created).getTime() + 2 * DAY_MS).toISOString(),
        outcome: "solved",
      }),
      makeTopic({ id: 2, createdAt: created, tags: ["api"] }),
      makeTopic({ id: 3, createdAt: created, tags: ["api"] }),
    ];

    const result = tagResolutionRanking(topics);

    expect(result[0].resolvedCount).toBe(1);
  });
});

// ---------------------------------------------------------------------------
// tagBacklogRanking (TD-12 – TD-13)
// ---------------------------------------------------------------------------

describe("tagBacklogRanking", () => {
  // TD-15
  it("returns an empty array when given no topics", () => {
    expect(tagBacklogRanking([])).toEqual([]);
  });

  // TD-12 — ordered by open count descending
  it("ranks tags by open topic count, highest first", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3), tags: ["api"] }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 4), tags: ["api"] }),
      makeTopic({ id: 3, createdAt: utcNoon(2025, 3, 5), tags: ["webhooks"] }),
    ];

    const result = tagBacklogRanking(topics);

    expect(result[0]).toEqual({ tag: "api", openCount: 2 });
    expect(result[1]).toEqual({ tag: "webhooks", openCount: 1 });
  });
});

// ---------------------------------------------------------------------------
// computeWeeklyBacklog (TD-16 – TD-23)
// ---------------------------------------------------------------------------

describe("computeWeeklyBacklog", () => {
  // TD-24
  it("returns an empty array when given no topics", () => {
    expect(computeWeeklyBacklog([], [])).toEqual([]);
  });

  // TD-17, TD-18 — topics in the same week produce one row
  it("groups topics from the same week into a single row", () => {
    // 2025-W10: Mon 2025-03-03 to Sun 2025-03-09
    const allTopics = [
      makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3) }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 7) }),
    ];

    const result = computeWeeklyBacklog(allTopics, []);

    expect(result).toHaveLength(1);
    expect(result[0].weekStart).toBe("2025-03-03");
  });

  // TD-18 — created count = unreplied + resolved in that week
  it("sets created to the total topic count for the week", () => {
    const allTopics = [
      makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3) }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 5) }),
      makeTopic({ id: 3, createdAt: utcNoon(2025, 3, 10) }), // different week
    ];

    const result = computeWeeklyBacklog(allTopics, []);

    const week10 = result.find((r) => r.weekStart === "2025-03-03");
    expect(week10?.created).toBe(2);
  });

  // TD-20 — stillOpen = unreplied topics in that week
  it("sets stillOpen to topics in allTopics that also appear in openTopics", () => {
    const open1 = makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3) });
    const resolved2 = makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 5) });
    const allTopics = [open1, resolved2];
    const openTopics = [open1];

    const [row] = computeWeeklyBacklog(allTopics, openTopics);

    expect(row.stillOpen).toBe(1);
  });

  // TD-19 — resolved = topics in allTopics not in openTopics
  it("sets resolved to topics in allTopics not present in openTopics", () => {
    const open1 = makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3) });
    const resolved2 = makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 5) });
    const allTopics = [open1, resolved2];
    const openTopics = [open1];

    const [row] = computeWeeklyBacklog(allTopics, openTopics);

    expect(row.resolved).toBe(1);
  });

  // TD-18, TD-19, TD-20 — created = resolved + stillOpen
  it("satisfies created = resolved + stillOpen for every row", () => {
    const open1 = makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3) });
    const open2 = makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 4) });
    const res3 = makeTopic({ id: 3, createdAt: utcNoon(2025, 3, 5) });
    const allTopics = [open1, open2, res3];
    const openTopics = [open1, open2];

    const [row] = computeWeeklyBacklog(allTopics, openTopics);

    expect(row.created).toBe(row.resolved + row.stillOpen);
  });

  // TD-21 — ordered newest first
  it("orders rows newest week first", () => {
    const allTopics = [
      makeTopic({ id: 1, createdAt: utcNoon(2025, 1, 6) }),  // W02
      makeTopic({ id: 2, createdAt: utcNoon(2025, 1, 20) }), // W04
      makeTopic({ id: 3, createdAt: utcNoon(2025, 1, 13) }), // W03
    ];

    const weeks = computeWeeklyBacklog(allTopics, []).map((r) => r.weekStart);

    expect(weeks).toEqual(["2025-01-20", "2025-01-13", "2025-01-06"]);
  });

  // TD-22 — only weeks with at least one topic
  it("only produces rows for weeks that contain at least one topic", () => {
    const allTopics = [
      makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3) }),  // W10
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 17) }), // W12 — skips W11
    ];

    const result = computeWeeklyBacklog(allTopics, []);

    expect(result).toHaveLength(2);
    const weeks = result.map((r) => r.weekStart);
    expect(weeks).not.toContain("2025-03-10"); // W11 is absent
  });

  // TD-17 — does not mutate either input array
  it("does not mutate either input array", () => {
    const open = makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3) });
    const allTopics = [open, makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 5) })];
    const openTopics = [open];
    const origAll = [...allTopics];
    const origOpen = [...openTopics];

    computeWeeklyBacklog(allTopics, openTopics);

    expect(allTopics).toEqual(origAll);
    expect(openTopics).toEqual(origOpen);
  });

  // openTopics may contain topics not in allTopics (edge case) — they are ignored
  it("ignores openTopics entries that are not present in allTopics", () => {
    const inAll = makeTopic({ id: 1, createdAt: utcNoon(2025, 3, 3) });
    const notInAll = makeTopic({ id: 99, createdAt: utcNoon(2025, 3, 3) });

    const [row] = computeWeeklyBacklog([inAll], [notInAll]);

    expect(row.stillOpen).toBe(0);
    expect(row.resolved).toBe(1);
  });
});
