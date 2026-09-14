# Database Mapping Report: Reference Schema to EkaRouter Schema

## 1. Context & Scope

This report maps the data structures observed in the 9Router reference database (`9router/src/lib/db/schema.js`) to the clean, normalized, and secured SQLite schema designed for EkaRouter as specified in `EkaRouter_Backend_PRD/09_DATABASE.md`.

## 2. Table-by-Table Mapping

| 9Router Reference Table | EkaRouter Target Table | Mapping / Transformation | Security & Design Notes |
| :--- | :--- | :--- | :--- |
| `settings` (JSON blob) | `settings` (key, value, updated_at) | Normalized key-value pairs | Clear scalar/JSON fields with specific typed keys |
| `providerConnections` | `providers`, `accounts`, `credentials` | Split monolith connection into normalized provider configuration, account entity, and encrypted credentials | In 9Router, `data` stored plaintext keys. In EkaRouter, `credentials` stores AES-GCM encrypted secrets |
| `providerNodes` | `providers` | Merged into `providers` table (`kind = 'custom'`) | Simplifies architecture; avoids redundant node table |
| `proxyPools` | `proxy_profiles` | Direct mapping with schema normalization | Password stored encrypted (`encrypted_password`) |
| `apiKeys` | `api_keys` | Transformed from plaintext key to SHA-256 hash + key prefix | 9Router stored raw keys; EkaRouter stores only cryptographic hashes |
| `combos` | `routes`, `route_items` | Normalized from JSON array to relational route + route_items table | Enables SQL queries for route item ordering, priorities, and weights |
| `usageHistory` | `usage_logs` | Direct mapping with typed columns | Stores compact tokens, latency, status, timestamps |
| `requestDetails` | `request_logs` | Direct mapping with metadata only | Request bodies are NOT stored by default to preserve RAM/disk |
| `_meta` | `_migrations` | Schema migration tracking table | Versioned integer migration steps |
| New in EkaRouter | `sessions` | Session tokens for dashboard management API | Stored as hashed tokens with expiration |
| New in EkaRouter | `quota_snapshots` | Periodic/cached quota snapshots per account | Normalized remaining requests/tokens |
| New in EkaRouter | `oauth_states` | Short-lived OAuth 2.0 PKCE state records | Ephemeral table with expiry and cryptographic verifiers |

## 3. Security Classification

1. `credentials.encrypted_access`, `credentials.encrypted_refresh`, `credentials.encrypted_secret`: **CRITICAL SECRET**. Must be encrypted at rest using a master key derived from machine ID or `EKAROUTER_SECRET_KEY` via AES-256-GCM.
2. `api_keys.hash`: **ONE-WAY HASH**. SHA-256 representation of the raw Bearer token `eka_live_...`.
3. `sessions.token_hash`: **ONE-WAY HASH**. SHA-256 representation of the session cookie.
4. `proxy_profiles.encrypted_password`: **SECRET**. Encrypted proxy auth.
