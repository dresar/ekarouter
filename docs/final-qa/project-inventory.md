# EkaRouter Project Inventory & Architecture Manifest

**Status:** Feature-Complete QA Audit  
**Date:** 2026-09-15  
**Version:** 1.0.0  
**Repository:** `github.com/dresar/ekarouter`

---

## 1. Application Modules (`internal/`)

| Package | Purpose | Primary Structs / Interfaces | Dependencies |
|---|---|---|---|
| `internal/app` | Application bootstrap, lifecycle orchestration, and dependency injection | `Application`, `Setup`, `Run` | All internal packages |
| `internal/audit` | Structured asynchronous audit logging with bounded in-memory buffer | `Logger`, `Event`, `Log()` | `database/sql` |
| `internal/auth` | Session token generation, SHA-256 token hashing, AES-256-GCM encryption | `CryptoService`, `GenerateApiKey`, `HashToken` | `crypto/aes`, `crypto/cipher` |
| `internal/cli` | Interactive terminal UI, TUI menu, subcommands (`serve`, `cli`, `import`, `backup`) | `HandleSubcommands`, `RunInteractiveMenu` | `internal/db`, `internal/auth` |
| `internal/config` | Environment variable parsing, validation, default settings | `Config`, `Load`, `DefaultConfig` | `os`, `strconv`, `time` |
| `internal/credpool` | Intelligent credential pooling, tiering, failure tracking, and health checks | `Store`, `Engine`, `HealthChecker`, `Pool` | `internal/providers`, `database/sql` |
| `internal/db` | SQLite connection manager with WAL mode, foreign keys, and migration runner | `DB`, `Open`, `Migrate`, `Backup`, `Import9Router` | `modernc.org/sqlite` |
| `internal/devtools` | OpenAPI 3.0/3.1 document parser, operation importer, and cURL generator | `ParseOpenAPIDocument`, `GenerateCurl` | `net/http` |
| `internal/executor` | Safe HTTP executor for request templates and tool actions with SSRF filters | `Executor`, `ExecuteTemplate`, `ExecuteTool` | `internal/platform`, `internal/vault` |
| `internal/freetier` | Catalog of AI & developer platform free tiers with verification status | `CatalogStore`, `SeedCatalog`, `FreeTierItem` | `database/sql` |
| `internal/gateway` | OpenAI-compatible AI gateway orchestrating model routing, retries, and token saving | `Gateway`, `Execute`, `ExecuteStream` | `internal/routing`, `internal/providers` |
| `internal/health` | HTTP health and readiness probes for infrastructure monitoring | `Checker`, `HealthHandler`, `ReadyHandler` | `database/sql` |
| `internal/httpapi` | Chi HTTP router, middleware, platform handlers, admin handlers, and proxy endpoints | `Server`, `NewServer`, `PlatformHandler`, `AdminHandler` | `github.com/go-chi/chi/v5` |
| `internal/limits` | Rate-limiting and quota tracking engine with SQLite persistence | `Engine`, `CheckLimit`, `RecordUsage` | `database/sql` |
| `internal/oauth` | PKCE OAuth2 state manager for external provider authorization flows | `Manager`, `GenerateAuthURL`, `Exchange` | `crypto/sha256` |
| `internal/platform` | Developer platform registry, adapter interfaces, SSRF-safe HTTP client | `Registry`, `BaseAdapter`, `NewSafeHTTPClient` | `net/http` |
| `internal/providers` | AI provider registry and adapters (OpenAI, Anthropic, Gemini, Groq, Ollama, etc.) | `Registry`, `Adapter`, `Request`, `Response` | `net/http` |
| `internal/proxy` | Outbound proxy client manager (HTTP, HTTPS, SOCKS5) with DNS cache and circuit breaker | `Manager`, `GetClient`, `Profile` | `net/http` |
| `internal/rbac` | Role-Based Access Control enforcing role permissions | `Service`, `HasPermission`, `AssignRole` | `database/sql` |
| `internal/rotator` | Intelligent credential rotation engine (Priority, Round Robin, Least Used, Health) | `Rotator`, `Select`, `SelectForPool` | `internal/vault` |
| `internal/routing` | AI model routing engine, aliases, fallback chains, and CooldownManager | `Router`, `CooldownManager`, `SelectTargets` | `database/sql`, `internal/providers` |
| `internal/scheduler` | Periodic background task runner (log cleanup, health checks, metrics rollup) | `Scheduler`, `Start`, `Stop` | `database/sql`, `time` |
| `internal/tokensaver` | Prompt compaction and whitespace optimization engine (off, safe, balanced) | `TokenSaver`, `Compact`, `Mode` | `regexp`, `strings` |
| `internal/usage` | Async token usage recorder and analytics aggregator | `Recorder`, `Record`, `GetSummary` | `database/sql` |
| `internal/vault` | Encrypted credential vault with master key rotation and versioning | `Vault`, `Store`, `Credential` | `crypto/aes` |
| `internal/webhooks` | Outbound webhook dispatcher with signature verification and retry queue | `Dispatcher`, `Send`, `VerifySignature` | `crypto/hmac` |

