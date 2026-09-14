# Reference Audit Report: 9Router Behavioral Analysis

## 1. Executive Summary

A deep audit was conducted on the cloned `9router` reference codebase (`c:\Users\NCN0C\Music\ekarouter\9router`). 9Router is implemented as a Node.js / Next.js / Express hybrid application with an embedded SQLite database (`src/lib/db`) using `better-sqlite3`, a custom translator layer (`open-sse`), an RTK token saver subsystem (`open-sse/rtk`), a multi-provider registry (`open-sse/providers/registry`), and a dashboard.

EkaRouter takes the observable behavioral specifications of 9Router and builds a lightweight, single-process, CGO-free Go backend without Node.js runtime dependencies.

## 2. Architecture & Modules Audited

### 2.1 Database & Persistence
- Reference engine: SQLite with WAL mode (`PRAGMA journal_mode = WAL; synchronous = NORMAL`).
- Tables: `_meta`, `settings`, `providerConnections`, `providerNodes`, `proxyPools`, `apiKeys`, `combos`, `kv`, `usageHistory`, `usageDaily`, `requestDetails`.
- Key weakness in reference: API keys and provider tokens stored in plaintext JSON within table columns (`data` column). EkaRouter replaces this with encrypted credentials (`AES-256-GCM`) and hashed API keys (`SHA-256`).

### 2.2 Provider Adapters & Registry
- Over 120 provider registrations observed in `open-sse/providers/registry/`.
- Four primary provider interaction paradigms:
  1. OpenAI-compatible (`/v1/chat/completions`, Bearer token, standard delta SSE chunks).
  2. Anthropic Messages (`/v1/messages`, `x-api-key`, system prompt separation, event-based SSE).
  3. Gemini Generative Language (`:generateContent`, `x-goog-api-key`, content parts structure).
  4. Custom / Self-hosted (configurable base URL and authentication headers).

### 2.3 Token Saver (RTK & Headroom)
- 9Router integrates an optional Headroom sidecar proxy (`/v1/compress`) and local RTK-style compaction filters:
  - `gitDiff`: Compacts diff hunks, counts added/removed lines, truncates hunks > 100 lines.
  - `gitLog`: Compresses commit lists to one-line summaries.
  - `grep` & `searchList`: Strips noisy context lines, retains file paths and line matches.
  - `dedupLog`: Collapses repetitive consecutive log lines into `[x repeats]`.
  - `buildOutput`: Filters successful compiler/build spam, preserves error outputs.
- Critical safety rule: Token saver must be fail-open. If compaction produces larger output or encounters parse issues, the original prompt must be preserved.

### 2.4 Routing, Combos & Account Pools
- Combos: Ordered list of provider/account/model targets.
- Strategies: Priority (first healthy) and Round-Robin.
- Failover conditions: HTTP 429 (rate limit), Quota exhaustion, Network timeout, 5xx server errors.
- Never failover on client 4xx (e.g. 400 Bad Request, 401 Unauthorized client key).

### 2.5 Outbound Proxy & Network Transport
- Reusable `http.Transport` pooling.
- Support for direct, HTTP proxy, HTTPS proxy, and SOCKS5.
- SSRF prevention: Restrict arbitrary loopback/internal addresses unless explicitly configured.
