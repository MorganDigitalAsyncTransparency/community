# AI Guidelines

Guidelines for AI tools contributing to this project.

For general code principles and delivery workflow, see the repository-level [CLAUDE.md](../CLAUDE.md). This file covers what is specific to discourse-observer.

## Project context

discourse-observer observes activity in a single Discourse forum — fetching data via the Discourse API, detecting changes, normalizing observations, and storing them for analysis.

- **Single forum per deployment.** Do not generate code that assumes multiple forums, tenants, or data sources.
- **Generic foundation.** Do not hardcode forum names, category IDs, tag names, or community-specific workflows. Forum-specific configuration belongs in `src/config/`.
- **Designed to be forked** for specific forums. Keep the core generic.

## Module boundaries

Each directory under `src/` has a single responsibility. See [ARCHITECTURE.md](ARCHITECTURE.md) for the full layer diagram and dependency flow.

| Module | Responsibility | May depend on |
| --- | --- | --- |
| `src/model/` | Domain types | nothing |
| `src/config/` | Configuration and adaptation | nothing |
| `src/discourse/` | Discourse API integration | `config/` |
| `src/storage/` | Persistence abstraction | `model/`, `config/` |
| `src/observer/` | Change detection and normalization | `discourse/`, `model/`, `config/`, `storage/` |

**Rules:**

- Do not mix responsibilities across modules.
- Discourse API data shapes, pagination, and authentication stay in `src/discourse/`. Other modules work with normalized types from `src/model/`.
- If unsure where something belongs, check the module README.

## Project-specific expectations

- Prefer pure transformations. Reserve side effects for the edges of the system.
- Do not introduce frameworks, ORMs, or libraries without a documented ADR.
- Do not generate speculative code that is not yet needed.
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
