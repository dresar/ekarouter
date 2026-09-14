# 05. Master API Contract & Schema Definitions

All JSON responses follow predictable envelope conventions depending on whether the route belongs to the Legacy Admin API (`/api/*`) or the Developer Platform API (`/api/v1/*`).

---

## 1. Global Response Envelopes

### Developer Platform Envelope (`/api/v1/*`)
All `/api/v1/*` routes return standardized success or error wrappers:

#### Success Response
```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "request_id": "req_01h7abc...",
    "total": 42
  }
}
```

#### Error Response
```json
{
  "success": false,
  "error": {
    "code": "invalid_request",
    "message": "Template name is required",
    "details": null
  },
  "meta": {
    "request_id": "req_01h7abc..."
  }
}
```

### Legacy Management Envelope (`/api/*`)
Direct object or array payloads with standard HTTP status codes:
- **Success**: Status `200 OK` or `201 Created` with direct JSON object/array.
- **Error**: Status `400`, `401`, `403`, `404`, `500` with:
```json
{
  "error": "human readable error description"
}
```

---

## 2. Core Entity Schemas

### 1. Provider (`/api/providers`)
```typescript
interface Provider {
  id: string;          // e.g. "prov_openai" or "prov_8f3a1b2c"
  key: string;         // Unique slug, e.g. "openai"
  name: string;        // Human title, e.g. "OpenAI Commercial"
  kind: string;        // Adapter kind: "openai", "anthropic", "gemini", etc.
  base_url: string;    // Base upstream URL: "https://api.openai.com/v1"
  enabled: boolean;    // Active flag
  created_at: string;  // ISO timestamp
  updated_at: string;  // ISO timestamp
}
```

### 2. Account (`/api/accounts`)
```typescript
interface Account {
  id: string;          // e.g. "acc_8f3a1b2c"
  provider_id: string; // Foreign key referencing providers.id
  name: string;        // e.g. "Production Master Key"
  auth_type: string;   // "api_key" | "oauth" | "custom"
  state: string;       // "active" | "cooling_down" | "disabled"
  priority: number;    // Routing priority (1 = highest, 10 = default)
  enabled: boolean;
}
```

### 3. Route & Combos (`/api/routes`)
```typescript
interface RouteItem {
  id: string;
  provider_id: string;
  account_id?: string;
  model_id?: string;
  priority: number;
  weight: number;
  timeout_ms: number;
  max_retries: number;
  enabled: boolean;
}

interface Route {
  id: string;
  name: string;        // Model ingress alias, e.g. "gpt-4o" or "claude-code"
  strategy: string;    // "priority" | "round_robin" | "least_used"
  enabled: boolean;
  item_count: number;
  items: RouteItem[];
}
```

### 4. Vault Credential (`/api/v1/credentials`)
```typescript
interface VaultCredential {
  id: string;
  name: string;
  credential_type: "api_key" | "bearer_token" | "basic_auth" | "oauth2";
  provider_id: string;
  project_id?: string;
  team_id?: string;
  environment: "production" | "staging" | "development";
  masked_value: string;      // "sk-****abcd" (plaintext is NEVER returned)
  priority: number;
  tags: string;              // Comma-separated tags
  status: "active" | "cooling_down" | "disabled" | "revoked";
  health_state: "healthy" | "degraded" | "unhealthy" | "unknown";
  cooldown_until?: string;   // ISO timestamp if currently in backoff
  request_count: number;
  error_count: number;
  last_used_at?: string;
  last_validated_at?: string;
  created_at: string;
}
```

### 5. Outbound Proxy Profile (`/api/proxies`)
```typescript
interface ProxyProfile {
  id: string;
  name: string;
  scheme: "http" | "https" | "socks5" | "relay";
  host: string;
  port: number;
  username?: string;
  enabled: boolean;
}
```

### 6. Generic Tool Definition (`/api/v1/tools`)
```typescript
interface ToolDefinition {
  id: string;
  provider_id: string;
  name: string;
  description: string;
  category: "developer" | "communication" | "monitoring" | "storage" | "payments" | "custom";
  method: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
  url_template: string;      // e.g. "https://api.resend.com/emails"
  timeout_ms: number;
  retry_count: number;
  created_at: string;
}
```

### 7. Ingress Client API Key (`/api/keys`)
```typescript
interface ApiKey {
  id: string;
  name: string;
  prefix: string;            // e.g. "eka_live_8a"
  scopes: string;            // "*" or comma-separated permissions
  enabled: boolean;
  created_at: string;
  last_used_at?: string;
}
```
*(Note: Full secret key is returned ONLY once inside `POST /api/keys` response under `{ "key": "eka_live_8a..." }`).*
