# discourse/

This module is responsible for raw Discourse API integration.

## Responsibility

All communication with the Discourse API happens here. This includes:

- Authenticating with the Discourse instance using API credentials
- Fetching topics, categories, tags, revisions, and other entities
- Handling pagination across list endpoints
- Respecting rate limits and retry logic
- Returning data in shapes close to the API response

## Boundaries

This module does **not**:

- Interpret or analyze the data it fetches
- Detect changes or produce observations
- Store data or manage state

It is a thin integration layer. Its job is to make reliable API calls and return the results.

## Current implementation

`Client` in `client.go` fetches topics (`/latest.json`), categories (`/categories.json`), and topic count (`/about.json`) from a Discourse-compatible HTTP API. It sends `Api-Key` and `Api-Username` headers when credentials are configured. Returns `model.RawTopic` and `model.RawCategory` values. Pagination timing (delay, retries) can be configured at construction via `WithPageConfig` and is applied internally when the observer calls through the `FetchClient` interface.

`mockserver/` provides Discourse-format JSON from the project's mock dataset, sorted by `bumped_at` descending to match real Discourse behavior. It exports a `Handler()` for standalone use (docker-compose service in dev mode) and `New()` / `NewWithPageSize()` for `httptest.Server` use in pipeline and sync integration tests.

## Design expectations

- Functions should map closely to Discourse API endpoints
- Error handling should distinguish between transient failures (network, rate limit) and permanent failures (auth, not found)
- The module should be testable with recorded or mocked API responses
- API credentials and the forum base URL come from `config/`
