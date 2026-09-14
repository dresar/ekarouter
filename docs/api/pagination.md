# EkaRouter Pagination & Filtering Guidelines

## 1. Overview

To ensure predictable latency, stable memory footprints, and responsive frontend dashboards, list endpoints across EkaRouter support standardized pagination and filtering.

---

## 2. Pagination Parameters

All paginated endpoints accept two primary query string parameters:

| Parameter | Type | Default | Minimum | Maximum | Description |
|---|---|---|---|---|---|
| `limit` | integer | `20` | `1` | `100` | Maximum number of records returned per page. |
| `offset` | integer | `0` | `0` | - | Zero-based index offset for record selection. |

### Example Request
```http
GET /api/v1/audit-logs?limit=50&offset=100 HTTP/1.1
Host: localhost:8080
Authorization: Bearer <token>
```

---

## 3. Response Envelope Structure

Paginated endpoints wrap records in a structured envelope containing both the record slice and pagination metadata:

```json
{
  "success": true,
  "data": [
    {
      "id": "evt_01j8k9m0",
      "action": "credential.rotated",
      "entity_id": "cred_123",
      "timestamp": "2026-09-15T04:30:00Z"
    }
  ],
  "pagination": {
    "total": 342,
    "limit": 50,
    "offset": 100,
    "has_more": true
  }
}
```

### Pagination Fields

- **`total`**: Total number of matching rows in the database matching the filter criteria.
- **`limit`**: Effective limit used for the query.
- **`offset`**: Current offset applied.
- **`has_more`**: Boolean flag indicating if subsequent records exist (`offset + limit < total`).

---

## 4. Sorting Parameters

Where supported (e.g. audit logs, usage logs, credentials list), sorting can be customized via:

| Parameter | Type | Default | Options | Description |
|---|---|---|---|---|
| `sort_by` | string | `created_at` | `created_at`, `updated_at`, `name`, `priority` | Column name to sort against. |
| `order` | string | `desc` | `asc`, `desc` | Sort direction. |

Example:
```http
GET /api/v1/credentials?sort_by=priority&order=asc
```

---

## 5. Standard Filter Parameters

Most entity lists support direct filtering:

| Filter | Query Parameter | Example |
|---|---|---|
| Filter by Provider | `provider_id` | `?provider_id=openai` |
| Filter by Environment | `environment` | `?environment=production` |
| Filter by Project | `project_id` | `?project_id=proj_web_app` |
| Filter by Status | `status` | `?status=active` (or `disabled`, `cooldown`) |
| Filter by Time Window | `since` & `until` | `?since=2026-09-01T00:00:00Z` |

---

## 6. Frontend Pagination Recipe (React / TanStack Query)

```typescript
import { useQuery } from '@tanstack/react-query';

interface PaginatedResponse<T> {
  success: boolean;
  data: T[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
    has_more: boolean;
  };
}

export function useAuditLogs(page: number, pageSize = 20) {
  const offset = page * pageSize;
  return useQuery<PaginatedResponse<AuditLog>>({
    queryKey: ['audit-logs', page, pageSize],
    queryFn: async () => {
      const res = await fetch(`/api/v1/audit-logs?limit=${pageSize}&offset=${offset}`, {
        credentials: 'include',
      });
      if (!res.ok) throw new Error('Failed to fetch audit logs');
      return res.json();
    },
    keepPreviousData: true,
  });
}
```
