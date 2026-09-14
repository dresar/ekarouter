# Httpapi Package

## Purpose
Exposes public OpenAI-compatible Gateway endpoints (`/v1/*`), management dashboard APIs (`/api/*`), and process health endpoints (`/health`, `/ready`).

## Files
- `router.go`: Chi router setup, global middleware registration, and route mounting.
- `middleware.go`: RequestID, MaxBodyLimit, CORS policy, Gateway API-Key validation (SHA-256), and Session token validation.
- `gateway_handlers.go`: Handlers for `/v1/models`, `/v1/chat/completions`, and `/v1/responses` with non-buffering SSE streaming.
- `admin_handlers.go`: Handlers for `/api/auth/*`, `/api/providers`, `/api/accounts`, `/api/routes`, `/api/keys`, `/api/usage`, and `/api/tokensaver/preview`.
- `httpapi_test.go`: End-to-end integration tests for routing, auth barriers, body limits, and responses.

## Allowed Responsibilities
- Parsing HTTP requests and writing standardized HTTP/SSE responses.
- Enforcing rate/body limits, CORS headers, and authentication tokens.

## Forbidden Responsibilities
- No direct provider wire transformations (delegated to `providers` package).
- No hardcoded secrets or plaintext credential exposure.
