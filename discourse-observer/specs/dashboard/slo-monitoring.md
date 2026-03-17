# SLO Monitoring — Dashboard View

This specification defines the requirements for SLO (Service Level Objective) monitoring: flagging topics that exceed configured thresholds and evaluating compliance rates per tag. It traces to UC-13 and UC-14 in [use-cases.md](../use-cases.md).

This file defines *what* the user sees and why. [dashboard-components.md](dashboard-components.md) defines *how* each component behaves to fulfill these requirements.

---

## Use case traceability

| Requirement | Use case |
|-------------|----------|
| SL-1 – SL-12 | UC-13: Flag topics exceeding SLO thresholds |
| SL-13 – SL-20 | UC-14: Evaluate SLO compliance |
| SL-21 – SL-25 | Cross-cutting: configuration, navigation, empty states |

---

## Configuration

**SL-21.** SLO thresholds are defined per tag in the unified configuration file (`config/tagConfig.json`), under each tag's optional `slo` field. Each `slo` object contains three numeric thresholds: `firstReplyHours`, `resolutionHours`, and `inactivityHours`. Tags without an explicit `slo` field inherit default thresholds from `defaults.slo`. All monitored tags participate in SLO monitoring.

**SL-21a.** Tags using default SLO thresholds (not explicitly configured) are indicated in the UI with an informational message — "(default thresholds)" — so that viewers do not mistake fallback values for agreed commitments.

**SL-22.** The configuration file is documented by a committed example (`config/tagConfig.example.json`). The runtime file is gitignored and created from the example during `make start`. There is no separate `sloThresholds.json`.

---

## Requirements

### Threshold violations (UC-13)

**SL-1.** The user sees topics that have exceeded one or more SLO thresholds, so that overdue work can be prioritized.

**SL-2.** Three threshold types are evaluated independently:

- **First reply:** elapsed time from `createdAt` to `firstReplyAt` (for resolved topics with `firstReplyAt`), or elapsed time from `createdAt` to now (for unreplied topics). Compared against `firstReplyHours`.
- **Resolution:** elapsed time from `createdAt` to `resolvedAt` (for resolved topics only). Compared against `resolutionHours`.
- **Inactivity:** elapsed time from `createdAt` to now, for unreplied topics only. Compared against `inactivityHours`.

**SL-3.** A topic is evaluated against the thresholds of each of its tags that appears in the SLO configuration. If any configured tag's threshold is exceeded, the topic appears in that threshold group.

**SL-4.** When a topic has multiple configured tags with different thresholds, the strictest (lowest) threshold for each type determines whether the topic is in violation.

**SL-5.** Violations are grouped into three sections by threshold type: first reply violations, resolution violations, and inactivity violations.

**SL-6.** Within each section, topics are sorted by excess time descending — the topic that exceeds the threshold by the largest margin appears first.

**SL-7.** Each violation row displays: the topic title, the tag whose threshold was exceeded, the threshold value, the actual elapsed time, and the excess time (actual minus threshold).

**SL-8.** Times in violation rows use the shared duration format (see RT-10 / RM-13): whole days for ≥ 24 hours, whole hours for < 24 hours, minimum "1h".

**SL-9.** Violation lists respect the active time period filter (UC-12). Only topics whose `createdAt` falls within the active period are evaluated.

**SL-10.** When a topic appears in multiple threshold groups (e.g. both first reply and inactivity), it is shown independently in each group.

**SL-11.** Topics with no configured tags are excluded from all violation checks.

**SL-12.** When a threshold group has no violations, it displays an empty-state message.

### Compliance rates (UC-14)

**SL-13.** The user sees what percentage of topics met each SLO threshold per monitored tag, so that team performance against expectations is visible.

**SL-14.** For each tag present in the SLO configuration, and for each threshold type, the compliance rate is: (topics within threshold / total evaluated topics) × 100, rounded to the nearest whole number.

**SL-15.** The evaluated topic set for each threshold type per tag:

- **First reply compliance:** resolved topics with `firstReplyAt` plus unreplied topics, all having the given tag. A topic is compliant if its first reply time (or time-since-creation for unreplied) is within `firstReplyHours`.
- **Resolution compliance:** resolved topics with the given tag. A topic is compliant if its resolution time is within `resolutionHours`.
- **Inactivity compliance:** unreplied topics with the given tag. A topic is compliant if its time-since-creation is within `inactivityHours`.

**SL-16.** Compliance is displayed as a table with one row per monitored tag and columns for: tag name, first reply compliance %, resolution compliance %, and inactivity compliance %.

**SL-17.** When a tag has no topics eligible for a threshold type, the cell displays "–" instead of a percentage.

**SL-18.** Compliance rates respect the active time period filter (UC-12). Only topics whose `createdAt` falls within the active period are evaluated.

**SL-19.** Tags are sorted alphabetically by tag name.

**SL-20.** When no monitored tags have any eligible topics, the compliance table is replaced by an empty-state message.

### Navigation and page layout

**SL-23.** SLO monitoring is accessible via a dedicated "SLO" page in the dashboard navigation, placed after "Distribution".

**SL-24.** The SLO page displays threshold violations (UC-13) above compliance rates (UC-14).

**SL-25.** When no SLO thresholds are configured (empty configuration), the entire page displays an empty-state message indicating that no thresholds are configured.
