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
- Know about internal domain types from `model/`
- Store data or manage state

It is a thin integration layer. Its job is to make reliable API calls and return the results.

## Design expectations

- Functions should map closely to Discourse API endpoints
- Error handling should distinguish between transient failures (network, rate limit) and permanent failures (auth, not found)
- The module should be testable with recorded or mocked API responses
- API credentials and the forum base URL come from `config/`
