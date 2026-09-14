# Routing Package

## Purpose
Manages model alias resolution, combo definitions, route item priority and round-robin ordering, and in-memory account cooldown / circuit breaking.

## Files
- `types.go`: Route, RouteItem, Account, and Target definitions.
- `cooldown.go`: Thread-safe `CooldownManager` providing bounded exponential backoff cooldowns for degraded accounts.
- `router.go`: `Router` handling model alias resolution, account state filtering, and target list construction.
- `routing_test.go`: Unit tests for alias mapping, priority ordering, round-robin rotation, and cooldown exclusions.

## Allowed Responsibilities
- Deciding target provider, account, model candidates based on strategy and health.
- Cooldown and failure bookkeeping.

## Forbidden Responsibilities
- No direct network calls, HTTP handlers, or provider format transformations.
