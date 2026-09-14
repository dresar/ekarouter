# Endpoint Details: Routing Combos & Fallback Targets

---

## 1. GET `/api/routes`

- **Purpose**: List all routing combos and model aliases with target priority chains.
- **Access**: Admin.
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  [
    {
      "id": "route_gpt4o",
      "name": "gpt-4o",
      "strategy": "priority",
      "enabled": true,
      "item_count": 2,
      "items": [
        {
          "id": "route_gpt4o_item_1",
          "provider_id": "prov_openai",
          "account_id": "acc_primary",
          "model_id": "gpt-4o",
          "priority": 1,
          "weight": 1,
          "timeout_ms": 60000,
          "max_retries": 2,
          "enabled": true
        },
        {
          "id": "route_gpt4o_item_2",
          "provider_id": "prov_openrouter",
          "account_id": "acc_backup",
          "model_id": "openai/gpt-4o",
          "priority": 2,
          "weight": 1,
          "timeout_ms": 60000,
          "max_retries": 2,
          "enabled": true
        }
      ]
    }
  ]
  ```

---

## 2. POST `/api/routes`

- **Purpose**: Create or update a routing combo.
- **Access**: Admin.
- **Request Body**:
  ```json
  {
    "id": "route_claude",
    "name": "claude-3-5-sonnet",
    "strategy": "round_robin",
    "items": [
      {
        "provider_id": "prov_anthropic",
        "account_id": "acc_ant_1",
        "priority": 1,
        "weight": 1
      },
      {
        "provider_id": "prov_anthropic",
        "account_id": "acc_ant_2",
        "priority": 1,
        "weight": 1
      }
    ]
  }
  ```
- **Strategies Supported**:
  - `priority`: Evaluates targets strictly in ascending priority order (1 first, then 2).
  - `round_robin`: Distributes traffic across accounts of equal priority.
- **Success Status**: `201 Created`
- **Response**: `{"status":"created","id":"route_claude"}`

---

## 3. DELETE `/api/routes/{id}`

- **Purpose**: Delete route definition and clear router cache.
- **Success Status**: `200 OK`
- **Response**: `{"status":"deleted"}`
