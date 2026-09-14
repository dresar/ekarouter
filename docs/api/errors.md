# EkaRouter Error Handling & Status Codes

## 1. Overview

EkaRouter enforces structured, predictable error responses across all endpoint groups. The format depends on whether the request targets the **AI Gateway (`/v1/*`)** or the **Management / Platform APIs (`/api/*`, `/api/v1/*`)**.

---

## 2. AI Gateway Error Format (`/v1/*`)

To maintain 100% drop-in compatibility with the OpenAI SDK and ecosystem tools (LangChain, LlamaIndex, Cursor, AutoGen), gateway errors adhere strictly to the OpenAI error envelope:

```json
{
  "error": {
    "message": "All provider accounts for model 'gpt-4o' are currently in cooldown",
    "type": "server_error",
    "param": null,
    "code": "all_providers_unavailable"
  }
}
```

### Common Gateway Error Types & Codes

| HTTP Status | Error Type | Error Code | Description |
|---|---|---|---|
| `400` | `invalid_request_error` | `missing_model` | Request body did not specify a valid `model`. |
| `400` | `invalid_request_error` | `invalid_json` | JSON payload cannot be parsed. |
| `401` | `authentication_error` | `invalid_api_key` | Bearer token is missing, malformed, or inactive. |
| `403` | `permission_error` | `key_revoked` | The client API key has been revoked by admin. |
| `404` | `not_found_error` | `model_not_found` | The requested model is not mapped to any active route. |
| `429` | `rate_limit_error` | `rate_limit_exceeded` | Client exceeded allotted RPM/TPM quota. |
| `502` | `upstream_error` | `bad_gateway` | Upstream provider returned invalid HTTP response or connection reset. |
| `503` | `service_unavailable` | `all_providers_unavailable` | All failover targets and rotation credentials are in cooldown. |
| `504` | `gateway_timeout` | `upstream_timeout` | Upstream provider failed to respond within the configured timeout. |

---

## 3. Platform & Management Error Format (`/api/v1/*`, `/api/*`)

Platform APIs return a standardized JSON error object with a descriptive string:

```json
{
  "success": false,
  "error": "foreign key constraint failed: invalid project_id"
}
```

or standard admin JSON:
```json
{
  "error": "invalid username or password"
}
```

---

## 4. Security & Safety Error Codes

When an incoming request triggers an automated security or SSRF filter, the request is rejected immediately before DNS resolution or network socket creation:

### 4.1 SSRF Protection (`403 Forbidden` / `400 Bad Request`)
If a Universal Proxy call (`/api/v1/proxy/...`) or Tool Execution (`/api/v1/tools/.../execute`) attempts to connect to local, private, or loopback IPs:
```json
{
  "success": false,
  "error": "request rejected: destination IP '127.0.0.1' is in restricted CIDR block (SSRF protection)"
}
```
**Restricted Target Blocks**:
- `127.0.0.0/8` (IPv4 loopback)
- `::1/128` (IPv6 loopback)
- `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` (Private RFC 1918 networks)
- `169.254.0.0/16` (Link-local & AWS/GCP cloud metadata endpoints `169.254.169.254`)
- `localhost`, `*.local`, `metadata.google.internal`

### 4.2 Body Limit Protection (`413 Payload Too Large`)
Incoming HTTP request bodies exceeding **10 MB** (`10,485,760 bytes`) are terminated immediately to prevent memory exhaustion DoS:
```json
{
  "error": "request body exceeds maximum limit of 10MB"
}
```

---

## 5. Client Handling Best Practices

1. **Check for `Retry-After` Header**: When receiving a `429 Too Many Requests` or `503 Service Unavailable`, inspect the `Retry-After` HTTP header for the cooldown backoff duration in seconds.
2. **Do Not Retry on 4xx**: Codes `400`, `401`, `403`, and `404` are terminal client errors. Retrying without altering credentials or payloads will always produce identical rejections.
3. **Exponential Jitter on 5xx**: If an upstream model is temporarily overloaded, EkaRouter automatically attempts fallback routes. If all fallback routes fail (returning `502` or `503`), clients should wait with randomized exponential jitter before retrying.
