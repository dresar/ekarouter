# EkaRouter Credentials Vault API

## 1. Overview

EkaRouter provides a secure credential vault where API keys, tokens, and secrets are encrypted at rest with AES-256-GCM.
Raw secrets are **never returned** by any GET endpoint. Only masked identifiers (e.g. `cf_s***90`) are exposed to clients and the frontend UI.

---

## 2. Endpoints

### Store Credential: `POST /api/v1/credentials`
**Request Body**:
```json
{
  "name": "Production Cloudflare Token",
  "provider_id": "cloudflare",
  "secret_value": "cf_live_token_secret_1234567890abcdef",
  "environment": "production",
  "project_id": "proj_123"
}
```
**Response (201 Created)**:
```json
{
  "success": true,
  "data": {
    "id": "cred_a1b2c3d4",
    "name": "Production Cloudflare Token",
    "provider_id": "cloudflare",
    "masked_value": "cf_l***ef",
    "environment": "production",
    "status": "active"
  }
}
```

### List Credentials: `GET /api/v1/credentials`
Returns all credentials in masked format.

### Get Credential: `GET /api/v1/credentials/{id}`
Returns details for a single credential (masked).

### Update Credential: `PATCH /api/v1/credentials/{id}`
Updates metadata such as `name`, `priority`, or `notes`.

### Enable / Disable:
- `POST /api/v1/credentials/{id}/enable`
- `POST /api/v1/credentials/{id}/disable`

### Rotate Credential: `POST /api/v1/credentials/{id}/rotate`
```json
{
  "new_secret": "cf_live_new_token_value_9876543210"
}
```
Re-encrypts with the new secret value and archives the previous secret version for auditing.

### Delete Credential: `DELETE /api/v1/credentials/{id}`
Permanently deletes the credential and any associated versions from the vault.
