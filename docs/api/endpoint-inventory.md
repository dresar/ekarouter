# EkaRouter Complete Endpoint Inventory

**Generated At:** 2026-09-15  
**Audit Phase:** Final Quality Assurance & Release Readiness  
**Target:** Frontend UI/UX Integration Contract

---

## 1. Summary Matrix

| Method | Endpoint | Purpose | Auth | Permission | Test Status | Notes |
|---|---|---|---|---|---|---|
| `GET` | `/health` | Overall system health check | None | Public | **PASS** | Returns 200 `{"status":"ok"}` |
| `GET` | `/ready` | Readiness check (validates DB connectivity) | None | Public | **PASS** | Returns 200 when SQLite connects |
| `GET` | `/live` | Liveness check for process orchestrator | None | Public | **PASS** | Returns 200 |
| `GET` | `/version` | System version string | None | Public | **PASS** | Returns 200 `{"version":"1.0.0","status":"ok"}` |
| `GET` | `/metrics` | Prometheus metrics endpoint | None | Public | **NOT_IMPLEMENTED** | Internal counters in DB |
| `POST` | `/v1/chat/completions` | OpenAI-compatible chat completion proxy | Bearer API Key | `gateway.execute` | **PASS** | Streaming SSE & JSON supported |
| `POST` | `/v1/responses` | Chat completions alternate endpoint | Bearer API Key | `gateway.execute` | **PASS** | Input string fallback supported |
| `GET` | `/v1/models` | List available models & routes | Bearer API Key | `gateway.read` | **PASS** | OpenAI model list format |
| `GET` | `/api/v1/providers` | List registered developer platform providers | Session / Token | `providers.read` | **PASS** | Returns 27 registered providers |
| `GET` | `/api/v1/providers/{id}` | Get specific platform provider metadata | Session / Token | `providers.read` | **PASS** | Returns provider metadata & schema |
| `GET` | `/api/v1/providers/{id}/capabilities` | List provider capabilities | Session / Token | `providers.read` | **PASS** | Lists supported auth & operations |
| `GET` | `/api/v1/providers/{id}/docs` | Retrieve provider official documentation URLs | Session / Token | `providers.read` | **PASS** | Returns URLs and usage notes |
| `POST` | `/api/v1/providers/{id}/validate` | Validate credentials against provider API | Session / Token | `providers.test` | **PASS** | Safe minimal validation |
| `POST` | `/api/v1/providers/{id}/health` | Perform live provider health check | Session / Token | `providers.test` | **PASS** | Returns latency and status |
| `GET` | `/api/v1/credentials` | List stored credentials (masked) | Session / Token | `credentials.read` | **PASS** | Secrets are always masked |
| `POST` | `/api/v1/credentials` | Create encrypted credential in vault | Session / Token | `credentials.create` | **PASS** | AES-256-GCM encrypted |
| `GET` | `/api/v1/credentials/{id}` | Get credential metadata & masked key | Session / Token | `credentials.read` | **PASS** | Never leaks raw secret |
| `PATCH` | `/api/v1/credentials/{id}` | Update credential metadata | Session / Token | `credentials.update` | **PASS** | Name, status, priority, env |
| `DELETE` | `/api/v1/credentials/{id}` | Delete credential and revoke versions | Session / Token | `credentials.delete` | **PASS** | Cascades assignments |
| `POST` | `/api/v1/credentials/{id}/test` | Test stored credential against provider | Session / Token | `credentials.test` | **PASS** | Validates connectivity |
| `POST` | `/api/v1/credentials/{id}/enable` | Enable disabled credential | Session / Token | `credentials.update` | **PASS** | Sets status to active |
| `POST` | `/api/v1/credentials/{id}/disable` | Disable active credential | Session / Token | `credentials.update` | **PASS** | Prevents router selection |
| `POST` | `/api/v1/credentials/{id}/rotate` | Rotate credential secret value | Session / Token | `credentials.rotate` | **PASS** | Creates version entry |
| `GET` | `/api/v1/credentials/{id}/health` | Read credential health logs | Session / Token | `credentials.read` | **PASS** | Health history |
| `GET` | `/api/v1/credentials/{id}/usage` | Read credential usage counters | Session / Token | `usage.read` | **PASS** | Request and error counts |
| `GET` | `/api/v1/credentials/{id}/events` | Read credential event history | Session / Token | `audit.read` | **PASS** | Status & rotation events |
| `GET` | `/api/v1/projects` | List developer projects | Session / Token | `projects.read` | **PASS** | Project list |
| `POST` | `/api/v1/projects` | Create a new project | Session / Token | `projects.create` | **PASS** | Generates project ID |
| `GET` | `/api/v1/projects/{id}` | Get project detail | Session / Token | `projects.read` | **PASS** | Project detail |
| `PATCH` | `/api/v1/projects/{id}` | Update project metadata | Session / Token | `projects.update` | **PASS** | Name, description |
| `DELETE` | `/api/v1/projects/{id}` | Delete project | Session / Token | `projects.delete` | **PASS** | Cascades environments |
| `GET` | `/api/v1/environments` | List project environments | Session / Token | `environments.read` | **PASS** | Filterable by project |
| `POST` | `/api/v1/environments` | Create new environment | Session / Token | `environments.create` | **PASS** | e.g. production, staging |
| `PATCH` | `/api/v1/environments/{id}` | Update environment | Session / Token | `environments.update` | **PASS** | Name, description |
| `DELETE` | `/api/v1/environments/{id}` | Delete environment | Session / Token | `environments.delete` | **PASS** | Clean removal |
| `GET` | `/api/v1/tools` | List registered executor tools | Session / Token | `tools.read` | **PASS** | Available tools |
| `POST` | `/api/v1/tools` | Register new platform tool | Session / Token | `tools.create` | **PASS** | Schema & definition |
| `GET` | `/api/v1/tools/{id}` | Get tool detail | Session / Token | `tools.read` | **PASS** | Tool detail |
| `GET` | `/api/v1/tools/{id}/schema` | Get tool parameter JSON schema | Session / Token | `tools.read` | **PASS** | JSON Schema spec |
| `POST` | `/api/v1/tools/{id}/execute` | Execute tool with safe SSRF check | Session / Token | `tools.execute` | **PASS** | Isolated executor |
| `POST` | `/api/v1/tools/{id}/test` | Dry-run / test tool execution | Session / Token | `tools.execute` | **PASS** | Validates execution |
| `GET` | `/api/v1/request-templates` | List request templates | Session / Token | `templates.read` | **PASS** | Pre-configured requests |
| `POST` | `/api/v1/request-templates` | Create request template | Session / Token | `templates.create` | **PASS** | Path, method, headers |
| `GET` | `/api/v1/request-templates/{id}` | Get request template | Session / Token | `templates.read` | **PASS** | Template detail |
| `PATCH` | `/api/v1/request-templates/{id}` | Update request template | Session / Token | `templates.update` | **PASS** | Template update |
| `DELETE` | `/api/v1/request-templates/{id}` | Delete request template | Session / Token | `templates.delete` | **PASS** | Clean deletion |
| `POST` | `/api/v1/request-templates/{id}/execute` | Execute request template | Session / Token | `templates.execute` | **PASS** | Templated HTTP call |
| `GET` | `/api/v1/usage` | Overall platform usage summary | Session / Token | `usage.read` | **PASS** | Tokens & requests |
| `GET` | `/api/v1/usage/summary` | Aggregate usage metrics | Session / Token | `usage.read` | **PASS** | Usage rollup |
| `GET` | `/api/v1/usage/providers` | Usage breakdown by provider | Session / Token | `usage.read` | **PASS** | Provider breakdown |
| `GET` | `/api/v1/usage/credentials` | Usage breakdown by credential | Session / Token | `usage.read` | **PASS** | Credential breakdown |
| `GET` | `/api/v1/usage/projects` | Usage breakdown by project | Session / Token | `usage.read` | **PASS** | Project breakdown |
| `GET` | `/api/v1/health` | Health summary of platform | Session / Token | `health.read` | **PASS** | Provider & cred health |
| `GET` | `/api/v1/audit-logs` | Query structured audit events | Session / Token | `audit.read` | **PASS** | Paginated audit log |
| `GET` | `/api/v1/events` | Alias for audit-logs | Session / Token | `audit.read` | **PASS** | Event log |
| `GET` | `/api/v1/system/settings` | Read platform system settings | Session / Token | `settings.read` | **PASS** | Settings map |
| `PATCH` | `/api/v1/system/settings` | Update system setting | Session / Token | `settings.write` | **PASS** | Key-value upsert |
| `PUT` | `/api/v1/system/settings` | Upsert system setting | Session / Token | `settings.write` | **PASS** | Key-value upsert |
| `GET` | `/api/v1/proxy/routes` | List configured proxy routes | Session / Token | `routes.read` | **PASS** | Route list |
| `POST` | `/api/v1/proxy/routes` | Create proxy route | Session / Token | `routes.create` | **PASS** | Route creation |
| `GET` | `/api/v1/proxy/routes/{id}` | Get proxy route detail | Session / Token | `routes.read` | **PASS** | Route detail |
| `DELETE` | `/api/v1/proxy/routes/{id}` | Delete proxy route | Session / Token | `routes.delete` | **PASS** | Route deletion |
| `ANY` | `/api/v1/proxy/{provider}/*` | Universal proxy dispatch with SSRF filter | Session / Token | `proxy.execute` | **PASS** | Forwarding with creds |
| `POST` | `/api/auth/login` | Administrator authentication | None | Public | **PASS** | Issues session cookie |
| `POST` | `/auth/login` | Root alias administrator login | None | Public | **PASS** | Issues session cookie |
| `POST` | `/api/auth/logout` | Revoke current admin session | Session Cookie | Admin | **PASS** | Clears session cookie |
| `POST` | `/auth/logout` | Root alias logout | Session Cookie | Admin | **PASS** | Clears session cookie |
| `GET` | `/api/auth/me` | Current authenticated admin identity | Session Cookie | Admin | **PASS** | Returns username |
| `GET` | `/auth/me` | Root alias admin identity | Session Cookie | Admin | **PASS** | Returns username |
| `GET` | `/api/auth/sessions` | List active admin sessions | Session Cookie | Admin | **PASS** | Session list |
| `GET` | `/auth/sessions` | Root alias list sessions | Session Cookie | Admin | **PASS** | Session list |
| `DELETE` | `/api/auth/sessions/{id}` | Revoke active admin session | Session Cookie | Admin | **PASS** | Revokes session |
| `DELETE` | `/auth/sessions/{id}` | Root alias revoke session | Session Cookie | Admin | **PASS** | Revokes session |
| `GET` | `/api/providers` | Admin list providers | Session Cookie | Admin | **PASS** | Provider registry |
| `POST` | `/api/providers` | Register or update provider | Session Cookie | Admin | **PASS** | Key, kind, base_url |
| `DELETE` | `/api/providers/{id}` | Delete registered provider | Session Cookie | Admin | **PASS** | Cascades accounts |
| `GET` | `/api/accounts` | Admin list provider accounts | Session Cookie | Admin | **PASS** | Account priority & state |
| `POST` | `/api/accounts` | Create provider account & credentials | Session Cookie | Admin | **PASS** | Encrypts secrets |
| `DELETE` | `/api/accounts/{id}` | Delete provider account | Session Cookie | Admin | **PASS** | Cascades routes |
| `POST` | `/api/accounts/oauth/start` | Begin PKCE OAuth2 flow | Session Cookie | Admin | **PASS** | Generates auth state |
| `POST` | `/api/accounts/oauth/callback` | Complete OAuth2 token exchange | Session Cookie | Admin | **PASS** | Encrypts refresh token |
| `GET` | `/api/credentials` | Admin list raw account credentials | Session Cookie | Admin | **PASS** | Key fingerprints |
| `GET` | `/api/credentials/{id}` | Admin get single account credential | Session Cookie | Admin | **PASS** | Account binding |
| `GET` | `/api/routes` | List configured AI model routes | Session Cookie | Admin | **PASS** | Model routing rules |
| `POST` | `/api/routes` | Create or update route & priority items | Session Cookie | Admin | **PASS** | Fallback chain setup |
| `GET` | `/api/routes/{id}` | Get AI model route detail | Session Cookie | Admin | **PASS** | Model route detail |
| `DELETE` | `/api/routes/{id}` | Delete AI model route | Session Cookie | Admin | **PASS** | Route deletion |
| `GET` | `/api/models` | List models in catalog | Session Cookie | Admin | **PASS** | Model context & caps |
| `POST` | `/api/models` | Add model to catalog | Session Cookie | Admin | **PASS** | Context limit, streaming |
| `DELETE` | `/api/models/{id}` | Remove model from catalog | Session Cookie | Admin | **PASS** | Clean removal |
| `GET` | `/api/proxies` | List outbound proxy profiles | Session Cookie | Admin | **PASS** | Password masked |
| `POST` | `/api/proxies` | Create or update proxy profile | Session Cookie | Admin | **PASS** | Auto-ID generated |
| `GET` | `/api/proxies/{id}` | Get proxy profile (password masked) | Session Cookie | Admin | **PASS** | Profile detail |
| `PUT` | `/api/proxies/{id}` | Update proxy profile | Session Cookie | Admin | **PASS** | Update profile |
| `PATCH` | `/api/proxies/{id}` | Patch proxy profile | Session Cookie | Admin | **PASS** | Patch profile |
| `DELETE` | `/api/proxies/{id}` | Delete proxy profile | Session Cookie | Admin | **PASS** | Deletion verified |
| `POST` | `/api/proxies/{id}/test` | Test outbound proxy connection | Session Cookie | Admin | **PASS** | Network probe |
| `POST` | `/api/proxies/{id}/enable` | Enable proxy profile | Session Cookie | Admin | **PASS** | Sets enabled |
| `POST` | `/api/proxies/{id}/disable` | Disable proxy profile | Session Cookie | Admin | **PASS** | Sets disabled |
| `GET` | `/api/keys` | List generated client API keys | Session Cookie | Admin | **PASS** | Prefix & last used |
| `POST` | `/api/keys` | Generate new client API key | Session Cookie | Admin | **PASS** | Returns full raw key once |
| `DELETE` | `/api/keys/{id}` | Revoke client API key | Session Cookie | Admin | **PASS** | Immediate revocation |
| `GET` | `/api/settings` | Read application system settings | Session Cookie | Admin | **PASS** | Key-value settings |
| `PUT` | `/api/settings` | Update system setting | Session Cookie | Admin | **PASS** | Upserts setting |
| `GET` | `/api/usage` | Query request logs & usage summary | Session Cookie | Admin | **PASS** | Aggregated usage |
| `POST` | `/api/tokensaver/preview` | Preview token compression on prompt | Session Cookie | Admin | **PASS** | Computes token delta |
| `GET` | `/api/backup` | List database backup files | Session Cookie | Admin | **PASS** | Backup directory |
| `POST` | `/api/backup` | Trigger instantaneous SQLite backup | Session Cookie | Admin | **PASS** | Online VACUUM INTO |
| `GET` | `/api/credential-pools` | List intelligent credential pools | Session Cookie | Admin | **PASS** | Pool list |
| `POST` | `/api/credential-pools` | Create credential pool | Session Cookie | Admin | **PASS** | Auto-ID, default policy |
| `GET` | `/api/credential-pools/{id}` | Get credential pool detail | Session Cookie | Admin | **PASS** | Pool status & env |
| `DELETE` | `/api/credential-pools/{id}` | Delete credential pool | Session Cookie | Admin | **PASS** | Cascades members |
| `POST` | `/api/credential-pools/{id}/pause` | Pause pool (halt traffic) | Session Cookie | Admin | **PASS** | Status paused |
| `POST` | `/api/credential-pools/{id}/resume` | Resume pool traffic | Session Cookie | Admin | **PASS** | Status active |
| `POST` | `/api/credential-pools/{id}/rotate` | Force manual rotation in pool | Session Cookie | Admin | **PASS** | Rotates active index |
| `GET` | `/api/free-tiers` | Query free tier directory | Session Cookie | Admin | **PASS** | Free quotas catalog |
| `GET` | `/api/free-tiers/categories` | Free tier categories | Session Cookie | Admin | **PASS** | Category breakdown |
| `GET` | `/api/free-tiers/verified` | List verified free tiers | Session Cookie | Admin | **PASS** | High confidence tiers |
| `GET` | `/api/devtools/templates` | List pre-built request templates | Session Cookie | Admin | **PASS** | Template gallery |
| `POST` | `/api/devtools/generate-curl` | Generate executable cURL from template | Session Cookie | Admin | **PASS** | Generates sanitized bash |
| `POST` | `/api/devtools/import-openapi` | Parse & preview OpenAPI document | Session Cookie | Admin | **PASS** | Previews endpoints |
| `POST` | `/api/devtools/confirm-openapi` | Save parsed OpenAPI operations | Session Cookie | Admin | **PASS** | Persists templates |

