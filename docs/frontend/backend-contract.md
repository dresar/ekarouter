# EkaRouter Backend-to-Frontend Contract & Type Specifications

**Target Audience**: Frontend Engineers, Full-stack Developers, UI Designers  
**Specification Version**: 1.0.0 (Release Candidate)  
**Backend Status**: Feature-Complete & Quality Assured  

---

## 1. System Topology & UI View Mapping

| Frontend View / Page | Primary Endpoints Called | Data Entities | Operations Supported |
|---|---|---|---|
| **Overview Dashboard** | `GET /api/v1/health`<br>`GET /api/v1/usage/summary`<br>`GET /api/v1/audit-logs?limit=10` | Health, Usage, Events | Real-time status cards, token charts, recent activity feed |
| **Credential Vault** | `GET /api/v1/credentials`<br>`POST /api/v1/credentials`<br>`PATCH /api/v1/credentials/{id}`<br>`POST /api/v1/credentials/{id}/rotate`<br>`DELETE /api/v1/credentials/{id}` | `Credential`, `SecretVersion` | Create, mask view, test connectivity, enable/disable, rotate, delete |
| **Provider Catalog** | `GET /api/v1/providers`<br>`GET /api/v1/providers/{id}/capabilities`<br>`POST /api/v1/providers/{id}/health` | `Provider`, `Capability` | Search 27+ providers, view documentation links, run health test |
| **Model Routing & Fallbacks** | `GET /api/routes`<br>`POST /api/routes`<br>`DELETE /api/routes/{id}`<br>`GET /api/models` | `Route`, `RouteItem`, `Model` | Configure priority order, weights, fallback chains, cooldown thresholds |
| **Credential Pools** | `GET /api/credential-pools`<br>`POST /api/credential-pools`<br>`POST /api/credential-pools/{id}/pause`<br>`POST /api/credential-pools/{id}/resume`<br>`POST /api/credential-pools/{id}/rotate` | `CredentialPool`, `PoolMember` | Pool lifecycle, rotation strategies (`least_used`, `priority`, `round_robin`) |
| **Outbound Proxies** | `GET /api/proxies`<br>`POST /api/proxies`<br>`DELETE /api/proxies/{id}`<br>`POST /api/proxies/{id}/test` | `ProxyProfile` | Manage HTTP/SOCKS5 proxies, test latency, mask auth passwords |
| **Client API Keys** | `GET /api/keys`<br>`POST /api/keys`<br>`DELETE /api/keys/{id}` | `APIKey` | Generate developer keys (`eka_live_*`), set rate limits, revoke keys |
| **Free Tier Directory** | `GET /api/free-tiers`<br>`GET /api/free-tiers/categories`<br>`GET /api/free-tiers/verified` | `FreeTier` | Filter zero-cost AI & platform quotas, copy setup snippets |
| **DevTools & OpenAPI** | `GET /api/devtools/templates`<br>`POST /api/devtools/generate-curl`<br>`POST /api/devtools/import-openapi`<br>`POST /api/devtools/confirm-openapi` | `RequestTemplate`, `OpenAPISpec` | Import Swagger/OpenAPI 3.x, preview operations, auto-generate cURL |
| **Audit Logs** | `GET /api/v1/audit-logs` | `AuditLog` | Paginated event ledger, action filtering, timestamps |
| **AI Playground** | `POST /v1/chat/completions`<br>`GET /v1/models` | `ChatMessage`, `CompletionChunk` | Interactive prompt test with streaming SSE & TokenSaver toggles |

---

## 2. TypeScript Data Interfaces

### 2.1 Core Entities

