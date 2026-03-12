# 2. Technology Choices

**Status:** Accepted
**Date:** 2026-03-12

## Context

The project foundation (ADR 0001) establishes a layered architecture for observing a single Discourse forum. We need to choose concrete technologies for each layer.

The choices are constrained by:

- The system must run on a small server with minimal resource usage
- Local development must be simple for contributors, including AI-assisted ones
- The Discourse forum has ~6k topics, ~38k posts, and up to ~1000 posts/day during working hours
- We want to avoid heavy enterprise infrastructure until the tool proves long-term value

## Alternatives Considered

### Enterprise platform (Java + OpenShift Golden Path)

Integrated DevSecOps tooling and enterprise deployment model. Rejected because the complexity and slow iteration cycle are disproportionate for an internal observability tool that may not survive its proving period.

### Direct frontend integration with Discourse API

The frontend calls the Discourse API directly, removing the need for a backend. Rejected because it offers no caching, makes it harder to control load on Discourse, exposes API credentials to the browser, and limits architectural evolution.

### Snapshot-based analytics (periodic full state dumps)

Periodically snapshot forum state and compute differences offline. Simpler to implement, but cannot reliably track the sequence of changes within a topic. Rejected because workflow visualization requires event-level granularity.

## Decision

We choose a lightweight stack optimized for simplicity and low operational cost:

**Backend:** A lightweight service responsible for polling Discourse, extracting change events, storing them, and exposing data for the frontend.

**Frontend:** React single-page application communicating only with the backend API.

**Storage:** SQLite with an event-based storage model. SQLite requires no separate database process and is sufficient for the expected data volume.

**Sync strategy:** Poll Discourse `/latest.json`, identify recently changed topics using `bumped_at`, fetch revisions for changed topics, and extract tag and category change events. Poll every 5 minutes during working hours, less frequently outside working hours.

**Deployment:** Docker for local development. Simple container deployment for production. No Kubernetes or heavy platform dependencies initially.

## Consequences

**Positive:**

- Low operational cost and minimal infrastructure requirements
- Simple local development with no external service dependencies beyond Docker
- Clear separation between data collection, storage, and presentation
- SQLite eliminates database administration overhead
- Easy to migrate to a heavier stack later if the tool proves its value

**Negative:**

- Requires building and maintaining a custom sync worker
- Event extraction logic must handle Discourse API quirks and edge cases
- Initial full sync of historical data may take significant time
- SQLite may become a bottleneck if data volume grows beyond expectations
