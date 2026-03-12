# Contributing to discourse-observer

## How to contribute

1. **Read the architecture first.** Before writing code, read [ARCHITECTURE.md](ARCHITECTURE.md) and the relevant module README in `src/`. Understand where your change belongs before starting.
2. **Keep PRs small and focused.** Each pull request should do one thing. A new API method, a model change, a bug fix — not all three at once. Small PRs are easier to review, easier to revert, and easier for AI tools to reason about.
3. **Write tests where practical.** Focus tests on transformation logic and observer behavior. Avoid tests that are fragile, depend on network calls, or test implementation details rather than outcomes.
4. **Document major decisions.** If your change introduces a new dependency, changes a module boundary, or makes an architectural trade-off, record it as an ADR in `docs/decisions/`. See [docs/decisions/README.md](docs/decisions/README.md) for the format.

## Clean code expectations

- **Small functions.** Each function should do one thing. If a function needs a comment explaining what a block does, that block should probably be its own function.
- **Pure transformations where possible.** Prefer functions that take input and return output without side effects. This makes code easier to test and reason about.
- **Clear naming.** Names should describe what something is or does, not how it works internally. Avoid abbreviations unless they are universally understood.
- **No premature abstraction.** Three similar lines of code are better than a helper that is used once. Extract shared logic only when the pattern is clear and repeated.

## Module boundaries

Each directory under `src/` has a specific responsibility described in its README. Respect these boundaries:

- Discourse API logic stays in `src/discourse/`
- Observation and change detection stays in `src/observer/`
- Normalized types and domain concepts stay in `src/model/`
- Forum-specific configuration stays in `src/config/`
- Persistence concerns stay in `src/storage/`

Do not mix responsibilities across modules. If you are unsure where something belongs, open an issue or discussion before writing code.

## Forum-specific assumptions

This project is a generic starter for single-forum deployments. When contributing:

- Do not hardcode forum names, category IDs, tag names, or team structures
- If you need forum-specific behavior, make it configurable through `src/config/`
- If a change only makes sense for one specific forum, document why in the PR and consider whether it should be in a fork instead
- Record any forum-specific assumptions in an ADR

## Commit messages

Write commit messages that explain **why** the change was made, not just what changed. The diff already shows what changed.

Good: `Add rate limiting to Discourse client to avoid API throttling`
Not helpful: `Update discourse client`

## Code review

All changes go through pull request review. Reviewers should check:

- Does the change respect module boundaries?
- Is the change tested where practical?
- Does the change introduce forum-specific assumptions?
- Is the change documented if it affects architecture?
