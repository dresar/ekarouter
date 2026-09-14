# Endpoint Details: Usage, Tokens & Analytics

---

## 1. GET `/api/usage`

- **Purpose**: Get aggregate token consumption, request counts, and model distribution from the gateway.
- **Access**: Authenticated.
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  {
    "total_requests": 1420,
    "successful_requests": 1398,
    "failed_requests": 22,
    "prompt_tokens": 850400,
    "completion_tokens": 312000,
    "total_tokens": 1162400,
    "recent_requests": [
      {
        "id": "req_01h8abc",
        "route_id": "gpt-4o",
        "provider_id": "prov_openai",
        "model_id": "gpt-4o",
        "prompt_tokens": 420,
        "completion_tokens": 150,
        "status_code": 200,
        "latency_ms": 680,
        "created_at": "2026-09-15T04:12:00Z"
      }
    ]
  }
  ```

---

## 2. GET `/api/v1/usage/summary`

- **Purpose**: Detailed Developer Platform usage breakdown.
- **Access**: Authenticated.
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  {
    "success": true,
    "data": {
      "total_requests": 1420,
      "providers_count": 8,
      "credentials_count": 14,
      "projects_count": 3
    }
  }
  ```

---

## 3. GET `/api/v1/usage/providers`

- **Purpose**: Requests and error rates grouped by upstream provider.
- **Access**: Authenticated.
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  {
    "success": true,
    "data": [
      {
        "provider_id": "openai",
        "requests": 950,
        "tokens": 820000,
        "errors": 12
      },
      {
        "provider_id": "anthropic",
        "requests": 470,
        "tokens": 342400,
        "errors": 10
      }
    ]
  }
  ```
