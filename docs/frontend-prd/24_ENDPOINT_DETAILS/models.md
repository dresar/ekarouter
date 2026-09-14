# Endpoint Details: AI Models & Public Discovery

---

## 1. GET `/v1/models`

- **Purpose**: OpenAI-compatible model catalog discovery for AI clients and IDEs.
- **Access**: Ingress Bearer Token (`eka_...`).
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  {
    "object": "list",
    "data": [
      {
        "id": "gpt-4o",
        "object": "model",
        "created": 1700000000,
        "owned_by": "ekarouter",
        "permission": []
      },
      {
        "id": "claude-3-5-sonnet",
        "object": "model",
        "created": 1700000000,
        "owned_by": "ekarouter",
        "permission": []
      }
    ]
  }
  ```

---

## 2. GET `/api/models`

- **Purpose**: Administrative listing of registered models and provider associations.
- **Access**: Admin.
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  [
    {
      "id": "m_gpt4o",
      "provider_id": "prov_openai",
      "external_name": "gpt-4o",
      "display_name": "GPT-4o Omni",
      "context_limit": 128000,
      "streaming": true,
      "enabled": true
    }
  ]
  ```

---

## 3. POST `/api/models`

- **Purpose**: Register or update custom model mapping.
- **Access**: Admin.
- **Request Body**:
  ```json
  {
    "id": "m_custom_1",
    "provider_id": "prov_openai",
    "external_name": "gpt-4o-mini",
    "display_name": "GPT-4o Mini",
    "context_limit": 128000,
    "streaming": true
  }
  ```
- **Success Status**: `201 Created`
- **Response**: `{"status":"created","id":"m_custom_1"}`
