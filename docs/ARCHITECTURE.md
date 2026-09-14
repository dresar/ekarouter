# EkaRouter Developer Platform Architecture

## 1. System Overview

EkaRouter is a lightweight, single-process, CGO-free universal API Gateway and Developer Management Platform written in Go. In addition to universal AI model routing, EkaRouter provides unified management, encrypted key vault storage, rate limiting, quota tracking, multi-strategy key rotation, generic external tool execution, inbound/outbound webhooks, and controlled API proxying across 11 major API categories.

```text
                               +-----------------------------------------------------+
                               |                   Client Applications               |
                               | (Curl / Postman / SDK / Automation / Web Dashboard) |
                               +-----------------------------------------------------+
                                                          |
                                                          | HTTP / HTTPS / SSE
                                                          v
                               +-----------------------------------------------------+
                               |              Chi Router & Middleware                |
                               |   - Recoverer          - Request ID                 |
                               |   - Body Limiter       - CORS Engine                |
                               |   - Session Auth       - Platform RBAC & Tokens     |
                               +-----------------------------------------------------+
                                      |                                    |
            +-------------------------+                                    +--------------------------+
            | OpenAI AI Gateway Path                                                                  | Unified Developer Platform Path
            v                                                                                         v
+-------------------------------+                                                      +-------------------------------+
|     /v1/chat/completions      |                                                      |          /api/v1/*            |
|       /v1/models              |                                                      |  - /providers     - /creds    |
|       /v1/responses           |                                                      |  - /tools         - /projects |
+-------------------------------+                                                      |  - /health        - /usage    |
            |                                                                          |  - /proxy/*       - /audit    |
            v                                                                          +-------------------------------+
+-------------------------------+                                                                     |
|    AI Router & Token Saver    |                                                                     v
|  - Heuristic compaction       |                                                      +-------------------------------+
|  - Multi-account failover     |                                                      |       Universal Platform      |
|  - Account pool rotation      |                                                      |  - Credential Vault (AES-GCM) |
|  - Cooldown manager           |                                                      |  - Rotator (Multi-strategy)   |
+-------------------------------+                                                      |  - Limits & Quota Engine      |
            |                                                                          |  - Tool Execution & Templates |
            |                                                                          |  - SSRF Protection Guard      |
            v                                                                          |  - In/Outbound Webhooks       |
+-------------------------------+                                                      +-------------------------------+
|      AI Provider Adapters     |                                                                     |
| OpenAI, Anthropic, Gemini,    |                                                                     v
| Cloudflare, Groq, Ollama, ... |                                                      +-------------------------------+
+-------------------------------+                                                      |    Category Provider Adapters |
            |                                                                          | Cloudflare, GitHub, Resend,   |
            |                                                                          | Sentry, Stripe, PostHog, ...  |
            |                                                                          +-------------------------------+
            |                                                                                         |
            +------------------------------------+----------------------------------------------------+
                                                 |
                                                 v
                               +-----------------------------------+
                               |     SQLite Database Engine        |
                               |    (Pure Go, WAL mode enabled)    |
                               |  - vault_credentials              |
                               |  - tool_definitions               |
                               |  - scheduled_tasks & task_runs    |
                               |  - quota_records & rate_limits    |
                               |  - audit_logs & webhooks          |
                               +-----------------------------------+
```

## 2. Key Architecture Pillars

### 2.1 CGO-Free & Standalone
- Uses `modernc.org/sqlite` allowing compilation to a single executable without requiring GCC, MSVC, or external runtime libraries.
- Compiles as a native Windows EXE (`ekarouter.exe`) or Linux binary.

### 2.2 Security by Construction
- **Zero Raw Secrets in Storage or Logs:** Authenticated AES-256-GCM symmetric encryption with unique 12-byte initialization vectors (nonces) and versioned ciphertext (`v1:<nonce>:<tag+ciphertext>`).
- **Secret Masking:** Plaintext secrets are masked on ingest (`sk-****abcd`, `ghp_****91xz`, `Bearer ****ef12`) and never returned in full.
- **Strict SSRF Guard:** All outbound network operations (tools, proxy, webhooks) check IP ranges, loopback (`127.0.0.1`), link-local (`169.254.0.0/16`), and AWS/GCP cloud metadata endpoints (`169.254.169.254`).

### 2.3 Modular Provider Framework
- Extensible `ProviderAdapter` contract with category tags, authentication configurations, explicit capability flags, credential validation, and health checks.
- 11 supported categories: Developer, Communication, Monitoring, Automation, Scraping & Data, Storage, Payments, Analytics, Maps, Security, and Custom.

### 2.4 Asynchronous Logging & Graceful Shutdown
- Usage recording, audit logs, and delivery events are buffered via non-blocking channels and processed by bounded background worker routines.
- System traps `SIGINT` and `SIGTERM`, shuts down the HTTP server gracefully within a 10-second deadline, cancels background workers, stops the scheduler, and closes database handles.
