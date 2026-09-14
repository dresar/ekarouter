# 01. Project Overview & Operational Context

---

## 1. System Identity & Mission

**EkaRouter** is an enterprise-grade, lightweight, single-binary Universal AI Gateway and Developer API Management Platform written in pure Go. 

The application solves two fundamental infrastructure challenges:
1. **AI Workload Routing & Failover**: Providing a single entry point compatible with the OpenAI API specification (`/v1/chat/completions`, `/v1/models`, `/v1/responses`) that translates requests across 50+ AI model providers, dynamically balances workloads via multi-strategy round-robin and priority policies, tracks token usage, and automatically recovers from provider rate limits (429) or authentication failures (401/403).
2. **Developer Infrastructure & API Key Vault**: Centralizing third-party developer APIs across 11 functional categories (Cloudflare, GitHub, Stripe, Twilio, AWS, Resend, etc.), storing credentials in an AES-256-GCM encrypted vault, executing SSRF-safe parameterized tool calls, enforcing quotas, and dispatching signed webhooks.

---

## 2. Operational Architecture

```text
[ Client Applications / AI IDEs / CLI ]
                   │
                   ▼
  ┌─────────────────────────────────────────────────────────┐
  │                   EkaRouter Gateway                     │
  │                   (Port 8080)                           │
  ├────────────────────────────┬────────────────────────────┤
  │   Public AI Ingress        │   Management Control Plane │
  │   /v1/chat/completions     │   /api/* & /api/v1/*       │
  │   (OpenAI Spec / SSE)      │   (JWT Session / RBAC)     │
  └─────────────┬──────────────┴──────────────┬─────────────┘
                │                             │
    ┌───────────▼───────────┐     ┌───────────▼─────────────┐
    │  Failover & Rotator   │     │  Platform Vault & SSRF  │
    │  - Priority Fallback  │     │  - AES-256-GCM Vault    │
    │  - Round-Robin Cursors│     │  - Rate Limiter & Quota │
    │  - Cooldown Circuit   │     │  - Generic Tool Caller  │
    │  - Token Saver Compactor    │  - HMAC Webhooks        │
    └───────────┬───────────┘     └───────────┬─────────────┘
                │                             │
    ┌───────────▼─────────────────────────────▼─────────────┐
    │              Embedded SQLite Database                 │
    │              (modernc.org/sqlite - WAL)               │
    └───────────────────────────────────────────────────────┘
```

---

## 3. High-Level Capabilities

### Universal AI Gateway
- **Ingress Routes**: `/v1/chat/completions`, `/v1/models`, `/v1/responses`.
- **Streaming**: Server-Sent Events (SSE) streaming with full reasoning/thinking token propagation.
- **Failover**: Automatic target rotation upon transient network errors, HTTP 429 quota exhaustion, or HTTP 401/403 credential rejection.
- **Token Saver**: Heuristic prompt compaction for git diffs, build logs, and code traces without stripping essential line numbers or identifiers.

### Developer API Platform
- **Encrypted Vault**: Symmetric AES-256-GCM encryption with versioned ciphertext (`v1:<base64>`) and deterministic secret masking.
- **Provider Categories**: 11 universal categories (Developer, Communication, Monitoring, Automation, Scraping, Storage, Payments, Analytics, Maps, Security, Custom).
- **SSRF Defense**: Strictly blocks loopback (`127.0.0.0/8`), link-local (`169.254.0.0/16`), RFC 1918 private subnets, and cloud instance metadata addresses (`169.254.169.254`).
- **Tool Calling**: Safe template variable interpolation `{{var}}` with CRLF injection sanitization.
- **Audit Logging**: Immutable append-only audit trail for administrative operations and tool executions.

---

## 4. Frontend Application Boundaries

The frontend will be built as an independent Single Page Application (SPA) or Server-Rendered web console (Next.js / Vite) deployed to Vercel. 

Key boundaries:
1. **Zero Direct Database Access**: The frontend interacts strictly over HTTP/HTTPS with the EkaRouter management APIs.
2. **Cross-Origin Deployment**: Frontend on `console.ekarouter.io` (or `*.vercel.app`) communicates with backend on `api.ekarouter.io` (or `http://localhost:8080`).
3. **Authentication Boundary**: Authentication state is maintained via Bearer JWT tokens or session cookies returned by `/api/auth/login`.
