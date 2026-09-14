# EkaRouter Authentication & Sessions

## 1. Overview

EkaRouter uses three authentication schemes depending on the endpoint group:

1. **Admin Console (`/api/*`)**: Session cookie (`session_token`) or `Authorization: Bearer <session_token>`.
2. **Platform APIs (`/api/v1/*`)**: Session cookie, Developer client token, or API key.
3. **AI Gateway (`/v1/*`)**: `Authorization: Bearer <api_key>`.

---

## 2. Administrator Authentication

### Login: `POST /api/auth/login`
```json
{
  "username": "admin",
  "password": "your_admin_password"
}
```
**Response (200 OK)**:
```json
{
  "status": "authenticated",
  "token": "sess_0123456789abcdef",
  "expires_at": "2026-09-16T04:45:00Z"
}
```
*Note*: Also sets an `HttpOnly`, `SameSite=Lax` cookie named `session_token`.

### Check Identity: `GET /api/auth/me`
**Response (200 OK)**:
```json
{
  "user": "admin"
}
```

### Logout: `POST /api/auth/logout`
Revokes the session in the database and clears the browser cookie.
```json
{
  "status": "logged_out"
}
```

---

## 3. Client API Keys

API keys are created by administrators via `POST /api/keys` and used to authenticate AI gateway requests:

```bash
curl -X POST https://api.ekarouter.dev/v1/chat/completions \
  -H "Authorization: Bearer eka_live_1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4o", "messages": [{"role": "user", "content": "Hello"}]}'
```