---

## 2. Registered AI Providers (23 Adapters)

All AI providers implement `providers.Adapter` in `internal/providers/`:

1. **`openai`** (`github.com/dresar/ekarouter/internal/providers/openai`): OpenAI official API (GPT-4o, o1, GPT-4 Turbo).
2. **`anthropic`** (`github.com/dresar/ekarouter/internal/providers/anthropic`): Claude 3.5 Sonnet, Claude 3 Opus, Claude 3.5 Haiku.
3. **`gemini`** (`github.com/dresar/ekarouter/internal/providers/gemini`): Google Gemini 1.5 Pro, Flash, 2.0 Flash (includes CLI & Antigravity variants).
4. **`groq`** (`github.com/dresar/ekarouter/internal/providers/groq`): Llama 3.3 70B, Mixtral 8x7B (ultra-low latency LPU).
5. **`cerebras`** (`github.com/dresar/ekarouter/internal/providers/cerebras`): Wafer-scale engine Llama 3.1 70B/8B.
6. **`openrouter`** (`github.com/dresar/ekarouter/internal/providers/openrouter`): Aggregated routing across 200+ models.
7. **`cloudflare`** (`github.com/dresar/ekarouter/internal/providers/cloudflare`): Cloudflare Workers AI serverless models.
8. **`nvidia`** (`github.com/dresar/ekarouter/internal/providers/nvidia`): NVIDIA NIM inference microservices.
9. **`ollama`** (`github.com/dresar/ekarouter/internal/providers/ollama`): Self-hosted local model inference.
10. **`huggingface`** (`github.com/dresar/ekarouter/internal/providers/huggingface`): Hugging Face Inference API / Endpoints.
11. **`airforce`** (`github.com/dresar/ekarouter/internal/providers/airforce`): Free-tier AI proxy provider.
12. **`bazaarlink`** (`github.com/dresar/ekarouter/internal/providers/bazaarlink`): Community AI routing gateway.
13. **`chutes`** (`github.com/dresar/ekarouter/internal/providers/chutes`): Decentralized GPU compute cluster.
14. **`coqui`** (`github.com/dresar/ekarouter/internal/providers/coqui`): TTS audio synthesis provider.
15. **`custom`** (`github.com/dresar/ekarouter/internal/providers/custom`): Custom backend endpoints (DeepSeek, Mistral, Together, vLLM, LocalAI).
16. **`devin`** (`github.com/dresar/ekarouter/internal/providers/devin`): Devin CLI & autonomous agent provider.
17. **`edgetts`** (`github.com/dresar/ekarouter/internal/providers/edgetts`): Microsoft Edge TTS engine.
18. **`kilogateway`** (`github.com/dresar/ekarouter/internal/providers/kilogateway`): High-throughput proxy gateway.
19. **`kimchi`** (`github.com/dresar/ekarouter/internal/providers/kimchi`): Lightweight community LLM gateway.
20. **`kiro`** (`github.com/dresar/ekarouter/internal/providers/kiro`): Specialized Asian LLM gateway.
21. **`mimofree`** (`github.com/dresar/ekarouter/internal/providers/mimofree`): Free tier aggregator.
22. **`opencode`** (`github.com/dresar/ekarouter/internal/providers/opencode`): Open-source code completion models.
23. **`searxng`** (`github.com/dresar/ekarouter/internal/providers/searxng`): Privacy-respecting meta-search integration.

---

## 3. Registered Developer Platform Providers (27 Adapters)

Configured in `internal/platform/catalog.go`:

