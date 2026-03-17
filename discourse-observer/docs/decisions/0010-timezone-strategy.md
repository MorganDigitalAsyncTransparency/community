# 10. Timezone Strategy

**Status:** Accepted
**Date:** 2026-03-18

## Context

The discourse-observer system collects topic data from a Discourse forum and presents it through a dashboard used by people in different timezones. Every timestamp entering the system — `created_at`, `first_replied_at`, `bumped_at`, and others — arrives from the Discourse API in ISO 8601 format with full UTC precision (e.g., `2026-03-15T14:32:07Z`). The system therefore knows the exact moment each event occurred and can represent it in any timezone.

The question this ADR addresses is: **in which timezone should timestamps be presented to the user on screen?**

This is purely a display-layer concern. Storage, computation, filtering boundaries, and duration arithmetic all operate on absolute UTC instants and are unaffected by the choice made here. The scope is limited to:

- Formatted dates and times shown in the UI (tables, cards, tooltips)
- Time-of-day bucketing for the peak activity heatmap (hour 0–23)
- Date labels on trend charts (x-axis labels showing days or week-start dates)
- Custom date range inputs (the calendar dates the user selects)

### Current state

The codebase currently uses UTC for all display. This was not a deliberate architectural decision — it emerged naturally from the implementation and is noted in individual specs:

- [time-period-filter.md](../../specs/dashboard/time-period-filter.md) TF-8: *"UTC boundaries are used because topic timestamps are stored in UTC, making the filter consistent and predictable regardless of the user's timezone."*
- [peak-activity.md](../../specs/dashboard/peak-activity.md) PA-5: *"Topics are assigned to (day, hour) slots based on their `createdAt` timestamp interpreted in UTC."*
- [topicFormatting.ts](../../frontend/src/components/topicFormatting.ts): `formatWeekLabel` and `formatDayLabel` explicitly pass `timeZone: "UTC"` to `toLocaleDateString`.
- [trendMetrics.ts](../../frontend/src/components/trendMetrics.ts): `mondayOf` uses `getUTCDay()` and `setUTCDate()`.
- [intakeMetrics.ts](../../frontend/src/components/intakeMetrics.ts): `dayOf` uses `toISOString().slice(0, 10)`, which is inherently UTC.
- [App.tsx](../../frontend/src/App.tsx): `formatSyncTime` uses `toLocaleString()` without an explicit timezone — this is the only place that implicitly uses the browser's local timezone.

Duration calculations (`formatDuration`, `formatAge`) are timezone-neutral since they operate on millisecond differences.

### Why this matters

When timestamps are displayed in UTC, a topic created at 23:30 local time on Tuesday in UTC+2 appears as Wednesday 21:30 UTC. This creates several friction points:

1. **Peak activity heatmap** — The hour distribution reflects UTC hours, not the hours when the team actually experiences activity. A team working CET (UTC+1) sees their 09:00 morning peak displayed as hour 8. This is misleading for staffing decisions, which is the stated purpose of UC-19.

2. **Daily and weekly chart labels** — A topic created late evening in a positive-offset timezone rolls into the next UTC day. Trend charts may attribute activity to different days than users expect.

3. **Custom date range filtering** — When a user picks "March 15" as a start date, the filter applies from `2026-03-15T00:00:00Z`. For a UTC+9 user, this excludes topics created on their local March 15 before 09:00 and includes topics from their March 14 after 15:00.

4. **"Last 7 days" semantics** — The rolling window is based on `Date.now()` which is timezone-neutral, so the window itself is correct. But label display ("March 10 – March 16") may not match the user's local calendar.

5. **Sync timestamp** — `formatSyncTime` already uses the browser's local timezone (no explicit `timeZone` option), which is inconsistent with the rest of the dashboard.

For a single-forum deployment where all users share roughly the same timezone, friction points 1–4 are manageable but surprising. For a globally used deployment, they become a source of confusion.

### Constraints

- The system is a **single-forum, single-deployment** application ([single-forum-scope.md](../../specs/single-forum-scope.md)). There is no multi-tenant requirement.
- The backend is not yet implemented. Any decision made now affects only specs and frontend code, making migration cost low.
- The Discourse API provides all timestamps in UTC. The forum itself has a configured timezone, but this is not exposed through the standard API.
- [operational-constraints.md](../../specs/operational-constraints.md) mentions "working hours" as a period of higher activity without defining whose timezone those hours refer to.
- The dashboard is a read-only analytics tool — there are no user accounts, authentication, or per-user preferences.
- No external date/time library is currently used. All formatting uses native `Date` and `Intl.DateTimeFormat`.

