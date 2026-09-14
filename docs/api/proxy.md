# EkaRouter AI & Outbound Proxy API

## 1. AI Gateway Proxy: `POST /v1/chat/completions`

### Non-Streaming Request:
```bash
curl -X POST https://api.ekarouter.dev/v1/chat/completions \
  -H "Authorization: Bearer eka_live_token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Hello!"}
    ],
    "stream": false
  }'
```

### Streaming Request (`stream: true`):
Returns an event stream (`Content-Type: text/event-stream`):
```text
data: {"id":"req_123","object":"chat.completion.chunk","choices":[{"delta":{"content":"Hello"}}]}

data: {"id":"req_123","object":"chat.completion.chunk","choices":[{"delta":{"content":" world!"}}]}

data: [DONE]
```

---

## 2. Universal Developer Platform Proxy: `/api/v1/proxy/{provider}/*`

Allows clients to send requests to any configured developer platform provider without managing upstream secrets on the client:

```bash
curl -X GET https://api.ekarouter.dev/api/v1/proxy/cloudflare/zones \
  -H "Authorization: Bearer session_token_or_client_token"
```
EkaRouter:
1. Validates authentication and permissions.
2. Performs SSRF verification on the destination endpoint.
3. Injects the encrypted credentials stored in the vault.
4. Forwards the request and pipes the upstream response back.

---

## 3. Outbound Proxy Profiles: `/api/proxies`

Used for routing EkaRouter's own outbound traffic through HTTP, HTTPS, or SOCKS5 proxies:
- `GET /api/proxies`: List configured proxies (passwords masked).
- `POST /api/proxies`: Create proxy profile.
- `POST /api/proxies/{id}/test`: Test proxy connectivity.
- `DELETE /api/proxies/{id}`: Remove proxy profile.
