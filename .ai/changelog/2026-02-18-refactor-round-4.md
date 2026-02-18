# Changelog: 2026-02-18 - Round 4 Refactoring (ISP & Cleanup)

## [iam] refactor: apply ISP to repositories
- Split `UserRepository` into `UserReader` and `UserWriter`.
- Split `RoleRepository` into `RoleReader` and `RoleWriter`.
- Refactored `AuthService` to use narrower interfaces.

## [initialize] chore: consolidate health routes
- Removed redundant global `/health` route in favor of health module's `/dashboards/health`.
- Cleaned up unused controller imports in `router.go`.

## [health] refactor: centralize thresholds
- Moved `DefaultSLAThreshold` to `domain/constants.go`.
