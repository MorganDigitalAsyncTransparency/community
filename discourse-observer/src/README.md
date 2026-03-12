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

Data flows through the layers in one direction:

```
discourse → observer → model → storage
                         ↑
                       config (read by all layers as needed)
```

- `discourse/` fetches raw data and returns it in API-adjacent shapes
- `observer/` receives raw data, normalizes it using `model/` types, detects changes, and passes observations to `storage/`
- `model/` defines types only — it has no dependencies on other layers
- `config/` provides configuration — it is read by other layers but does not depend on them
- `storage/` receives observations and persists them — it depends only on `model/` types

## Guiding principles

- Each layer has a README explaining its responsibility
- Dependencies flow inward (toward `model/`), not outward
- Discourse API details do not leak beyond `discourse/`
- Forum-specific assumptions are isolated in `config/`
- New layers (API server, event extraction, dashboard) will be added alongside these as the project grows
