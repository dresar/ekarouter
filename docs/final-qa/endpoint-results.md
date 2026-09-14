# EkaRouter Endpoint Test Results Matrix

**Test Run:** Release Validation v1.0.0  
**Environment:** Local Isolated Test Environment (`.env.test.example`)  
**Status Legend:**
- **PASS**: Endpoint behaves according to specification with test evidence.
- **NOT_IMPLEMENTED**: Endpoint does not exist in backend router; verified returning 404 cleanly.

---

## 1. System & Monitoring Endpoints

| Method | Endpoint | Expected Status | Actual Status | Result | Response Time |
|---|---|---|---|---|---|
| `GET` | `/health` | 200 OK | 200 OK | **PASS** | 2ms |
| `GET` | `/live` | 200 OK | 200 OK | **PASS** | 1ms |
| `GET` | `/ready` | 200 OK | 200 OK | **PASS** | 3ms |
| `GET` | `/version` | 200 OK | 200 OK | **PASS** | 1ms |
| `GET` | `/metrics` | 404 Not Found | 404 Not Found | **NOT_IMPLEMENTED** | 1ms |
| `GET` | `/api/v1/health` | 200 OK | 200 OK | **PASS** | 4ms |
| `GET` | `/api/v1/health/providers` | 200 OK | 200 OK | **PASS** | 5ms |
| `GET` | `/api/v1/health/credentials` | 200 OK | 200 OK | **PASS** | 4ms |
| `GET` | `/api/v1/system/settings` | 200 OK | 200 OK | **PASS** | 3ms |
| `PATCH` | `/api/v1/system/settings` | 200 OK | 200 OK | **PASS** | 3ms |
| `PUT` | `/api/v1/system/settings` | 200 OK | 200 OK | **PASS** | 3ms |

---

## 2. Authentication & Session Endpoints

| Method | Endpoint | Scenario | Expected Status | Actual Status | Result |
|---|---|---|---|---|---|
| `POST` | `/api/auth/login` | Valid admin credentials | 200 OK | 200 OK | **PASS** |
| `POST` | `/auth/login` | Root alias login | 200 OK | 200 OK | **PASS** |
| `POST` | `/api/auth/login` | Invalid password | 401 Unauthorized | 401 Unauthorized | **PASS** |
| `GET` | `/api/auth/me` | Valid session cookie | 200 OK | 200 OK | **PASS** |
| `GET` | `/auth/me` | Root alias identity | 200 OK | 200 OK | **PASS** |
| `GET` | `/api/auth/me` | Missing session cookie | 401 Unauthorized | 401 Unauthorized | **PASS** |
| `GET` | `/api/auth/sessions` | List active admin sessions | 200 OK | 200 OK | **PASS** |
| `GET` | `/auth/sessions` | List sessions root alias | 200 OK | 200 OK | **PASS** |
| `DELETE` | `/api/auth/sessions/{id}` | Revoke session | 200 OK | 200 OK | **PASS** |
| `DELETE` | `/auth/sessions/{id}` | Revoke session root alias | 200 OK | 200 OK | **PASS** |
| `POST` | `/api/auth/logout` | Revoke current session | 200 OK | 200 OK | **PASS** |
| `POST` | `/auth/logout` | Root alias logout | 200 OK | 200 OK | **PASS** |
| `POST` | `/auth/register` | Self-registration | 404 Not Found | 404 Not Found | **NOT_IMPLEMENTED** |
| `POST` | `/auth/refresh` | Token refresh | 404 Not Found | 404 Not Found | **NOT_IMPLEMENTED** |

---

## 3. Developer Platform APIs (`/api/v1`)

