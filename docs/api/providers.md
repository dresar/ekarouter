# EkaRouter Providers API

## 1. Developer Platform Providers (`/api/v1/providers`)

### List Providers: `GET /api/v1/providers`
Returns metadata, required credential schemas, and documentation links for all 27 registered platform providers.

**Response (200 OK)**:
```json
{
  "success": true,
  "data": [
    {
      "id": "cloudflare",
      "name": "Cloudflare",
      "category": "developer",
      "description": "Cloudflare API for DNS, Workers, Pages, and Edge rules",
      "base_url": "https://api.cloudflare.com/client/v4",
      "auth_type": "bearer_token",
      "required_credentials": ["api_token"],
      "supported_operations": ["list_zones", "dns_records", "workers", "purge_cache"],
      "website_url": "https://www.cloudflare.com",
      "docs_url": "https://developers.cloudflare.com/api/",
      "free_tier_status": "available",
      "enabled": true
    }
  ]
}
```

### Provider Details: `GET /api/v1/providers/{id}`
Returns details for a specific provider by ID (e.g. `cloudflare`, `github`, `supabase`, `resend`, `stripe`).

### Capabilities: `GET /api/v1/providers/{id}/capabilities`
Returns list of provider capability flags (e.g. `bearer_auth`, `resource_listing`, `request_proxy`, `health_check`).

### Docs: `GET /api/v1/providers/{id}/docs`
Returns official documentation and API reference links.

### Health Check: `POST /api/v1/providers/{id}/health`
Performs an active connectivity probe against the provider endpoint.