## Alternatives Considered

### Alternative A — UTC everywhere (current behavior)

Display all timestamps, date labels, and time-of-day bucketing in UTC. The user sees UTC times throughout the dashboard.

**How it works:**

- All `toLocaleDateString` and `toLocaleString` calls receive `timeZone: "UTC"`.
- Peak activity heatmap uses `getUTCHours()` for bucketing.
- Custom date range boundaries are midnight-to-midnight UTC.
- No timezone configuration or detection needed.

**Strengths:**

- Simplest possible implementation — already in place.
- Zero ambiguity about what a displayed time means.
- Filter boundaries, chart labels, and stored data are all in the same timezone — no conversion errors possible.
- Deterministic across all browsers and environments.
- No configuration surface to maintain.

**Weaknesses:**

- Displayed times do not match the user's wall clock or calendar. A topic created "this morning" may show as yesterday or tomorrow depending on offset.
- Peak activity heatmap is misleading for any team not in UTC±0. The hour distribution does not reflect local work patterns, undermining UC-19's stated goal of informing staffing decisions.
- Custom date range picks are unintuitive — "today" in the date picker may not correspond to the UTC "today" shown in the data.
- The `formatSyncTime` inconsistency (browser-local) would need to be resolved by forcing it to UTC too, making the sync time less useful.

**Best suited for:** Teams operating in or near UTC, or deployments where absolute consistency is valued over local relevance.

### Alternative B — Browser-local timezone (auto-detected)

Display all timestamps in the user's browser timezone, detected via `Intl.DateTimeFormat().resolvedOptions().timeZone`.

**How it works:**

- All `toLocaleDateString` and `toLocaleString` calls omit the `timeZone` option (or explicitly pass the detected zone).
- Peak activity heatmap converts UTC timestamps to local hour before bucketing.
- Custom date range boundaries are interpreted as midnight-to-midnight in the browser's timezone, then converted to UTC for filtering.
- No configuration needed — the browser provides the timezone automatically.

**Strengths:**

- Dates and times match the user's wall clock. "Today" means today. "09:00" means 09:00 local time.
- Peak activity heatmap reflects actual local work patterns, directly supporting staffing decisions.
- Custom date range is intuitive — picking "March 15" means the user's local March 15.
- Zero configuration — the browser handles detection automatically.
- `formatSyncTime` already works this way, so the dashboard becomes internally consistent.

**Weaknesses:**

- Two users in different timezones see different data for the same date range. "Last 7 days" covers a different UTC window depending on the viewer's offset. This makes it harder to share findings or compare notes.
- Daily and weekly chart labels shift between users. A topic may appear under "Monday" for one user and "Tuesday" for another.
- The peak activity heatmap becomes viewer-dependent — the same data produces different heat patterns for users in different timezones. If the dashboard is used by a distributed team to make shared staffing decisions, they may see conflicting pictures.
- Custom range filtering becomes more complex: the frontend must convert local date boundaries to UTC before comparing against stored timestamps.
- DST transitions introduce edge cases — a day near a DST change has 23 or 25 hours, which affects daily bucketing and "last 7 days" calculations (7 × 24h ≠ 7 calendar days).
- Testing becomes harder — tests must account for timezone-dependent behavior or mock the browser timezone.
- Server-rendered or cached views would not reflect the correct timezone.

**Best suited for:** Dashboards with a single user or where each user's local perspective is more important than cross-user consistency.

### Alternative C — Configured forum timezone

A single timezone is configured per deployment (e.g., `"Europe/Stockholm"`) and all display uses that timezone. The setting is stored in a configuration file alongside existing forum settings.

**How it works:**

- A new field (e.g., `displayTimezone: "Europe/Stockholm"`) is added to the deployment configuration.
- All `toLocaleDateString` and `toLocaleString` calls pass this configured timezone.
- Peak activity heatmap converts UTC timestamps to the configured timezone's hour before bucketing.
- Custom date range boundaries are midnight-to-midnight in the configured timezone.
- The frontend reads the timezone from the configuration (served via API or embedded in the page).

**Strengths:**

