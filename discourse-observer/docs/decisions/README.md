# Architecture Decision Records (ADRs)

This directory contains architecture decision records for discourse-observer.

## What is an ADR?

An ADR is a short document that captures a single architectural decision. Each ADR records what was decided, why, what alternatives were considered, and what consequences follow. ADRs are numbered, immutable once accepted, and accumulate over time to form a decision log for the project.

## Why this repository uses ADRs

Software projects accumulate decisions over time. Without a record of what was decided and why, contributors — both human and AI — are left guessing at intent. ADRs prevent this by providing lightweight, searchable documentation of the reasoning behind architectural choices.

This project is designed for AI-assisted contribution. ADRs give AI tools the structured context they need to make decisions consistent with prior choices. They also help new contributors understand the project without digging through git history.

## When to write an ADR

Write an ADR when:

- Adding or removing a dependency
- Changing module boundaries or responsibilities
- Choosing (or rejecting) a framework, library, or platform
- Making a trade-off that future contributors might question
- Establishing a convention that is not obvious from the code alone

Do not write an ADR for routine code changes, bug fixes, or implementation details that are clear from the code.

## ADR format

This repository uses a modernized Nygard-style ADR format. It extends the original format from Michael Nygard's [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) with an explicit **Alternatives Considered** section.

Each ADR contains these sections in this order:

| Section                     | Purpose                                                      |
| --------------------------- | ------------------------------------------------------------ |
| **Title**                   | \`# [Number]. [Title]\` — one decision per ADR               |
| **Status**                  | Current lifecycle state (see below)                          |
| **Date**                    | Date the ADR was created or last updated (\`YYYY-MM-DD\`)    |
| **Context**                 | The situation, forces, and problem that require a decision   |
| **Alternatives Considered** | Options that were evaluated, with brief trade-off notes      |
| **Decision**                | What was decided and why                                     |
| **Consequences**            | What follows from this decision — both positive and negative |

A template is available at [template.md](template.md). Copy it when creating a new ADR.

## Status model

Each ADR has exactly one status:

- **Proposed** — Under discussion, not yet accepted
- **Accepted** — The decision has been made and is in effect
- **Superseded** — Replaced by a later ADR
- **Deprecated** — No longer relevant due to project changes

Status transitions:

\`\`\`text
Proposed → Accepted → Superseded
                    → Deprecated
\`\`\`

An ADR may be created directly as \`Accepted\` when the decision is already made.

## How superseding works

When a decision is replaced:

1. Create a new ADR with the new decision
2. In the new ADR's **Context**, reference the ADR it replaces
3. Update the old ADR's **status line only** to \`Superseded by [NNNN](NNNN-new-decision.md)\`

The old ADR is never deleted and its content is never modified. It remains in the log as an immutable record of what was decided and why. Only the status line may change.

## Checking for conflicts before writing a new ADR

Before writing a new ADR, read all existing Accepted ADRs and check whether the new decision contradicts or replaces any of them. If it does:

- Note the conflict in the new ADR's **Context**
- Plan to supersede the affected ADR once the new one is accepted

Never modify the content of an Accepted ADR to retroactively align it with a newer decision. The status line is the only permitted change.

## Naming convention

ADRs are numbered sequentially:

\`\`\`text
0001-project-foundation.md
0002-technology-choices.md
0003-programming-languages.md
\`\`\`

Use lowercase with hyphens. The number prefix ensures chronological ordering. The name should describe the decision topic, not the outcome.

## Current ADRs

| Number                                     | Title                   | Status   |
| ------------------------------------------ | ----------------------- | -------- |
| [0001](0001-project-foundation.md)         | Project Foundation      | Accepted |
| [0002](0002-technology-choices.md)         | Technology Choices      | Accepted |
| [0003](0003-programming-languages.md)      | Programming Languages   | Accepted |
| [0004](0004-code-quality-tooling.md)       | Code Quality Tooling    | Accepted |
| [0005](0005-storage-format.md)             | Raw Data Storage Format | Proposed |
| [0006](0006-analytical-storage.md)         | Analytical Storage      | Proposed |
