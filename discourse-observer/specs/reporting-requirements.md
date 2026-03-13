# Reporting Requirements

This document defines what questions the reporting layer must be able to answer and what data those answers depend on. It is a fact source for future technical decisions, not a UI specification.

## Forum workflow

Support activity on the forum is tracked through tags. All support-related topics carry one of the configured support tags. A dedicated team monitors these tags. As the team investigates a topic, the specific tag may change to reflect what was discovered — there are no separate intermediate-state tags, no assignment tags, and no escalation markers visible in the forum. Topics are either open or closed.

Topics close in one of two ways:

- **Solved** — The Discourse solved plugin marks a reply as the accepted answer. The topic has a proper resolution.
- **Self-closed** — No one engaged meaningfully. The topic is closed without a real answer and carries a configured closed tag to distinguish it from solved topics.

When a topic is escalated to an external system (backlog, issue tracker), a link appears in the closing post. What happens in that external system is not tracked here.

Topics without any tag at all are anomalies. Most topics are tagged. Untagged topics may have slipped through the tagging convention and risk being missed by the team.

## Questions the system must answer

The reporting layer exists to answer these questions. They are grouped by what they describe, not by who asks them. Any viewer — team member, manager, or external observer — may be interested in any of these.

### Queue and status

- Which support topics have been open the longest without any reply?
- Which support topics have no reply yet?
- How many topics currently have no tag at all, and what share of total topics does that represent?

Note: the current data model does not include user identity, so it is not possible to distinguish team replies from community replies. These questions use any reply as a proxy. Distinguishing team responses would require user identity data to be added to the model.

### Response times

- What is the median time from topic creation to the first reply?
- What is the median time from topic creation to the topic being marked solved?

### Outcomes

- How many topics are solved versus self-closed per period?
- What share of support topics receive a real answer?

### Trends

- Are response times and resolution times improving or worsening over time?

### Distribution and bottlenecks

- Which monitored tags have the highest volume of topics?
- Which monitored tags have the longest average handling time?
- Are there tag areas where open topics accumulate without resolution?

## Time horizons

All questions can be scoped to the following periods:

- Last 7 days
- Last 30 days
- Last year
- All time
- Custom date range

## SLA thresholds

Three thresholds define when a topic is considered late or stalled: first response time, resolution time, and stalled (no activity for a defined period). They are used to flag topics that need attention and to evaluate team performance.

All thresholds are configurable per monitored tag, allowing different tag groups to carry different expectations. The tags to monitor and their associated thresholds are defined in a config file, not hardcoded.

## Output

The reporting layer produces a locally hosted website. It reads from the analytical store defined in [ADR 0006](../docs/decisions/0006-analytical-storage.md) and does not query Discourse directly.
