# Endpoint Details: AI Providers & Upstream Backends

---

## 1. GET `/api/providers`

- **Purpose**: List all configured upstream AI providers.
- **Access**: Admin.
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  [
    {
      "id": "prov_openai",
      "key": "openai",
      "name": "OpenAI Official",
      "kind": "openai",
      "base_url": "https://api.openai.com/v1",
      "enabled": true,
      "created_at": "2026-09-14 12:00:00",
      "updated_at": "2026-09-14 12:00:00"
    }
  ]
  ```

---

## 2. POST `/api/providers`

- **Purpose**: Create or update provider configuration.
- **Access**: Admin.
- **Request Body**:
  ```json
  {
    "id": "prov_anthropic",
    "key": "anthropic",
    "name": "Anthropic Claude",
    "kind": "anthropic",
    "base_url": "https://api.anthropic.com/v1",
    "enabled": true
  }
  ```
- **Validation**:
  - `id`: Optional. If empty, backend auto-generates `prov_UUID`.
  - `key`: Optional. Defaults to `id`.
  - `name`: Required string.
  - `kind`: Required enum (`openai`, `anthropic`, `gemini`, `groq`, `cerebras`, `openrouter`, `huggingface`, `cloudflare`, `custom`, etc.).
  - `base_url`: Valid HTTP/HTTPS upstream URL.
- **Success Status**: `201 Created`
- **Response Schema**:
  ```json
  {
    "status": "created",
    "id": "prov_anthropic"
  }
  ```

---

## 3. DELETE `/api/providers/{id}`

- **Purpose**: Delete AI provider and cascade deletion of related accounts.
- **Success Status**: `200 OK`
- **Response**: `{"status":"deleted"}`
