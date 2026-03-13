# Purpose

## What discourse-observer does

discourse-observer exists to watch a single Discourse forum and make its activity understandable. It fetches data from the Discourse API, normalizes it into a clean internal model, and prepares it for analysis, event extraction, and reporting.

## Why this project exists

The primary motivation is understanding how support work flows through a Discourse forum. Topics are created, triaged, moved between categories, tagged, responded to, and eventually resolved. Understanding these patterns — response times, topic lifecycle, category health, workflow bottlenecks — requires structured observation rather than ad-hoc API queries.

Some of these workflow events (category changes, tag reassignments, title edits) cannot be reliably captured from snapshots of current state alone. Revisions and history matter: the movement of a topic through categories and tags over time is what makes the support workflow visible.

This project provides the structured observation layer. It separates the concerns of fetching data, detecting changes, modeling domain concepts, and storing results so that each concern can evolve independently.

While the project is built as a generic foundation (not hardcoded to any specific forum), the design is informed by the needs of support-focused forums where workflow visibility is the primary goal.

## Intended usage

One deployment of discourse-observer watches one Discourse forum. The project is designed as a generic starting point:

- A team running a Discourse-based support forum can fork this project, configure it for their forum, and build analysis tools on top of it
- A community manager can adapt it to track engagement patterns and community health
- A developer relations team can extend it to observe developer forum activity and feed data into their own reporting tools

Each of these uses starts from the same foundation but adapts it for their specific forum's categories, tags, workflows, and reporting needs.

## What this project does not do

- It does not prescribe a specific runtime platform or hosting model
- It does not assume which parts of forum activity matter most — that is a decision for the team adapting it

## Direction

The expected evolution of this project is:

1. **Foundation** (current stage) — Project structure, documentation, architecture boundaries
2. **Discourse client** — Minimal API integration to fetch topics, categories, tags
3. **Observer logic** — Change detection, normalization, observation lifecycle
4. **Internal model** — Normalized domain types independent of the API
5. **Storage** — Persisting observations for later retrieval
6. **Event extraction** — Deriving meaningful events from observation history
7. **Backend API** — Serving processed data to consumers
8. **Dashboard / Reporting** — Visualizing patterns and trends

Each stage builds on the previous one. The foundation stage ensures that all future stages have a clear place to land.