- All users see the same times and dates — shared reference frame for discussion, reports, and staffing decisions.
- Peak activity heatmap reflects the forum community's local work patterns, directly supporting UC-19.
- Matches the mental model of "this forum operates in timezone X" — which is how Discourse itself works (every Discourse instance has a configured timezone).
- Custom date range is intuitive for users in or near the configured zone.
- Only one timezone conversion to implement and test, not N (one per user).
- Simpler than per-user timezone — no detection, no per-session state, no conflicting views.
- The `Intl.DateTimeFormat` API natively supports IANA timezone identifiers, so implementation uses standard browser APIs.

**Weaknesses:**

- Users far from the configured timezone see times that don't match their wall clock. For a globally distributed team, this is the same problem as Alternative A but shifted.
- Requires a configuration value that someone must set and maintain. If the forum's community shifts timezones (e.g., the team moves from Stockholm to New York), the configuration must be updated.
- DST transitions in the configured timezone affect display — the same UTC instant maps to different local hours depending on whether DST is active. The `Intl` API handles this correctly, but it means displayed hours are not uniformly spaced.
- Adds a deployment-time decision that does not exist today.

**Best suited for:** Single-community forums where the community operates primarily in one timezone. This is the most common Discourse deployment pattern.

### Alternative D — User-selectable timezone

The dashboard provides a timezone selector (dropdown or similar) that lets each user choose their preferred display timezone. The selection persists in browser local storage.

**How it works:**

- A timezone picker component is added to the dashboard UI.
- The selected IANA timezone identifier is stored in `localStorage`.
- All display formatting uses the selected timezone (same mechanism as Alternative C, but per-user).
- Defaults to the browser's detected timezone (Alternative B behavior) if no selection is stored.

**Strengths:**

- Maximum flexibility — each user sees times in their preferred timezone.
- Combines the intuitiveness of local time with the ability to switch to a shared reference when collaborating.
- A user can set it to match the forum timezone for consistency with colleagues, or to their own timezone for personal use.
- localStorage persistence means the choice survives page reloads without requiring user accounts.

**Weaknesses:**

- Highest implementation complexity — requires a timezone picker component, localStorage management, and propagation of the selected timezone to every formatting call.
- The IANA timezone database contains ~400+ entries. A usable picker needs grouping, search, or a curated subset.
- Two users discussing the same data may be looking at different timezone views without realizing it, unless the active timezone is prominently displayed.
- Same DST complexity as Alternatives B and C.
- Adds UI surface area and cognitive load for a feature that may not be needed if the team is co-located.
- Testing must cover the default (no selection), explicit selection, and timezone changes mid-session.

**Best suited for:** Multi-user dashboards with a geographically distributed audience and no single dominant timezone.

### Alternative E — Hybrid: UTC for ranges, configured timezone for clock times

Dates (March 15, week of March 10) remain in UTC. Clock times (14:32, hour 8) and time-of-day bucketing use a configured forum timezone.

**How it works:**

- Date labels on charts, custom range inputs, and filter boundaries continue using UTC days.
- The peak activity heatmap converts to the configured timezone before bucketing by hour.
- Timestamps that include a clock component (e.g., "Mar 15, 14:32") display in the configured timezone.
- The dashboard indicates which timezone applies where, e.g., "(UTC)" next to date ranges and "(CET)" next to clock times.

**Strengths:**

- Filter boundaries and date-based grouping remain simple and unambiguous (UTC days).
- The peak activity heatmap — the feature most affected by timezone choice — shows meaningful local hours.
- Preserves the current behavior for date ranges and trend charts while fixing the most confusing display.
- Smaller implementation scope than full timezone conversion.

**Weaknesses:**

- Two timezone references in the same UI is inherently confusing. A topic "created March 15 at 01:30 CET" might sit under the "March 14" UTC bucket in a trend chart.
- The inconsistency requires explanation — timezone labels or tooltips become necessary.
- Custom date range remains unintuitive for non-UTC users (same problem as Alternative A for this specific interaction).
- More complex to explain in documentation and to new contributors than a single consistent rule.
- Edge cases multiply — any feature that combines a date and a time must decide which rule applies.

**Best suited for:** Situations where full timezone conversion is deemed too costly but the peak activity heatmap must reflect local hours. A pragmatic compromise, not an ideal end state.

## Narrowing the scope

Alternatives A through E treat timezone display as a system-wide concern. After reviewing the dashboard views against the friction points listed in Context, a more targeted analysis emerges:

