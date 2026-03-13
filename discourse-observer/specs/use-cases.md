# Use Cases

This document describes what users need from discourse-observer. Each use case is a concrete goal that a person wants to achieve using the system. Use cases drive the creation of module specifications but are not specifications themselves — they describe needs, not solutions.

## Terminology

- **Support topic** — a forum topic carrying one of the configured support tags.
- **Solved** — a topic where a reply has been marked as the accepted answer.
- **Self-closed** — a topic that carries a configured closed tag but has no accepted answer. Solved topics also carry the closed tag, so the distinction is the absence of an accepted answer.
- **Untagged topic** — a topic without any tag. These are anomalies that may have been missed by the team.
- **First reply** — the earliest reply on a topic. The system does not distinguish team replies from community replies.
- **SLI (service level indicator)** — a measured metric such as time to first reply or time to resolution.
- **SLO (service level objective)** — a target for an SLI, defined per monitored tag. These are internal goals, not contractual obligations.

---

## Queue visibility

### UC-1: Identify topics waiting longest for a reply

**Goal:** Find support topics that have been open the longest without any reply, so that the most neglected topics can be addressed first.
**Expected result:** A list of unreplied support topics, ordered by age (oldest first), showing topic title, creation date, and associated tag.

### UC-2: See all unreplied support topics

**Goal:** Get a complete view of support topics that have received no reply yet, to understand current queue size.
**Expected result:** A count and list of all unreplied support topics, filterable by time period.

### UC-3: Detect untagged topics

**Goal:** Identify topics that have no tag at all, which may have slipped through the tagging convention and risk being missed.
**Expected result:** A count of untagged topics, the share they represent of all topics, and a list of individual untagged topics.

---

## Response times

### UC-4: Measure time to first reply

**Goal:** Understand how quickly support topics receive their first reply, to evaluate team responsiveness.
**Expected result:** Median time from topic creation to first reply, for a selected time period.

### UC-5: Measure time to resolution

**Goal:** Understand how quickly support topics reach a resolution, to evaluate end-to-end handling speed.
**Expected result:** Median time from topic creation to the topic being marked solved, for a selected time period.

---

## Outcomes

### UC-6: Compare solved versus self-closed topics

**Goal:** Understand how many topics receive a real answer versus being closed without resolution, to gauge support effectiveness.
**Expected result:** Count and ratio of solved versus self-closed topics for a selected time period.

### UC-7: Measure answer rate

**Goal:** Know what share of support topics receive a real answer, to track overall support quality.
**Expected result:** Percentage of support topics that were solved (not self-closed) for a selected time period.

---

## Trends

### UC-8: Track response time trends

**Goal:** See whether response times and resolution times are improving or worsening over time, to identify positive or negative momentum.
**Expected result:** Response time and resolution time metrics shown across multiple consecutive periods, with enough history to identify a trend direction.

---

## Distribution and bottlenecks

### UC-9: Identify highest-volume tag areas

**Goal:** Know which monitored tags generate the most support topics, to understand where demand is concentrated.
**Expected result:** A ranking of monitored tags by topic count for a selected time period.

### UC-10: Identify slowest tag areas

**Goal:** Know which monitored tags have the longest average handling time, to find areas where the team may need more capacity or expertise.
**Expected result:** A ranking of monitored tags by average time to resolution for a selected time period.

### UC-11: Detect accumulating backlogs

**Goal:** Spot tag areas where open topics are accumulating without resolution, indicating a growing backlog.
**Expected result:** A list of monitored tags with their count of currently open topics, highlighting tags where the open count is growing over time.

---

## Time filtering

### UC-12: Filter by time period

**Goal:** Scope any of the above use cases to a specific time window, to focus on recent activity or compare periods.
**Expected result:** All metrics and lists can be filtered to last 7 days, last 30 days, last year, all time, or a custom date range.

---

## SLO monitoring

### UC-13: Flag topics exceeding SLO thresholds

**Goal:** Identify topics that have exceeded configured thresholds for first response time, resolution time, or inactivity, so they can be prioritized.
**Expected result:** A list of topics that exceed one or more SLO thresholds, grouped by threshold type, showing how far each topic exceeds the threshold. Thresholds are configurable per monitored tag.

### UC-14: Evaluate SLO compliance

**Goal:** Understand what proportion of topics meet SLO thresholds, to evaluate team performance against expectations.
**Expected result:** For each monitored tag and threshold type, the percentage of topics that were handled within the configured threshold for a selected time period.
