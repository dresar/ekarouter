# Database Architecture & Schema

EkaRouter persists all configuration, credentials, routing policies, tool definitions, logs, and state inside SQLite (`data/ekarouter.db`) using pure Go (`modernc.org/sqlite`), operating with WAL (Write-Ahead Logging) and strict foreign key constraints enabled.

## Migration Structure

Schema versioning is managed via sequential SQL files under `migrations/`:
1. `0001_initial.sql`: Core schema (`providers`, `accounts`, `credentials`, `models`, `routes`, `route_items`, `api_keys`, `sessions`, `proxy_profiles`, `usage_logs`, `request_logs`, `quota_snapshots`, `oauth_states`).
2. `0002_credential_pools.sql`: Credential pools, rotation policies, free-tier catalog, provider templates, request templates, API operations, projects, quota limits.
3. `0003_developer_platform.sql`: Multi-tenant user RBAC, client tokens, vault credentials, credential versions, cooldowns, webhook deliveries, tool definitions, tool executions, scheduled tasks, task runs, and immutable audit logs.
4. `0004_platform_extensions.sql`: Complete developer platform extensions (`credential_assignments`, `credential_permissions`, `credential_health`, `credential_usage`, `api_requests`, `api_request_attempts`, `api_request_logs`, `usage_records`, `routing_rules`, `tool_actions`, `openapi_documents`, `notifications`, `system_settings`, unique index on `quota_records`).

## Core Tables for Developer Platform

### `vault_credentials`
Stores encrypted credentials for third-party developer APIs.
- `id`: UUID primary key.
- `name`: Human-readable label.
- `credential_type`: `api_key`, `bearer_token`, `personal_access_token`, `oauth_token`, `basic_auth`, etc.
- `provider_id`: ID of target provider (e.g., `cloudflare`, `github`, `stripe`).
- `project_id`: Optional foreign key to `projects`.
- `environment`: Deployment environment (`production`, `staging`, `development`).
- `encrypted_value`: Versioned AES-256-GCM ciphertext (`v1:<base64>`).
- `masked_value`: Safe display string (`sk-****abcd`, `ghp_****91xz`).
- `priority`: Integer priority for selection (lower = higher priority).
- `health_state`: `unknown`, `healthy`, `degraded`, `unhealthy`, `expired`.
- `cooldown_until`: Timestamp until which the credential cannot be used.

### `tool_definitions` & `tool_executions`
Manages reusable external API operations with parameterized URL/header/query templates and execution audit logs.

### `audit_logs`
Immutable record of sensitive operations (`credential.create`, `credential.rotate`, `tool.execute`, `login`). Contains actor ID, action, resource, IP address, user-agent, and status.
