# Single Forum Scope — Verification

This document defines how to verify that the single-forum scope constraint holds.

Spec: [single-forum-scope.md](single-forum-scope.md)

---

## What to verify

The system assumes exactly one Discourse forum per deployment. This constraint is structural — it affects configuration, data model, storage, and observation logic.

---

## Verification steps

### 1. Configuration accepts exactly one forum

- Confirm that `.env.example` defines a single `DISCOURSE_BASE_URL`, a single `DISCOURSE_API_TOKEN`, and a single `DISCOURSE_API_USER`.
- Confirm that `backend/config/` does not support arrays, maps, or loops over multiple forum configurations.
- There is no per-forum routing, tenant ID, or namespace disambiguation in the configuration layer.

### 2. Data model has no multi-forum awareness

- Confirm that domain types in `backend/model/` do not include a forum identifier field.
- Confirm that storage schemas (NDJSON files and SQLite tables) do not partition data by forum.

### 3. Observation logic targets one forum

- Confirm that the observer fetches from a single base URL and does not iterate over multiple forum endpoints.
- Confirm that API client initialization takes a single set of credentials.

### 4. No multi-tenant infrastructure

- Confirm that `docker-compose.yml` defines one backend instance, not a per-forum fleet.
- Confirm that there is no tenant routing, per-forum database selection, or cross-forum data isolation logic.

---

## When to re-run

Re-run this verification when:

- New configuration options are added
- The data model changes
- Storage schemas are created or modified
- The deployment topology changes
