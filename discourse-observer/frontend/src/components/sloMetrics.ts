// Spec: specs/dashboard/slo-monitoring.md
// Tests: tests/dashboard/slo-monitoring.unit.test.ts

import type { Topic } from "../mock/data";
import type { SloThresholds, SloConfig } from "./tagFilter";

export type { SloConfig };
export type TagSloThresholds = SloThresholds;

export interface Violation {
  topicId: number;
  topicTitle: string;
  tag: string;
  thresholdMs: number;
  actualMs: number;
  excessMs: number;
}

export interface ViolationGroups {
  firstReply: Violation[];
  resolution: Violation[];
  inactivity: Violation[];
}

export interface TagCompliance {
  tag: string;
  firstReplyPercent: number | null;
  resolutionPercent: number | null;
  inactivityPercent: number | null;
}

const HOUR_MS = 3_600_000;

// SL-4: find the strictest (lowest) threshold across all configured tags for a topic
function strictestThreshold(
  tags: string[],
  config: SloConfig,
  field: keyof TagSloThresholds
): { tag: string; hours: number } | null {
  let best: { tag: string; hours: number } | null = null;

  for (const tag of tags) {
    const thresholds = config[tag];
    if (!thresholds) continue;
    if (!best || thresholds[field] < best.hours) {
      best = { tag, hours: thresholds[field] };
    }
  }

  return best;
}

function makeViolation(
  topic: Topic,
  tag: string,
  thresholdMs: number,
  actualMs: number
): Violation {
  return {
    topicId: topic.id,
    topicTitle: topic.title,
    tag,
    thresholdMs,
    actualMs,
    excessMs: actualMs - thresholdMs,
  };
}

// SL-6: sort by excess descending
function sortByExcess(violations: Violation[]): Violation[] {
  return [...violations].sort((a, b) => b.excessMs - a.excessMs);
}

// SL-1 through SL-12: find all topics exceeding SLO thresholds
export function findViolations(
  resolvedTopics: Topic[],
  unrepliedTopics: Topic[],
  config: SloConfig,
  now: number
): ViolationGroups {
  const firstReply: Violation[] = [];
  const resolution: Violation[] = [];
  const inactivity: Violation[] = [];

  // Resolved topics: check first reply and resolution
  for (const topic of resolvedTopics) {
    // SL-2: first reply for resolved topics with firstReplyAt
    if (topic.firstReplyAt) {
      const strict = strictestThreshold(topic.tags, config, "firstReplyHours");
      if (strict) {
        const actual = new Date(topic.firstReplyAt).getTime() - new Date(topic.createdAt).getTime();
        const threshold = strict.hours * HOUR_MS;
        if (actual > threshold) {
          firstReply.push(makeViolation(topic, strict.tag, threshold, actual));
        }
      }
    }

    // SL-2: resolution
    if (topic.resolvedAt) {
      const strict = strictestThreshold(topic.tags, config, "resolutionHours");
      if (strict) {
        const actual = new Date(topic.resolvedAt).getTime() - new Date(topic.createdAt).getTime();
        const threshold = strict.hours * HOUR_MS;
        if (actual > threshold) {
          resolution.push(makeViolation(topic, strict.tag, threshold, actual));
        }
      }
    }
  }

  // Unreplied topics: check first reply and inactivity
  for (const topic of unrepliedTopics) {
    const elapsed = now - new Date(topic.createdAt).getTime();

    // SL-2: first reply for unreplied topics (time since creation)
    const firstReplyStrict = strictestThreshold(topic.tags, config, "firstReplyHours");
    if (firstReplyStrict) {
      const threshold = firstReplyStrict.hours * HOUR_MS;
      if (elapsed > threshold) {
        firstReply.push(makeViolation(topic, firstReplyStrict.tag, threshold, elapsed));
      }
    }

    // SL-2: inactivity
    const inactivityStrict = strictestThreshold(topic.tags, config, "inactivityHours");
    if (inactivityStrict) {
      const threshold = inactivityStrict.hours * HOUR_MS;
      if (elapsed > threshold) {
        inactivity.push(makeViolation(topic, inactivityStrict.tag, threshold, elapsed));
      }
    }
  }

  return {
    firstReply: sortByExcess(firstReply),
    resolution: sortByExcess(resolution),
    inactivity: sortByExcess(inactivity),
  };
}

// SL-13 through SL-20: compute compliance rates per tag
export function computeCompliance(
  resolvedTopics: Topic[],
  unrepliedTopics: Topic[],
  config: SloConfig,
  now: number
): TagCompliance[] {
  const tags = Object.keys(config).sort(); // SL-19

  if (tags.length === 0) return [];

  return tags.map((tag) => {
    const thresholds = config[tag];

    // Collect topics that have this tag
    const resolvedWithTag = resolvedTopics.filter((t) => t.tags.includes(tag));
    const unrepliedWithTag = unrepliedTopics.filter((t) => t.tags.includes(tag));

    // SL-15: first reply — resolved with firstReplyAt + unreplied
    const firstReplyEligible = [
      ...resolvedWithTag.filter((t) => t.firstReplyAt),
      ...unrepliedWithTag,
    ];
    const firstReplyPercent = firstReplyEligible.length === 0
      ? null
      : computePercent(firstReplyEligible, (t) => {
          const actual = t.firstReplyAt
            ? new Date(t.firstReplyAt).getTime() - new Date(t.createdAt).getTime()
            : now - new Date(t.createdAt).getTime();
          return actual <= thresholds.firstReplyHours * HOUR_MS;
        });

    // SL-15: resolution — resolved topics only
    const resolutionPercent = resolvedWithTag.length === 0
      ? null
      : computePercent(resolvedWithTag, (t) => {
          if (!t.resolvedAt) return false;
          const actual = new Date(t.resolvedAt).getTime() - new Date(t.createdAt).getTime();
          return actual <= thresholds.resolutionHours * HOUR_MS;
        });

    // SL-15: inactivity — unreplied topics only
    const inactivityPercent = unrepliedWithTag.length === 0
      ? null
      : computePercent(unrepliedWithTag, (t) => {
          const elapsed = now - new Date(t.createdAt).getTime();
          return elapsed <= thresholds.inactivityHours * HOUR_MS;
        });

    return { tag, firstReplyPercent, resolutionPercent, inactivityPercent };
  });
}

function computePercent(
  topics: Topic[],
  isCompliant: (t: Topic) => boolean
): number {
  const compliant = topics.filter(isCompliant).length;
  return Math.round((compliant / topics.length) * 100);
}
