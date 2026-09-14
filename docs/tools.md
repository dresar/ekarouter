# Generic API Tool Execution Engine

A Tool represents an executable third-party API operation (e.g., sending an email via Resend, checking a domain on Cloudflare, or creating an issue on GitHub).

## Defining a Tool

```http
POST /api/v1/tools
Content-Type: application/json
Authorization: Bearer <token>

{
  "provider_id": "resend",
  "name": "Send Email",
  "description": "Send a transactional email",
  "category": "communication",
  "method": "POST",
  "url_template": "https://api.resend.com/emails",
  "headers_template": {
    "Content-Type": "application/json"
  },
  "timeout_ms": 15000
}
```

## Executing a Tool

```http
POST /api/v1/tools/{id}/execute
Content-Type: application/json
Authorization: Bearer <token>

{
  "variables": {},
  "body": {
    "from": "onboarding@resend.dev",
    "to": "user@example.com",
    "subject": "Welcome to EkaRouter",
    "html": "<strong>EkaRouter Developer Platform active.</strong>"
  }
}
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "e3b0c442-98fc-1c14-9afb-4c8996fb9242",
    "tool_id": "tool_resend_send",
    "status": "success",
    "status_code": 200,
    "latency_ms": 235,
    "executed_at": "2026-09-15T04:20:00Z"
  },
  "meta": {
    "request_id": "req_837192"
  }
}
```
All executions are persisted in `tool_executions` and recorded in `audit_logs`.