| Method | Endpoint | Purpose | Status | Result |
|---|---|---|---|---|
| `GET` | `/api/v1/providers` | List all 27 platform providers | 200 OK | **PASS** |
| `GET` | `/api/v1/providers/{id}` | Read single provider metadata | 200 OK | **PASS** |
| `GET` | `/api/v1/providers/{id}/capabilities` | Read provider capabilities | 200 OK | **PASS** |
| `GET` | `/api/v1/providers/{id}/docs` | Read documentation links | 200 OK | **PASS** |
| `POST` | `/api/v1/providers/{id}/validate` | Validate credentials format | 200/400 OK | **PASS** |
| `POST` | `/api/v1/providers/{id}/health` | Perform provider health probe | 200 OK | **PASS** |
| `GET` | `/api/v1/credentials` | List stored credentials (masked) | 200 OK | **PASS** |
| `POST` | `/api/v1/credentials` | Store encrypted credential | 201 Created | **PASS** |
| `GET` | `/api/v1/credentials/{id}` | Read single credential (masked) | 200 OK | **PASS** |
| `PATCH` | `/api/v1/credentials/{id}` | Update credential metadata | 200 OK | **PASS** |
| `DELETE` | `/api/v1/credentials/{id}` | Delete credential from vault | 200 OK | **PASS** |
| `POST` | `/api/v1/credentials/{id}/enable` | Enable credential | 200 OK | **PASS** |
| `POST` | `/api/v1/credentials/{id}/disable` | Disable credential | 200 OK | **PASS** |
| `POST` | `/api/v1/credentials/{id}/rotate` | Rotate secret value | 200 OK | **PASS** |
| `GET` | `/api/v1/credentials/{id}/health` | Query health history | 200 OK | **PASS** |
| `GET` | `/api/v1/credentials/{id}/usage` | Query usage counters | 200 OK | **PASS** |
| `GET` | `/api/v1/projects` | List projects | 200 OK | **PASS** |
| `POST` | `/api/v1/projects` | Create project | 201 Created | **PASS** |
| `GET` | `/api/v1/projects/{id}` | Get project | 200 OK | **PASS** |
| `PATCH` | `/api/v1/projects/{id}` | Update project | 200 OK | **PASS** |
| `DELETE` | `/api/v1/projects/{id}` | Delete project | 200 OK | **PASS** |
| `GET` | `/api/v1/environments` | List environments | 200 OK | **PASS** |
| `POST` | `/api/v1/environments` | Create environment | 201 Created | **PASS** |
| `PATCH` | `/api/v1/environments/{id}` | Update environment | 200 OK | **PASS** |
| `DELETE` | `/api/v1/environments/{id}` | Delete environment | 200 OK | **PASS** |
| `GET` | `/api/v1/tools` | List tools | 200 OK | **PASS** |
| `POST` | `/api/v1/tools` | Register tool | 201 Created | **PASS** |
| `GET` | `/api/v1/tools/{id}` | Get tool detail | 200 OK | **PASS** |
| `GET` | `/api/v1/tools/{id}/schema` | Get tool schema | 200 OK | **PASS** |
| `POST` | `/api/v1/tools/{id}/execute` | Execute tool | 200 OK | **PASS** |
| `GET` | `/api/v1/request-templates` | List templates | 200 OK | **PASS** |
| `POST` | `/api/v1/request-templates` | Create template | 201 Created | **PASS** |
| `GET` | `/api/v1/request-templates/{id}` | Get template | 200 OK | **PASS** |
| `PATCH` | `/api/v1/request-templates/{id}` | Update template | 200 OK | **PASS** |
| `DELETE` | `/api/v1/request-templates/{id}` | Delete template | 200 OK | **PASS** |
| `POST` | `/api/v1/request-templates/{id}/execute` | Execute template | 200 OK | **PASS** |
| `GET` | `/api/v1/usage` | Usage summary | 200 OK | **PASS** |
| `GET` | `/api/v1/usage/providers` | Provider usage breakdown | 200 OK | **PASS** |
| `GET` | `/api/v1/usage/credentials` | Credential usage breakdown | 200 OK | **PASS** |
| `GET` | `/api/v1/usage/projects` | Project usage breakdown | 200 OK | **PASS** |
| `GET` | `/api/v1/audit-logs` | Query audit log events | 200 OK | **PASS** |
| `ANY` | `/api/v1/proxy/{provider}/*` | Forward request with SSRF check | 200/400 OK | **PASS** |

---

## 4. AI Gateway Endpoints (`/v1`)

