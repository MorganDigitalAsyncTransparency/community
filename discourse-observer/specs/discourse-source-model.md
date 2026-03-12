# Discourse Source Model

This document describes the initial set of Discourse data that the observer will work with. It defines what data is fetched from the Discourse API and which fields are relevant for observation.

## Guiding principle

Only model what is needed. The Discourse API exposes extensive data for each entity. The source model should capture only the fields necessary for observation, normalization, and change detection. Additional fields can be added incrementally as new observation needs emerge.

## Core entities

### Topics

Topics are the primary unit of activity in a Discourse forum. A topic represents a conversation thread.

Relevant fields:

- **id** — Unique topic identifier
- **title** — Topic title (may change over time)
- **created_at** — When the topic was created
- **updated_at** — When the topic was last updated (by any activity)
- **category_id** — The category the topic belongs to
- **tags** — List of tags assigned to the topic
- **status** — Whether the topic is open, closed, archived, or otherwise flagged
- **posts_count** — Number of posts in the topic
- **reply_count** — Number of replies
- **views** — View count
- **slug** — URL-friendly identifier

### Categories

Categories organize topics into groups. A forum typically has a fixed set of categories, though they can change over time.

Relevant fields:

- **id** — Unique category identifier
- **name** — Category name
- **slug** — URL-friendly identifier
- **parent_category_id** — If this is a subcategory, the parent's ID
- **description** — Category description
- **topic_count** — Number of topics in the category

### Tags

Tags provide cross-cutting classification for topics. A topic can have multiple tags.

Relevant fields:

- **id** — Unique tag identifier
- **name** — Tag name (used as the primary identifier in most API contexts)
- **topic_count** — Number of topics using this tag

### Revisions

Revisions capture edits to topic posts. They are important for understanding how topic content evolves.

Relevant fields:

- **post_id** — The post that was revised
- **revision_number** — Sequential revision number
- **created_at** — When the revision was made
- **title_changes** — Whether the topic title changed in this revision
- **body_changes** — Whether the post body changed
- **category_changes** — Whether the category changed (for first-post revisions)
- **tag_changes** — Whether tags changed

### Timestamps

All Discourse entities include timestamps that are essential for observation:

- **created_at** — When the entity was first created
- **updated_at** — When the entity was last modified
- **bumped_at** (topics) — When the topic was last bumped in the topic list
- **last_posted_at** (topics) — When the most recent post was added

Timestamps are returned by the Discourse API in ISO 8601 format and should be stored as UTC internally.

## What is not modeled initially

The following Discourse concepts exist but are not part of the initial source model. They can be added later if observation needs require them:

- **Users** — User profiles and activity. May be needed for contributor analysis.
- **Posts** (full content) — Individual post bodies. Topics and post counts may be sufficient initially.
- **Likes and reactions** — Engagement signals. Useful for analytics but not core to observation.
- **Groups** — User groups and membership. Relevant if group-based workflows need tracking.
- **Badges** — Gamification data. Rarely relevant for support observation.
- **Private messages** — Not accessible through standard API credentials and generally out of scope.

## API considerations

- The Discourse API uses cursor-based or page-based pagination depending on the endpoint
- Rate limits vary by endpoint and authentication level
- Some fields are only available with admin-level API keys
- The API may return additional fields not listed here; these should be ignored during normalization rather than causing errors
