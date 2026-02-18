# Changelog: 2026-02-18 - Round 3 Refactoring

## [workflow] refactor: decouple notification dependency
- Defined `Notifier` interface in domain repository to prevent leaky abstractions.
- Implemented `LocalNotifierAdapter` to bridge to the `Notification` module.
- Updated `ChannelHandler` to use the unified `Notifier` interface.

## [notification] refactor: centralize domain constants
- Consolidated hard-coded status strings (`sent`, `failed`) and channels into `domain/constants.go`.
- Standardized `DefaultLanguage` fallback.

## [workflow] chore: centralize executor constants
- Moved polling interval and batch size defaults to domain constants.
