# Authentication Reference

EkaRouter supports multiple authentication methods depending on the endpoint:

## 1. Web Session Authentication (`/api/*`)
Administrators authenticate via `POST /api/auth/login`:
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin12345"}'
```
Returns a `session_token` cookie and JSON payload. Subsequent requests pass either the cookie or `Authorization: Bearer <session_token>`.

## 2. Developer Personal Access Tokens (`eka_pat_*`)
Developers and automated CI/CD systems authenticate against `/api/v1/*` endpoints using Client Tokens:
```bash
curl -X GET http://localhost:8080/api/v1/providers \
  -H "Authorization: Bearer eka_pat_1234567890abcdef..."
```

## 3. AI Gateway API Keys (`eka_live_*`)
AI generation requests to `/v1/chat/completions` use Gateway API keys:
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer eka_live_abcdef1234567890..." \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-3-7-sonnet","messages":[{"role":"user","content":"ping"}]}'
```
