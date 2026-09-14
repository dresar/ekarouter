# 03. Backend Feature Inventory & Verification Matrix

---

## 1. Feature Verification Matrix

This matrix maps every intended feature to its actual backend implementation state.

- `PASS`: Verified working in backend code, covered by tests, and exposed via API.
- `NEEDS_BACKEND_SUPPORT`: Backend data model or stub exists, but endpoint requires polish or query parameters.
- `NOT_IMPLEMENTED`: Feature requested in high-level concepts but completely absent in Go backend code.
- `BLOCKED`: Feature cannot function without external third-party services.

| Feature Area | Specific Capability | Status | Backend Location | Frontend Handling |
| :--- | :--- | :--- | :--- | :--- |
| **Authentication** | Username / Password Login | `PASS` | `internal/httpapi/admin_auth.go` | Form login to `/api/auth/login` |
| **Authentication** | JWT Bearer & Cookie Session | `PASS` | `internal/httpapi/admin_auth.go` | Store token in memory / cookie |
| **Authentication** | Current User Profile (`/me`) | `PASS` | `internal/httpapi/admin_auth.go` | Topbar user badge |
| **AI Gateway** | OpenAI Chat (`/v1/chat/completions`) | `PASS` | `internal/gateway/gateway.go` | API playground / status check |
| **AI Gateway** | SSE Streaming Completions | `PASS` | `internal/gateway/gateway.go` | EventSource / Fetch reader |
| **AI Gateway** | Model Catalog (`/v1/models`) | `PASS` | `internal/gateway/gateway.go` | Model selection dropdowns |
| **AI Gateway** | Anthropic Ingress (`/v1/messages`) | `NOT_IMPLEMENTED` | - | Omit from frontend or mark planned |
| **AI Gateway** | Vector Embeddings (`/v1/embeddings`) | `NOT_IMPLEMENTED` | - | Omit from frontend |
| **AI Providers** | List, Create, Delete Providers | `PASS` | `internal/httpapi/admin_entities.go` | Dedicated Provider Management page |
| **AI Accounts** | Account CRUD & AES Credentials | `PASS` | `internal/httpapi/admin_entities.go` | Account list & create drawer/page |
| **AI Accounts** | OAuth PKCE Start & Callback | `PASS` | `internal/oauth/oauth.go` | OAuth connect button & redirect |
| **AI Routes** | Combo / Route CRUD & Ordering | `PASS` | `internal/httpapi/admin_entities.go` | Routing list & visual target list |
| **AI Routes** | Multi-Strategy Rotation | `PASS` | `internal/routing/router.go` | Strategy picker (Priority, RR) |
| **AI Routes** | Auto-Failover (Auth / Quota 429) | `PASS` | `internal/gateway/gateway.go` | Cooldown indicators & badges |
| **Gateway Keys** | Ingress API Key Generation & Revoke | `PASS` | `internal/httpapi/admin_entities.go` | API Key Management table |
| **Outbound Proxies** | Proxy Profile CRUD | `PASS` | `internal/httpapi/admin_entities.go` | Proxy table & edit page |
| **Outbound Proxies** | Live Proxy Connectivity Testing | `PASS` | `internal/proxy/proxy.go` | "Test Connection" button |
| **Token Saver** | Text Compaction Preview | `PASS` | `internal/tokensaver/tokensaver.go` | Interactive comparison playground |
| **Database Backup** | Online SQLite VACUUM Backup | `PASS` | `internal/httpapi/admin_entities.go` | Trigger backup button & file list |
| **Platform Vault** | AES-256-GCM Credential Storage | `PASS` | `internal/vault/vault.go` | Credential Vault page |
| **Platform Vault** | Masked Display (`sk-****abcd`) | `PASS` | `internal/vault/vault.go` | Masked token viewer with copy |
| **Platform Vault** | Credential Connectivity Testing | `PASS` | `internal/platform/adapters.go` | Individual credential test action |
| **Developer Projects**| Project & Environment CRUD | `PASS` | `internal/httpapi/admin_platform.go` | Workspace / project filter |
| **Generic Tools** | Parameterized Tool Definition | `PASS` | `internal/httpapi/admin_platform.go` | Tool catalog & schema viewer |
| **Generic Tools** | Safe Tool Execution & SSRF Guard | `PASS` | `internal/executor/executor.go` | Tool execution test panel |
| **Request Templates**| HTTP Template CRUD & Execution | `PASS` | `internal/httpapi/admin_platform.go` | Request Template manager |
| **Platform Usage** | Aggregate Summary & Breakdown | `PASS` | `internal/httpapi/admin_platform.go` | Usage overview charts & metrics |
| **System Health** | Liveness & Readiness Probes | `PASS` | `internal/health/checker.go` | Status indicator pills in topbar |
| **Audit Logs** | Immutable Log Querying | `PASS` | `internal/httpapi/admin_platform.go` | Audit Log data table |
| **Free Tier Catalog**| Browse & Filter Free Tiers | `PASS` | `internal/freetier/catalog.go` | Community Free Tier directory |
| **OpenAPI Import** | Document Preview & Schema Confirm | `PASS` | `internal/devtools/templates.go` | OpenAPI spec import flow |
| **Billing / Cost** | Exact Monetary Invoicing | `NOT_IMPLEMENTED` | - | Display estimated tokens only |
| **WebSocket Logs** | Live WebSocket Streaming Logs | `NOT_IMPLEMENTED` | - | Use polling (5s - 10s) |

---

## 2. Invariants for Frontend Construction

1. **No Estimated Cost as Actual Billing**: The backend tracks token counts (`prompt_tokens`, `completion_tokens`). Any monetary cost calculations on the frontend must be explicitly labeled **"Estimated"** and calculated via client-side multipliers.
2. **Polling Over WebSockets**: The backend currently serves standard HTTP JSON endpoints and SSE for `/v1/chat/completions`. Real-time telemetry (health, active cooldowns, recent logs) must use configurable polling (e.g., SWR / React Query interval) rather than attempting WebSocket connections.
3. **No Unimplemented Form Controls**: Do not add UI inputs for Anthropic `/v1/messages` direct pass-through or Vector embeddings until the backend exposes those routes.
