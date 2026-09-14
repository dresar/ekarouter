# Credential Vault & Secret Management

The EkaRouter Vault provides encrypted, versioned storage for all API keys, access tokens, webhook signing secrets, and cloud credentials.

## API Endpoints

### 1. List Credentials
```http
GET /api/v1/credentials?provider_id=cloudflare&environment=production
Authorization: Bearer <token>
```

Response:
```json
{
  "success": true,
  "data": [
    {
      "id": "c7a8b9d0-1234-5678-9abc-def012345678",
      "name": "Cloudflare DNS Production",
      "credential_type": "api_key",
      "provider_id": "cloudflare",
      "environment": "production",
      "masked_value": "cf_a****wxyz",
      "priority": 1,
      "status": "active",
      "health_state": "healthy",
      "request_count": 420,
      "error_count": 0
    }
  ],
  "meta": {
    "request_id": "req_01928374"
  }
}
```

### 2. Create Credential
```http
POST /api/v1/credentials
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "GitHub Actions Automation Token",
  "credential_type": "personal_access_token",
  "provider_id": "github",
  "environment": "production",
  "secret_value": "ghp_1234567890abcdef1234567890abcdef",
  "priority": 1,
  "tags": "ci,deployments",
  "notes": "Scoped to repo workflow and dispatch"
}
```

### 3. Rotate Credential
```http
POST /api/v1/credentials/{id}/rotate
Content-Type: application/json
Authorization: Bearer <token>

{
  "new_secret": "ghp_9876543210fedcba9876543210fedcba"
}
```

### 4. Test Connectivity
```http
POST /api/v1/credentials/{id}/test
Authorization: Bearer <token>
```
Runs a real-time connectivity verification against the upstream provider and updates the credential's health state.
