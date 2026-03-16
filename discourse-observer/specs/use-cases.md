# Use Cases

This document describes what users need from discourse-observer. Each use case is a concrete goal that a person wants to achieve using the system. Use cases drive the creation of module specifications but are not specifications themselves — they describe needs, not solutions.

## Terminology

- **Support topic** — a forum topic carrying one of the configured support tags.
- **Solved** — a topic where a reply has been marked as the accepted answer.
- **Self-closed** — a topic that carries a configured closed tag but has no accepted answer. Solved topics also carry the closed tag, so the distinction is the absence of an accepted answer.
- **Stalled** — an open topic that has received at least one reply but has had no activity for an extended period without reaching a resolution. Stalled topics are not explicitly closed — they have simply gone quiet.
- **Untagged topic** — a topic without any tag. These are anomalies that may have been missed by the team.
- **First reply** — the earliest reply on a topic. The system does not distinguish team replies from community replies.
- **Monitored tag** — a support tag defined in the tag configuration file. Only topics carrying a monitored tag are counted in metrics.
- **Area** — a named grouping of related monitored tags, defined in the tag configuration file. Each area has one primary tag shown at the top; remaining tags in the area are sorted alphabetically.
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
**Status:** Implemented — see [specs/dashboard/response-time-trends.md](dashboard/response-time-trends.md).

---

## Distribution and bottlenecks

### UC-9: Identify highest-volume tag areas

**Goal:** Know which monitored tags generate the most support topics, to understand where demand is concentrated.
**Expected result:** A ranking of monitored tags by topic count for a selected time period.
**Status:** Implemented — see [specs/dashboard/tag-distribution.md](dashboard/tag-distribution.md).

### UC-10: Identify slowest tag areas

**Goal:** Know which monitored tags have the longest average handling time, to find areas where the team may need more capacity or expertise.
**Expected result:** A ranking of monitored tags by average time to resolution for a selected time period.
**Status:** Implemented — see [specs/dashboard/tag-distribution.md](dashboard/tag-distribution.md).

### UC-11: Detect accumulating backlogs

**Goal:** Spot tag areas where open topics are accumulating without resolution, indicating a growing backlog.
**Expected result:** A list of monitored tags with their count of currently open topics and a weekly trend table showing created, resolved, and still-open counts per week.
**Status:** Implemented — see [specs/dashboard/tag-distribution.md](dashboard/tag-distribution.md).

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

---

## Tag and area selection

### UC-15: Filter dashboard by tag

**Goal:** Focus all dashboard metrics, lists, and charts on a single monitored tag, to evaluate that specific area's performance in isolation.
**Expected result:** All metrics and lists reflect only topics carrying the selected tag. When no tag is selected, data covers all monitored tags aggregated together.
**Status:** Implemented — see [specs/dashboard/tag-area-filter.md](dashboard/tag-area-filter.md).

### UC-16: Navigate tags by area

**Goal:** Find the relevant tag quickly when many tags are configured, by browsing within a named area grouping.
**Expected result:** An area selector narrows the visible tag list to tags belonging to that area. Each area's primary tag appears first; remaining tags are sorted alphabetically. Selecting an area does not itself select a tag — it only filters the tag list.
**Status:** Implemented — see [specs/dashboard/tag-area-filter.md](dashboard/tag-area-filter.md).

---

## Volume

### UC-17: Track topic intake over time

**Goal:** See how many new support topics are created per period, to understand demand and provide context for response metrics.
**Expected result:** Topic count shown per day or week for a selected time period, broken down by tag when a tag is selected.

---

## Activity patterns

### UC-18: Detect stalled topics

**Goal:** Identify open topics that have received at least one reply but have gone quiet without resolution, so that conversations at risk of being abandoned can be followed up.
**Expected result:** A list of open topics with at least one reply and no activity for a configurable number of days, sorted by time since last activity (oldest first), showing topic title, tag, and days since last activity.

### UC-19: Identify peak activity periods

**Goal:** Know when support topics typically arrive and when activity is highest, to inform staffing decisions and understand whether SLO misses cluster at specific times.
**Expected result:** A breakdown of topic creation by day of week and hour of day, showing where demand is concentrated.

---

## Response time distribution

### UC-20: Understand response time spread

**Goal:** See not just the median response time but how response times are distributed, to identify whether the median hides a long tail of slow responses.
**Expected result:** A distribution of time-to-first-reply and time-to-resolution values, showing how many topics fall into each time bracket for a selected period.
