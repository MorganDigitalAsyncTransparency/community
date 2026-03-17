// Spec: specs/dashboard/slo-monitoring.md

import { describe, expect, it } from "vitest";
import {
  findViolations,
  computeCompliance,
  type Violation,
  type TagCompliance,
} from "../../frontend/src/components/sloMetrics";
import type { SloConfig } from "../../frontend/src/components/tagFilter";
import type { Topic } from "../../frontend/src/mock/data";

const HOUR_MS = 3_600_000;

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

function hoursAfter(base: string, hours: number): string {
  return new Date(new Date(base).getTime() + hours * HOUR_MS).toISOString();
}

const CONFIG: SloConfig = {
  api: { firstReplyHours: 4, resolutionHours: 48, inactivityHours: 24 },
  plugin: { firstReplyHours: 8, resolutionHours: 72, inactivityHours: 48 },
};

// ---------------------------------------------------------------------------
// findViolations — threshold violation detection (SL-2, SL-3, SL-4, SL-6, SL-9, SL-11)
// ---------------------------------------------------------------------------

describe("findViolations", () => {
  const now = new Date(utcNoon(2026, 3, 16)).getTime();

  // SL-2: first reply — resolved topic with firstReplyAt exceeding threshold
  it("detects first reply violation for resolved topic", () => {
    const created = utcNoon(2026, 3, 10);
    const topic = makeTopic({
      id: 1,
      createdAt: created,
      tags: ["api"],
      firstReplyAt: hoursAfter(created, 6), // 6h > 4h threshold
      resolvedAt: hoursAfter(created, 10),
      outcome: "solved",
    });
    const result = findViolations([topic], [], CONFIG, now);
    expect(result.firstReply.length).toBe(1);
    expect(result.firstReply[0].topicId).toBe(1);
    expect(result.firstReply[0].excessMs).toBeGreaterThan(0);
  });

  // SL-2: first reply — unreplied topic where time since creation exceeds threshold
  it("detects first reply violation for unreplied topic", () => {
    const topic = makeTopic({
      id: 2,
      createdAt: utcNoon(2026, 3, 15), // ~24h ago at now, > 4h threshold
      tags: ["api"],
    });
    const result = findViolations([], [topic], CONFIG, now);
    expect(result.firstReply.length).toBe(1);
  });

  // SL-2: resolution — resolved topic where resolution time exceeds threshold
  it("detects resolution violation", () => {
    const created = utcNoon(2026, 3, 10);
    const topic = makeTopic({
      id: 3,
      createdAt: created,
      tags: ["api"],
      firstReplyAt: hoursAfter(created, 2),
      resolvedAt: hoursAfter(created, 72), // 72h > 48h threshold
      outcome: "solved",
    });
    const result = findViolations([topic], [], CONFIG, now);
    expect(result.resolution.length).toBe(1);
  });

  // SL-2: inactivity — unreplied topic where time since creation exceeds inactivity threshold
  it("detects inactivity violation for unreplied topic", () => {
    const topic = makeTopic({
      id: 4,
      createdAt: utcNoon(2026, 3, 14), // ~48h ago, > 24h threshold
      tags: ["api"],
    });
    const result = findViolations([], [topic], CONFIG, now);
    expect(result.inactivity.length).toBe(1);
  });

  // SL-11: topics with no configured tags are excluded
  it("excludes topics with no configured tags", () => {
    const topic = makeTopic({
      id: 5,
      createdAt: utcNoon(2026, 1, 1),
      tags: ["unknown-tag"],
    });
    const result = findViolations([], [topic], CONFIG, now);
    expect(result.firstReply.length).toBe(0);
    expect(result.inactivity.length).toBe(0);
  });

  // SL-4: strictest threshold applies when topic has multiple configured tags
  it("uses strictest threshold across multiple configured tags", () => {
    const created = utcNoon(2026, 3, 15);
    const topic = makeTopic({
      id: 6,
      createdAt: created,
      tags: ["api", "plugin"], // api: 4h, plugin: 8h — strictest is 4h
      firstReplyAt: hoursAfter(created, 5), // 5h: violates api (4h) but not plugin (8h)
      resolvedAt: hoursAfter(created, 10),
      outcome: "solved",
    });
    const result = findViolations([topic], [], CONFIG, now);
    expect(result.firstReply.length).toBe(1);
    expect(result.firstReply[0].tag).toBe("api");
    expect(result.firstReply[0].thresholdMs).toBe(4 * HOUR_MS);
  });

  // SL-6: sorted by excess time descending (worst first)
  it("sorts violations by excess time descending", () => {
    const created1 = utcNoon(2026, 3, 10);
    const created2 = utcNoon(2026, 3, 10);
    const topic1 = makeTopic({
      id: 10,
      createdAt: created1,
      tags: ["api"],
      firstReplyAt: hoursAfter(created1, 6), // 2h excess
      resolvedAt: hoursAfter(created1, 10),
      outcome: "solved",
    });
    const topic2 = makeTopic({
      id: 11,
      createdAt: created2,
      tags: ["api"],
      firstReplyAt: hoursAfter(created2, 20), // 16h excess
      resolvedAt: hoursAfter(created2, 30),
      outcome: "solved",
    });
    const result = findViolations([topic1, topic2], [], CONFIG, now);
    expect(result.firstReply[0].topicId).toBe(11);
    expect(result.firstReply[1].topicId).toBe(10);
  });

  // SL-2: no violation when within threshold
  it("does not flag topics within threshold", () => {
    const created = utcNoon(2026, 3, 10);
    const topic = makeTopic({
      id: 20,
      createdAt: created,
      tags: ["api"],
      firstReplyAt: hoursAfter(created, 2), // 2h < 4h threshold
      resolvedAt: hoursAfter(created, 24), // 24h < 48h threshold
      outcome: "solved",
    });
    const result = findViolations([topic], [], CONFIG, now);
    expect(result.firstReply.length).toBe(0);
    expect(result.resolution.length).toBe(0);
  });

  // SL-9: empty config returns no violations
  it("returns empty results for empty config", () => {
    const topic = makeTopic({
      id: 30,
      createdAt: utcNoon(2026, 1, 1),
      tags: ["api"],
    });
    const result = findViolations([], [topic], {}, now);
    expect(result.firstReply.length).toBe(0);
    expect(result.resolution.length).toBe(0);
    expect(result.inactivity.length).toBe(0);
  });
});

