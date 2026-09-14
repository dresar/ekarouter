# 02. Backend Technical Audit & Stack Verification

---

## 1. Runtime & Technology Stack

- **Language & Runtime**: Go (`go1.24+`)
- **HTTP Routing Framework**: `github.com/go-chi/chi/v5` (v5.3.2)
- **Database Engine**: `modernc.org/sqlite` (v1.58.0) — 100% Pure Go, zero CGO dependencies.
- **Database Concurrency**: Journal Mode `WAL` (Write-Ahead Logging), `busy_timeout = 5000ms`, `foreign_keys = ON`.
- **Encryption Library**: Standard Go `crypto/aes`, `crypto/cipher` (AES-256-GCM authenticated cipher).
- **Password Hashing**: `golang.org/x/crypto/pbkdf2` with HMAC-SHA256, 100,000 iterations, 32-byte salt.
- **UUID Generation**: `github.com/google/uuid` (v1.6.0).

---

## 2. Default Configuration & Networking

| Setting | Default Value | Environment Variable | Notes |
| :--- | :--- | :--- | :--- |
| **Port** | `8080` | `PORT` | Listening interface `0.0.0.0:8080` |
| **Database Path** | `data/ekarouter.db` | `DATABASE_PATH` | Automatically created on startup |
| **Admin User** | `admin` | `ADMIN_USER` | Initial administrative credential |
| **Admin Password** | `admin12345` | `ADMIN_PASSWORD` | Must be updated in production |
| **Secret Key** | `12345678901234567890123456789012` | `SECRET_KEY` | 32-character AES-GCM master key |
| **CORS Origins** | `*` | `CORS_ALLOWED_ORIGINS` | Configurable comma-separated origins |
| **Log Level** | `info` | `LOG_LEVEL` | `debug`, `info`, `warn`, `error` |

---

## 3. Database Schema Inventory

The backend database migrations live in `migrations/*.sql` and execute automatically in numeric order on startup:

### 1. `0001_initial.sql`
- `schema_migrations`: Version tracking table.
- `settings`: Global system settings (`key`, `value`, `updated_at`).
- `providers`: AI providers (`id`, `key`, `name`, `kind`, `base_url`, `enabled`).
- `accounts`: Provider accounts (`id`, `provider_id`, `name`, `auth_type`, `state`, `priority`, `enabled`).
- `credentials`: Encrypted account credentials (`id`, `account_id`, `encrypted_access`, `encrypted_secret`).
- `models`: Model catalog entries (`id`, `provider_id`, `external_name`, `display_name`, `context_limit`, `streaming`).
- `routes`: Routing combos (`id`, `name`, `strategy`, `enabled`).
- `route_items`: Targets per route (`id`, `route_id`, `provider_id`, `account_id`, `model_id`, `priority`, `weight`).
- `api_keys`: Ingress gateway client tokens (`id`, `name`, `prefix`, `hash`, `scopes`, `enabled`).
- `proxy_profiles`: Outbound proxies (`id`, `name`, `scheme`, `host`, `port`, `username`, `encrypted_password`).
- `usage_logs`: Real-time request and token consumption logs (`id`, `route_id`, `provider_id`, `model_id`, `prompt_tokens`, `completion_tokens`).

### 2. `0002_credential_pools.sql`
- `credential_pools`: High-availability pools (`id`, `name`, `owner_id`, `provider_id`, `environment`, `status`).
- `pool_members`: Accounts attached to pools (`id`, `pool_id`, `account_id`, `priority`, `weight`, `status`, `cooldown_until`).
- `rotation_policies`: Pool rotation strategies (`id`, `pool_id`, `strategy`, `cooldown_seconds`, `max_retries`).
- `credential_health_checks`: Health check logs per member (`id`, `member_id`, `pool_id`, `status`, `latency_ms`).
- `credential_events`: Pool state transitions (`id`, `pool_id`, `event_type`, `details`).
- `rotation_decisions`: Rotation audit log (`id`, `pool_id`, `strategy`, `selected_member_id`, `reason`).
- `provider_templates`: Developer templates (`id`, `provider_name`, `category`, `auth_type`, `base_url`).
- `request_templates`: Pre-built HTTP execution definitions (`id`, `name`, `method`, `path`, `headers`, `body_schema`).
- `rate_limit_configs`: Rate limiter configuration (`id`, `reference_type`, `reference_id`, `limit_value`, `window_seconds`).
- `quota_records`: Quota consumption counters (`id`, `reference_type`, `reference_id`, `metric`, `used_value`, `limit_value`).
- `free_tier_catalog`: Discovered free tier offerings (`id`, `provider_id`, `category`, `free_tier_description`, `verified`).

### 3. `0003_developer_platform.sql`
- `vault_credentials`: Enterprise credential vault (`id`, `name`, `credential_type`, `provider_id`, `environment`, `encrypted_value`, `masked_value`, `priority`, `tags`, `status`, `health_state`).
- `projects`: Developer workspace containers (`id`, `name`, `description`, `owner_id`).
- `environments`: Deployment stages (`id`, `project_id`, `name`, `slug`).
- `tool_definitions`: Generic parameter-driven tools (`id`, `provider_id`, `name`, `method`, `url_template`, `timeout_ms`).
- `tool_executions`: Tool execution log (`id`, `tool_id`, `status`, `status_code`, `latency_ms`, `error_message`).
- `webhooks`: Inbound/outbound webhook subscriptions (`id`, `name`, `url`, `event_types`, `secret_hash`, `status`).
- `webhook_deliveries`: Delivery attempt logs (`id`, `webhook_id`, `status_code`, `latency_ms`, `response_snippet`).
- `audit_records`: Immutable compliance logs (`id`, `actor_id`, `action`, `resource_type`, `resource_id`, `result`).
- `rbac_users`: Platform users (`id`, `username`, `password_hash`, `role`, `status`).
- `rbac_tokens`: Scoped personal access tokens (`id`, `user_id`, `name`, `token_prefix`, `token_hash`, `scopes`).

### 4. `0004_platform_extensions.sql`
- `credential_assignments`: Explicit project-environment mapping for credentials.
- `credential_permissions`: Fine-grained credential access rights.
- `credential_health`: Live health status and latency logs for vault credentials.
- `credential_usage`: Aggregate request and token usage counters for vault credentials.
- `api_requests`: Outbound reverse proxy and tool execution tracking.
- `api_request_attempts`: Retry and failover attempt logging.
- `api_request_logs`: Request and response header/body inspection snippets.
- `usage_records`: Partitioned usage counters by reference and date.
- `routing_rules`: Conditional routing expressions.
- `tool_actions`: Parameterized actions attached to tool definitions.
- `openapi_documents`: Imported OpenAPI 3.x specifications and operation metadata.
- `notifications`: Administrative notification alerts.
- `system_settings`: Key-value configuration entries.

---

## 4. Local Execution & Testing Verification

### Starting Server Locally
```powershell
go run ./cmd/ekarouter serve
```

### Running Test Suite
```powershell
go test -v -count=1 ./...
```
Verification confirmed: **All 30+ internal packages pass with exit code 0.**
