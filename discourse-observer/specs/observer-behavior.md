# Observer Behavior

This document describes the expected behavior of the observation layer at a high level. It defines what the observer does, not how it is implemented.

## Role

The observer is the core logic layer of discourse-observer. It sits between the raw Discourse API integration and the internal domain model. Its job is to take raw forum data and produce structured, meaningful observations.

## Expected behavior

### Fetch data from Discourse

The observer coordinates data retrieval from a single Discourse forum through the discourse client layer. It determines what to fetch and when to fetch it. The observer does not call the Discourse API directly — it delegates to the discourse module. Pagination, rate limits, and retries are handled by the discourse adapter, not the observer.

Fetching includes:

- Topics (new and updated)
- Topic revisions and edits
- Categories and category changes
- Tags and tag assignments

User activity (authorship, assignment) is not part of the initial model but may be added later if observation needs require it (see [discourse-source-model.md](discourse-source-model.md)).

### Why revisions matter

A snapshot of a topic's current state shows where it ended up, but not how it got there. In support workflows, the path matters: a topic may be created in one category, moved to another during triage, re-tagged as it is escalated, and moved again before resolution. These transitions are the workflow.

Some of these changes are only visible through revision history. Without revisions, the observer would see a topic's current category and tags but not the sequence of changes that brought it there. Since understanding workflow movement is a primary goal, the observer must fetch and preserve revision data, not just current state.

### Normalize and model data

Raw Discourse API responses contain more data than needed, in shapes dictated by the API rather than by the project's domain. The observer transforms this raw data into normalized internal types defined in the model module.

Normalization includes:

- Extracting only the fields that matter for observation
- Converting API-specific formats (timestamps, IDs, nested structures) into internal representations
- Establishing relationships between entities (topic belongs to category, topic has tags)

### Detect changes

The observer compares current data with previously observed data to detect meaningful changes:

- New topics appearing
- Topics changing category or tags
- Topics receiving new replies
- Topics being closed, archived, or otherwise changing status
- Edits to topic content

Change detection produces observations — records of what changed, when, and in what context.

### Expose reusable observer functionality

The observer exposes its functionality as composable operations that can be used in different contexts:

- A scheduled sync that runs periodically
- An on-demand fetch for specific topics or categories
- A backfill operation for historical data

These operations share the same normalization and change detection logic but differ in scope and trigger.

## Support for later event extraction and analytics

The observer does not perform analytics itself, but it produces observations in a form that supports downstream analysis:

- Observations include timestamps for ordering and windowing
- Observations reference stable internal IDs for joining and grouping
- Observations capture both the current state and the nature of the change
- Observations are stored through the storage abstraction for later retrieval

This design allows a future event extraction layer to derive higher-level events (response time patterns, topic lifecycle events, category health metrics) from the observation stream without re-fetching data from Discourse.

## Boundaries

The observer does **not**:

- Serve HTTP requests or expose an API
- Render dashboards or reports
- Make decisions about what observations mean (that is for analytics)
- Store data directly (it uses the storage abstraction)
- Know about forum-specific categories, tags, or workflows (it receives config values)
