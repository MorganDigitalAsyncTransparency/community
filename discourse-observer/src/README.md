# Source Code

This directory contains the application source code for discourse-observer, organized into layers with distinct responsibilities.

## Layers

| Directory | Responsibility |
|-----------|---------------|
| `discourse/` | Raw Discourse API integration — fetching, authentication, pagination |
| `observer/` | Turning raw Discourse data into structured observations and detecting changes |
| `model/` | Internal normalized types and domain concepts, independent of the API |
| `config/` | Forum-specific configuration and adaptation points |
| `storage/` | Abstraction for persisting observed data |

## How layers interact

See [ARCHITECTURE.md](../ARCHITECTURE.md) for the data flow diagram, dependency direction, and detailed layer descriptions.

The key rule: dependencies flow inward toward `model/`, never outward. Each layer has its own README explaining its responsibility and boundaries.
