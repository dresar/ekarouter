# Endpoint Details: Ingress API Keys & Client Access

---

## 1. GET `/api/keys`

- **Purpose**: List all generated client API keys.
- **Access**: Admin.
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  [
    {
      "id": "key_8f3a1b2c",
      "name": "Production AI Agent",
      "prefix": "eka_live_8a",
      "scopes": "*",
      "enabled": true,
      "created_at": "2026-09-14T10:00:00Z",
      "last_used_at": "2026-09-15T04:20:00Z"
    }
  ]
  ```
- **Security Invariant**: Never returns the full plaintext key or key hash.

---

## 2. POST `/api/keys`

- **Purpose**: Generate a new client API key for accessing `/v1/*` gateway endpoints.
- **Access**: Admin.
- **Request Body**:
  ```json
  {
    "name": "Development Cursor Instance",
    "scopes": "*"
  }
  ```
- **Success Status**: `201 Created`
- **Response Schema**:
  ```json
  {
    "status": "created",
    "id": "key_8f3a1b2c",
    "key": "eka_live_8a92f01bc45d2e7a1024bc"
  }
  ```
- **Crucial UI Rule**: The field `key` contains the full unmasked key and is returned **ONLY ONCE**. The frontend must instruct the user to copy and store it immediately.

---

## 3. DELETE `/api/keys/{id}`

- **Purpose**: Permanently revoke an API key.
- **Access**: Admin.
- **Parameters**: `id` (path parameter, string).
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  {
    "status": "deleted"
  }
  ```