- **Queue tables, response metrics, SLO monitoring, tag distribution, response time distribution** — these show durations (e.g., "4h", "2d"), counts, and percentages. Timezone is irrelevant to these displays.
- **Trend charts (response time trends, topic intake)** — these group by day or week. A topic shifting from one UTC day to the next at the boundary is a minor edge case that does not change the trend shape. All users seeing the same day/week boundaries is more valuable than each seeing their own.
- **Time period filter** — rolling windows (`Date.now() - 7 * 86_400_000`) are timezone-neutral. Custom date ranges using UTC boundaries are predictable and consistent. The slight mismatch with a user's local calendar is an acceptable trade-off for cross-user consistency.
- **Peak activity heatmap** — this is the one view where hour-of-day meaning is timezone-dependent. Its stated purpose (UC-19) is to inform staffing decisions, which are inherently local-time questions. This is where UTC causes real friction.

This means UTC is the correct choice for the system as a whole, and the problem reduces to: **how should the peak activity heatmap help users in different timezones read hour-of-day patterns?**

This reframing opens an alternative that the system-wide analysis would not surface — one that leaves the data untouched and adds timezone context as a visual reading aid.

### Alternative F — User-added timezone header rows on the heatmap

The heatmap keeps its UTC data grid unchanged. Users can add up to three additional timezone header rows above the UTC hour header, each showing offset hour labels for a chosen timezone. The UTC row is always present and visually marked as the primary reference. Timezone selections are persisted in a cookie with explicit user consent.

**How it works:**

- The heatmap renders a UTC hour header row (0–23) by default, with an "Add timezone" button.
- Clicking the button opens a timezone picker (e.g., a searchable dropdown of common IANA timezones or a curated list grouped by region). The user selects a timezone, and a new header row appears above the UTC row showing offset hours for that timezone.
- Each added timezone row has a remove button so the user can delete it individually.
- The "Add timezone" button is hidden once three timezone rows are present. It reappears when one is removed.
- The UTC row is always present, always at the bottom of the header area, and cannot be removed.

**Persistence via cookie with consent:**

- When a user adds their first timezone, a modal appears explaining that the selection will be stored in a browser cookie so it persists across visits.
- The modal offers two choices: **Accept** (store the cookie, persist the selection) or **Deny** (the timezone is shown for the current session only and lost on page reload).
- If the user chose Deny, the consent modal appears again the next time they add a timezone — giving them the opportunity to change their mind.
- If the user chose Accept, subsequent timezone additions are persisted without re-prompting.
- The cookie stores the list of selected timezone identifiers (e.g., `["Europe/Berlin", "Asia/Kolkata"]`). No personal data beyond the timezone preference is stored.

**Data grid and header interaction:**

- The data grid (7 rows × 24 columns) is unchanged — it still shows topic counts bucketed by UTC day and hour. No re-bucketing occurs.
- Each timezone header row is purely a reading aid — a ruler the user holds against the same data to find "my 09:00".
- Hour labels are computed from the IANA timezone identifier using `Intl.DateTimeFormat` to resolve the current UTC offset, which means the displayed offset automatically reflects whether DST is active.
- The three-timezone maximum keeps the header area readable.

**Visual sketch:**

```text
                0    1    2    3    4    5    6    7    8    9   10  ...  23
  IST (+5:30)  5:30 6:30 7:30 8:30 9:30 10:30 ...                       ✕
  CET (+1)      1    2    3    4    5    6    7    8    9   10   11  ...  ✕
  UTC            0    1    2    3    4    5    6    7    8    9   10  ...  23
  Mon            0    0    1    2    5    8   12   15   14   11    7  ...   0
  Tue            ...
                                                              [+ Add timezone]
```

**Strengths:**

- The underlying data is unchanged. All users see the same counts in the same cells. No re-bucketing, no data transformation.
- Each user personalizes their own view — a CET user adds CET, an IST user adds IST. No deployment-time configuration needed.
- Multiple timezones are visible simultaneously. A user working with teams in several zones can add all of them and compare patterns at a glance.
- The UTC row remains the anchor — anyone looking at server logs or raw data can correlate directly.
- Using IANA timezone identifiers with `Intl.DateTimeFormat` means DST is handled automatically — the header shows the correct offset for the current date without manual configuration.
- Cookie consent flow respects user choice. Users who decline cookies still get the feature for the current session.
- No spec changes needed to the data model, bucketing logic, or any other dashboard view. Only the heatmap's header area and the consent flow are affected.
- No backend changes required — the cookie is read and written entirely by the frontend.

**Weaknesses:**

