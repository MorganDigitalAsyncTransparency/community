# Copilot Instructions for discourse-observer

## Project purpose

discourse-observer is a generic starter project for observing activity from a single Discourse forum. It fetches data from the Discourse API, normalizes it, detects changes, and stores observations for later analysis and reporting.

## Scope

- **Single forum per deployment.** Do not generate code that assumes multiple forums, tenants, or data sources. Each deployment connects to one Discourse instance.
- **Generic foundation.** Do not hardcode forum names, category IDs, tag names, or community-specific workflows. Forum-specific configuration belongs in `src/config/`.

## Architecture boundaries

The source code is organized into layers with strict boundaries:

| Module | Responsibility | Depends on |
|--------|---------------|------------|
| `src/discourse/` | Discourse API calls | `config/` |
| `src/observer/` | Normalization and change detection | `discourse/`, `model/`, `config/`, `storage/` |
| `src/model/` | Domain types (no dependencies) | nothing |
| `src/config/` | Configuration and adaptation | nothing |
| `src/storage/` | Persistence abstraction | `model/`, `config/` |

When generating code:
- Put API integration in `src/discourse/`
- Put observation logic in `src/observer/`
- Put types in `src/model/`
- Put configuration in `src/config/`
- Put persistence in `src/storage/`
- Do not mix these responsibilities

## Code expectations

- Keep functions small and focused on one task
- Prefer pure transformations over side-effectful code
- Do not introduce frameworks, ORMs, or libraries without a documented reason
- Do not add error handling for scenarios that cannot occur
- Do not create abstractions for patterns that appear only once
- Name functions and variables for what they represent, not how they work internally

## Documentation

- When generating code that changes module boundaries, update the relevant README
- When making architectural decisions, suggest creating an ADR in `docs/decisions/`
- Write commit messages that explain why, not just what

## Testing

- Focus tests on transformation logic and observer behavior
- Use deterministic inputs (fixtures, recorded responses), not live API calls
- Do not write tests that assert on implementation details