```typescript
export type CredentialStatus = 'active' | 'cooldown' | 'disabled' | 'expired';

export interface Credential {
  id: string;
  name: string;
  provider_id: string;
  masked_value: string; // e.g. "sk-proj-***4b" - NEVER plaintext
  environment: 'production' | 'staging' | 'development';
  project_id?: string;
  status: CredentialStatus;
  priority: number; // Lower number = higher priority (1 is highest)
  weight: number;
  consecutive_failures: number;
  is_in_cooldown: boolean;
  cooldown_expires_at: string | null;
  last_used_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface ProviderCapability {
  supports_streaming: boolean;
  supports_tools: boolean;
  supports_vision: boolean;
  supports_json_mode: boolean;
  max_context_window: number;
}

export interface Provider {
  id: string;
  name: string;
  category: 'ai' | 'developer' | 'cloud' | 'custom';
  base_url: string;
  auth_type: 'bearer' | 'basic' | 'custom_header' | 'oauth2';
  auth_header?: string;
  doc_url: string;
  capabilities: ProviderCapability;
  is_system: boolean;
  created_at: string;
}

export interface RouteItem {
  id: string;
  route_id: string;
  provider_id: string;
  account_id: string;
  priority: number; // 1 = Primary, 2 = Secondary, etc.
  weight: number;
  model_mapping?: string; // e.g. "claude-3-5-sonnet" -> "openai/gpt-4o"
  is_active: boolean;
}

export interface Route {
  id: string;
  model_pattern: string; // e.g. "gpt-4o*", "claude-*"
  strategy: 'priority' | 'round_robin' | 'least_used' | 'random' | 'weighted';
  timeout_seconds: number;
  retry_count: number;
  items: RouteItem[];
  created_at: string;
}

export interface ProxyProfile {
  id: string;
  name: string;
  protocol: 'http' | 'https' | 'socks5';
  host: string;
  port: number;
  username?: string;
  password?: string; // MASKED in GET responses: "********"
  is_active: boolean;
  last_latency_ms?: number;
  created_at: string;
}

export interface APIKey {
  id: string;
  name: string;
  prefix: string; // e.g. "eka_live_ab12"
  project_id?: string;
  rate_limit_rpm: number;
  rate_limit_tpm: number;
  expires_at: string | null;
  last_used_at: string | null;
  is_active: boolean;
  created_at: string;
}

export interface AuditLog {
  id: string;
  actor: string;
  action: string; // e.g. "credential.created", "route.deleted"
  entity_type: string;
  entity_id: string;
  metadata: Record<string, any>;
  ip_address: string;
  timestamp: string;
}

export interface FreeTier {
  id: string;
  provider_name: string;
  category: string;
  model_name: string;
  free_quota_description: string;
  rpm_limit: number;
  tpm_limit: number;
  requires_credit_card: boolean;
  verified: boolean;
  documentation_url: string;
}
```

---

## 3. Credential State Transition Model

Frontend components must handle the four distinct lifecycle states of credentials:

```
          [Create / Enable]
                |
                v
        +---------------+
+-----> |    active     | <-----+
|       +---------------+       |
|               |               |
|      (Consecutive 429/5xx)    | (Cooldown Timer Expires
|               |               |  & Health Check Passes)
|               v               |
|       +---------------+       |
|       |   cooldown    | ------+
|       +---------------+
|               |
|         (Manual Disable)
|               |
|               v
|       +---------------+
+------ |   disabled    |
        +---------------+
```

---

## 4. Input Validation & Form Constraints

Frontend forms must enforce the following validation limits before submitting to the API:

| Form / Field | Requirement | Min | Max | Pattern / Regex | Notes |
|---|---|---|---|---|---|
| **Login Username** | Required | 3 | 50 | `^[a-zA-Z0-9_-]+$` | Alphanumeric only |
| **Login Password** | Required | 8 | 128 | - | Minimum 8 characters |
| **Credential Name** | Required | 3 | 100 | - | Human readable label |
| **Secret Value** | Required on Create | 8 | 4096 | - | Never pre-populated on Edit |
| **Priority** | Required | 1 | 1000 | Integer | 1 is highest priority |
| **Proxy Host** | Required | 1 | 255 | Hostname / IP | Blocked: loopback, RFC1918 |
| **Proxy Port** | Required | 1 | 65535 | Integer | Valid TCP port |
| **Route Model** | Required | 1 | 128 | `^[a-zA-Z0-9_.*-]+$` | Wildcards supported (`gpt-*`) |
| **API Key Name** | Required | 3 | 100 | - | Team or application name |

---

## 5. Standard Error Handling UI Contract

All non-2xx responses deliver structured errors. The frontend must parse errors as follows:

```typescript
export interface BackendErrorResponse {
  success?: false;
  error: string | {
    message: string;
    type?: string;
    code?: string;
    param?: string;
  };
}

export function extractErrorMessage(payload: BackendErrorResponse): string {
  if (typeof payload.error === 'string') {
    return payload.error;
  }
  if (payload.error && typeof payload.error.message === 'string') {
    return payload.error.message;
  }
  return 'Unknown error occurred';
}
```
