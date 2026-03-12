# CLAUDE.md

## Main Rule

The codebase must always move toward greater clarity, lower complexity, and stronger maintainability.

When in doubt, choose the option that:

- makes the code easier to understand
- makes intent more explicit
- reduces hidden coupling
- keeps responsibilities smaller and clearer
- improves changeability without adding unnecessary abstraction

Do not add complexity unless it solves a real problem.
Do not keep confusing code just because it already exists.
Always leave the codebase cleaner than you found it.

All other rules in this document are subordinate to this principle.

---

# 1. Clean Code Foundation

This repository follows a clean code mindset inspired by:

- clarity of intent
- separation of concerns
- small focused units
- low coupling
- high cohesion
- explicit boundaries
- testability
- continuous refactoring
- the boy scout rule

The goal is not to satisfy style rules mechanically.
The goal is to produce code that is easy to read, easy to change, and hard to misuse.

If a rule is followed mechanically but makes the design worse, the design is still wrong.

---

# 2. Non-Negotiable Principles

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

If something adds complexity without clear value, it should not be added.

---

# 3. Hard Code Rules

These are the default rules for production code.

## Functions

- A function must do one thing.
- A function must do one thing at one level of abstraction.
- If a function needs “and”, “then”, or “also” in its description, it probably does more than one thing.
- Prefer functions under 20 lines.
- Functions over 30 lines require strong justification.
- Functions over 50 lines should be treated as a design failure unless there is a very clear reason.
- Avoid more than 3 parameters where possible.
- 4 or more parameters should usually be replaced by a small object or clearer structure.
- Boolean flag parameters are not allowed unless clearly justified. They usually mean the function does more than one thing.
- A function should have a single, clear reason to change.

## Files

- A file should represent one coherent responsibility.
- Prefer files under 200 lines.
- Files over 300 lines require review and justification.
- Files over 500 lines should be treated as a refactoring target.
- Large files are a design smell even when the code inside is “organized”.

## Nesting

- Avoid deep nesting.
- Prefer no more than 3 levels of indentation.
- If logic is deeper than that, extract intent into smaller functions.

## Conditionals

- Keep conditionals simple.
- Replace repeated conditional logic with clearer structure when appropriate.
- Avoid long if/else chains when a more explicit model would make intent clearer.
- Hide complex branching behind well-named functions or domain concepts.

## Comments

- Do not write comments that explain confusing code when the code can be improved instead.
- Use comments for intent, constraints, warnings, and non-obvious context.
- Do not leave outdated comments.
- Redundant comments should be removed.

## Naming

- Names must reveal intent.
- Names must reflect domain meaning where relevant.
- Avoid vague names like `data`, `utils`, `helper`, `temp`, `misc`, `manager`, or `thing` unless they are genuinely correct.
- Abbreviations should be avoided unless they are well-established in the domain.
- A name should explain why something exists, not just what type of thing it is.

## Duplication

- Duplication is a design problem.
- Remove duplicated business rules aggressively.
- Do not create shared abstractions too early; remove duplication when the shared concept is real, not imagined.

## Side Effects

- Hidden side effects are dangerous.
- A function name must not suggest a read operation if it also mutates state.
- Mutation must be explicit and local where possible.
- Avoid mixing query and command behavior in the same function.

---

# 4. Responsibility Rules

Apply single responsibility at every level:

- a function should have one responsibility
- a module should have one responsibility
- a file should have one responsibility
- a component should have one responsibility
- a class or service should have one responsibility

“Responsibility” means one reason to change.

If one unit changes for unrelated reasons, it is too broad.

---

# 5. Boundaries and Dependency Direction

Code should depend inward toward stable concepts, not outward toward volatile details.

Prefer:

- domain logic separated from framework details
- parsing separated from business rules
- rendering separated from data shaping
- configuration separated from behavior
- side-effecting code separated from pure logic

Frameworks, APIs, file systems, browsers, and external services are details.
Business rules and core behavior are more important than details.

Do not let low-level details dictate the structure of the whole codebase.

---

# 6. General Working Rules

Before making changes:

- understand the request
- check whether it conflicts with existing architecture, requirements, or conventions
- challenge unclear, weak, or contradictory requests
- improve the problem definition before changing the code where needed
- do not guess when important constraints are missing
- do not start implementation before the scope and direction are clear

When working in an existing codebase:

- respect existing patterns unless there is a good reason to change them
- if changing a pattern, make that decision explicit
- avoid mixing unrelated refactors with feature work
- keep changes traceable and reviewable
- improve naming, structure, and clarity when touching code, as long as the work remains scoped and safe

