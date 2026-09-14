# EkaRouter Developer Platform: Comprehensive System Audit

**Audit Date:** 2026-09-15  
**Auditor:** Lead Backend Architect & Senior Go Platform Engineer  
**Project:** EkaRouter (`github.com/dresar/ekarouter`)

---

## 1. Executive Summary

EkaRouter currently operates as a high-performance, single-process Universal AI Gateway in pure Go (CGO-free via `modernc.org/sqlite`). It provides OpenAI-compatible interfaces (`/v1/chat/completions`, `/v1/models`, `/v1/responses`), intelligent model routing, multi-account pooling with cooldown management, token saver heuristics, outbound HTTP/SOCKS5 proxy chaining with SSRF protection, and SQLite persistence.

The objective of this architectural expansion is to transform EkaRouter into a comprehensive, enterprise-grade **Developer Platform & Universal API Management Hub**. This hub extends beyond AI models to encompass external developer APIs, cloud infrastructure, SaaS services, automation platforms, databases, communication gateways, monitoring systems, and custom endpoints—while preserving the existing AI engine and maintaining standalone single-binary deployment.

---

## 2. Current Architecture & Component Inventory

### 2.1 Core Runtime & Server
- **Entrypoint:** `cmd/ekarouter/main.go` parses flags, initializes database, runs migrations, sets up signal traps (`SIGINT`, `SIGTERM`), and starts Chi HTTP server.
- **Router:** Chi v5 (`internal/httpapi/router.go`) providing middleware (`Recoverer`, `RequestID`, `BodyLimit`, `CORS`, `GatewayAuth`, `SessionAuth`).
- **Config:** `internal/config/config.go` loads environment variables (`EKAROUTER_*`) with safe production validation.

### 2.2 Persistence & Migrations
- **Engine:** SQLite via `modernc.org/sqlite` (pure Go, zero CGO requirement).
- **WAL Mode:** Enabled with foreign keys and transactional migration runner (`internal/db/db.go`).
- **Current Migrations:**
  - `0001_initial.sql`: Core schema for `providers`, `accounts`, `credentials`, `models`, `routes`, `route_items`, `api_keys`, `sessions`, `proxy_profiles`, `usage_logs`, `request_logs`, `quota_snapshots`, `oauth_states`.
  - `0002_credential_pools.sql`: Enhanced credential pooling (`credential_pools`, `pool_members`, `rotation_policies`, `credential_health_checks`, `credential_events`), `free_tier_catalog`, `provider_templates`, `request_templates`, `api_operations`, `projects`, `project_credentials`, `credential_aliases`, `quota_limits`, `rotation_decisions`.

### 2.3 Security & Cryptography
- **Encryption:** `internal/auth/auth.go` uses AES-256-GCM authenticated encryption for secret storage.
- **Key Derivation:** SHA-256 hash of master secret key (`EKAROUTER_SECRET_KEY`).
- **Token Hashing:** SHA-256 hex encoding for session tokens and API keys (`eka_live_*`).
- **SSRF Defense:** `internal/proxy/proxy.go` filters loopback (`127.0.0.0/8`), RFC 1918 private IPv4 (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), link-local (`169.254.0.0/16`), IPv6 loopback (`::1`), and cloud metadata IP (`169.254.169.254`).

### 2.4 Existing AI Providers & Gateway
- **Registry:** `internal/providers/registry.go` registers AI adapters (OpenAI, Anthropic, Gemini, Cloudflare, Groq, Cerebras, OpenRouter, HuggingFace, Ollama, etc.).
- **Gateway Pipeline:** `internal/gateway/gateway.go` orchestrates request normalization, token saving, fallback routing across accounts, error classification, SSE streaming chunks, and usage recording.

---

## 3. Reusable Subsystems & Abstractions

1. **CryptoService (`internal/auth/auth.go`):** Reusable for encrypting third-party API keys, bearer tokens, OAuth client secrets, and webhook secrets. Needs versioned payload tagging (`v1:...`).
2. **SSRF Guard (`internal/proxy/proxy.go`):** Direct reuse for generic HTTP tool execution, OpenAPI URL fetching, and custom provider health checks.
3. **Database Migration Runner (`internal/db/db.go`):** Pure-Go SQL migration runner with schema version tracking in `schema_migrations`.
4. **Credential Pool Store & Engine (`internal/credpool`):** Solid foundation for priority, least-used, round-robin, and health-based rotation strategies.
5. **Usage & Latency Recorder (`internal/usage/usage.go`):** Bounded async worker queue for usage persistence without blocking HTTP response paths.

---

## 4. Gap Analysis: Missing Developer Platform Capabilities