- Day-of-week rows remain in UTC. For a user reading the IST header, the "Monday" row includes their Sunday evening and excludes their Monday evening. This is a known simplification — the heatmap answers "which hours are busiest" more accurately than "which day is busiest," and day-of-week patterns are less affected by timezone offset than hour-of-day patterns.
- The cookie consent modal adds a small interaction cost on first use. This is unavoidable if cookie persistence is desired without assuming consent.
- The timezone picker needs a curated or searchable list of IANA timezones. The full database has ~400+ entries, so some UX work is needed to make selection fast (grouped by region, searchable, or showing common offsets).
- Half-hour offsets (IST +5:30, Nepal +5:45) produce non-integer hour labels. The display must handle this gracefully — showing "5:30", "13:30" rather than rounding.
- The visual design needs care to distinguish timezone header rows from the data grid. Color, typography, or separator lines must make the boundary clear so users do not mistake a header row for a data row.
- Implementation is more involved than a static config file approach — it requires a timezone picker component, cookie read/write logic, and a consent modal. However, the scope remains confined to the peak activity view.

## Comparison

### System-wide alternatives (A–E)

| Concern | A: UTC | B: Browser-local | C: Forum timezone | D: User-selectable | E: Hybrid |
|---|---|---|---|---|---|
| Times match user's clock | No | Yes | For users in/near the zone | Yes (if configured) | Partially |
| All users see the same data | Yes | No | Yes | No | Mostly |
| Peak heatmap reflects local patterns | No | Per viewer | Yes | Per viewer | Yes |
| Custom range is intuitive | No | Yes | For users in/near the zone | Yes | No |
| Implementation complexity | None | Low-medium | Low-medium | Medium-high | Medium |
| Configuration required | None | None | One setting | UI component + storage | One setting |
| DST edge cases | None | Yes | Yes | Yes | Yes (partial) |
| Testing complexity | Low | Medium | Medium | High | Medium-high |
| Cross-user consistency | Full | None | Full | None | Mostly |
| Explanation burden | Low | Low | Low | Low | High |

### Heatmap-scoped alternative (F) compared to system-wide approaches

| Concern | A: UTC | F: Multi-header | B–E (best case) |
|---|---|---|---|
| Users can read local hours on heatmap | No | Up to 3 user-chosen timezones + UTC | Yes (one timezone) |
| All users see the same heatmap data | Yes | Yes (grid unchanged, headers personalized) | B, D, E: No. C: Yes |
| UTC remains visible on heatmap | Yes | Yes (always present, cannot be removed) | B: No. C: No. D: only if selected. E: only if local = UTC |
| Multiple timezones visible simultaneously | No | Yes | No |
| Rest of dashboard unchanged | Yes | Yes | No — all views affected |
| Implementation complexity | None | Medium (picker, consent, cookie) | Low-medium to high |
| Configuration required | None | None (user self-service) | Varies |
| DST handling needed | No | Yes (automatic via Intl API) | B, C, D, E: Yes (same) |
| Spec changes required | None | PA-5 addendum, PA-7 addendum | Multiple specs across dashboard |

### Interaction with existing specs

| Spec | Reference | Impact per alternative |
|---|---|---|
| [time-period-filter.md](../../specs/dashboard/time-period-filter.md) TF-8 | Custom range uses UTC boundaries | A, F: unchanged. B, C, D: boundaries shift to display timezone. |
| [peak-activity.md](../../specs/dashboard/peak-activity.md) PA-5 | Bucketing uses UTC day and hour | A: unchanged. F: bucketing unchanged, header rows added. B, C, D, E: re-bucketing in different timezone. |
| [peak-activity.md](../../specs/dashboard/peak-activity.md) PA-2, PA-7 | Grid layout and column headers (0–23) | A: unchanged. F: additional header rows above hour numbers. B, C, D, E: header row changes timezone label. |
| [topic-intake.md](../../specs/dashboard/topic-intake.md) TI-2 | Daily bucketing via `toISOString().slice(0, 10)` | A, F: unchanged. B, C, D: day boundaries shift. |
| [response-time-trends.md](../../specs/dashboard/response-time-trends.md) RT-2 | Weekly bucketing uses UTC Monday | A, F: unchanged. B, C, D: week start shifts. |
| [dashboard-components.md](../../specs/dashboard/dashboard-components.md) | `formatWeekLabel`, `formatDayLabel` use `timeZone: "UTC"` | A, F: unchanged. B, C, D, E: timezone parameter changes. |
| [operational-constraints.md](../../specs/operational-constraints.md) | "working hours" undefined timezone | F: user-chosen timezone headers make individual working hours readable. C, E: configured timezone resolves this. |

