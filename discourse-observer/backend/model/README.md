# model/

This module contains internal normalized types and domain concepts.

## Responsibility

The model defines the project's own understanding of forum activity, independent of the Discourse API's data shapes. Types defined here are used throughout the project by the observer, storage, and any future layers.

## Boundaries

This module:

- Defines types, interfaces, and domain constants only
- Has **no dependencies** on other modules in the project
- Does not import from `discourse/`, `observer/`, `config/`, or `storage/`

It is a leaf dependency — everything else depends on it, but it depends on nothing.

## Design expectations

- Types should represent normalized domain concepts, not raw API shapes
- Fields should be named for what they mean in the project's domain, not what the API calls them
- Types should be small and composable rather than large and monolithic
- Changes to the model affect downstream modules, so changes should be intentional and documented
