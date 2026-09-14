# EkaRouter Models & Routing API

## 1. Overview

EkaRouter provides an OpenAI-compatible model catalog at `/v1/models`. It aggregates both physical upstream provider models and virtual fallback routes.

---

## 2. Endpoints

### List Models: `GET /v1/models`
**Authentication**: `Authorization: Bearer <api_key>`
**Response (200 OK)**:
```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4o",
      "object": "model",
      "created": 1700000000,
      "owned_by": "ekarouter"
    },
    {
      "id": "claude-3-5-sonnet",
      "object": "model",
      "created": 1700000000,
      "owned_by": "ekarouter"
    }
  ]
}
```

### Admin Model Catalog: `GET /api/models`
**Authentication**: Session Cookie / Admin Bearer
Returns full configuration parameters (context limit, capabilities, provider mapping).

### Add Model: `POST /api/models`
```json
{
  "provider_id": "prov_openai",
  "external_name": "gpt-4o",
  "display_name": "OpenAI GPT-4o",
  "context_limit": 128000,
  "streaming": true,
  "enabled": true
}
```

### Delete Model: `DELETE /api/models/{id}`
Removes a model from the administrative catalog.

---

## 3. Route Configuration: `POST /api/routes`
Creates an intelligent fallback route:
```json
{
  "name": "fast-code",
  "strategy": "priority",
  "items": [
    {
      "provider_id": "prov_openai",
      "account_id": "acc_primary",
      "model_id": "mod_gpt4o",
      "priority": 1,
      "timeout_ms": 30000
    },
    {
      "provider_id": "prov_anthropic",
      "account_id": "acc_secondary",
      "model_id": "mod_claude",
      "priority": 10,
      "timeout_ms": 30000
    }
  ]
}
```
Clients can then request `model: "fast-code"` on `/v1/chat/completions` and EkaRouter will automatically handle priority and fallback!
