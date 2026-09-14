# EkaRouter Deep System Audit: Backend, Rotations, Endpoints & Architectural Gaps

**Date:** 2026-09-15  
**Auditor:** Antigravity AI Platform Architect  
**Workspace:** `c:\Users\NCN0C\Music\ekarouter\ekarouter`  
**Target:** Complete Backend Architecture, Rotation Engine, Endpoint Inventory & Docs

---

## 1. Executive Summary

A comprehensive, end-to-end architectural and code-level audit was conducted across the entire EkaRouter repository. The audit covered all 22 internal packages, database migrations (`0001_initial.sql`, `0002_credential_pools.sql`, `0003_developer_platform.sql`), routing and failover mechanics, credential rotation algorithms, HTTP API endpoints, and existing documentation in `docs/`.

While EkaRouter compiles cleanly and passes all unit tests, the audit identified **critical rotation bugs**, **severe failover cut-offs**, **five completely dormant subsystems** (`vault`, `platform`, `rotator`, `executor`, `limits`), and **schema fragmentation** across three distinct database eras.

---

## 2. Deep Rotation Mechanics Audit

EkaRouter contains three distinct rotation systems operating in parallel:

### 2.1 Gateway & Routing Rotation (`internal/routing` + `internal/gateway`)

| Mechanism | Implementation | Status | Audit Finding |
|---|---|---|---|
| **Route Item Round-Robin** | `router.go:selectFromRoute` | Functional | Uses atomic cursor per route name to cycle candidate items `copy(rotated, eligible[idx:])`. Correctly preserves relative item order. |
| **Route Item Priority** | `router.go:selectFromRoute` | Functional | Stable sort by `item.Priority ASC`. |
| **Wildcard Account Selection** | `router.go:selectFromRoute` | ⚠️ Degraded | When a route item has `AccountID == ""` (wildcard for a provider), matching accounts are **always** sorted strictly by priority. Even if `route.Strategy == StrategyRoundRobin`, accounts inside a wildcard item are never rotated. |
| **Dynamic Model Requests** | `router.go:selectDynamicTargets` | ❌ **BUG (No Rotation)** | For direct model requests (e.g. `gpt-4o`, `claude-3-5-sonnet`) not bound to an explicit route, accounts are sorted strictly by priority (`eligible[i].Priority < eligible[j].Priority`). If multiple accounts have equal priority, Account 1 always takes 100% of the traffic. |
| **Upstream 401/403 Failover** | `gateway.go:117-120` | 🚨 **CRITICAL BUG** | When an upstream account returns `401 Unauthorized` (expired, invalid, or revoked key), `ClassifyHTTPError` assigns `ErrorClassAuth`. In `gateway.go`, `!providers.IsTransient(execErr)` evaluates to `true`, causing the gateway to **immediately abort** (`return nil, execErr`). It **never** attempts the remaining accounts in the route, and fails to mark the account in cooldown! |
| **Cooldown Recovery** | `cooldown.go` | ⚠️ Ephemeral | In-memory map `entries map[string]*entry`. All cooldown state and failure counts are lost upon process restart. |

### 2.2 Credential Pool Rotation (`internal/credpool`)

| Mechanism | Implementation | Status | Audit Finding |
|---|---|---|---|
| **Priority Fallback** | `engine.go:selectPriorityFallback` | Functional | Picks first eligible member sorted by priority. |
| **Round Robin** | `engine.go:selectRoundRobin` | Functional | Uses per-pool cursor with atomic increment. |
| **Weighted Round Robin** | `engine.go:selectWeightedRoundRobin` | ⚠️ Misnamed | Implemented as weighted random (`rand.Intn(totalWeight)`), not deterministic smooth weighted round-robin. |
| **Least Recently Used (LRU)** | `engine.go:selectLRU` | ❌ **Flawed Logic** | If `eligible[0].LastUsedAt == nil` and `eligible[1].LastUsedAt == nil`, the loop immediately breaks on `i=1` and selects `eligible[1]`, skipping `eligible[0]`. |
| **Quota Aware** | `engine.go:selectQuotaAware` | Functional | Calculates score based on `1000 - TotalRequests - (FailureCount * 10)`. |
| **Lowest Failure Rate** | `engine.go:selectLowestFailureRate` | Functional | Calculates `FailureCount / TotalRequests`. |
| **Environment Aware** | `pool.go:StrategyEnvironmentAware` | ❌ **UNIMPLEMENTED** | Constant defined in `pool.go`, but omitted from `switch policy.Strategy` in `engine.go`. Falls back to default priority. |
| **Member Cooldown Recovery** | `pool.go:IsAvailable` vs `store.go:SetMemberCooldown` | 🚨 **CRITICAL BUG (Permanent Trap)** | When `SetMemberCooldown` runs, it updates SQLite `status = 'cooling_down'`. However, `m.IsAvailable()` starts with: `if m.Status != StatusActive && m.Status != StatusValid { return false }`. Because status is `'cooling_down'`, it returns `false` forever! Even after `time.Now().After(*m.CooldownUntil)`, the member is permanently dead because no code ever resets `status` back to `active`. |
| **Failure Classification** | `failure.go:classifyProviderError` | ❌ **BUG** | Sets `ShouldRotate: false` on `ErrorClassAuth`. If an API key in a pool is revoked or invalid, the pool refuses to rotate to other available keys. |