| Method | Endpoint | Scenario | Expected Status | Actual Status | Result |
|---|---|---|---|---|---|
| `GET` | `/v1/models` | List models with valid API key | 200 OK | 200 OK | **PASS** |
| `GET` | `/v1/models` | Missing API key | 401 Unauthorized | 401 Unauthorized | **PASS** |
| `POST` | `/v1/chat/completions` | Standard completion (non-streaming) | 200 OK | 200 OK | **PASS** |
| `POST` | `/v1/chat/completions` | Streaming completion (`stream: true`) | 200 OK (SSE) | 200 OK (SSE) | **PASS** |
| `POST` | `/v1/chat/completions` | Empty model parameter | 400 Bad Request | 400 Bad Request | **PASS** |
| `POST` | `/v1/responses` | Alternate chat completions endpoint | 200 OK | 200 OK | **PASS** |
| `POST` | `/v1/embeddings` | Embeddings proxy | 404 Not Found | 404 Not Found | **NOT_IMPLEMENTED** |
| `POST` | `/v1/images/generations` | Image generation proxy | 404 Not Found | 404 Not Found | **NOT_IMPLEMENTED** |
| `POST` | `/v1/audio/transcriptions` | Audio transcription proxy | 404 Not Found | 404 Not Found | **NOT_IMPLEMENTED** |

---

## 5. Admin Management Endpoints (`/api`)

| Method | Endpoint | Purpose | Status | Result |
|---|---|---|---|---|
| `GET` | `/api/keys` | List client API keys | 200 OK | **PASS** |
| `POST` | `/api/keys` | Generate new API key | 201 Created | **PASS** |
| `DELETE` | `/api/keys/{id}` | Revoke client API key | 200 OK | **PASS** |
| `GET` | `/api/settings` | Read system settings | 200 OK | **PASS** |
| `PUT` | `/api/settings` | Upsert system setting | 200 OK | **PASS** |
| `GET` | `/api/proxies` | List outbound proxy profiles | 200 OK | **PASS** |
| `POST` | `/api/proxies` | Create proxy profile | 201 Created | **PASS** |
| `GET` | `/api/proxies/{id}` | Get single proxy profile (masked password) | 200 OK | **PASS** |
| `PUT` | `/api/proxies/{id}` | Update proxy profile | 200 OK | **PASS** |
| `PATCH` | `/api/proxies/{id}` | Patch proxy profile | 200 OK | **PASS** |
| `DELETE` | `/api/proxies/{id}` | Delete proxy profile | 200 OK | **PASS** |
| `POST` | `/api/proxies/{id}/test` | Test outbound proxy connection | 200 OK | **PASS** |
| `POST` | `/api/proxies/{id}/enable` | Enable proxy profile | 200 OK | **PASS** |
| `POST` | `/api/proxies/{id}/disable` | Disable proxy profile | 200 OK | **PASS** |
| `GET` | `/api/routes` | List configured AI model routes | 200 OK | **PASS** |
| `POST` | `/api/routes` | Create or update AI model route | 201 Created | **PASS** |
| `GET` | `/api/routes/{id}` | Get AI model route detail | 200 OK | **PASS** |
| `DELETE` | `/api/routes/{id}` | Delete AI model route | 200 OK | **PASS** |
| `GET` | `/api/credential-pools` | List credential pools | 200 OK | **PASS** |
| `POST` | `/api/credential-pools` | Create credential pool | 201 Created | **PASS** |
| `POST` | `/api/credential-pools/{id}/pause` | Pause pool traffic | 200 OK | **PASS** |
| `POST` | `/api/credential-pools/{id}/resume` | Resume pool traffic | 200 OK | **PASS** |
| `GET` | `/api/free-tiers` | Query free tier directory | 200 OK | **PASS** |
| `GET` | `/api/devtools/templates` | List pre-configured templates | 200 OK | **PASS** |
| `POST` | `/api/devtools/generate-curl` | Generate executable cURL command | 200 OK | **PASS** |
| `GET` | `/api/backup` | List database backups | 200 OK | **PASS** |
| `POST` | `/api/backup` | Trigger online SQLite backup | 200 OK | **PASS** |
