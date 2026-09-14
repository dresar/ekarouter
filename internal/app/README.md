# App Package

## Purpose
Dependency injection, subsystem lifecycle orchestration, database state hydration, and HTTP server lifecycle management with graceful shutdown.

## Files
- `app.go`: Application bootstrap, subsystem wiring (DB, crypto, usage recorder, tokensaver, proxy manager, adapters, router, gateway, server), and graceful signal/context shutdown.
- `app_test.go`: Integration tests verifying clean application startup, endpoint accessibility, and graceful shutdown.

## Allowed Responsibilities
- Composing internal subsystems and binding them to the HTTP server.
- Orchestrating graceful shutdown timeouts, connection draining, and resource cleanup.

## Forbidden Responsibilities
- No business logic or protocol translation details.
