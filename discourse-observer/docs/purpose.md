# Purpose

## What discourse-observer does

discourse-observer exists to watch a single Discourse forum and make its activity understandable. It fetches data from the Discourse API, normalizes it into a clean internal model, and prepares it for analysis, event extraction, and reporting.

## Why this project exists

Discourse forums generate a continuous stream of activity: topics are created, edited, tagged, categorized, and responded to. Understanding patterns in this activity — response times, topic lifecycle, category health, contributor patterns — requires structured observation rather than ad-hoc API queries.

This project provides the structured observation layer. It separates the concerns of fetching data, detecting changes, modeling domain concepts, and storing results so that each concern can evolve independently.

## Intended usage

One deployment of discourse-observer watches one Discourse forum. The project is designed as a generic starting point:

- A team running a Discourse-based support forum can fork this project, configure it for their forum, and build analysis tools on top of it
- A community manager can adapt it to track engagement patterns specific to their community
- A developer relations team can extend it to observe developer forum activity and feed it into their own reporting tools

Each of these uses starts from the same foundation but adapts it for their specific forum's categories, tags, workflows, and reporting needs.

## What this project does not do

- It does not support multiple forums in a single deployment
- It does not provide a ready-made dashboard (that comes later)
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
