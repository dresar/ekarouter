# Endpoint Details: Outbound Proxies & Connectivity

---

## 1. GET `/api/proxies`

- **Purpose**: List outbound proxy configurations.
- **Access**: Admin.
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  [
    {
      "id": "px_datacenter_1",
      "name": "US Residential Proxy",
      "scheme": "http",
      "host": "104.28.16.20",
      "port": 8080,
      "username": "proxy_user",
      "enabled": true
    }
  ]
  ```

---

## 2. POST `/api/proxies`

- **Purpose**: Create or update outbound proxy profile.
- **Access**: Admin.
- **Request Body**:
  ```json
  {
    "id": "px_datacenter_1",
    "name": "US Residential Proxy",
    "scheme": "http",
    "host": "104.28.16.20",
    "port": 8080,
    "username": "proxy_user",
    "password": "secret_password"
  }
  ```
- **Validation**:
  - `scheme`: `http` | `https` | `socks5` | `relay`.
  - `host`: Valid domain or IP (SSRF validated).
  - `port`: Positive integer (`1` - `65535`).
- **Success Status**: `201 Created`
- **Response**: `{"status":"created","id":"px_datacenter_1"}`

---

## 3. POST `/api/proxies/{id}/test`

- **Purpose**: Execute an active network probe through the proxy to determine reachability and latency.
- **Access**: Admin.
- **Parameters**: `id` (path parameter, string).
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  {
    "ok": true,
    "status_code": 200,
    "latency_ms": 42,
    "error": ""
  }
  ```
- **Failure State (Proxy Offline)**:
  ```json
  {
    "ok": false,
    "status_code": 0,
    "latency_ms": 5000,
    "error": "connection timed out"
  }
  ```
