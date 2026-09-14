# Usage Statistics & Observability

EkaRouter provides usage tracking across providers, credentials, and projects.

## Usage Endpoints

### 1. Global Usage Summary
```http
GET /api/v1/usage
Authorization: Bearer <token>
```

Response:
```json
{
  "success": true,
  "data": {
    "total_requests": 1420,
    "total_errors": 12,
    "active_credentials": 18
  },
  "meta": {
    "request_id": "req_847120"
  }
}
```

### 2. Credential Usage Stats
```http
GET /api/v1/credentials/{id}/usage
Authorization: Bearer <token>
```
Reports total requests, errors, and last-used timestamp for the given key.
