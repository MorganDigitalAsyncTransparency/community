# Operational Constraints — Verification

This document defines how to verify that the operational constraints are respected in the implementation.

Spec: [operational-constraints.md](operational-constraints.md)

---

## What to verify

The target Discourse server has limited resources. The observer must respect rate limits, use incremental synchronization, and keep request frequency conservative.

---

## Verification steps

### 1. Polling interval is configurable

- Confirm that the polling interval is defined in configuration, not hardcoded.
- Confirm that the default interval is conservative (5 minutes or more).

### 2. Incremental synchronization

- Confirm that the observer fetches only data that has changed since the last observation, using timestamps, IDs, or API cursors.
- Confirm that the observer does not perform full re-fetches of all data on every polling cycle.

### 3. Rate limit respect

- Confirm that the Discourse API client respects rate limit headers returned by the server.
- Confirm that there is margin between the observer's request rate and the server's rate limit.
- Confirm that the observer does not issue parallel API requests unless explicitly configured to do so.

### 4. No excessive load

- Confirm that sync duration and request counts are logged, making load problems visible.
- Confirm that a single polling cycle does not issue an unbounded number of API requests.

### 5. Sequential processing by default

- Confirm that the observer processes data sequentially, not in parallel, unless the configuration explicitly enables parallelism.

---

## When to re-run

Re-run this verification when:

- The sync strategy changes
- New API endpoints are added to the fetch cycle
- Polling frequency or concurrency configuration is modified
- The target forum grows significantly in size
