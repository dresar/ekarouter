# EkaRouter Environments API

## 1. Overview

Environments segment credentials and routing targets into stages such as `production`, `staging`, `preview`, and `development`.

---

## 2. Endpoints

### List Environments: `GET /api/v1/environments`
Optional query parameter: `?project_id=proj_123`

### Create Environment: `POST /api/v1/environments`
```json
{
  "name": "staging",
  "project_id": "proj_123456",
  "description": "Pre-release testing environment"
}
```
**Response (201 Created)**:
```json
{
  "success": true,
  "data": {
    "id": "env_987654",
    "name": "staging",
    "project_id": "proj_123456",
    "description": "Pre-release testing environment"
  }
}
```

### Update Environment: `PATCH /api/v1/environments/{id}`
```json
{
  "description": "Updated environment description"
}
```

### Delete Environment: `DELETE /api/v1/environments/{id}`
Removes the environment.
