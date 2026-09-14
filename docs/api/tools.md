# EkaRouter Tools & Execution API

## 1. Overview

EkaRouter tools allow autonomous agents and developers to define parameterized API actions (e.g. Purge CDN Cache, Create Database Branch, Trigger Deployment) with strict schema validation and SSRF defenses.

---

## 2. Endpoints

### List Tools: `GET /api/v1/tools`
Returns all registered tools.

### Register Tool: `POST /api/v1/tools`
```json
{
  "name": "Purge Cloudflare Cache",
  "provider_id": "cloudflare",
  "action_type": "http",
  "description": "Purges all cached files from Cloudflare edge",
  "configuration": {
    "method": "POST",
    "path": "/zones/{{zone_id}}/purge_cache",
    "body_template": "{\"purge_everything\": true}"
  }
}
```

### Get Tool Schema: `GET /api/v1/tools/{id}/schema`
Returns the JSON Schema defining the tool's required and optional input arguments.

### Execute Tool: `POST /api/v1/tools/{id}/execute`
Executes the tool using the bound credentials in the vault:
```json
{
  "parameters": {
    "zone_id": "023e105f4ecef8ad9ca31a8372d0c353"
  }
}
```
**Response (200 OK)**:
```json
{
  "success": true,
  "execution_id": "exec_987654",
  "output": {
    "result": {"id": "023e105f4ecef8ad9ca31a8372d0c353"}
  }
}
```

### Test Tool: `POST /api/v1/tools/{id}/test`
Dry-runs parameter substitution without executing destructive external effects.