### 2.3 Rotator Subsystem (`internal/rotator`)

| Mechanism | Implementation | Status | Audit Finding |
|---|---|---|---|
| **Strategies** | Priority, RoundRobin, Random, LeastUsed, LowestErrorRate, HealthBased | Functional in unit tests | Implemented over `[]*vault.Credential`. |
| **Global Cursor Leak** | `rotator.go:rrIndex` | ⚠️ Concurrency Defect | Single uint64 counter shared across all pools, providers, and categories. |
| **Integration** | `rotator` package | ❌ **DISCONNECTED** | Not wired to HTTP router, gateway, or CLI. Exists as an isolated library. |

---

## 3. Endpoints & Handlers Audit

### 3.1 Active & Operational Endpoints

```
[System Health]
GET  /health                                 -> Checker.HealthHandler
GET  /ready                                  -> Checker.ReadyHandler

[OpenAI Gateway]
GET  /v1/models                              -> GatewayHandler.ListModels
POST /v1/chat/completions                    -> GatewayHandler.ChatCompletions (Streaming & Non-streaming)
POST /v1/responses                           -> GatewayHandler.Responses

[Admin Management API]
POST /api/auth/login                         -> AdminHandler.Login
POST /api/auth/logout                        -> AdminHandler.Logout
GET  /api/auth/me                            -> AdminHandler.Me
GET  /api/providers                          -> AdminHandler.ListProviders
POST /api/providers                          -> AdminHandler.CreateProvider
DELETE /api/providers/{id}                   -> AdminHandler.DeleteProvider
GET  /api/accounts                           -> AdminHandler.ListAccounts
POST /api/accounts                           -> AdminHandler.CreateAccount
DELETE /api/accounts/{id}                    -> AdminHandler.DeleteAccount
POST /api/accounts/oauth/start               -> AdminHandler.OAuthStart
POST /api/accounts/oauth/callback            -> AdminHandler.OAuthCallback
GET  /api/credentials                        -> AdminHandler.ListCredentials
GET  /api/credentials/{id}                   -> AdminHandler.GetCredential
GET  /api/routes                             -> AdminHandler.ListRoutes
POST /api/routes                             -> AdminHandler.CreateRoute
DELETE /api/routes/{id}                      -> AdminHandler.DeleteRoute
GET  /api/models                             -> AdminHandler.ListModelsAdmin
POST /api/models                             -> AdminHandler.CreateModel
DELETE /api/models/{id}                      -> AdminHandler.DeleteModel
GET  /api/proxies                            -> AdminHandler.ListProxyProfiles
POST /api/proxies                            -> AdminHandler.CreateProxyProfile
DELETE /api/proxies/{id}                     -> AdminHandler.DeleteProxyProfile
POST /api/proxies/{id}/test                  -> AdminHandler.TestProxyProfile
GET  /api/keys                               -> AdminHandler.ListApiKeys
POST /api/keys                               -> AdminHandler.CreateApiKey
DELETE /api/keys/{id}                        -> AdminHandler.DeleteApiKey
GET  /api/settings                           -> AdminHandler.GetSettings
PUT  /api/settings                           -> AdminHandler.UpdateSetting
GET  /api/usage                              -> AdminHandler.GetUsageSummary
POST /api/tokensaver/preview                 -> AdminHandler.PreviewTokenSaver
GET  /api/backup                             -> AdminHandler.ListBackups
POST /api/backup                             -> AdminHandler.CreateBackup

[Credential Pools API]
GET  /api/credential-pools                   -> AdminHandler.ListPools
POST /api/credential-pools                   -> AdminHandler.CreatePool
GET  /api/credential-pools/{id}              -> AdminHandler.GetPool
DELETE /api/credential-pools/{id}           -> AdminHandler.DeletePool
GET  /api/credential-pools/{id}/credentials  -> AdminHandler.ListPoolMembers
POST /api/credential-pools/{id}/credentials  -> AdminHandler.AddPoolMember
DELETE /api/credential-pools/{id}/credentials/{mid} -> AdminHandler.RemovePoolMember
GET  /api/credential-pools/{id}/policy       -> AdminHandler.GetPoolPolicy
PUT  /api/credential-pools/{id}/policy       -> AdminHandler.UpdatePoolPolicy
POST /api/credential-pools/{id}/health       -> AdminHandler.PoolHealthCheck
POST /api/credential-pools/{id}/rotate       -> AdminHandler.PoolRotate
POST /api/credential-pools/{id}/pause        -> AdminHandler.PausePool
POST /api/credential-pools/{id}/resume       -> AdminHandler.ResumePool
GET  /api/credential-pools/{id}/usage        -> AdminHandler.PoolUsage

[Free Tiers API]
GET  /api/free-tiers                         -> AdminHandler.ListFreeTiers
GET  /api/free-tiers/categories              -> AdminHandler.ListFreeTierCategories
GET  /api/free-tiers/verified                -> AdminHandler.ListVerifiedFreeTiers
POST /api/free-tiers/refresh                 -> AdminHandler.RefreshFreeTiers (Stub only)
GET  /api/free-tiers/{id}                    -> AdminHandler.GetFreeTier
GET  /api/free-tiers/{id}/sources            -> AdminHandler.GetFreeTierSources

[DevTools API]
GET  /api/devtools/templates                 -> AdminHandler.ListTemplates
POST /api/devtools/templates                 -> AdminHandler.CreateTemplate
POST /api/devtools/import-openapi            -> AdminHandler.ImportOpenAPI
POST /api/devtools/confirm-openapi           -> AdminHandler.ConfirmOpenAPIImport
POST /api/devtools/generate-curl             -> AdminHandler.GenerateCurl
```

