# EkaRouter Database & Migration Report

**Date:** 2026-09-15  
**Engine:** SQLite via `modernc.org/sqlite` (Pure Go, Zero CGO)  
**Storage Mode:** WAL (Write-Ahead Logging), Foreign Keys Enabled, Synchronous Normal

---

## 1. Migration History & Idempotency

| Version | Migration File | Tables Created / Modified | Idempotent |
|---|---|---|---|
| **0001** | `0001_initial.sql` | `schema_migrations`, `settings`, `providers`, `accounts`, `credentials`, `models`, `routes`, `route_items`, `api_keys`, `sessions`, `proxy_profiles`, `usage_logs`, `request_logs`, `quota_snapshots`, `oauth_states` | **YES** |
| **0002** | `0002_credential_pools.sql` | `free_tier_catalog`, `free_tier_sources`, `credential_pools`, `pool_members`, `rotation_policies`, `pool_health_events`, `credential_usage_snapshots`, `credential_audit_logs` | **YES** |
| **0003** | `0003_developer_platform.sql` | `users`, `teams`, `team_members`, `environments`, `roles`, `permissions`, `role_permissions`, `projects`, `vault_credentials`, `credential_versions`, `client_tokens`, `tools`, `tool_executions`, `audit_events`, `webhooks`, `webhook_deliveries`, `rate_limits`, `request_quotas` | **YES** |
| **0004** | `0004_platform_extensions.sql` | `credential_assignments`, `credential_permissions`, `credential_health`, `credential_usage`, `api_requests`, `api_request_attempts`, `api_request_logs`, `usage_records`, `routing_rules`, `tool_actions`, `openapi_documents`, `notifications`, `system_settings` | **YES** |

---

## 2. Foreign Key Integrity & Cascades

Verified in `TestDatabaseMigrationsAndIntegrity/Foreign_key_cascade_behavior`:
- `providers` -> `accounts` (ON DELETE CASCADE)
- `accounts` -> `credentials` (ON DELETE CASCADE)
- `routes` -> `route_items` (ON DELETE CASCADE)
- `projects` -> `environments` (ON DELETE CASCADE)
- `credential_pools` -> `pool_members` (ON DELETE CASCADE)

When a parent entity is deleted, SQLite cascading rules cleanly delete all children records without orphan rows remaining.

---

## 3. Concurrency & WAL Performance

The database connection pool is configured with:
- `db.SetMaxOpenConns(10)`
- `db.SetMaxIdleConns(5)`
- `db.SetConnMaxLifetime(time.Hour)`
- Connection DSN includes `_pragma=busy_timeout(5000)` and `_pragma=journal_mode(WAL)`.

This configuration completely prevents `database is locked` errors during concurrent read/write operations from worker pools.
