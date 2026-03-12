# Architecture Decision Records (ADRs)

This directory contains architecture decision records for discourse-observer. ADRs capture significant decisions that affect the structure, dependencies, or direction of the project.

## Why ADRs

Software projects accumulate decisions over time. Without a record of what was decided and why, contributors — both human and AI — are left guessing at intent. ADRs prevent this by providing lightweight, searchable documentation of the reasoning behind architectural choices.

## When to write an ADR

Write an ADR when:

- Adding or removing a dependency
- Changing module boundaries or responsibilities
- Choosing (or rejecting) a framework, library, or platform
- Making a trade-off that future contributors might question
- Establishing a convention that is not obvious from the code alone

Do not write an ADR for routine code changes, bug fixes, or implementation details that are clear from the code.

## Naming convention

ADRs are numbered sequentially:

```
0001-project-foundation.md
0002-discourse-client-design.md
0003-storage-backend-choice.md
```

Use lowercase with hyphens. The number prefix ensures chronological ordering. The name should describe the decision topic, not the outcome.

## Statuses

Each ADR has a status:

- **Proposed** — Under discussion, not yet accepted
- **Accepted** — The decision has been made and is in effect
- **Superseded** — Replaced by a later ADR (link to the replacement)
- **Deprecated** — No longer relevant due to project changes

## Template

```markdown
# [Number]. [Title]

**Status:** Accepted | Proposed | Superseded | Deprecated
**Date:** YYYY-MM-DD

## Context

What is the situation? What forces are at play? What problem needs a decision?

## Decision

What was decided and why.

## Consequences

What follows from this decision — both positive and negative.
```

## Current ADRs

| Number | Title | Status |
|--------|-------|--------|
| [0001](0001-project-foundation.md) | Project Foundation | Accepted |
