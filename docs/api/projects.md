# EkaRouter Projects API

## 1. Overview

Projects provide multi-tenant grouping for credentials, environments, routing rules, and usage tracking.

---

## 2. Endpoints

### List Projects: `GET /api/v1/projects`
Returns all projects accessible to the authenticated user.

### Create Project: `POST /api/v1/projects`
```json
{
  "name": "Production E-Commerce",
  "description": "Customer-facing web store"
}
```
**Response (201 Created)**:
```json
{
  "success": true,
  "data": {
    "id": "proj_123456",
    "name": "Production E-Commerce",
    "description": "Customer-facing web store",
    "created_at": "2026-09-15T04:45:00Z"
  }
}
```

### Get Project: `GET /api/v1/projects/{id}`
Returns project details by ID.

### Update Project: `PATCH /api/v1/projects/{id}`
```json
{
  "name": "Updated Project Name",
  "description": "Updated project description"
}
```

### Delete Project: `DELETE /api/v1/projects/{id}`
Cascades deletion to environments, tool bindings, and credential assignments.