### 3.2 Incomplete & Stub Endpoints

1. **`POST /api/free-tiers/refresh`**:
   - Only marks non-verified database records as `unverified`. Does not execute any network verification or health check against upstream catalog sources.
2. **`POST /api/accounts/oauth/start`**:
   - Generates PKCE code challenge and transaction ID, but returns relative redirect `auth_url: "/oauth/authorize?..."`. There is no `/oauth/authorize` route registered in the HTTP server, causing browser 404.
3. **`POST /api/providers`, `POST /api/accounts`, `POST /api/routes`**:
   - Raw `INSERT` statements fail with SQLite `UNIQUE constraint failed` if an ID already exists. Does not support idempotent updates (`ON CONFLICT DO UPDATE`).

### 3.3 Completely Missing Endpoints (Dormant Subsystems)

The following capabilities have full backend implementations in `internal/` but **zero exposure in `internal/httpapi/router.go`**:

| Subsystem | Underlying Package | Missing HTTP Endpoints |
|---|---|---|
| **Universal Credential Vault** | `internal/vault` | `GET /api/v1/vault/credentials`<br>`POST /api/v1/vault/credentials`<br>`GET /api/v1/vault/credentials/{id}`<br>`DELETE /api/v1/vault/credentials/{id}`<br>`POST /api/v1/vault/credentials/{id}/rotate`<br>`GET /api/v1/vault/credentials/{id}/versions` |
| **Universal Platform Catalog** | `internal/platform` | `GET /api/v1/platform/providers`<br>`GET /api/v1/platform/providers/categories`<br>`GET /api/v1/platform/providers/{id}`<br>`POST /api/v1/platform/providers/{id}/health`<br>`POST /api/v1/platform/providers/{id}/validate` |
| **Tool Execution Engine** | `internal/executor` | `GET /api/v1/tools`<br>`POST /api/v1/tools`<br>`GET /api/v1/tools/{id}`<br>`POST /api/v1/tools/{id}/execute` |
| **Rate Limit & Quota Engine** | `internal/limits` | `GET /api/v1/limits/quotas`<br>`POST /api/v1/limits/quotas`<br>`POST /api/v1/limits/rate-limit/check` |
| **Controlled API Proxy** | Architecture PRD | `ALL /api/v1/proxy/{provider}/*` (Dynamic reverse proxy with secret injection) |
| **Audit Logs** | Schema 0003 | `GET /api/v1/audit-logs` |
| **Native Anthropic Ingress** | PRD Layer 7 | `POST /v1/messages` (Claude native format) |

