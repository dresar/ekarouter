# Gateway Package

## Purpose
Core AI gateway execution pipeline orchestrating request normalization, token saver compaction, route resolution, multi-account retry and fallback loops, streaming event propagation, and non-blocking usage recording.

## Files
- `gateway.go`: `Gateway` struct, `Execute`, `ExecuteStream`, token saver integration, transient error fallback, and usage logging.
- `gateway_test.go`: Unit tests for fallback on 429, no-fallback on client 400 errors, and SSE stream chunks.

## Allowed Responsibilities
- Coordinating the complete gateway request/response lifecycle.
- Bounded retries and intelligent fallback across configured providers and accounts.
- Propagating context cancellation to upstream HTTP requests.

## Forbidden Responsibilities
- No direct database schema DDL or provider-specific JSON wire format handling.