| Capability Domain | Current State | Required State |
|---|---|---|
| **Multi-Category Provider Registry** | AI providers only (`internal/providers`) | Universal Provider Registry supporting 11 categories: Developer, Communication, Monitoring, Automation, Scraping, Storage, Payments, Analytics, Maps, Security, Custom. |
| **Provider Metadata & Capabilities** | Basic model/endpoint structs | Formal capability declaration: `api_key_auth`, `bearer_auth`, `oauth2`, `webhook`, `quota_api`, `health_check`, `request_proxy`, etc. |
| **Credential Vault** | Account-tied credentials only | Unified Credential Vault with versioning, masking (`sk-****abcd`), environment isolation (`dev`, `staging`, `prod`), project assignment, and permission scopes. |
| **RBAC & Multi-Tenancy** | Single admin password + session | Users, Teams, Roles (`admin`, `developer`, `viewer`), Granular Permissions (`credentials.read`, `tools.execute`, etc.), and Client Tokens. |
| **Generic Tool Execution Engine** | Non-existent (only AI completions) | Parameterized HTTP tool caller with path/header/query/body templating, credential injection, SSRF validation, and timeout control. |
| **Request Templates** | Table schema exists, handlers rudimentary | Executable templates with safe variable substitution `{{var}}`, header injection sanitization, and output transformation. |
| **OpenAPI 3.x Parser** | Placeholder parser in `internal/devtools` | Full JSON/YAML OpenAPI 3.x importer parsing paths, methods, parameters, and generating executable Tool Definitions. |
| **OAuth Lifecycle** | PKCE flow initiated for AI providers | Universal OAuth 2.0 connection lifecycle with refresh token rotation, state validation, and encrypted token storage. |
| **Webhook System** | None | Inbound webhook signature verification (HMAC-SHA256) and outbound event dispatch with retry backoff and delivery logging. |
| **Controlled API Proxy** | Forward proxy for AI requests only | Dynamic reverse proxy route `/api/v1/proxy/{provider}/*` injecting credentials, stripping sensitive headers, and tracking usage. |
| **Quota & Rate-Limit Engine** | Basic quota snapshots in DB | Active sliding-window/token-bucket rate limiter and quota tracker with standardized states (`available`, `limited`, `exhausted`, `suspended`). |
| **Audit Logging** | Generic request logs | Dedicated immutable audit trail recording actor, action, resource, IP, user-agent, and status without secret leakage. |
| **Background Scheduler** | Ad-hoc healthcheck ticker | Lightweight goroutine scheduler for periodic provider health checks, cooldown recovery, token refresh, and retention pruning. |

---

## 5. Security & Risk Assessment

1. **Secret Exposure in Logs:**
   - *Risk:* Raw API keys or authorization headers appearing in logs or error traces.
   - *Mitigation:* Mandatory secret masking (`MaskCredential`) before output; strict logging sanitizer redacting `Authorization`, `x-api-key`, and tokens.
2. **SSRF in Generic Tools & Custom Providers:**
   - *Risk:* Outbound HTTP requests targeting AWS metadata (`169.254.169.254`), local Kubernetes endpoints, or internal RFC1918 hosts.
   - *Mitigation:* Strict URL parsing, DNS resolution checking, IP range exclusion, and protocol allowlisting (`http`, `https` only).
3. **CRLF & Header Injection:**
   - *Risk:* Variable interpolation in request headers allowing HTTP response splitting.
   - *Mitigation:* Enforce header key and value sanitization stripping `\r` and `\n`.
4. **Database Migration Safety:**
   - *Risk:* Disrupting existing SQLite databases during migration.
   - *Mitigation:* Incremental migration `0003_developer_platform.sql` using `CREATE TABLE IF NOT EXISTS` and `CREATE INDEX IF NOT EXISTS`. Zero breaking changes to `0001` or `0002` tables.

---

## 6. Implementation Strategy & Milestones

- **Phase 1: Architecture, Database & Universal Vault**
  - Schema migration `0003_developer_platform.sql`.
  - Credential Vault package with authenticated AES-256-GCM, versioned payloads, and uniform masking.
  - User, Role, Team, and Project RBAC domain models and stores.
- **Phase 2: Universal Provider Integration Framework**
  - Universal Provider Registry with 11 categories and formal capability discovery.
  - Concrete adapter implementations for premier providers across categories (Cloudflare, GitHub, Resend, Sentry, Stripe, PostHog, VirusTotal, etc.).
- **Phase 3: Rate Limiting, Quota & Rotation Engine**
  - Multi-strategy rotation (priority, round-robin, least-used, health-based).
  - Sliding-window rate limiter & cooldown manager with persistent status.
- **Phase 4: Tool Execution Engine, Templates & SSRF Guard**
  - Generic tool executor with safe template interpolation.
  - Controlled API proxy `/api/v1/proxy/*`.
  - OpenAPI 3.x document import and operation generator.
- **Phase 5: OAuth Framework, Webhooks & Background Scheduler**
  - OAuth 2.0 token manager with auto-refresh.
  - Webhook delivery manager with HMAC-SHA256 signature verification.
  - Background scheduler for health checks, cooldown recovery, and log retention.
- **Phase 6: Unified Developer API (REST Endpoints)**
  - Full RESTful `/api/v1/*` endpoints with Chi router, validation, pagination, and audit logging.
- **Phase 7: Test Suites, Windows EXE Build & Documentation**
  - Comprehensive unit and integration tests (mock HTTP servers, zero external network dependency).
  - Standalone Windows binary build verification (`ekarouter.exe`).
  - Full documentation suite in `docs/`.