---

# 7. Requirements-Driven Development

Before writing implementation code for any meaningful change:

- convert the request into clear requirements
- write requirements as desired state, not as a description of edits
- make requirements understandable to someone unfamiliar with the current implementation
- separate background and motivation from requirements themselves

Good requirements describe:

- what the system must do
- what users or maintainers must be able to do
- what constraints must hold
- what quality expectations apply

Bad requirements describe:

- code diffs
- temporary implementation details
- vague intentions without observable meaning

---

# 8. Preferred Delivery Order

For significant work, follow this order:

## Phase 0 — Alignment

- understand the request
- identify ambiguity, risk, and hidden assumptions
- challenge bad ideas when necessary
- suggest a better approach when appropriate
- agree on scope before proceeding

If the requested work is too large for one focused session or one coherent pull request:

- split it into smaller work packages
- define a sensible implementation order
- stop after producing the breakdown if needed

Do not rush into implementation when decomposition is the better choice.

## Phase 1 — Requirements

- write or update requirements for the agreed scope
- make them concrete, testable where possible, and easy to understand

## Phase 2 — Design / Documentation

- record how the requirements are expected to be fulfilled
- update architecture, workflow, design, or operational documentation where relevant

## Phase 3 — Validation Strategy

- define how the change will be verified
- prefer automated verification where practical
- if something cannot reasonably be automated, define a manual verification step

## Phase 4 — Implementation

- implement the smallest coherent solution that satisfies the requirements
- avoid unrelated cleanup unless truly necessary

## Phase 5 — Review and Consistency Check

- check that requirements, documentation, tests, and implementation still match
- fix mismatches before considering the work complete

## Phase 6 — Final Review

Review from multiple perspectives:

- developer maintainability
- user clarity
- new contributor readability
- consistency with the rest of the project
- edge cases and failure handling
- whether unnecessary complexity was introduced

Repeat until no meaningful issues remain, or diminishing returns are clear.

---

# 9. Testing and Verification

Testing should match the nature of the system.

Prefer automated verification for:

- business rules
- data transformations
- parsing
- validation
- regressions
- integration points that can be tested reliably

When something cannot or should not be automated, define a concrete manual verification step.

Tests should be:

- readable
- focused
- deterministic
- maintainable
- tied to behavior, not brittle implementation trivia

Tests are part of the design.
Untestable code is usually a design warning.

Code should be structured so that important logic can be tested without needing the whole runtime environment.

---

# 10. Documentation Expectations

Documentation is part of the product.

When a change affects:

- system behavior
- architecture
- workflows
- constraints
- operating assumptions
- important decisions

update the relevant documentation.

Documentation should explain:

- what exists
- why it exists
- how to change it safely
- how to verify changes

Do not let documentation drift away from reality.

---

# 11. Change Discipline

Each change should be as small as reasonably possible while still being coherent.

A change should have:

- a clear purpose
- clear boundaries
- understandable reasoning
- a verification approach

Avoid mixing these together unless necessary:

- feature work
- refactoring
- formatting-only edits
- infrastructure changes
- documentation rewrites
- renaming unrelated code

Smaller, clearer changes are easier to review, test, and trust.

---

# 12. Refactoring Rule

Refactoring is not separate from delivery.
It is part of responsible delivery.

When touching code, improve obvious clarity problems if you can do so safely without expanding scope too far.

Examples:

- rename confusing identifiers
- split oversized functions
- remove dead code
- reduce duplication
- clarify control flow
- strengthen boundaries between concerns
- extract pure logic from side-effect-heavy code

Do not perform broad refactors without stating that intention explicitly.

---

# 13. Boy Scout Rule

Always leave the code a little better than you found it.

This can mean:

- better names
- smaller functions
- less duplication
- clearer structure
- fewer hidden side effects
- more accurate documentation
- stronger tests
- removed dead code

Small improvements, done continuously, are preferred over occasional giant cleanup efforts.

---

# 14. Pragmatism Rule

These rules are strong defaults, not excuses for mindless rule-following.

Do not game line limits by splitting code into meaningless wrappers.
Do not create fake abstractions just to satisfy structure rules.
Do not move complexity around and pretend it disappeared.

The purpose of every rule is better design.

When breaking a rule is the right choice:

- do it consciously
- keep the exception small
- make the reason clear in code or review context
- do not let the exception become the norm

---

# 15. Final Rule

The project must remain understandable.

A future contributor should be able to open this repository and answer:

- what the system does
- where things belong
- how to change it safely
- how to verify those changes
- why important decisions were made

If a change makes those answers harder, it is probably the wrong change.