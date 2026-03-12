# Operational Constraints

This document describes the operational context that informs design decisions around polling, synchronization, and resource usage.

## Forum size and growth

The target forum is moderate in size and growing. It is not large enough to require heavy infrastructure, distributed processing, or aggressive optimization from the start. The design should handle current volumes comfortably and scale incrementally as the forum grows.

This means:

- Start with simple, sequential processing
- Avoid premature optimization for scale that does not yet exist
- Design storage and sync patterns that can be improved later without architectural changes

## Server constraints

The Discourse server has limited resources. The observer must not place unnecessary load on it. Design decisions should favor:

- Incremental synchronization over full re-fetches
- Respecting API rate limits with margin
- Fetching only what has changed since the last observation
- Avoiding parallel API requests unless the server can handle them
- Keeping request frequency conservative by default

If the observer causes noticeable load on the forum, the design has failed its constraints.

## Activity patterns

Most forum activity occurs during working hours. Outside of those hours, activity drops significantly.

This pattern has design implications:

- Sync frequency does not need to be uniform across the day
- A future optimization is to poll more frequently during active hours and less frequently during quiet periods
- The initial implementation can use a fixed polling interval, but the config layer should allow interval adjustment so that time-based scheduling can be added later without restructuring

## What this means for implementation

These constraints do not require complex solutions. They require awareness:

- Default to the simplest sync strategy that respects the server
- Make polling intervals configurable from the start
- Prefer incremental fetching patterns (using timestamps, IDs, or API cursors)
- Do not build time-based scheduling in the first iteration, but do not make it hard to add later
- Monitor and log sync duration and request counts so that problems become visible early
