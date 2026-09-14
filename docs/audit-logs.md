# Audit Logging Engine

EkaRouter maintains an immutable append-only audit trail in the `audit_logs` table for compliance and security auditing.

## Audited Events
- User authentication (`login`, `logout`, `login_failed`).
- Credential lifecycle (`credential.create`, `credential.update`, `credential.delete`, `credential.rotate`, `credential.test`).
- Tool execution (`tool.execute`).
- System configuration changes.

## Querying Audit Logs
```http
GET /api/v1/audit-logs?action=credential.create&limit=25
Authorization: Bearer <token>
```

Response:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "actor_id": "usr_admin",
      "actor_type": "user",
      "action": "credential.create",
      "resource_type": "credential",
      "resource_id": "cred_xyz123",
      "project_id": "proj_prod",
      "ip_address": "198.51.100.42",
      "user_agent": "Mozilla/5.0",
      "request_id": "req_847120",
      "result": "success",
      "created_at": "2026-09-15T04:10:00Z"
    }
  ],
  "meta": {
    "request_id": "req_918234"
  }
}
```
Credentials and sensitive parameters are never recorded in audit payloads.
