# EkaRouter Credential Rotation API

## 1. Overview

EkaRouter provides automatic rotation across multi-key pools as well as manual on-demand secret rotation.

---

## 2. Manual Credential Rotation: `POST /api/v1/credentials/{id}/rotate`

Updates a single credential's secret in the vault while retaining historical version audit records:

```json
{
  "new_secret": "new_super_secret_value_here"
}
```
**Response (200 OK)**:
```json
{
  "success": true,
  "message": "Credential rotated successfully",
  "data": {
    "id": "cred_12345",
    "masked_value": "new_***re",
    "version": 2
  }
}
```

---

## 3. Credential Pool Rotation: `POST /api/credential-pools/{id}/rotate`

Forces the rotation engine to advance its cursor to the next eligible credential within a pool:
**Response (200 OK)**:
```json
{
  "status": "rotated",
  "pool_id": "pool_openai_prod",
  "active_member_id": "pm_5678"
}
```

---

## 4. Pool Traffic Management
- `POST /api/credential-pools/{id}/pause`: Pauses traffic routing to all keys in the pool.
- `POST /api/credential-pools/{id}/resume`: Resumes traffic routing.