| Provider ID | Provider Name | Category | Auth Type | Primary Use Case |
|---|---|---|---|---|
| `cloudflare` | Cloudflare | Developer | Bearer Token | DNS, Workers, Pages, Edge Rules |
| `github` | GitHub | Developer | Personal Token | Repos, Issues, Workflows, Releases |
| `vercel` | Vercel | Developer | Bearer Token | Deployments, Projects |
| `supabase` | Supabase | Developer | Bearer Token | PostgreSQL, Auth, Edge Functions |
| `neon` | Neon Serverless Postgres | Developer | Bearer Token | Serverless DB branching & connections |
| `resend` | Resend | Communication | Bearer Token | Transactional email delivery |
| `sendgrid` | SendGrid | Communication | Bearer Token | Marketing & transactional email |
| `twilio` | Twilio | Communication | Basic Auth | SMS, Voice calls, WhatsApp |
| `discord` | Discord | Communication | Bot Token | Webhooks, Channels, Messages |
| `slack` | Slack | Communication | Bot Token | Channels, Workspace alerts |
| `sentry` | Sentry | Monitoring | Bearer Token | Error tracking, Performance metrics |
| `betterstack` | Better Stack | Monitoring | Bearer Token | Uptime monitoring, Incidents |
| `webhook` | Generic Webhook | Automation | Custom Header | Outbound event dispatching |
| `firecrawl` | Firecrawl | Scraping & Data | Bearer Token | Web crawling & markdown conversion |
| `tavily` | Tavily Search API | Scraping & Data | API Key Header | LLM-optimized real-time search |
| `serpapi` | SerpAPI | Scraping & Data | Custom Header | Google search scraping |
| `jina` | Jina AI Reader | Scraping & Data | Bearer Token | Webpage reader & grounding |
| `cloudflare-r2` | Cloudflare R2 | Storage | Bearer Token | S3-compatible zero-egress object storage |
| `stripe` | Stripe | Payments | Bearer Token | Invoicing, Checkout, Subscriptions |
| `midtrans` | Midtrans | Payments | Basic Auth | Southeast Asian QRIS, VA, E-Wallets |
| `posthog` | PostHog | Analytics | Personal Token | Event capture, Feature flags |
| `umami` | Umami | Analytics | Bearer Token | Privacy-preserving web analytics |
| `mapbox` | Mapbox | Maps & Geo | Custom Header | Geocoding, Directions, Static maps |
| `virustotal` | VirusTotal | Security | API Key Header | URL and file reputation analysis |
| `abuseipdb` | AbuseIPDB | Security | API Key Header | IP abuse reports & blacklist check |
| `ipinfo` | IPinfo | Security | Bearer Token | IP geolocation and ASN data |
| `generic-rest` | Generic REST API | Custom | Custom Header | Arbitrary HTTP endpoints with SSRF check |

---

## 4. API Route Groups

1. **System & Probes (`/` root)**:
   - `GET /health` — Overall health check
   - `GET /ready` — Database readiness probe
   - `GET /live` — Liveness probe

2. **OpenAI-Compatible AI Gateway (`/v1`)**:
   - Protected by `GatewayAuthMiddleware` (Bearer API Key with `api_keys` hash)
   - `GET /v1/models` — Model catalog
   - `POST /v1/chat/completions` — Chat completions (streaming SSE & non-streaming)
   - `POST /v1/responses` — Chat completions alias

3. **Developer Platform Management (`/api/v1`)**:
   - Protected by `PlatformAuthMiddleware` (Session Cookie, Developer Client Token, or API Key)
   - Providers: `GET /api/v1/providers`, `GET /api/v1/providers/{id}`, capabilities, docs, validation, health
   - Credentials: `GET /api/v1/credentials`, `POST /api/v1/credentials`, get, update, delete, enable, disable, rotate, health, usage, events
   - Projects: `GET /api/v1/projects`, `POST /api/v1/projects`, get, patch, delete
   - Environments: `GET /api/v1/environments`, `POST /api/v1/environments`, patch, delete
   - Tools: `GET /api/v1/tools`, `POST /api/v1/tools`, schema, execute, test
   - Request Templates: `GET /api/v1/request-templates`, `POST`, get, patch, delete, execute
   - Usage & Analytics: `GET /api/v1/usage`, `GET /api/v1/usage/summary`, providers, credentials, projects
   - Health: `GET /api/v1/health`, `GET /api/v1/health/providers`, `GET /api/v1/health/credentials`
   - Audit Logs: `GET /api/v1/audit-logs`, `GET /api/v1/events`
   - Universal Proxy: `ANY /api/v1/proxy/{provider}/*`

4. **Admin Console & Entity Management (`/api`)**:
   - Authentication: `POST /api/auth/login` (Public)
   - Protected by `SessionAuthMiddleware`:
     - `POST /api/auth/logout`, `GET /api/auth/me`
     - Providers: `GET /api/providers`, `POST`, `DELETE /{id}`
     - Accounts & OAuth: `GET /api/accounts`, `POST`, `DELETE /{id}`, `POST /api/accounts/oauth/start`, `callback`
     - Credentials: `GET /api/credentials`, `GET /api/credentials/{id}`
     - Routes: `GET /api/routes`, `POST`, `DELETE /{id}`
     - Models: `GET /api/models`, `POST`, `DELETE /{id}`
     - Proxies: `GET /api/proxies`, `POST`, `DELETE /{id}`, `POST /{id}/test` (and `/api/proxy-profiles` aliases)
     - API Keys: `GET /api/keys`, `POST`, `DELETE /{id}`
     - System Settings: `GET /api/settings`, `PUT /api/settings`
     - Usage & Backup: `GET /api/usage`, `POST /api/tokensaver/preview`, `GET /api/backup`, `POST /api/backup`
     - Credential Pools: `GET /api/credential-pools`, `POST`, `GET /{id}`, `DELETE /{id}`, members, policy, health, rotate, pause, resume, usage
     - Free Tiers: `GET /api/free-tiers`, `categories`, `verified`, `refresh`, `{id}`, `{id}/sources`
     - Devtools: `GET /api/devtools/templates`, `POST`, `import-openapi`, `confirm-openapi`, `generate-curl`

