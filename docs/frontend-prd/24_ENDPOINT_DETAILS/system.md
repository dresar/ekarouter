# Endpoint Details: System Probes & Liveness

---

## 1. GET `/health`

- **Purpose**: High-level process liveness check.
- **Access**: Public.
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  {
    "status": "ok",
    "database": "ok",
    "uptime": "4h12m30s"
  }
  ```

---

## 2. GET `/ready`

- **Purpose**: Kubernetes/Docker readiness probe confirming SQLite connection pool is healthy and migrations are applied.
- **Access**: Public.
- **Success Status**: `200 OK`
- **Response**: `{"status":"ready"}`

---

## 3. GET `/live`

- **Purpose**: Minimal fast probe verifying the HTTP server loop is accepting connections.
- **Access**: Public.
- **Success Status**: `200 OK`
- **Response**: `{"status":"alive"}`
