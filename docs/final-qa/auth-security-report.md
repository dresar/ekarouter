# EkaRouter Authentication & Access Control Report

**Date:** 2026-09-15  
**Version:** 1.0.0  
**Scope:** Session Management, Token Hashing, RBAC, Privilege Separation

---

## 1. Authentication Schemes

EkaRouter implements three distinct authentication mechanisms:

1. **Administrator Sessions (`SessionAuthMiddleware`)**:
   - Cookie-based (`session_token`) with `HttpOnly`, `SameSite=Lax`, and 24-hour expiration.
   - Also accepts `Authorization: Bearer <session_token>`.
   - Stored in SQLite as SHA-256 hash in `sessions` table.
   - Revocation explicitly sets `revoked_at = CURRENT_TIMESTAMP`.

2. **Client Developer Tokens (`PlatformAuthMiddleware`)**:
   - Bound to specific user IDs and project IDs in `client_tokens` table.
   - Scoped to project-level resources (`/api/v1/*`).
   - Strictly prohibited from accessing administrative routes (`/api/settings`, `/api/credential-pools`, etc.).

3. **Gateway API Keys (`GatewayAuthMiddleware`)**:
   - Used for inference on `/v1/chat/completions` and `/v1/models`.
   - Stored as SHA-256 hashes in `api_keys` table.
   - Updates `last_used_at` timestamp asynchronously.

---

## 2. RBAC Roles and Scopes

| Role | Scope / Permitted Paths | Excluded Paths |
|---|---|---|
| **Admin** | Full access to `/api/*`, `/api/v1/*`, `/v1/*` | None |
| **Developer** | Scoped access to `/api/v1/projects`, `/api/v1/credentials`, `/api/v1/tools` | `/api/settings`, `/api/keys`, `/api/credential-pools` |
| **Gateway Client** | Inference access to `/v1/chat/completions`, `/v1/models` | All `/api/*` and `/api/v1/*` management routes |
| **Unauthenticated** | `/health`, `/live`, `/ready`, `POST /api/auth/login` | All protected endpoints |
