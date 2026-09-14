# EkaRouter Architecture Specification

## 1. Overview

EkaRouter is a lightweight, single-process, CGO-free universal AI Gateway written in Go. It provides OpenAI-compatible interfaces (`/v1/models`, `/v1/chat/completions`, `/v1/responses`), intelligent routing, multi-account pooling, automated failover, token saver heuristic compactions, outbound proxy support with SSRF protection, and SQLite persistence.

## 2. Package Boundaries & Dependency Graph

```text
cmd/ekarouter
    └── internal/app
            ├── internal/config
            ├── internal/db
            ├── internal/auth
            ├── internal/proxy
            ├── internal/tokensaver
            ├── internal/providers
            ├── internal/routing
            ├── internal/gateway
            ├── internal/usage
            ├── internal/health
            └── internal/httpapi
```

- `internal/config`: Loads and validates application configuration from CLI flags and environment variables.
- `internal/db`: Pure-Go SQLite driver (`modernc.org/sqlite`) connection pool, WAL configuration, and schema migrations.
- `internal/auth`: Cryptographic hashing (SHA-256 for API keys and sessions) and AES-256-GCM authenticated encryption for provider secrets.
- `internal/proxy`: Reusable `http.Transport` instances with connection reuse, timeouts, proxy profile support, and strict SSRF destination validation.
- `internal/tokensaver`: Safe CPU-bounded text compactions (git diff, grep, find, logs, etc.) running fail-open before upstream dispatch.
- `internal/providers`: Provider abstraction and adapters for OpenAI-compatible, Anthropic-style, Gemini-style, and Custom HTTP endpoints.
- `internal/routing`: Route definitions, combos, model alias mapping, account pool rotation (priority and round-robin), cooldown management, and error classification.
- `internal/gateway`: Core orchestration pipeline: request normalization, token saver execution, route and account selection, bounded retries with exponential backoff, SSE streaming chunks, client context cancellation propagation, and asynchronous usage logging.
- `internal/usage`: Usage and request metadata persistence with bounded retention and background pruning.
- `internal/health`: Liveness (`/health`) and readiness (`/ready`) endpoints.
- `internal/httpapi`: Chi router, HTTP middleware (request ID, CORS, max body limit, authentication), and API endpoints.

## 3. Strict Architectural Rules

1. **No Source Comments**: In accordance with the `/nokomen` standard, Go source files must contain zero comments. Explanations reside entirely in Markdown files.
2. **File Size Limit**: No source file may exceed 1000 lines. The preferred range is 100-500 lines per file.
3. **No Dumping Grounds**: Avoid generic `utils` or `helpers` packages.
4. **Folder Documentation**: Every directory containing Go code must have a concise `README.md`.
5. **Fail-Open Token Saver**: Token saving operations must never fail the user request. If a transform fails or grows in size, the original prompt is preserved.
6. **Streaming Integrity**: Never buffer entire streaming responses. SSE events must be decoded and piped directly to the client as neutral chunks.