---

## 2. Detailed Endpoint Specifications

### Gateway: `POST /v1/chat/completions`
- **Purpose**: Unified OpenAI-compatible chat completion proxy routing requests across providers with automatic retry, rotation, and prompt compaction.
- **Authentication**: `Authorization: Bearer <API_KEY>`
- **Headers**:
  - `Content-Type: application/json` (Required)
  - `X-Request-ID`: Optional client tracing ID
  - `X-Token-Saver`: Optional `off` to bypass compaction
- **Request Body**:
  ```json
  {
    "model": "gpt-4o",
    "messages": [
      {"role": "user", "content": "Hello, EkaRouter!"}
    ],
    "stream": false,
    "temperature": 0.7
  }
  ```
- **Responses**:
  - `200 OK`: Non-streaming JSON or `text/event-stream` chunks
  - `400 Bad Request`: Missing model or invalid JSON
  - `401 Unauthorized`: Invalid or disabled API key
  - `502 Bad Gateway`: All upstream route targets failed or in cooldown

### Platform Credentials: `POST /api/v1/credentials`
- **Purpose**: Securely stores a third-party provider credential encrypted with AES-256-GCM.
- **Authentication**: `Authorization: Bearer <SESSION_TOKEN_OR_CLIENT_TOKEN>`
- **Request Body**:
  ```json
  {
    "name": "Production Cloudflare Token",
    "provider_id": "cloudflare",
    "secret_value": "secret_token_value_here",
    "environment": "production",
    "project_id": "proj_123"
  }
  ```
- **Responses**:
  - `201 Created`:
    ```json
    {
      "success": true,
      "data": {
        "id": "cred_abc123",
        "name": "Production Cloudflare Token",
        "provider_id": "cloudflare",
        "masked_value": "secr***here",
        "environment": "production",
        "status": "active"
      }
    }
    ```
  - `400 Bad Request`: Missing name, provider_id, or secret_value
  - `401 Unauthorized`: Missing or invalid session/token
