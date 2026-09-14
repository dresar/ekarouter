# Endpoint Details: Audit Logs & Request Inspection

---

## 1. GET `/api/v1/audit-logs`

- **Purpose**: Query append-only immutable audit trail for compliance and operations.
- **Access**: Admin.
- **Query Parameters**:
  - `limit`: Integer (default `50`, max `200`).
  - `actor_id`: Filter by actor username or key.
  - `action`: Filter by action type (e.g. `credential.create`, `route.update`, `tool.execute`).
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": 1042,
        "actor_id": "admin",
        "action": "provider.create",
        "resource_type": "provider",
        "resource_id": "prov_anthropic",
        "project_id": "",
        "request_id": "req_01h8abc",
        "result": "success",
        "created_at": "2026-09-15T04:25:00Z"
      }
    ]
  }
  ```

---

## 2. Real-Time Telemetry Constraints

- **No WebSocket Stream**: There is no live WebSocket log tail in the current backend.
- **Polling Strategy**: The frontend should fetch new log records on a 10-second polling interval with a manual "Pause Auto-Refresh" switch.
