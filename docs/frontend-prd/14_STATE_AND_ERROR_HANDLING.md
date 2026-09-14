# 14. State Management, Offline Probes & Error Protocols

---

## 1. Six Mandatory UI States

Every view, data table, and interactive widget must explicitly support all six fundamental states:

```text
┌────────────────────────────────────────────────────────┐
│ 1. Loading State     (Skeleton animation / shimmer)    │
├────────────────────────────────────────────────────────┤
│ 2. Empty State       (Zero items illustration & action)│
├────────────────────────────────────────────────────────┤
│ 3. Error State       (Status code, message, retry btn) │
├────────────────────────────────────────────────────────┤
│ 4. Success State     (Populated data grid / feedback)  │
├────────────────────────────────────────────────────────┤
│ 5. Disabled State    (Read-only or permission locked)  │
├────────────────────────────────────────────────────────┤
│ 6. Retrying State    (Exponential backoff with timer)  │
└────────────────────────────────────────────────────────┘
```

---

## 2. HTTP Status Code Handling Matrix

| Status Code | Backend Meaning | Client Action | User Feedback |
| :--- | :--- | :--- | :--- |
| **`400 Bad Request`** | Validation or JSON parse error | Keep form open; highlight invalid field | Form field helper text in Rose |
| **`401 Unauthorized`** | Expired or missing session token | Clear token storage; redirect to `/login` | Toast: *"Session expired. Please log in again."* |
| **`403 Forbidden`** | Insufficient RBAC permission | Block action | Inline alert: *"Admin permissions required."* |
| **`404 Not Found`** | Resource missing or deleted | Redirect or show 404 container | Empty state: *"Resource does not exist."* |
| **`429 Too Many Req`** | Rate limit or quota hit | Pause outbound requests; show backoff | Banner: *"Rate limit reached. Retrying in Xs."* |
| **`500 Server Error`** | Unhandled backend fault | Log error trace; provide retry button | Error banner with *"Retry"* action |
| **`502 Bad Gateway`** | Upstream provider connection failed | Mark provider degraded; suggest fallback | Warning badge on affected provider |
| **`Network Offline`** | Backend unreachable / CORS failure | Persistent Topbar Alert | Topbar: *"Backend disconnected. Reconnecting..."* |

---

## 3. Optimistic Updates & Cache Invalidation

When mutating resources via `POST`, `PATCH`, or `DELETE`:
1. **Optimistic Row Update**: Update local list state immediately for instant feedback.
2. **Revalidation**: Trigger background refetch on the parent query key upon server response.
3. **Rollback on Error**: If the server returns a 4xx/5xx code, revert the optimistic state change and display an error toast.