// ---------------------------------------------------------------------------
// computeCompliance — SLO compliance rates (SL-14, SL-15, SL-17, SL-19)
// ---------------------------------------------------------------------------

describe("computeCompliance", () => {
  const now = new Date(utcNoon(2026, 3, 16)).getTime();

  // SL-14: basic compliance percentage calculation
  it("computes compliance percentages per tag", () => {
    const created = utcNoon(2026, 3, 10);
    const compliant = makeTopic({
      id: 1,
      createdAt: created,
      tags: ["api"],
      firstReplyAt: hoursAfter(created, 2), // within 4h
      resolvedAt: hoursAfter(created, 24), // within 48h
      outcome: "solved",
    });
    const violating = makeTopic({
      id: 2,
      createdAt: created,
      tags: ["api"],
      firstReplyAt: hoursAfter(created, 6), // exceeds 4h
      resolvedAt: hoursAfter(created, 72), // exceeds 48h
      outcome: "solved",
    });
    const result = computeCompliance([compliant, violating], [], CONFIG, now);
    const api = result.find((r) => r.tag === "api");
    expect(api).toBeDefined();
    expect(api!.firstReplyPercent).toBe(50); // 1 of 2
    expect(api!.resolutionPercent).toBe(50); // 1 of 2
  });

  // SL-17: "–" equivalent (null) when no topics are eligible for a threshold type
  it("returns null for threshold types with no eligible topics", () => {
    const result = computeCompliance([], [], CONFIG, now);
    const api = result.find((r) => r.tag === "api");
    expect(api).toBeDefined();
    expect(api!.firstReplyPercent).toBeNull();
    expect(api!.resolutionPercent).toBeNull();
    expect(api!.inactivityPercent).toBeNull();
  });

  // SL-15: inactivity compliance uses unreplied topics only
  it("evaluates inactivity compliance from unreplied topics", () => {
    const recentUnreplied = makeTopic({
      id: 10,
      createdAt: hoursAfter(utcNoon(2026, 3, 16), -2), // 2h ago, within 24h
      tags: ["api"],
    });
    const oldUnreplied = makeTopic({
      id: 11,
      createdAt: utcNoon(2026, 3, 10), // ~144h ago, exceeds 24h
      tags: ["api"],
    });
    const result = computeCompliance([], [recentUnreplied, oldUnreplied], CONFIG, now);
    const api = result.find((r) => r.tag === "api");
    expect(api!.inactivityPercent).toBe(50);
  });

  // SL-19: tags sorted alphabetically
  it("sorts tags alphabetically", () => {
    const result = computeCompliance([], [], CONFIG, now);
    expect(result[0].tag).toBe("api");
    expect(result[1].tag).toBe("plugin");
  });

  // SL-14: empty config returns empty array
  it("returns empty array for empty config", () => {
    const result = computeCompliance([], [], {}, now);
    expect(result).toEqual([]);
  });

  // SL-15: first reply compliance includes unreplied topics
  it("includes unreplied topics in first reply compliance", () => {
    const recentUnreplied = makeTopic({
      id: 20,
      createdAt: hoursAfter(utcNoon(2026, 3, 16), -2), // 2h ago, within 4h
      tags: ["api"],
    });
    const result = computeCompliance([], [recentUnreplied], CONFIG, now);
    const api = result.find((r) => r.tag === "api");
    expect(api!.firstReplyPercent).toBe(100); // within threshold
  });
});
