# Usage Package

## Purpose
Provides asynchronous background recording of request metadata, token consumption logs, latency measurements, retention cleanup pruning, and aggregated analytical reporting.

## Files
- `usage.go`: `Recorder` worker loop, database inserts for `usage_logs` and `request_logs`, prune logic, and summary aggregations.
- `usage_test.go`: Unit tests for record queueing, persistence, summary calculations, and pruning.

## Allowed Responsibilities
- Non-blocking logging of completed gateway request metrics.
- Retention cleanup and summary statistics queries.

## Forbidden Responsibilities
- No raw prompt/response body persistence (privacy & memory safety).
