# Scaffold Next Step

You are helping develop discourse-observer, a project for observing activity from a single Discourse forum.

The project foundation is in place: folder structure, documentation, architecture boundaries, and contributor guidance. The next step is to begin the first implementation layer.

## Your task

Scaffold the first implementation code for discourse-observer. This likely includes:

### 1. Discourse client abstraction

Create a minimal client in `src/discourse/` that can:
- Connect to a Discourse instance using a base URL and API key from `src/config/`
- Fetch a list of recent topics
- Fetch a single topic by ID
- Handle basic error cases (auth failure, not found, rate limited)

Keep the client minimal. Do not build every API endpoint — start with topics only.

### 2. Minimal source model

Create initial types in `src/model/` that represent:
- A normalized topic (independent of the API response shape)
- A category reference (ID and name)
- A tag reference (name)
- An observation record (what was observed, when, what changed)

These types should be clean, small, and composable.

### 3. Testable transformation flow

Create initial observer logic in `src/observer/` that:
- Takes a raw topic response (as returned by the discourse client)
- Transforms it into a normalized topic using model types
- Returns a structured observation

Write tests in `tests/` that verify this transformation with fixture data.

## Constraints

- Do not introduce a web framework, ORM, or external dependencies beyond an HTTP client
- Do not build a full sync loop or scheduler yet
- Do not build storage persistence yet (the storage interface can remain abstract)
- Keep the implementation language consistent with any decisions already recorded in the project
- Update module READMEs if you add new files or change responsibilities
- Create an ADR in `docs/decisions/` for any significant decisions you make (language choice, HTTP client choice, etc.)
