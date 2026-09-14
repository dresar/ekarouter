# EkaRouter OAuth 2.0 PKCE Integration API

## 1. Overview

EkaRouter supports OAuth 2.0 authorization code flows secured with **PKCE (Proof Key for Code Exchange, RFC 7636)** for external developer platforms (such as GitHub, Google Cloud, Anthropic Workspaces, Slack, and Discord).

This allows administrators and platform developers to connect third-party accounts without ever manually copying and pasting long-lived client secrets into plain configuration files.

---

## 2. OAuth Workflow Architecture

```
User (Browser)               EkaRouter Backend              Third-Party OAuth Provider
     |                               |                                      |
     |-- 1. POST /oauth/start ------>|                                      |
     |   (provider_id: github)       |-- Generates PKCE code_verifier       |
     |                               |   and cryptographically random state |
     |<-- 2. Returns Auth URL, ------|                                      |
     |       State & Redirect URI    |                                      |
     |                               |                                      |
     |-- 3. Redirects to Auth URL ----------------------------------------->|
     |                                                                      |-- User grants consent
     |<-- 4. Redirects to Frontend with Code & State -----------------------|
     |                                                                      |
     |-- 5. POST /oauth/callback --->|                                      |
     |   (code, state)               |-- Validates state in DB              |
     |                               |-- 6. POST /token (code + verifier) ->|
     |                               |<-- 7. Returns Access & Refresh Token-|
     |                               |-- Encrypts tokens in Vault (AES-256) |
     |<-- 8. Connection Saved (200) -|                                      |
```

---

## 3. Endpoints

### 3.1 Start OAuth Flow: `POST /api/accounts/oauth/start`
Initializes a new PKCE session and generates the authorization URL.

- **URL**: `/api/accounts/oauth/start`
- **Method**: `POST`
- **Authentication**: Admin Session Cookie or Platform Token
- **Request Body**:
```json
{
  "provider_id": "github",
  "redirect_uri": "http://localhost:3000/oauth/callback",
  "scopes": ["read:user", "repo"]
}
```
- **Response Body (200 OK)**:
```json
{
  "success": true,
  "data": {
    "provider_id": "github",
    "state": "st_9b7c8d9e2f1a4c5b6e7d8c9b",
    "auth_url": "https://github.com/login/oauth/authorize?client_id=12345&state=st_9b7c8d9e2f1a4c5b6e7d8c9b&redirect_uri=http%3A%2F%2Flocalhost%3A3000%2Foauth%2Fcallback&scope=read%3Auser+repo&code_challenge=xyz789...&code_challenge_method=S256"
  }
}
```

### 3.2 Complete OAuth Callback: `POST /api/accounts/oauth/callback`
Exchanges the authorization code and stored PKCE code verifier for OAuth tokens, storing them securely in the credential vault.

- **URL**: `/api/accounts/oauth/callback`
- **Method**: `POST`
- **Authentication**: Admin Session Cookie or Platform Token
- **Request Body**:
```json
{
  "provider_id": "github",
  "state": "st_9b7c8d9e2f1a4c5b6e7d8c9b",
  "code": "8e3b2f1a0c4d5e6f"
}
```
- **Response Body (200 OK)**:
```json
{
  "success": true,
  "data": {
    "connection_id": "conn_01j8k9m012",
    "provider_id": "github",
    "account_name": "developer@company.org",
    "scopes": ["read:user", "repo"],
    "expires_at": "2026-09-15T12:55:00Z",
    "status": "connected"
  }
}
```

---

## 4. Background Token Auto-Refresh

EkaRouter includes an integrated background scheduler that:
1. Scans `oauth_connections` for access tokens expiring within the next 10 minutes.
2. Uses the stored, encrypted `refresh_token` to perform token refresh with the upstream provider.
3. Automatically updates the encrypted token in the vault without interrupting in-flight API gateway traffic or proxy operations.
