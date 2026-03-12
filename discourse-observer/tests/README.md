# Tests

This directory contains tests for discourse-observer.

## Testing philosophy

### Focus on transformation logic

The most valuable tests verify that data transformations produce correct results. Given known input from the Discourse API, does the observer produce the expected normalized output? Given two snapshots, does change detection identify the right differences?

### Focus on observer behavior

Test that the observer does what it should: fetches the right data, normalizes it correctly, detects changes accurately, and produces well-formed observations. These are the core behaviors that must remain correct as the project evolves.

### Avoid fragile tests

Do not write tests that:

- Depend on live API calls to a real Discourse instance
- Break when unrelated fields are added to data structures
- Assert on implementation details rather than observable behavior
- Require complex setup or teardown that is hard to maintain

### Prefer deterministic tests

Tests should produce the same result every time they run. Use recorded API responses, fixture data, or constructed inputs rather than depending on external state. Randomized or time-dependent tests should be clearly marked and isolated.

## Organization

Tests should mirror the source structure:

```
tests/
  discourse/    # Tests for API integration (using recorded responses)
  observer/     # Tests for observation logic and change detection
  model/        # Tests for model validation and type behavior
  config/       # Tests for configuration loading and defaults
  storage/      # Tests for storage interface implementations
```

## Running tests

Test runner and commands will be established when the first implementation code is added. The CI workflow in `.github/workflows/ci.yml` will run tests automatically on pull requests.
