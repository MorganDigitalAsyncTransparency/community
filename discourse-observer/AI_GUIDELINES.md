# AI Guidelines

Guidelines for AI tools contributing to this project.

For general code principles and delivery workflow, see the repository-level [CLAUDE.md](../CLAUDE.md). This file covers what is specific to discourse-observer.

## Project context

discourse-observer observes activity in a single Discourse forum — fetching data via the Discourse API, detecting changes, normalizing observations, and storing them for analysis.

- **Single forum per deployment.** Do not generate code that assumes multiple forums, tenants, or data sources.
- **Generic foundation.** Do not hardcode forum names, category IDs, tag names, or community-specific workflows. Forum-specific configuration belongs in `backend/config/`.
- **Designed to be forked.** Keep the core generic; forum-specific adaptation belongs in a fork.

## Module boundaries

Each directory under `backend/` has a single responsibility. See [ARCHITECTURE.md](ARCHITECTURE.md) for the full layer diagram and dependency flow.

| Module | Responsibility | May depend on |
| --- | --- | --- |
| `backend/model/` | Domain types | nothing |
| `backend/config/` | Configuration and adaptation | nothing |
| `backend/discourse/` | Discourse API integration | `model/`, `config/` |
| `backend/storage/` | Persistence abstraction | `model/`, `config/` |
| `backend/observer/` | Change detection and normalization | `model/` |

**Rules:**

- `observer` defines interfaces for fetching and storing data. `discourse` and `storage` implement those interfaces. At runtime they are injected into the observer. The observer never imports `discourse` or `storage` directly.
- Do not mix responsibilities across modules.
- Discourse API data shapes, pagination, and authentication stay in `backend/discourse/`. Other modules work with normalized types from `backend/model/`.
- If unsure where something belongs, check the module README.

## Project-specific expectations

- Prefer pure transformations. Reserve side effects for the edges of the system.
- Do not introduce frameworks, ORMs, or libraries without a documented ADR.
- Do not generate code that is not yet needed.
- Do not create abstractions for patterns that appear only once.
- Do not add configuration complexity before the feature it configures exists.
- Do not generate placeholder implementations that will need to be entirely rewritten.

## Documentation

- When changing module boundaries, update the module README and [ARCHITECTURE.md](ARCHITECTURE.md).
- When making architectural decisions, record them as an ADR in `docs/decisions/`.
- Commit messages explain why, not just what.

## Testing

- Focus tests on transformation logic and observer behavior.
- Use deterministic inputs (fixtures, recorded responses), not live API calls.
- Test behavior, not implementation details.
