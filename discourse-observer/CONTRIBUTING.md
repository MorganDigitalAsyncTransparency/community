# Contributing to discourse-observer

## Before you start

1. Read [ARCHITECTURE.md](ARCHITECTURE.md) and the relevant module README in `src/` to understand where your change belongs before starting.
2. Read [AI_GUIDELINES.md](AI_GUIDELINES.md) for module boundaries, scope constraints, and coding expectations.
3. Read [docs/documentation-strategy.md](docs/documentation-strategy.md) for how specs, tests, and source files are organized and linked.
4. Run `npm run setup` to install dependencies and activate git hooks.

## How to contribute

1. **Keep PRs small and focused.** Each pull request should do one thing. A new API method, a model change, a bug fix — not all three at once.
2. **Write tests where practical.** Focus tests on transformation logic and observer behavior. Use deterministic inputs, not live API calls. Test behavior, not implementation details.
3. **Document major decisions.** If your change introduces a new dependency, changes a module boundary, or makes an architectural trade-off, record it as an ADR in `docs/decisions/`. See [docs/decisions/README.md](docs/decisions/README.md) for the format.

## Forum-specific assumptions

This project is a generic starter for single-forum deployments. When contributing:

- Do not hardcode forum names, category IDs, tag names, team structures, or community-specific workflows.
- If you need forum-specific behavior, make it configurable through `src/config/`.
- If a change only makes sense for one specific forum, document why in the PR and consider whether it should be in a fork instead.

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
