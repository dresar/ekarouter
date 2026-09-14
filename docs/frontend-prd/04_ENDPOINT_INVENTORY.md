# 04. Comprehensive Backend Endpoint Inventory

Complete inventory of all HTTP routes discovered, inspected, and verified in the EkaRouter backend codebase.

---

## 1. System Health & Probes (`/`)

| Method | Path | Auth | Purpose | Response Format | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `GET` | `/health` | Public | Process health summary | `{"status":"ok","database":"ok","uptime":"..."}` | 200 OK |
| `GET` | `/ready` | Public | Database readiness probe | `{"status":"ready"}` | 200 OK |
| `GET` | `/live` | Public | Liveness ping | `{"status":"alive"}` | 200 OK |

---

## 2. Public AI Gateway (`/v1/*`)

| Method | Path | Auth | Purpose | Response Format | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `GET` | `/v1/models` | Bearer API Key | OpenAI-compatible model list | `{"object":"list","data":[{"id":"gpt-4o",...}]}` | 200 OK |
| `POST` | `/v1/chat/completions` | Bearer API Key | Standard & Streaming SSE Chat | JSON Object or SSE Stream `data: {...}` | 200 OK |
| `POST` | `/v1/responses` | Bearer API Key | Normalized response format | `{"id":"...","model":"...","output":"..."}` | 200 OK |

---

## 3. Administrative Management API (`/api/*`)

Requires `Authorization: Bearer <token>` or valid `ekarouter_session` cookie.

### Authentication (`/api/auth/*`)
- `POST /api/auth/login`: Accepts `{"username":"...","password":"..."}`, returns JWT token and user info.
- `POST /api/auth/logout`: Invalidates session cookie.
- `GET /api/auth/me`: Returns current authenticated user metadata.

### Providers (`/api/providers`)
- `GET /api/providers`: Returns array of configured AI providers.
- `POST /api/providers`: Upserts provider (`id`, `key`, `name`, `kind`, `base_url`, `enabled`).
- `DELETE /api/providers/{id}`: Removes provider and cascades related accounts.

### Accounts (`/api/accounts`)
- `GET /api/accounts`: Lists all provider accounts with priority and state.
- `POST /api/accounts`: Upserts account with API keys/secret keys (AES-256-GCM encrypted).
- `DELETE /api/accounts/{id}`: Removes account and associated credentials.
- `POST /api/accounts/oauth/start`: Initiates OAuth PKCE authorization.
- `POST /api/accounts/oauth/callback`: Handles OAuth redirect and token storage.

### Routes & Combos (`/api/routes`)
- `GET /api/routes`: Lists model routes with child `items` (priority, weights, retries).
- `POST /api/routes`: Upserts route definition and atomically refreshes route items.
- `DELETE /api/routes/{id}`: Deletes route and clears cached router definitions.

### Models (`/api/models`)
- `GET /api/models`: Lists registered models across all providers.
- `POST /api/models`: Registers custom model alias or mapping.
- `DELETE /api/models/{id}`: Removes model registration.

### Gateway API Keys (`/api/keys`)
- `GET /api/keys`: Returns active client keys (`prefix`, `name`, `scopes`, `last_used_at`).
- `POST /api/keys`: Generates new API key; returns full plaintext key once (`eka_...`).
- `DELETE /api/keys/{id}`: Permanently revokes key.

### Outbound Proxies (`/api/proxies` & `/api/proxy-profiles`)
- `GET /api/proxies`: Lists configured outbound proxy profiles.
- `POST /api/proxies`: Creates or updates proxy (`scheme`, `host`, `port`, `username`, `password`).
- `DELETE /api/proxies/{id}`: Deletes proxy profile.
- `POST /api/proxies/{id}/test`: Dispatches live HTTP probe to test proxy connectivity and latency.

