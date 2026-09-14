# EkaRouter

EkaRouter is a high-performance, lightweight, single-process Universal AI Gateway written in pure Go. It delivers OpenAI-compatible endpoints (`/v1/models`, `/v1/chat/completions`, `/v1/responses`), intelligent combo routing, multi-account rotation, automated fallback, safe token saving compactions, SSRF-protected outbound proxying, and pure-Go SQLite persistence without external dependencies or CGO requirements.

## Key Features

- **Universal OpenAI Compatibility**: Drop-in replacement for OpenAI `/v1/chat/completions` (with real-time SSE streaming) and `/v1/models`.
- **Multi-Provider Adapters**: Native translations for OpenAI-compatible, Anthropic Claude (`/v1/messages`), Google Gemini (`generateContent`), and custom backends.
- **Intelligent Routing & Failover**: Priority and round-robin strategies with automatic cooldown circuit breaking on transient errors (429 rate limits, 5xx server errors, timeouts).
- **Safe Token Saver**: RTK-inspired heuristic compactions (Git diffs, deduplicated logs, build output, truncation) preserving file paths, line numbers, and error traces fail-open.
- **Enterprise Security**: Secrets encrypted at rest with AES-256-GCM; API keys hashed with SHA-256; strict SSRF IP validation on outbound requests.
- **Pure Go & CGO-Free**: Uses `modernc.org/sqlite` for 100% CGO-free builds on Windows (`.exe`) and Linux (`amd64`, `arm64`) running under a 500 MB VPS memory footprint.
- **Zero Source Comments**: Compliant with the strict `/nokomen` standard. All documentation lives in Markdown files.

## Project Structure

```text
ekarouter/
├── cmd/
│   └── ekarouter/       # CLI entrypoint with flag parsing and lifecycle signal traps
├── internal/
│   ├── app/             # Application dependency injection and graceful shutdown
│   ├── auth/            # AES-256-GCM crypto service and SHA-256 token hashing
│   ├── config/          # Environment and flag configuration loader
│   ├── db/              # SQLite connection pool, WAL configuration, and migrations
│   ├── gateway/         # Core AI gateway pipeline, retry/fallback, and SSE streaming
│   ├── health/          # Process liveness (/health) and readiness (/ready) checks
│   ├── httpapi/         # Chi HTTP router, CORS/auth middleware, and API handlers
│   ├── oauth/           # Ephemeral PKCE OAuth state and anti-stampede refresh locks
│   ├── providers/       # Abstract provider interface and OpenAI/Anthropic/Gemini adapters
│   ├── proxy/           # Reusable outbound HTTP transports and SSRF protection
│   ├── routing/         # Combos, model alias resolution, and account cooldown manager
│   ├── tokensaver/      # Heuristic text compaction algorithms (safe, balanced, off)
│   └── usage/           # Asynchronous non-blocking usage and request metadata logger
├── migrations/          # Versioned SQL database migrations
├── scripts/             # Build scripts, smoke tests, and benchmarks
└── docs/                # Architecture specifications, agent reports, and reference audits
```

## Quick Start

### Build Native Binaries

```powershell
powershell -ExecutionPolicy Bypass -File ./scripts/build.ps1
```

Or on Linux:

```bash
./scripts/build.sh
```

### Run the Server

```bash
./bin/ekarouter.exe -port 8080 -db data/ekarouter.db
```

### Smoke Test

```powershell
powershell -ExecutionPolicy Bypass -File ./scripts/smoke_test.ps1
```

## Configuration

| Environment Variable | Default | Description |
| :--- | :--- | :--- |
| `EKAROUTER_PORT` | `8080` | HTTP listen port |
| `EKAROUTER_HOST` | `0.0.0.0` | Bind address |
| `EKAROUTER_DB_PATH` | `data/ekarouter.db` | SQLite database file location |
| `EKAROUTER_SECRET_KEY` | *(auto-generated)* | 256-bit encryption key for stored secrets |
| `EKAROUTER_ADMIN_USER` | `admin` | Admin dashboard username |
| `EKAROUTER_ADMIN_PASSWORD` | `admin12345` | Admin dashboard password |
| `EKAROUTER_TOKEN_SAVER_MODE`| `safe` | Token saver mode (`off`, `safe`, `balanced`) |
| `EKAROUTER_ALLOW_LOCAL_PROVIDERS` | `false` | Set to `true` to allow local IPs (Ollama) bypassing SSRF check |
