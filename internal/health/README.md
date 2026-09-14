# Health Package

## Purpose
Provides liveness (`/health`) and readiness (`/ready`) HTTP endpoints for process monitoring and orchestration.

## Files
- `health.go`: `Checker` with uptime tracking and database ping readiness checks.
- `health_test.go`: Unit tests for `/health` and `/ready` endpoints.

## Allowed Responsibilities
- Returning HTTP 200 for process liveness and database connectivity checks.

## Forbidden Responsibilities
- No business routing, provider communication, or heavy computational tasks.
