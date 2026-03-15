# config/

This module handles forum-specific configuration and adaptation points.

## Responsibility

Every deployment of discourse-observer targets a specific Discourse forum. This module is where that specificity lives:

- Forum base URL and API credentials
- Polling intervals and sync schedules
- Which categories or tags to observe (if scoped)
- Forum-specific mappings (category names to internal labels, tag groupings)
- Feature flags for optional observation behaviors

## Boundaries

This module:

- Is read by other modules but does not depend on them
- Does not contain observation logic, API calls, or data types
- Provides configuration values, not behavior

## Design expectations

- Configuration should be loadable from environment variables, configuration files, or both
- Secrets (API keys) should never be committed to the repository
- Default values should be sensible for getting started quickly
- Forum-specific mappings should be clearly separated from generic configuration
- Adding a new configuration option should not require changes to unrelated modules