### Migration cost

For alternatives A–E, the frontend changes span formatting functions, bucketing functions, filter boundary calculations, and tests across multiple files. For Alternative F, changes are confined to:

1. The `PeakActivity` component — render timezone header rows, "Add timezone" button, and per-row remove buttons.
2. A timezone picker component — searchable or grouped list of IANA timezones.
3. A cookie consent modal — shown on first timezone addition, with Accept/Deny flow.
4. Cookie read/write logic — persist and restore the list of selected timezone identifiers.
5. Offset computation — use `Intl.DateTimeFormat` with the selected IANA identifier to resolve the current UTC offset (DST-aware).
6. The peak activity spec — add requirements for the timezone header rows and consent flow.
7. Tests — verify header row offset arithmetic, add/remove interactions, cookie persistence, and consent states.

No changes to bucketing logic, filter logic, other dashboard views, or backend. No external library required — `Intl.DateTimeFormat` and `document.cookie` are native browser APIs.

## Decision

We choose a two-part strategy:

1. **UTC is the system-wide standard for all dashboard views.** Storage, computation, filtering, date labels, trend bucketing, and custom date ranges all use UTC. This is already implemented and the analysis of Alternatives A–E confirms it is the right default: it guarantees cross-user consistency, eliminates conversion errors, and keeps the implementation simple. The only view where UTC causes meaningful friction is the peak activity heatmap.

2. **The peak activity heatmap gets user-added timezone header rows (Alternative F).** Users can add up to three timezone headers above the UTC row, each showing offset hour labels for an IANA timezone of their choice. The data grid remains UTC-bucketed and identical for all users. Timezone selections are persisted via a cookie with explicit consent. This solves the core problem — making hour-of-day patterns readable for users in any timezone — without introducing system-wide timezone complexity.

Alternative F was chosen over the system-wide alternatives because the problem is narrower than initially assumed. Only the heatmap's hour axis is timezone-sensitive; all other dashboard views work well in UTC. Alternatives B–E would introduce timezone conversion across the entire dashboard for a problem that exists in one component. Alternative F keeps the blast radius small: no bucketing changes, no filter changes, no backend changes.

Alternative F was chosen over the simpler static-config variant because the dashboard has no user accounts and serves a potentially distributed audience. A deployment-time config file would require someone to predict which timezones matter. A user-driven UI lets each person add their own timezone without coordinating with the deployment owner.

## Consequences

**Positive:**

- UTC remains the single source of truth across the entire system. No ambiguity about date boundaries, no cross-user data differences, no timezone conversion in storage or computation.
- The peak activity heatmap becomes useful for distributed teams. A user in any timezone can add their local zone and immediately read "my 09:00 is this column."
- Multiple timezones are visible simultaneously — a manager overseeing teams in CET, IST, and PST can add all three and compare patterns in one view.
- The implementation scope is confined to the `PeakActivity` component, a timezone picker, a cookie consent modal, and cookie read/write logic. No other dashboard views are touched.
- DST is handled automatically by using IANA timezone identifiers with `Intl.DateTimeFormat` to compute offsets.
- Users who decline cookies still get the feature for the current session — the consent requirement does not block functionality.

**Negative:**

- Day-of-week rows remain in UTC. Users reading a non-UTC header row may see activity attributed to the "wrong" day near midnight boundaries. This is documented as a known simplification.
- The timezone picker requires UX design work — the IANA database has ~400+ entries, so grouping, searching, or curating is needed.
- The cookie consent modal adds interaction friction on first use. This is inherent to any cookie-based persistence approach.
- Half-hour offsets (IST +5:30, Nepal +5:45) require the header to display non-integer hour labels, which adds visual complexity.
- The `formatSyncTime` inconsistency noted in Context (browser-local while the rest of the dashboard is UTC) is not addressed by this decision. It should be resolved separately — either by making it explicitly UTC or by accepting browser-local as appropriate for a "last synced" timestamp.

**Specs that need updating when this ADR is implemented:**

- [peak-activity.md](../../specs/dashboard/peak-activity.md) — add requirements for timezone header rows (add/remove interaction, three-row limit, UTC always present), cookie consent flow, and visual design constraints.
- [dashboard-components.md](../../specs/dashboard/dashboard-components.md) — document the timezone picker and consent modal components.
- [traceability.md](../../specs/dashboard/traceability.md) — add requirement-to-code mappings for the new components.