### System Settings & Backup (`/api/settings` & `/api/backup`)
- `GET /api/settings`: Returns key-value system settings map.
- `PUT /api/settings`: Updates single setting key (`{"key":"theme","value":"dark"}`).
- `GET /api/usage`: Returns aggregate gateway usage and token counts.
- `POST /api/tokensaver/preview`: Previews heuristic token reduction on provided text.
- `GET /api/backup`: Returns recent SQLite backup files.
- `POST /api/backup`: Executes online SQLite `VACUUM INTO` backup.

---

## 4. Developer Platform API (`/api/v1/*`)

Requires `Authorization: Bearer <token>` or developer client token (`eka_pat_*`).

### Platform Providers (`/api/v1/providers`)
- `GET /api/v1/providers`: Lists catalog of 11 universal provider categories.
- `GET /api/v1/providers/{id}`: Returns provider details and supported capabilities.
- `POST /api/v1/providers/{id}/validate`: Validates credential format.
- `POST /api/v1/providers/{id}/health`: Checks provider reachability.
- `GET /api/v1/providers/{id}/capabilities`: Returns explicit feature flags.
- `GET /api/v1/providers/{id}/docs`: Returns reference links and free tier documentation.

### Credential Vault (`/api/v1/credentials`)
- `GET /api/v1/credentials`: Lists vault credentials with automatic secret masking.
- `POST /api/v1/credentials`: Encrypts and stores credential secret.
- `GET /api/v1/credentials/{id}`: Returns masked credential metadata.
- `PATCH /api/v1/credentials/{id}`: Updates priority, tags, or notes.
- `DELETE /api/v1/credentials/{id}`: Permanently removes credential from vault.
- `POST /api/v1/credentials/{id}/test`: Tests credential against provider API.
- `POST /api/v1/credentials/{id}/rotate`: Rotates encrypted secret.
- `GET /api/v1/credentials/{id}/usage`: Returns request count and error totals.
- `GET /api/v1/credentials/{id}/health`: Returns health state and active cooldown expiry.

### Workspaces & Projects (`/api/v1/projects` & `/api/v1/environments`)
- `GET /api/v1/projects`: Lists developer projects.
- `POST /api/v1/projects`: Creates project.
- `GET /api/v1/environments`: Lists deployment environments (`production`, `staging`, `dev`).
- `POST /api/v1/environments`: Creates environment.

### Generic Tools & Request Templates (`/api/v1/tools` & `/api/v1/request-templates`)
- `GET /api/v1/tools`: Lists registered generic HTTP tools.
- `POST /api/v1/tools`: Creates tool definition.
- `GET /api/v1/tools/{id}`: Returns tool details.
- `GET /api/v1/tools/{id}/schema`: Returns parameter schema for tool.
- `POST /api/v1/tools/{id}/execute`: Executes tool with variable substitution and SSRF validation.
- `GET /api/v1/request-templates`: Lists reusable HTTP request templates.
- `POST /api/v1/request-templates`: Creates request template.
- `GET /api/v1/request-templates/{id}`: Returns template details.
- `PATCH /api/v1/request-templates/{id}`: Updates request template fields.
- `DELETE /api/v1/request-templates/{id}`: Deletes request template.
- `POST /api/v1/request-templates/{id}/execute`: Executes request template safely.

### Telemetry, Audit & Proxying
- `GET /api/v1/usage`: Usage summary metrics.
- `GET /api/v1/usage/summary`: Detailed usage aggregate.
- `GET /api/v1/usage/providers`: Usage grouped by provider.
- `GET /api/v1/usage/credentials`: Usage grouped by credential.
- `GET /api/v1/usage/projects`: Usage grouped by project.
- `GET /api/v1/health`: Platform health summary.
- `GET /api/v1/health/providers`: Health status per provider.
- `GET /api/v1/health/credentials`: Health status per vault credential.
- `GET /api/v1/audit-logs`: Paginated compliance audit log.
- `ALL /api/v1/proxy/{provider}/*`: Secure outbound reverse proxy injecting vault credentials.