---

## 4. Database Schema Fragmentation Analysis

The SQLite database contains 3 separate schema layers that do not cross-reference cleanly:

```
┌─────────────────────────────────────────────────────────────┐
│ 0001_initial.sql (AI Gateway Schema)                        │
│ providers -> accounts -> credentials                        │
│ routes -> route_items -> models                             │
│ proxy_profiles, api_keys, sessions, usage_logs              │
└─────────────────────────────────────────────────────────────┘
                               ▲
                               │ Disjoint (different entities)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ 0002_credential_pools.sql (Pool Engine Schema)              │
│ credential_pools -> pool_members (references accounts.id)   │
│ rotation_policies, credential_health_checks                 │
│ free_tier_catalog, provider_templates, request_templates    │
└─────────────────────────────────────────────────────────────┘
                               ▲
                               │ Disjoint (different entities)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ 0003_developer_platform.sql (Universal Vault & Tools Schema)│
│ users -> teams -> team_members                              │
│ vault_credentials -> credential_versions                    │
│ tool_definitions -> tool_executions                         │
│ webhooks, audit_logs, quota_records, rate_limit_records     │
└─────────────────────────────────────────────────────────────┘
```

- **Problem:** When an operator adds an account via `/api/accounts`, it writes to `accounts` and `credentials` (Schema 0001).
- `internal/vault` writes to `vault_credentials` (Schema 0003).
- `internal/credpool` references `accounts.id` (Schema 0001) in `pool_members`.
- `internal/executor` looks up `vault_credentials` (Schema 0003).
- **Result:** The system cannot rotate a credential from the Vault into the AI Gateway without bridging code.

---

## 5. Prioritized Remediation Roadmap

### Phase 1: Critical Bug Fixes (Stability & Failover)
1. **Fix Upstream Auth Failover (`internal/gateway/gateway.go`)**:
   - When upstream returns `ErrorClassAuth`, mark the specific target account in cooldown (`MarkFailure`), log the account error, and **continue the loop** to the next target in the route instead of immediately returning.
2. **Fix Cooldown Trap (`internal/credpool/pool.go`)**:
   - In `Member.IsAvailable()`, allow `StatusCoolingDown` if `m.CooldownUntil != nil && time.Now().After(*m.CooldownUntil)`.
3. **Fix Dynamic Target Balancing (`internal/routing/router.go`)**:
   - Add round-robin cursor support to `selectDynamicTargets` for accounts sharing the same top priority.
4. **Fix LRU Selection (`internal/credpool/engine.go`)**:
   - Properly compare `LastUsedAt` timestamps and handle multiple `nil` cases without premature termination.

### Phase 2: Dormant Subsystem Integration (API Exposure)
1. **Wire `internal/vault` to HTTP API**:
   - Expose `/api/v1/vault/*` endpoints with AES-256-GCM encryption and uniform masking.
2. **Wire `internal/platform` Catalog**:
   - Expose `/api/v1/platform/*` endpoints for listing, filtering, and health checking 50+ providers.
3. **Wire `internal/executor` Tools**:
   - Expose `/api/v1/tools` and `/api/v1/tools/{id}/execute` with SSRF validation and parameter interpolation.
4. **Wire `internal/limits` Quota**:
   - Expose `/api/v1/limits/*` endpoints for quota inspection and sliding-window rate checks.

### Phase 3: Schema Reconciliation
1. **Unified Credential Bridge**:
   - Allow `vault_credentials` to serve as upstream credentials for `accounts` and `pool_members`.
