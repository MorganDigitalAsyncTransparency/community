# CLAUDE.md

## Main Rule

The codebase must always move toward greater clarity, lower complexity, and stronger maintainability.

When in doubt, choose the option that makes intent more explicit, reduces hidden coupling, and keeps responsibilities smaller. Do not add complexity unless it solves a real problem. Do not keep confusing code just because it already exists.

All other rules in this document are subordinate to this principle.

---

## 1. Core Principles

- Clarity over cleverness.
- Simplicity over unnecessary abstraction.
- Readability over compression.
- Maintainability over short-term speed.
- Explicit intent over implicit behavior.
- Small responsibilities over large mixed concerns.
- Consistency over novelty.
- Requirements first.
- Documentation and implementation must stay aligned.
- Quality is part of the work, not a final extra step.

The goal is code that is easy to read, easy to change, and hard to misuse. If a rule is followed mechanically but makes the design worse, the design is still wrong.

---

## 2. Code Rules

### Functions

- A function must do one thing, at one level of abstraction.
- If a function needs "and", "then", or "also" in its description, it does more than one thing.
- Prefer under 20 lines. Over 30 requires justification. Over 50 is a design failure unless clearly warranted.
- Avoid more than 3 parameters. 4+ should be replaced by a small object.
- Boolean flag parameters are not allowed unless clearly justified.

### Files

- A file should represent one coherent responsibility.
- Prefer under 200 lines. Over 300 requires justification. Over 500 is a refactoring target.

### Structure

- No more than 3 levels of nesting. Extract deeper logic into well-named functions.
- Keep conditionals simple. Hide complex branching behind domain concepts.
- Names must reveal intent and reflect domain meaning. Avoid `data`, `utils`, `helper`, `temp`, `misc`, `manager`.

### Duplication and Side Effects

- Remove duplicated business rules aggressively, but do not create shared abstractions until the shared concept is real.
- Hidden side effects are dangerous. A function name must not suggest a read if it also mutates.
- Do not mix query and command behavior in the same function.

### Comments

- Do not write comments that explain confusing code — improve the code instead.
- Use comments for intent, constraints, warnings, and non-obvious context.
- Remove outdated and redundant comments.

---

## 3. Boundaries and Dependencies

Code should depend inward toward stable concepts, not outward toward volatile details.

- Domain logic separated from framework details.
- Parsing separated from business rules.
- Configuration separated from behavior.
- Side-effecting code separated from pure logic.

Frameworks, APIs, file systems, and external services are details. Do not let them dictate the structure of the codebase.

---

## 4. Delivery Workflow

For significant work, follow these phases in order.

**Execution ownership:** Phase 0 and Phase 7 are checkpoints. In Phase 0, present your understanding of the task and wait for confirmation before proceeding. In Phase 7, present the completed work and wait for confirmation before creating the PR. Between these checkpoints, run phases end-to-end without stopping for check-ins. The role of the reviewer is to evaluate a finished deliverable, not to co-drive each step.

**Phase commits:** After any phase that produces a stable artifact (documentation, design decisions, implementation, review fixes), commit before moving to the next phase. This makes it possible to identify where something went wrong and return to a known-good state without losing earlier work.

### Phase 0 — Alignment

This phase is analysis only — no file changes, no branch creation, no commits.

- Understand the request. Identify ambiguity, risk, and hidden assumptions.
- Challenge bad ideas. Suggest better approaches when appropriate.
- Present your interpretation of the scope and wait for confirmation before proceeding.
- If the work is too large for one coherent pull request, split it into smaller packages and define implementation order.

### Phase 1 — Requirements

- Create a feature branch before any work begins.
- Convert the request into clear requirements describing desired state, not code diffs.
- Make requirements concrete, testable, and understandable to someone unfamiliar with the current implementation.

### Phase 2 — Design and Documentation

- Record how the requirements will be fulfilled.
- Update architecture, workflow, or design documentation where relevant.

### Phase 3 — Validation Strategy

- Define how the change will be verified.
- Prefer automated verification. When not practical, define a concrete manual step.

### Phase 4 — Implementation

- Implement the smallest coherent solution that satisfies the requirements.
- Avoid unrelated cleanup unless truly necessary.

### Phase 5 — Review and Consistency Check

- Verify that requirements, documentation, tests, and implementation still match.
- Fix mismatches before considering the work complete.

### Phase 6 — Review

Review from multiple perspectives: maintainability, user clarity, new contributor readability, consistency, edge cases, and whether unnecessary complexity was introduced.

If Phase 6 identifies issues, fix them and return to Phase 5. Repeat until both phases pass with nothing to fix.

### Phase 7 — Rebase and Pull Request

- Rebase the branch onto the latest main. Resolve conflicts if any, then re-run tests and linters to confirm the branch is clean.
- Create a PR with a short imperative title (under 70 characters) and a body containing summary bullets, a test plan checklist, and a link to any related issue.
- When the task originates from a GitHub Issue, include `Closes #<number>` in the PR body.

### Phase 8 — CI, Merge, and Cleanup

- Wait for all CI checks to pass before merging. If any check fails, investigate and fix on the branch.
- Merge the PR.
- Switch to main, pull, and delete the local feature branch.

---

## 5. Testing

Prefer automated verification for business rules, data transformations, parsing, validation, regressions, and integration points.

Tests should be readable, focused, deterministic, maintainable, and tied to behavior — not implementation details.

Tests are part of the design. Untestable code is usually a design warning. Structure code so important logic can be tested without the full runtime environment.

---

## 6. Documentation

Documentation is part of the product. When a change affects system behavior, architecture, workflows, constraints, or important decisions, update the relevant documentation.

Documentation should explain what exists, why it exists, how to change it safely, and how to verify changes. Do not let documentation drift from reality.

---

## 7. Change Discipline

Each change should be as small as reasonably possible while still being coherent, with a clear purpose, clear boundaries, and a verification approach.

Avoid mixing feature work, refactoring, formatting edits, infrastructure changes, and documentation rewrites in the same change unless necessary.

When touching code, improve obvious clarity problems — rename confusing identifiers, split oversized functions, remove dead code, reduce duplication — as long as it stays scoped and safe. Do not perform broad refactors without stating that intention explicitly.

Always leave the code a little better than you found it.

---

## 8. Pragmatism

These rules are strong defaults, not excuses for mindless rule-following.

Do not game line limits with meaningless wrappers. Do not create fake abstractions to satisfy structure rules. Do not move complexity around and pretend it disappeared.

When breaking a rule is the right choice, do it consciously, keep the exception small, and make the reason clear.

---

## 9. Final Rule

The project must remain understandable. A future contributor should be able to open this repository and answer: what the system does, where things belong, how to change it safely, how to verify those changes, and why important decisions were made.

If a change makes those answers harder, it is probably the wrong change.