---

## 5. Configuration Variables

| Variable | Type | Default | Description |
|---|---|---|---|
| `EKAROUTER_ENV` | string | `production` | Deployment mode (`development`, `test`, `production`) |
| `EKAROUTER_HOST` | string | `0.0.0.0` | Bind IP address |
| `EKAROUTER_PORT` | int | `8080` | HTTP port |
| `EKAROUTER_DB_PATH` | string | `data/ekarouter.db` | Path to SQLite database file |
| `EKAROUTER_SECRET_KEY` | string | (auto-generated) | 32-char AES-256-GCM master encryption key |
| `EKAROUTER_ADMIN_USER` | string | `admin` | Default administrator username |
| `EKAROUTER_ADMIN_PASSWORD` | string | `admin12345` | Default administrator password |
| `EKAROUTER_ALLOW_LOCAL_PROVIDERS` | bool | `false` | When true, permits connections to RFC 1918 / localhost |
| `EKAROUTER_TOKEN_SAVER_MODE` | string | `safe` | Prompt compaction mode (`off`, `safe`, `balanced`) |
| `EKAROUTER_MAX_BODY_BYTES` | int64 | `10485760` (10MB) | Max incoming HTTP request body size |
| `EKAROUTER_CORS_ORIGINS` | string | `*` | Comma-separated allowed CORS origins |
| `EKAROUTER_LOG_RETENTION_DAYS` | int | `30` | Auto-purge retention period for request & audit logs |
| `EKAROUTER_ENABLE_LIVE_PROVIDER_TESTS` | bool | `false` | Explicit opt-in flag for testing real external provider APIs |

---

## 6. Background Workers & Schedulers

1. **`scheduler.Scheduler`**: Runs periodic cron tasks every 60 seconds.
   - Automatically purges expired request and audit logs older than `EKAROUTER_LOG_RETENTION_DAYS`.
   - Runs background health probes on registered developer platform credentials.
2. **`usage.Recorder`**: Asynchronous ring-buffered token usage writer preventing request-path latency spikes.
3. **`audit.Logger`**: Asynchronous bounded-channel writer persisting structured security and modification logs.
4. **`routing.CooldownManager`**: In-memory exponential backoff circuit breaker for failing AI provider accounts.

---

## 7. Database Tables (31 Tables across 4 Migrations)

- **Migration 0001 (`0001_initial.sql`)**:
  `schema_migrations`, `settings`, `providers`, `accounts`, `credentials`, `models`, `routes`, `route_items`, `api_keys`, `sessions`, `proxy_profiles`, `usage_logs`, `request_logs`, `quota_snapshots`, `oauth_states`
- **Migration 0002 (`0002_credential_pools.sql`)**:
  `free_tier_catalog`, `free_tier_sources`, `credential_pools`, `pool_members`, `rotation_policies`, `pool_health_events`, `credential_usage_snapshots`, `credential_audit_logs`
- **Migration 0003 (`0003_developer_platform.sql`)**:
  `users`, `teams`, `team_members`, `environments`, `roles`, `permissions`, `role_permissions`, `projects`, `vault_credentials`, `credential_versions`, `client_tokens`, `tools`, `tool_executions`, `audit_events`, `webhooks`, `webhook_deliveries`, `rate_limits`, `request_quotas`
- **Migration 0004 (`0004_platform_extensions.sql`)**:
  `credential_assignments`, `credential_permissions`, `credential_health`, `credential_usage`, `api_requests`, `api_request_attempts`, `api_request_logs`, `usage_records`, `routing_rules`, `tool_actions`, `openapi_documents`, `notifications`, `system_settings`

---

## 8. Known Limitations

1. **Race Detection on Windows**: Go race detector (`-race`) requires CGO and MinGW/GCC, which is not present by default on pure Windows environments. SQLite runs via `modernc.org/sqlite` in pure Go. Local high-concurrency stress tests are used to verify thread-safety.
2. **User/Team Self-Service Registration**: Currently managed via administrative platform APIs and CLI seeding; public signup routes (`POST /auth/register`) are not exposed by design.
3. **Streaming Formats**: AI gateway supports SSE (`text/event-stream`) streaming for Chat Completions. Image generation, audio transcription, and embeddings proxy endpoints are not yet registered on `/v1`.
